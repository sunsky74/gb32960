package codec

import (
	"errors"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/frame"
	"github.com/sunsky74/gb32960/types"
	"github.com/sunsky74/gb32960/utils"
)

// ProtocolCodec 是顶层帧编解码器。
// Decode:校验 BCC、解析帧头字段、解密数据单元(仅 0x01)。
// Encode:构造帧头、加密数据单元、追加 BCC。
var ProtocolCodec = &protocolMessageCodec{}

type protocolMessageCodec struct{}

// Decode 从 r 解析一个完整的 GB/T 32960 协议帧。
//
// 线格式布局:Header(2B) | Cmd(1B) | Response(1B) | VIN(17B) | Encryption(1B)
//
//	| PayloadLen(2B) | Payload(N B) | BCC(1B)
//
// BCC 覆盖帧的 [2:] 字节(即 2 字节帧头之后的全部内容)。
// 仅支持 EncryptionNone(0x01,直通);其他任何模式
// 都返回 api.ErrEncryptionNotSupported。
func (c *protocolMessageCodec) Decode(r api.Reader) (api.Message, error) {
	m := &frame.ProtocolMessage{}

	// 1. 读取版本帧头(2 字节)。
	header := r.ReadUint16()
	// fix 2026-09-17: 2016.md 表2 L31(同表B.1 L614)、2025.md 表2 L32 ——
	// 起始符固定为 "##"(0x23,0x23,2016 版)或 "$$"(0x24,0x24,2025 版)。
	// 此前任意 2 字节都会被接受,未知帧头还会被 GBTVersionByHeader
	// 静默按 2016 版解析;起始符位于 BCC 覆盖范围之外,必须在此显式校验。
	// 先上报读取下溢,保持既有的 ErrBufferUnderflow 语义。
	if err := r.Err(); err != nil {
		return nil, err
	}
	if header != 0x2323 && header != 0x2424 {
		return nil, api.ErrInvalidHeader
	}
	m.Version = types.GBTVersionByHeader(header)

	// 2. 读取剩余字节(从命令字段到 BCC,含两端)。
	bodyWithBCC := r.ReadBytes(r.Remaining())

	// 3. 校验 BCC:除最后一个字节外所有字节的异或值必须等于最后一个字节。
	// 空帧体意味着帧头下溢,或根本没有帧体,
	// 统一上报为 ErrBufferUnderflow(按 api.Reader 契约,容忍 r.Err() 存在)。
	if len(bodyWithBCC) == 0 {
		return nil, api.ErrBufferUnderflow
	}
	bodyBytes := bodyWithBCC[:len(bodyWithBCC)-1]
	expectedBCC := utils.CalcBCC(bodyBytes)
	actualBCC := bodyWithBCC[len(bodyWithBCC)-1]
	if expectedBCC != actualBCC {
		return nil, api.ErrBCCMismatch
	}
	m.CheckCode = actualBCC

	// 4. 从 bodyBytes 解析协议帧头字段。
	bodyReader := utils.NewByteReader(bodyBytes)
	cmd := bodyReader.ReadUint8()
	// 按版本分发:V2025 有自己的命令表;V2016 为默认。
	// (audit 2026-07-31:早期草稿硬编码了 CommandV2016ByCode,会把
	// V2025 帧错误分类。对齐参考实现的 switch 写法。)
	switch m.Version {
	case api.V2025:
		m.RequestType = types.CommandV2025ByCode(cmd)
	default:
		m.RequestType = types.CommandV2016ByCode(cmd)
	}
	m.ResponseType = types.ResponseByCode(bodyReader.ReadUint8())
	m.VIN = bodyReader.ReadString(17)
	m.Encryption = types.EncryptionType(bodyReader.ReadUint8())
	m.PayloadLength = int(bodyReader.ReadUint16())

	// audit 2026-09-17:GB/T 32960 表 2 将数据单元长度上限设为 65531 字节。
	if m.PayloadLength > 65531 {
		return nil, api.ErrLengthMismatch
	}

	// 5. 读取数据单元,并拒绝非直通加密。
	// 对 0x01(EncryptionNone)而言,线上字节就是明文数据单元。
	encrypted := bodyReader.ReadBytes(m.PayloadLength)
	if m.Encryption != types.EncryptionNone {
		return nil, api.ErrEncryptionNotSupported
	}
	m.RawBytes = encrypted // 0x01:encrypted 即明文

	// 上报解析帧体字段期间记录的任何下溢
	// (与 codec/gbt2016/platform_login_codec.go 中的模式一致)。
	if err := bodyReader.Err(); err != nil {
		return nil, err
	}
	// audit 2026-09-17:超出声明数据单元长度的尾随字节意味着
	// 长度字段与帧体不一致;此前被静默丢弃。
	if bodyReader.Remaining() != 0 {
		return nil, api.ErrLengthMismatch
	}
	return m, nil
}

// Encode 将 msg 的完整 GB/T 32960 协议帧写入 w。
//
// 与 Plan 2 示例的偏差(已上报):Plan 2 Part B 的示例调用
// `w.Bytes()[2:]` 计算 BCC 范围,但 api.Writer 接口
// (见 api/api.go)没有 Bytes() 方法,只有具体类型 *utils.ByteWriter
// 才有。为了在与任意 api.Writer 协作的同时保持帧布局不变,
// 我们把 frame.ProtocolMessage.Bytes() 镜像到本地的
// *utils.ByteWriter(它确实暴露 Bytes()),在其 [2:]
// 切片上计算 BCC,然后通过 w.WriteBytes 输出组装好的帧。产出的线上
// 字节与 frame.ProtocolMessage.Bytes() 逐字节一致。
func (c *protocolMessageCodec) Encode(w api.Writer, msg api.Message) error {
	m, ok := msg.(*frame.ProtocolMessage)
	if !ok {
		return errors.New("gb32960: message is not *frame.ProtocolMessage")
	}
	code, ok := frame.CommandCode(m.RequestType)
	if !ok {
		return errors.New("gb32960: RequestType is nil or unknown command type")
	}

	// audit 2026-09-17:nil Payload 仅在命令没有帧体类型时合法
	// (0x07/0x08 及预留命令,0 字节数据单元)。与
	// frame.ProtocolMessage.Bytes() 完全一致。
	var payload []byte
	if m.Payload != nil {
		var err error
		payload, err = m.Payload.Bytes()
		if err != nil {
			return err
		}
	} else if frame.PayloadType(m.Version, code) != nil {
		return errors.New("gb32960: Payload is nil for a command that requires a body")
	} else if len(m.RawBytes) > 0 {
		// fix 2026-09-17: 2016.md 表3 L65(0xC0~0xFE 平台交换自定义数据,
		// 同 2025.md 表3 L62)—— 无注册消息体类型的命令,其数据单元必须
		// 以解码时保存的原始字节透传。此处与 frame.ProtocolMessage.Bytes()
		// 保持逐字节一致,避免重编码写出空数据单元。
		payload = m.RawBytes
	}
	if m.Encryption != types.EncryptionNone {
		return api.ErrEncryptionNotSupported
	}
	// audit 2026-09-17:GB/T 32960 表 2 将数据单元长度上限设为 65531 字节。
	if len(payload) > 65531 {
		return api.ErrLengthMismatch
	}

	// 构建到本地 writer,使 BCC 范围([2:])可观测。
	// 这与 frame.ProtocolMessage.Bytes() 完全一致。
	local := utils.NewByteWriter()
	local.WriteString(types.Header(m.Version), 2)
	local.WriteUint8(code)
	local.WriteUint8(m.ResponseType.Code())
	local.WriteString(m.VIN, 17)
	local.WriteUint8(byte(m.Encryption))

	local.WriteUint16(uint16(len(payload)))
	local.WriteBytes(payload)

	// BCC:从 byte[2](2 字节帧头之后)到数据单元末尾的所有字节异或。
	bccRange := local.Bytes()[2:]
	local.WriteUint8(utils.CalcBCC(bccRange))

	// 将组装好的帧输出到调用方提供的 writer。
	w.WriteBytes(local.Bytes())
	return nil
}

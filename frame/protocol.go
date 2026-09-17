// Package frame 包含协议帧类型(ProtocolMessage)以及
// 命令到消息类型的分发(payloadType)。它位于单独的
// 包中,以打破循环导入:ProtocolMessage 引用 gbt2016
// 和 gbt2025 类型,而 gbt2016/gbt2025 又引用 model(用于 BeanTime)。
// 把帧类型放在这里,使其可以导入所有 model 子包,
// 而不会形成回到 model 的循环。
package frame

import (
	"errors"
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/model/gbt2016"
	"github.com/sunsky74/gb32960/model/gbt2025"
	"github.com/sunsky74/gb32960/types"
	"github.com/sunsky74/gb32960/utils"
)

// ProtocolMessage 表示一个完整的 GB/T 32960 协议帧。
type ProtocolMessage struct {
	Version       api.GBTVersion
	RequestType   any // *types.CommandV2016 或 *types.CommandV2025
	ResponseType  types.ResponseType
	VIN           string
	Encryption    types.EncryptionType
	PayloadLength int
	RawBytes      []byte
	Payload       model.MessageBody
	CheckCode     byte
}

// PayloadType 返回(版本,命令)对所对应的 Go 结构体类型,
// 与参考 Java 实现(getV2016Body / getV2025Body)一致。
// 对于没有可解码消息体的命令(Heartbeat、ClockCorrect、
// 直通的配置/控制),返回 nil。
func PayloadType(v api.GBTVersion, cmdCode byte) reflect.Type {
	switch v {
	case api.V2016:
		switch cmdCode {
		case 0x01:
			return reflect.TypeOf((*gbt2016.VehicleLogin)(nil)).Elem()
		case 0x02, 0x03:
			return reflect.TypeOf((*gbt2016.RealTimeData)(nil)).Elem()
		case 0x04:
			return reflect.TypeOf((*gbt2016.VehicleLogout)(nil)).Elem()
		case 0x05:
			// audit 2026-09-17:V2016 的 0x05/0x06 此前缺失,导致平台
			// 登入/登出帧尽管已注册编解码器,仍解码出 nil 的 Payload。
			return reflect.TypeOf((*gbt2016.PlatformLogin)(nil)).Elem()
		case 0x06:
			return reflect.TypeOf((*gbt2016.PlatformLogout)(nil)).Elem()
		}
	case api.V2025:
		switch cmdCode {
		case 0x01:
			return reflect.TypeOf((*gbt2025.VehicleLoginV2025)(nil)).Elem()
		case 0x02, 0x03:
			return reflect.TypeOf((*gbt2025.RealTimeV2025Data)(nil)).Elem()
		case 0x04:
			return reflect.TypeOf((*gbt2016.VehicleLogout)(nil)).Elem()
		case 0x05:
			// audit 2026-09-17:V2025 注册的是 PlatformLoginV2025/PlatformLogoutV2025,
			// 因此旧的 *gbt2016.Platform* 条目会让 GetCodec 返回 nil(ErrCodecNotFound)。
			return reflect.TypeOf((*gbt2025.PlatformLoginV2025)(nil)).Elem()
		case 0x06:
			return reflect.TypeOf((*gbt2025.PlatformLogoutV2025)(nil)).Elem()
		case 0x09:
			return reflect.TypeOf((*gbt2025.VehicleActivate)(nil)).Elem()
		case 0x0A:
			return reflect.TypeOf((*gbt2025.VehicleActivateResponse)(nil)).Elem()
		case 0x0B:
			return reflect.TypeOf((*gbt2025.KeyExchangeData)(nil)).Elem()
		}
	}
	return nil
}

// CommandCode 从 RequestType 中提取线上的命令字节。
func CommandCode(rt any) (byte, bool) {
	switch c := rt.(type) {
	case *types.CommandV2016:
		// audit 2026-09-17:CommandV2016ByCode 对未知字节
		// (0x00/0xFF)返回带类型的 nil,它仍会命中此类型 switch。
		// 解引用前必须先做防护,此前这里会 panic(可被远端触发的 DoS)。
		if c == nil {
			return 0, false
		}
		return c.Code, true
	case *types.CommandV2025:
		// audit 2026-09-17:参见上面的 *types.CommandV2016(带类型 nil 防护)。
		if c == nil {
			return 0, false
		}
		return c.Code, true
	}
	return 0, false
}

// Bytes 将整个协议帧编码为字节。
func (m *ProtocolMessage) Bytes() ([]byte, error) {
	w := utils.NewByteWriter()
	w.WriteString(types.Header(m.Version), 2)
	code, ok := CommandCode(m.RequestType)
	if !ok {
		return nil, errors.New("gb32960: RequestType is nil or unknown command type")
	}
	w.WriteUint8(code)
	w.WriteUint8(m.ResponseType.Code())
	w.WriteString(m.VIN, 17)
	w.WriteUint8(byte(m.Encryption))

	// audit 2026-09-17:仅当命令没有消息体类型时,
	// Payload 为 nil 才是合法的。0x07 心跳 / 0x08 校时
	// 按 GB/T 32960 附录 B 携带 0 字节数据单元,预留命令同理。
	// 有消息体类型的命令必须携带消息体,否则我们会发出畸形的帧。
	var payload []byte
	if m.Payload != nil {
		var err error
		payload, err = m.Payload.Bytes()
		if err != nil {
			return nil, err
		}
	} else if PayloadType(m.Version, code) != nil {
		return nil, errors.New("gb32960: Payload is nil for a command that requires a body")
	} else if len(m.RawBytes) > 0 {
		// fix 2026-09-17: 2016.md 表3 L65(0xC0~0xFE 平台交换自定义数据,
		// 含预留/未映射命令;同 2025.md 表3 L62)—— 命令没有注册消息体
		// 类型时,数据单元必须以解码时保存的原始字节透传,否则
		// 重编码会写出空数据单元。心跳 0x07/0x08 的 RawBytes 长度为 0,
		// 既有 0 字节数据单元行为保持不变。
		payload = m.RawBytes
	}

	if m.Encryption != types.EncryptionNone {
		return nil, api.ErrEncryptionNotSupported
	}

	// audit 2026-09-17:GB/T 32960 表2 将数据单元长度上限设为 65531 字节。
	if len(payload) > 65531 {
		return nil, api.ErrLengthMismatch
	}

	w.WriteUint16(uint16(len(payload)))
	w.WriteBytes(payload)

	bccRange := w.Bytes()[2:]
	w.WriteUint8(utils.CalcBCC(bccRange))

	return w.Bytes(), nil
}

// DecodePayload 将 RawBytes 解码到 Payload 字段。
func (m *ProtocolMessage) DecodePayload() error {
	if m.RawBytes == nil {
		return api.ErrBufferUnderflow
	}
	code, ok := CommandCode(m.RequestType)
	if !ok {
		return errors.New("gb32960: RequestType is nil or unknown command type")
	}

	msgType := PayloadType(m.Version, code)
	if msgType == nil {
		return nil
	}

	codec := api.GetCodec(m.Version, msgType)
	if codec == nil {
		return api.ErrCodecNotFound
	}

	r := utils.NewByteReader(m.RawBytes)
	decoded, err := codec.Decode(r)
	if err != nil {
		return err
	}

	var ok2 bool
	m.Payload, ok2 = decoded.(model.MessageBody)
	if !ok2 {
		return errors.New("gb32960: decoded message does not implement MessageBody")
	}
	return nil
}

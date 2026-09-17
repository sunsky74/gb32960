package types

// ResponseType 表示命令/应答方向。
// 取值与参考 Java 实现(ResponseType)对齐:
//
//	SUCCESS=0x01, FAILED=0x02, VIN_DUP=0x03, VIN_NOT_EXIST=0x04,
//	SIGN_ERR=0x05, STRUCTURE_ERR=0x06, DECODE_ERR=0x07, COMMAND=0xFE.
//
// 重要:Java 中 0x01 表示 SUCCESS(不是 "command")。早期草稿
// 搞乱了这些语义,见 audit 2026-07-31。
type ResponseType byte

const (
	ResponseSuccess      ResponseType = 0x01 // 成功；接收到的消息正确
	ResponseFailed       ResponseType = 0x02 // 错误；设置未成功
	ResponseVINDup       ResponseType = 0x03 // VIN 重复
	ResponseVINNotExist  ResponseType = 0x04 // VIN 不存在
	ResponseSignErr      ResponseType = 0x05 // 验签错误
	ResponseStructureErr ResponseType = 0x06 // 数据结构错误
	ResponseDecodeErr    ResponseType = 0x07 // 解密错误
	ResponseCommand      ResponseType = 0xFE // 命令；表示数据包为命令包，而非应答包
)

// Code 返回线格式字节。
func (r ResponseType) Code() byte { return byte(r) }

// ResponseByCode 返回给定线格式字节对应的 ResponseType。
// GB/T 32960-2016 表4 只定义了 0x01/0x02/0x03/0xFE;0x04-0x07 是 V2025
// 新增的;其余一律为预留 (reserved)。
// audit 2026-09-17:未知字节原样返回(ResponseType(code)),
// 而不是映射为 ResponseCommand:旧映射把畸形或未设置的帧伪装成合法命令包,
// 并在重编码时改写了该字节。
func ResponseByCode(code byte) ResponseType {
	switch code {
	case 0x01:
		return ResponseSuccess
	case 0x02:
		return ResponseFailed
	case 0x03:
		return ResponseVINDup
	case 0x04:
		return ResponseVINNotExist
	case 0x05:
		return ResponseSignErr
	case 0x06:
		return ResponseStructureErr
	case 0x07:
		return ResponseDecodeErr
	case 0xFE:
		return ResponseCommand
	default:
		return ResponseType(code)
	}
}

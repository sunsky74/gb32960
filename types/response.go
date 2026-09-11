package types

// ResponseType indicates command/response direction.
// Values aligned with the reference Java implementation (ResponseType):
//   SUCCESS=0x01, FAILED=0x02, VIN_DUP=0x03, VIN_NOT_EXIST=0x04,
//   SIGN_ERR=0x05, STRUCTURE_ERR=0x06, DECODE_ERR=0x07, COMMAND=0xFE.
// IMPORTANT: 0x01 means SUCCESS in Java (not "command"). Earlier draft
// scrambled these semantics — see audit 2026-07-31.
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

// Code returns the wire format byte.
func (r ResponseType) Code() byte { return byte(r) }

// ResponseByCode returns the ResponseType for a given wire byte.
// Unknown bytes default to ResponseCommand (matches Java behavior of
// treating unrecognized response bytes as command packets on the wire).
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
		return ResponseCommand
	}
}

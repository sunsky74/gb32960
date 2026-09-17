package api

import "errors"

// 用于生产可观测性的哨兵错误。
var (
	// 当某个消息类型没有注册编解码器时,返回 ErrCodecNotFound。
	ErrCodecNotFound = errors.New("gb32960: codec not found for message type")

	// 当协议帧的 BCC 校验码不正确时,返回 ErrBCCMismatch。
	ErrBCCMismatch = errors.New("gb32960: BCC checksum mismatch")

	// 当在 TLV 流中遇到无法识别的实时数据类型时,返回 ErrUnknownTLVType。
	ErrUnknownTLVType = errors.New("gb32960: unknown realtime data type in TLV stream")

	// 当字节不足以完成一次解码时,返回 ErrBufferUnderflow。
	ErrBufferUnderflow = errors.New("gb32960: buffer underflow during decode")

	// 当请求非直通加密模式时,返回 ErrEncryptionNotSupported。
	ErrEncryptionNotSupported = errors.New("gb32960: encryption mode not supported (only 0x01 pass-through is implemented)")

	// 当声明的数据单元长度与
	// 实际帧字节不一致,或超出规范范围 0~65531 时,返回 ErrLengthMismatch。
	ErrLengthMismatch = errors.New("gb32960: data unit length mismatch or out of range")

	// 当协议帧起始符不是 "##"(0x23,0x23;GB/T 32960-2016)
	// 或 "$$"(0x24,0x24;GB/T 32960-2025)时,返回 ErrInvalidHeader。
	ErrInvalidHeader = errors.New(`gb32960: invalid start flag (want "##" 0x23,0x23 or "$$" 0x24,0x24)`)
)

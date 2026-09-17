// Package api 定义 GB/T 32960 协议库的共享接口与类型。
// 它打破了 model 与 codec 包之间的循环依赖。
package api

// GBTVersion 表示协议版本。
type GBTVersion int

const (
	V2016 GBTVersion = 2016
	V2025 GBTVersion = 2025
)

// Message 是所有协议消息类型的标记接口。
// 每个 model 结构体(VehicleLogin、RealTimeData 等)都满足此接口。
type Message interface{}

// Reader 抽象了解码操作的字节读取。
// 方法只返回值(不返回 error),以便于链式调用。
// 实现必须做边界检查,并在下溢时返回零值,
// 同时记录可通过 Err() 获取的错误。编解码器应在
// Decode 末尾调用 r.Err() 并将其转换为 api.ErrBufferUnderflow。
// (Audit 2026-07-31:早期草稿在下溢时 panic;Java ByteBuffer
// 则抛出可恢复的 BufferUnderflowException。)
//
// 方法名使用 ReadUint8(而非 ReadByte),以避免与
// 标准库 io.ByteReader 接口冲突,该接口强制要求返回 (byte, error)。
// 我们这种不带 error 的签名是有意为之,与 ReadUint16/32 系列保持一致。
type Reader interface {
	Remaining() int
	ReadUint8() byte
	ReadUint16() uint16
	ReadUint32() uint32
	ReadString(n int) string
	ReadBytes(n int) []byte
	Err() error

	// Consumed 返回从位置 0 到当前读取位置之间已读取的字节。
	// 返回的切片是底层缓冲区的别名,如需保留请先复制。
	// 带签名的编解码器在解码时用它捕获 SignData
	// (签名之前的线格式字节),对应 Java
	// ByteBuffer 基于 readerIndex 的 dumpBytes。
	Consumed() []byte
}

// Writer 抽象了编码操作的字节写入。
// 方法名使用 WriteUint8(而非 WriteByte),以避免与
// 标准库 io.ByteWriter 接口冲突,该接口强制要求返回 error。
type Writer interface {
	WriteUint8(b byte)
	WriteUint16(v uint16)
	WriteUint32(v uint32)
	WriteString(s string, n int)
	WriteBytes(b []byte)
}

// Codecer 是消息体编解码器的接口。
// 每个消息类型都有一个对应的 Codecer 实现。
type Codecer interface {
	Decode(r Reader) (Message, error)
	Encode(w Writer, msg Message) error
}

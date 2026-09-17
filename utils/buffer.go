// Package utils 提供 GB/T 32960 协议处理的底层工具。
package utils

import (
	"bytes"
	"encoding/binary"

	"github.com/sunsky74/gb32960/api"
)

// 编译期接口检查
var _ api.Reader = (*ByteReader)(nil)
var _ api.Writer = (*ByteWriter)(nil)

// ByteReader 从字节切片读取二进制数据。
// 发生下溢时它返回零值并把错误记录在 err 中;
// 调用方在一系列读取之后检查 Err()。这对应 Java
// ByteBuffer 的 BufferUnderflowException 语义,同时无需让每个
// Reader 方法都返回 (T, error)。
type ByteReader struct {
	buf []byte
	pos int
	err error
}

// NewByteReader 从字节切片创建一个 ByteReader。
func NewByteReader(data []byte) *ByteReader {
	return &ByteReader{buf: data, pos: 0}
}

// Err 返回先前读取过程中记录的错误(通常为 api.ErrBufferUnderflow),
// 若所有读取都成功则返回 nil。编解码器应在 Decode 末尾调用它。
func (r *ByteReader) Err() error { return r.err }

// Consumed 返回从位置 0 到当前读取位置之间已读取的字节。
// 返回的切片是底层缓冲区的别名,
// 若要在下一次读取之后继续保留,请先复制。
func (r *ByteReader) Consumed() []byte {
	return r.buf[:r.pos]
}

// Remaining 返回未读取的字节数。
func (r *ByteReader) Remaining() int {
	return len(r.buf) - r.pos
}

// canRead 在 n 个字节不可用时记录下溢错误。
func (r *ByteReader) canRead(n int) bool {
	if r.err != nil {
		return false
	}
	if r.pos+n > len(r.buf) {
		r.err = api.ErrBufferUnderflow
		return false
	}
	return true
}

// ReadUint8 读取单个字节。下溢时返回 0 并记录 ErrBufferUnderflow。
// 方法名与 ReadUint16/32 系列保持一致,并避免与
// 标准库 io.ByteReader 接口冲突(该接口强制要求返回 (byte, error))。
func (r *ByteReader) ReadUint8() byte {
	if !r.canRead(1) {
		return 0
	}
	b := r.buf[r.pos]
	r.pos++
	return b
}

// ReadUint16 读取大端序 uint16。下溢时返回 0 并记录 ErrBufferUnderflow。
func (r *ByteReader) ReadUint16() uint16 {
	if !r.canRead(2) {
		return 0
	}
	v := binary.BigEndian.Uint16(r.buf[r.pos:])
	r.pos += 2
	return v
}

// ReadUint32 读取大端序 uint32。下溢时返回 0 并记录 ErrBufferUnderflow。
func (r *ByteReader) ReadUint32() uint32 {
	if !r.canRead(4) {
		return 0
	}
	v := binary.BigEndian.Uint32(r.buf[r.pos:])
	r.pos += 4
	return v
}

// ReadString 读取 n 个字节并返回去掉尾部填充的字符串。
// 尾部空格和 NUL 字节(0x00)都会被去除:GB/T 32960-2016 表1
// 将 STRING 定义为 ASCII,为空时以 0 作为终止符,而真实终端
// (以及参考 Java 实现)用空格填充定长字段,
// 为了互通,两者都要去除。
// 下溢时返回 "" 并记录 ErrBufferUnderflow。
func (r *ByteReader) ReadString(n int) string {
	if !r.canRead(n) {
		return ""
	}
	s := string(r.buf[r.pos : r.pos+n])
	r.pos += n
	// audit 2026-09-17 (L3):去除尾部空格和 NUL(表1 为空时的
	// 0 终止符 + 为对齐 Java 的空格填充)。
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == 0) {
		s = s[:len(s)-1]
	}
	return s
}

// ReadBytes 读取 n 个字节(返回副本)。
// 下溢时返回 nil 并记录 ErrBufferUnderflow。
func (r *ByteReader) ReadBytes(n int) []byte {
	if !r.canRead(n) {
		return nil
	}
	b := make([]byte, n)
	copy(b, r.buf[r.pos:r.pos+n])
	r.pos += n
	return b
}

// ByteWriter 通过二进制写入构建字节切片。
type ByteWriter struct {
	buf bytes.Buffer
}

// NewByteWriter 创建一个新的 ByteWriter。
func NewByteWriter() *ByteWriter {
	return &ByteWriter{}
}

// WriteUint8 写入单个字节。
// 方法名避免与标准库 io.ByteWriter 接口冲突
// (该接口强制要求返回 error)。
func (w *ByteWriter) WriteUint8(b byte) {
	w.buf.WriteByte(b)
}

// WriteUint16 写入大端序 uint16。
func (w *ByteWriter) WriteUint16(v uint16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, v)
	w.buf.Write(b)
}

// WriteUint32 写入大端序 uint32。
func (w *ByteWriter) WriteUint32(v uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	w.buf.Write(b)
}

// WriteString 写入定长字符串,必要时用尾部空格填充。
// 填充使用空格(不是 NUL),以便与参考 Java
// 实现和金样报文互通;ReadString 在解码时两者都接受
// (audit 2026-09-17, L3)。
func (w *ByteWriter) WriteString(s string, n int) {
	b := make([]byte, n)
	copy(b, s)
	// 用空格填充剩余字节(与 Java 行为一致)
	for i := len(s); i < n; i++ {
		b[i] = ' '
	}
	w.buf.Write(b)
}

// WriteBytes 写入原始字节。
func (w *ByteWriter) WriteBytes(b []byte) {
	w.buf.Write(b)
}

// Bytes 返回累积的字节。
func (w *ByteWriter) Bytes() []byte {
	return w.buf.Bytes()
}

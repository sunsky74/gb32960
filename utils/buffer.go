// Package utils provides low-level utilities for GB/T 32960 protocol handling.
package utils

import (
	"bytes"
	"encoding/binary"

	"github.com/sunsky74/gb32960/api"
)

// Compile-time interface check
var _ api.Reader = (*ByteReader)(nil)
var _ api.Writer = (*ByteWriter)(nil)

// ByteReader reads binary data from a byte slice.
// On underflow it returns the zero value and records the error in err;
// callers check Err() after a sequence of reads. This mirrors Java
// ByteBuffer's BufferUnderflowException semantics without forcing every
// Reader method to return (T, error).
type ByteReader struct {
	buf []byte
	pos int
	err error
}

// NewByteReader creates a ByteReader from a byte slice.
func NewByteReader(data []byte) *ByteReader {
	return &ByteReader{buf: data, pos: 0}
}

// Err returns any error recorded during prior reads (typically api.ErrBufferUnderflow),
// or nil if all reads succeeded. Codecs should call this at the end of Decode.
func (r *ByteReader) Err() error { return r.err }

// Consumed returns the bytes read so far, from position 0 to the current read
// position. The returned slice aliases the underlying buffer — copy before
// retaining it beyond the next read.
func (r *ByteReader) Consumed() []byte {
	return r.buf[:r.pos]
}

// Remaining returns the number of unread bytes.
func (r *ByteReader) Remaining() int {
	return len(r.buf) - r.pos
}

// canRead records an underflow error if n bytes are not available.
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

// ReadUint8 reads a single byte. Returns 0 and records ErrBufferUnderflow on underflow.
// Method name aligns with the ReadUint16/32 family and avoids colliding with the
// stdlib io.ByteReader interface (which mandates a (byte, error) return).
func (r *ByteReader) ReadUint8() byte {
	if !r.canRead(1) {
		return 0
	}
	b := r.buf[r.pos]
	r.pos++
	return b
}

// ReadUint16 reads a big-endian uint16. Returns 0 and records ErrBufferUnderflow on underflow.
func (r *ByteReader) ReadUint16() uint16 {
	if !r.canRead(2) {
		return 0
	}
	v := binary.BigEndian.Uint16(r.buf[r.pos:])
	r.pos += 2
	return v
}

// ReadUint32 reads a big-endian uint32. Returns 0 and records ErrBufferUnderflow on underflow.
func (r *ByteReader) ReadUint32() uint32 {
	if !r.canRead(4) {
		return 0
	}
	v := binary.BigEndian.Uint32(r.buf[r.pos:])
	r.pos += 4
	return v
}

// ReadString reads n bytes and returns a trimmed string.
// Returns "" and records ErrBufferUnderflow on underflow.
func (r *ByteReader) ReadString(n int) string {
	if !r.canRead(n) {
		return ""
	}
	s := string(r.buf[r.pos : r.pos+n])
	r.pos += n
	// Trim trailing spaces (matching Java behavior)
	for len(s) > 0 && s[len(s)-1] == ' ' {
		s = s[:len(s)-1]
	}
	return s
}

// ReadBytes reads n bytes (returns a copy).
// Returns nil and records ErrBufferUnderflow on underflow.
func (r *ByteReader) ReadBytes(n int) []byte {
	if !r.canRead(n) {
		return nil
	}
	b := make([]byte, n)
	copy(b, r.buf[r.pos:r.pos+n])
	r.pos += n
	return b
}

// ByteWriter builds a byte slice from binary writes.
type ByteWriter struct {
	buf bytes.Buffer
}

// NewByteWriter creates a new ByteWriter.
func NewByteWriter() *ByteWriter {
	return &ByteWriter{}
}

// WriteUint8 writes a single byte.
// Method name avoids colliding with the stdlib io.ByteWriter interface
// (which mandates an error return).
func (w *ByteWriter) WriteUint8(b byte) {
	w.buf.WriteByte(b)
}

// WriteUint16 writes a big-endian uint16.
func (w *ByteWriter) WriteUint16(v uint16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, v)
	w.buf.Write(b)
}

// WriteUint32 writes a big-endian uint32.
func (w *ByteWriter) WriteUint32(v uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	w.buf.Write(b)
}

// WriteString writes a fixed-length string, padding with trailing spaces if needed.
func (w *ByteWriter) WriteString(s string, n int) {
	b := make([]byte, n)
	copy(b, s)
	// Pad remaining bytes with spaces (matching Java behavior)
	for i := len(s); i < n; i++ {
		b[i] = ' '
	}
	w.buf.Write(b)
}

// WriteBytes writes raw bytes.
func (w *ByteWriter) WriteBytes(b []byte) {
	w.buf.Write(b)
}

// Bytes returns the accumulated bytes.
func (w *ByteWriter) Bytes() []byte {
	return w.buf.Bytes()
}

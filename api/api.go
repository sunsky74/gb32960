// Package api defines shared interfaces and types for the GB/T 32960 protocol library.
// It breaks the circular dependency between model and codec packages.
package api

// GBTVersion represents the protocol version.
type GBTVersion int

const (
	V2016 GBTVersion = 2016
	V2025 GBTVersion = 2025
)

// Message is the marker interface for all protocol message types.
// Every model struct (VehicleLogin, RealTimeData, etc.) satisfies this interface.
type Message interface{}

// Reader abstracts byte reading for decoding operations.
// Methods return only the value (no error) for ergonomic chaining.
// Implementations MUST bounds-check and, on underflow, return the zero
// value and record the error retrievable via Err(). Codecs should call
// r.Err() at the end of Decode and convert it to api.ErrBufferUnderflow.
// (Audit 2026-07-31: earlier draft panicked on underflow; Java ByteBuffer
// throws a recoverable BufferUnderflowException instead.)
//
// Method names use ReadUint8 (not ReadByte) to avoid colliding with the
// standard io.ByteReader interface, which mandates a (byte, error) return.
// Our no-error signature is intentional and matches the ReadUint16/32 family.
type Reader interface {
	Remaining() int
	ReadUint8() byte
	ReadUint16() uint16
	ReadUint32() uint32
	ReadString(n int) string
	ReadBytes(n int) []byte
	Err() error

	// Consumed returns the bytes read so far, from position 0 to the current
	// read position. The returned slice ALIASES the underlying buffer — copy
	// before retaining. Used by signature-bearing codecs to capture SignData
	// (the wire bytes preceding the signature) during decode, mirroring Java
	// ByteBuffer's readerIndex-based dumpBytes.
	Consumed() []byte
}

// Writer abstracts byte writing for encoding operations.
// Method names use WriteUint8 (not WriteByte) to avoid colliding with the
// standard io.ByteWriter interface, which mandates an error return.
type Writer interface {
	WriteUint8(b byte)
	WriteUint16(v uint16)
	WriteUint32(v uint32)
	WriteString(s string, n int)
	WriteBytes(b []byte)
}

// Codecer is the interface for message body codecs.
// Each message type has a corresponding Codecer implementation.
type Codecer interface {
	Decode(r Reader) (Message, error)
	Encode(w Writer, msg Message) error
}

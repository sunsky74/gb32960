package api

import "errors"

// Sentinel errors for production observability.
var (
	// ErrCodecNotFound is returned when no codec is registered for a message type.
	ErrCodecNotFound = errors.New("gb32960: codec not found for message type")

	// ErrBCCMismatch is returned when the BCC checksum of a protocol frame is incorrect.
	ErrBCCMismatch = errors.New("gb32960: BCC checksum mismatch")

	// ErrUnknownTLVType is returned when an unrecognized realtime data type is encountered in TLV stream.
	ErrUnknownTLVType = errors.New("gb32960: unknown realtime data type in TLV stream")

	// ErrBufferUnderflow is returned when there are insufficient bytes to complete a decode.
	ErrBufferUnderflow = errors.New("gb32960: buffer underflow during decode")

	// ErrEncryptionNotSupported is returned when a non-pass-through encryption mode is requested.
	ErrEncryptionNotSupported = errors.New("gb32960: encryption mode not supported (only 0x01 pass-through is implemented)")
)

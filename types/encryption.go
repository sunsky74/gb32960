package types

// EncryptionType represents the encryption mode of a protocol frame.
type EncryptionType byte

const (
	EncryptionNone      EncryptionType = 0x01 // No encryption (pass-through)
	EncryptionRSA       EncryptionType = 0x02 // RSA encryption (not implemented)
	EncryptionAES128    EncryptionType = 0x03 // AES-128 encryption (not implemented)
	EncryptionException EncryptionType = 0xFE // Encryption exception
	EncryptionInvalid   EncryptionType = 0xFF // Invalid
)

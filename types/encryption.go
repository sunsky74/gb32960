package types

// EncryptionType 表示协议帧的加密方式。
type EncryptionType byte

const (
	EncryptionNone      EncryptionType = 0x01 // 不加密(直通)
	EncryptionRSA       EncryptionType = 0x02 // RSA 加密(未实现)
	EncryptionAES128    EncryptionType = 0x03 // AES-128 加密(未实现)
	EncryptionException EncryptionType = 0xFE // 加密异常
	EncryptionInvalid   EncryptionType = 0xFF // 无效
)

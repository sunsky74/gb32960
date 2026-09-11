package utils

import "encoding/hex"

// HexToBytes decodes a hex string to bytes.
func HexToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// BytesToHex encodes bytes to a hex string.
func BytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

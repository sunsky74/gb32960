package utils

import "encoding/hex"

// HexToBytes 将十六进制字符串解码为字节。
func HexToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// BytesToHex 将字节编码为十六进制字符串。
func BytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

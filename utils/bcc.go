package utils

// CalcBCC computes the XOR checksum over all bytes.
// Equivalent to Java: ByteCheckUtils.getBcc(data)
func CalcBCC(data []byte) byte {
	var b byte
	for _, v := range data {
		b ^= v
	}
	return b
}

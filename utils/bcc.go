package utils

// CalcBCC 对所有字节计算异或校验码。
// 等价于 Java:ByteCheckUtils.getBcc(data)
func CalcBCC(data []byte) byte {
	var b byte
	for _, v := range data {
		b ^= v
	}
	return b
}

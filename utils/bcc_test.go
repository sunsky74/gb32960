package utils

import "testing"

func TestCalcBCC(t *testing.T) {
	tests := []struct {
		data     []byte
		expected byte
	}{
		{[]byte{0x01, 0x02, 0x03}, 0x00},       // 1^2^3 = 0
		{[]byte{0xFF, 0xFF, 0xFF}, 0xFF},       // 0xFF ^ 0xFF ^ 0xFF = 0xFF
		{[]byte{0x00, 0x00, 0x00, 0x01}, 0x01}, // 0^0^0^1 = 1
		{[]byte{}, 0x00},                       // 空
		{[]byte{0xAA, 0x55}, 0xFF},             // 0xAA ^ 0x55 = 0xFF
	}

	for _, tt := range tests {
		got := CalcBCC(tt.data)
		if got != tt.expected {
			t.Errorf("CalcBCC(%X) = 0x%02X, want 0x%02X", tt.data, got, tt.expected)
		}
	}
}

func TestCalcBCC_SingleByte(t *testing.T) {
	if got := CalcBCC([]byte{0x42}); got != 0x42 {
		t.Errorf("single byte: want 0x42, got 0x%02X", got)
	}
}

func TestCalcBCC_Repeating(t *testing.T) {
	// 相同字节出现偶数次时异或结果为 0x00
	if got := CalcBCC([]byte{0x55, 0x55}); got != 0x00 {
		t.Errorf("repeating even: want 0x00, got 0x%02X", got)
	}
	// 三个相同字节
	if got := CalcBCC([]byte{0xAA, 0xAA, 0xAA}); got != 0xAA {
		t.Errorf("repeating odd: want 0xAA, got 0x%02X", got)
	}
}

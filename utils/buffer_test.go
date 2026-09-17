package utils

import (
	"testing"
)

func TestByteReader_Basic(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	r := NewByteReader(data)

	if r.Remaining() != 5 {
		t.Fatalf("expected 5 remaining, got %d", r.Remaining())
	}

	if b := r.ReadUint8(); b != 0x01 {
		t.Errorf("expected 0x01, got 0x%02X", b)
	}
	if r.Remaining() != 4 {
		t.Errorf("expected 4 remaining, got %d", r.Remaining())
	}
}

func TestByteReader_Uint16_BigEndian(t *testing.T) {
	data := []byte{0x12, 0x34}
	r := NewByteReader(data)

	if v := r.ReadUint16(); v != 0x1234 {
		t.Errorf("expected 0x1234, got 0x%04X", v)
	}
}

func TestByteReader_ReadString_TrimsSpaces(t *testing.T) {
	data := []byte("HELLO           ") // 15 字节,"HELLO" + 10 个空格
	r := NewByteReader(data)

	s := r.ReadString(15)
	if s != "HELLO" {
		t.Errorf("expected 'HELLO', got '%s'", s)
	}
}

// TestByteReader_ReadString_TrimsNULAndSpaces 锁定 audit 2026-09-17 (L3)
// 契约:GB/T 32960-2016 表1 将 STRING 定义为为空时带 0x00 终止符,
// 而真实终端用 ' ' 或 0x00 填充定长字段,
// ReadString 必须去除这两种尾部填充。
func TestByteReader_ReadString_TrimsNULAndSpaces(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		want string
	}{
		{"NUL padding", []byte("AB\x00\x00"), "AB"},
		{"space padding", []byte("AB  "), "AB"},
		{"mixed padding", []byte("A\x00 "), "A"},
		{"all NUL", []byte{0x00, 0x00, 0x00}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewByteReader(tc.in)
			if got := r.ReadString(len(tc.in)); got != tc.want {
				t.Errorf("ReadString(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestByteWriter_WriteString_PadsSpaces(t *testing.T) {
	w := NewByteWriter()
	w.WriteString("HI", 5)
	result := w.Bytes()

	expected := []byte{'H', 'I', ' ', ' ', ' '}
	if len(result) != 5 {
		t.Fatalf("expected 5 bytes, got %d", len(result))
	}
	for i, b := range result {
		if b != expected[i] {
			t.Errorf("byte[%d]: expected %d, got %d", i, expected[i], b)
		}
	}
}

func TestByteWriter_Roundtrip(t *testing.T) {
	w := NewByteWriter()
	w.WriteUint8(0xAA)
	w.WriteUint16(0x1234)
	w.WriteUint32(0xDEADBEEF)

	data := w.Bytes()
	r := NewByteReader(data)

	if b := r.ReadUint8(); b != 0xAA {
		t.Errorf("byte roundtrip failed: expected 0xAA, got 0x%02X", b)
	}
	if v := r.ReadUint16(); v != 0x1234 {
		t.Errorf("uint16 roundtrip failed: expected 0x1234, got 0x%04X", v)
	}
	if v := r.ReadUint32(); v != 0xDEADBEEF {
		t.Errorf("uint32 roundtrip failed: expected 0xDEADBEEF, got 0x%08X", v)
	}
}

// TestByteReader_Underflow 验证 audit 2026-07-31 加入的
// 边界检查策略:越界读取返回零值,并
// 通过 Err() 记录 api.ErrBufferUnderflow,而不是 panic。
func TestByteReader_Underflow(t *testing.T) {
	r := NewByteReader([]byte{0x01}) // 只有 1 个字节可用

	// 首次读取成功
	if b := r.ReadUint8(); b != 0x01 {
		t.Fatalf("expected 0x01, got 0x%02X", b)
	}
	if err := r.Err(); err != nil {
		t.Fatalf("unexpected err after valid read: %v", err)
	}

	// 后续读取发生下溢:返回零值,不 panic
	if v := r.ReadUint16(); v != 0 {
		t.Errorf("underflow ReadUint16: expected 0, got 0x%04X", v)
	}
	if v := r.ReadUint32(); v != 0 {
		t.Errorf("underflow ReadUint32: expected 0, got 0x%08X", v)
	}
	if s := r.ReadString(5); s != "" {
		t.Errorf("underflow ReadString: expected empty, got %q", s)
	}

	// 此时 Err() 必须报告下溢
	if err := r.Err(); err == nil {
		t.Fatal("expected ErrBufferUnderflow after underflow reads, got nil")
	}
}

package codec_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/model/gbt2016"
	"github.com/sunsky74/gb32960/utils"

	// 空导入以触发编解码器 init() 注册。
	_ "github.com/sunsky74/gb32960/codec/gbt2016"
)

// TestPlatformLogin_Roundtrip 是 Gate-0 验收测试:
// 编码 → 解码 → 重编码必须逐字节一致,且所有
// 字段都必须完整通过往返。
func TestPlatformLogin_Roundtrip(t *testing.T) {
	original := &gbt2016.PlatformLogin{
		BeanTime:  model.BeanTime{Year: 26, Month: 7, Day: 31, Hour: 10, Minute: 30, Second: 0},
		SerialNum: 1,
		Username:  "admin",
		Password:  "secret123",
		Cipher:    0x01, // 0x01 = NONE 直通(与 Java NonCipher 默认值一致)
	}

	// 通过注册的编解码器编码(委托给 modelutil.DefaultBytes)。
	encoded, err := original.Bytes()
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	const wantLen = 6 + 2 + 12 + 20 + 1 // BeanTime + SerialNum + Username + Password + Cipher(线格式各字段长度)
	if len(encoded) != wantLen {
		t.Fatalf("encoded length: got %d, want %d (bytes=%X)", len(encoded), wantLen, encoded)
	}

	// 通过反射类型查找的编解码器解码。
	codec := api.GetCodec(api.V2016, reflect.TypeOf((*gbt2016.PlatformLogin)(nil)).Elem())
	if codec == nil {
		t.Fatal("codec not found for PlatformLogin")
	}
	decoded, err := codec.Decode(utils.NewByteReader(encoded))
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	result := decoded.(*gbt2016.PlatformLogin)

	// 校验字段。
	if result.BeanTime != original.BeanTime {
		t.Errorf("BeanTime mismatch: got %v, want %v", result.BeanTime, original.BeanTime)
	}
	if result.SerialNum != original.SerialNum {
		t.Errorf("SerialNum mismatch: got %d, want %d", result.SerialNum, original.SerialNum)
	}
	if result.Username != original.Username {
		t.Errorf("Username mismatch: got %q, want %q", result.Username, original.Username)
	}
	if result.Password != original.Password {
		t.Errorf("Password mismatch: got %q, want %q", result.Password, original.Password)
	}
	if result.Cipher != original.Cipher {
		t.Errorf("Cipher mismatch: got 0x%02X, want 0x%02X", result.Cipher, original.Cipher)
	}

	// 重编码并校验字节稳定性。
	reEncoded, err := result.Bytes()
	if err != nil {
		t.Fatalf("re-encode failed: %v", err)
	}
	if !bytes.Equal(encoded, reEncoded) {
		t.Errorf("re-encode differs:\n  original  = %X\n  reEncoded = %X", encoded, reEncoded)
	}
}

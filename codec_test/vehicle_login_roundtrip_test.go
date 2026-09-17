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
	// 父包 gbt2016 注册 VehicleLogin 和 RealTimeData
	// 分发表;realtime 子包也必须导入,这样其
	// init() 才会注册分发表要查找的各子编解码器。
	_ "github.com/sunsky74/gb32960/codec/gbt2016"
	_ "github.com/sunsky74/gb32960/codec/gbt2016/realtime"
)

// TestVehicleLogin_Roundtrip 是 Plan 2 Part C2 Step 2 验收测试:
// 编码 → 解码 → 重编码必须逐字节一致,且 ICCID 和 Codes
// 字段必须完整通过往返(包括对填充字节的右截断)。
func TestVehicleLogin_Roundtrip(t *testing.T) {
	original := &gbt2016.VehicleLogin{
		BeanTime:  model.BeanTime{Year: 26, Month: 7, Day: 31, Hour: 10, Minute: 30, Second: 0},
		SerialNum: 1,
		ICCID:     "89860000000000000001",
		Count:     2,
		Length:    3,
		Codes:     []string{"001", "002"},
	}

	encoded, err := original.Bytes()
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// 线格式布局:BeanTime(6) + SerialNum(2) + ICCID(20) + Count(1) + Length(1) + 2*3 = 34
	const wantLen = 6 + 2 + 20 + 1 + 1 + 2*3
	if len(encoded) != wantLen {
		t.Fatalf("encoded length: got %d, want %d (bytes=%X)", len(encoded), wantLen, encoded)
	}

	codec := api.GetCodec(api.V2016, reflect.TypeOf((*gbt2016.VehicleLogin)(nil)).Elem())
	if codec == nil {
		t.Fatal("codec not found for VehicleLogin")
	}
	decoded, err := codec.Decode(utils.NewByteReader(encoded))
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	result := decoded.(*gbt2016.VehicleLogin)

	// 校验字段。
	if result.BeanTime != original.BeanTime {
		t.Errorf("BeanTime mismatch: got %v, want %v", result.BeanTime, original.BeanTime)
	}
	if result.SerialNum != original.SerialNum {
		t.Errorf("SerialNum mismatch: got %d, want %d", result.SerialNum, original.SerialNum)
	}
	if result.ICCID != original.ICCID {
		t.Errorf("ICCID mismatch: got %q, want %q", result.ICCID, original.ICCID)
	}
	if result.Count != original.Count {
		t.Errorf("Count mismatch: got %d, want %d", result.Count, original.Count)
	}
	if result.Length != original.Length {
		t.Errorf("Length mismatch: got %d, want %d", result.Length, original.Length)
	}
	if len(result.Codes) != len(original.Codes) {
		t.Fatalf("Codes length: got %d, want %d", len(result.Codes), len(original.Codes))
	}
	for i, want := range original.Codes {
		if result.Codes[i] != want {
			t.Errorf("Codes[%d]: got %q, want %q", i, result.Codes[i], want)
		}
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

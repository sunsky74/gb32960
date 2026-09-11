package codec_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/model/gbt2016"
	"github.com/sunsky74/gb32960/utils"

	// Blank-import to trigger codec init() registration.
	// The parent gbt2016 package registers VehicleLogin + the RealTimeData
	// dispatch map; the realtime sub-package must also be imported so its
	// init() registers the sub-codecs that the dispatch map looks up.
	_ "github.com/sunsky74/gb32960/codec/gbt2016"
	_ "github.com/sunsky74/gb32960/codec/gbt2016/realtime"
)

// TestVehicleLogin_Roundtrip is the Plan 2 Part C2 Step 2 acceptance test:
// encode → decode → re-encode must be byte-identical, and the ICCID and Codes
// fields must survive the round trip (including right-trim of pad bytes).
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

	// Wire layout: BeanTime(6) + SerialNum(2) + ICCID(20) + Count(1) + Length(1) + 2*3 = 34
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

	// Verify fields.
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

	// Re-encode and verify byte stability.
	reEncoded, err := result.Bytes()
	if err != nil {
		t.Fatalf("re-encode failed: %v", err)
	}
	if !bytes.Equal(encoded, reEncoded) {
		t.Errorf("re-encode differs:\n  original  = %X\n  reEncoded = %X", encoded, reEncoded)
	}
}

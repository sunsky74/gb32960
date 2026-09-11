package codec_test

import (
	"math"
	"testing"

	"github.com/sunsky74/gb32960/codec"
	"github.com/sunsky74/gb32960/types"
)

func TestValueConverter_Speed(t *testing.T) {
	vc := codec.SpeedConverter
	if got := vc.Decode(500); math.Abs(got-50.0) > 0.01 {
		t.Errorf("Decode(500) = %f, want 50.0", got)
	}
	if got := vc.Decode(65534); got != 65534.0 {
		t.Errorf("Decode(65534) = %f, want 65534.0 (passthrough)", got)
	}
	if got := vc.Decode(65535); got != 65535.0 {
		t.Errorf("Decode(65535) = %f, want 65535.0 (passthrough)", got)
	}
	if got := vc.Encode(50.0); got != 500 {
		t.Errorf("Encode(50.0) = %d, want 500", got)
	}
}

func TestValueConverter_Current(t *testing.T) {
	vc := codec.CurrentConverter2016
	// raw=10500 → (10500/10 - 1000) = 50.0
	if got := vc.Decode(10500); math.Abs(got-50.0) > 0.01 {
		t.Errorf("Decode(10500) = %f, want 50.0", got)
	}
	if got := vc.Decode(65534); got != 65534.0 {
		t.Errorf("Decode(65534) = %f, want 65534.0", got)
	}
}

func TestValueConverter_NegativeOffset(t *testing.T) {
	vc := codec.CurrentConverterChargeElectric
	// Offset fixed -1000→+1000 (simulator audit 2026-08-26, see converter doc):
	// raw=100 → (100/10 - 1000) = -990.0, matching GB/T 32960 and the golden
	// packet where chargeable current raw equals vehicle current raw (2.5A).
	if got := vc.Decode(100); math.Abs(got-(-990.0)) > 0.01 {
		t.Errorf("Decode(100) = %f, want -990.0", got)
	}
	// Encode: val=-990.0 → (-990+1000)*10 = 100
	if got := vc.Encode(-990.0); got != 100 {
		t.Errorf("Encode(-990.0) = %d, want 100", got)
	}
}

func TestValueConverter_DataErrorValue_IsInvalid(t *testing.T) {
	ev := types.DataErrorValue{Error: 65534, Invalid: 65535}
	if !ev.IsInvalid(65534) {
		t.Error("65534 should be invalid")
	}
	if !ev.IsInvalid(65535) {
		t.Error("65535 should be invalid")
	}
	if ev.IsInvalid(100) {
		t.Error("100 should NOT be invalid")
	}
}

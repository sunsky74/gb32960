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
	// 原始值=10500 → (10500/10 - 1000) = 50.0
	if got := vc.Decode(10500); math.Abs(got-50.0) > 0.01 {
		t.Errorf("Decode(10500) = %f, want 50.0", got)
	}
	if got := vc.Decode(65534); got != 65534.0 {
		t.Errorf("Decode(65534) = %f, want 65534.0", got)
	}
}

func TestValueConverter_NegativeOffset(t *testing.T) {
	vc := codec.CurrentConverterChargeElectric
	// 偏移量已从 -1000 固定为 +1000(模拟器 audit 2026-08-26,见转换器文档):
	// 原始值=100 → (100/10 - 1000) = -990.0,与 GB/T 32960 及金样
	// 报文一致,其中可充电电流原始值等于车辆电流原始值(2.5A)。
	if got := vc.Decode(100); math.Abs(got-(-990.0)) > 0.01 {
		t.Errorf("Decode(100) = %f, want -990.0", got)
	}
	// 编码:val=-990.0 → (-990+1000)*10 = 100
	if got := vc.Encode(-990.0); got != 100 {
		t.Errorf("Encode(-990.0) = %d, want 100", got)
	}
}

// TestValueConverter_RoundTripStability 锁定 audit 2026-09-17 (H4) 修复:
// 那些 float64 解码乘积刚好落在目标整数之下的原始值
// (如燃料原始值 29 -> 0.29 -> 28.999999999999996)必须仍能 Encode
// 回精确的原始值。下面每个样本在修复前都失败。
func TestValueConverter_RoundTripStability(t *testing.T) {
	cases := []struct {
		name string
		vc   codec.ValueConverter
		raws []int64
	}{
		{"FuelConsumptionRate", codec.FuelConsumptionRateConverter, []int64{29, 57, 58, 113, 114, 115, 116, 201}},
		{"MotorTorque2016", codec.MotorTorqueConverter2016, []int64{1, 3, 6}},
		{"Current2016", codec.CurrentConverter2016, []int64{3, 4, 8, 9}},
		{"BatteryVoltage", codec.BatteryVoltageConverter, []int64{1001, 1003, 1005}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, raw := range c.raws {
				decoded := c.vc.Decode(raw)
				if got := c.vc.Encode(decoded); got != raw {
					t.Errorf("round-trip: raw %d -> decode %v -> encode %d, want %d", raw, decoded, got, raw)
				}
			}
		})
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

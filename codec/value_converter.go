package codec

import (
	"github.com/sunsky74/gb32960/types"
)

// ValueConverter converts between raw wire values and decoded float64 values.
// Uses the same formula as Java ValueConverter.java:
//
//	Decode: decoded = (raw / scale) - offset   [BigDecimal.divide(scale, precision, HALF_UP)]
//	Encode: raw = (val + offset) * scale       [BigDecimal.multiply then .longValue() — TRUNCATION toward zero]
//
// IMPORTANT (audit 2026-07-31): Java ValueConverter.encode returns
// `result.longValue()` which TRUNCATES toward zero. Earlier draft used
// `math.Round` — that was WRONG and diverged by 1 raw ULP for any product
// with fractional part >= 0.5. Plan 3 Layer B golden generator uses
// `result.longValue()` too, so truncation is what makes Go match Java.
type ValueConverter struct {
	ErrValue  types.DataErrorValue
	Scale     float64
	Offset    float64
	Precision int // decimal places for display
}

// Decode converts a raw wire value to a decoded float64.
// If raw is an error/invalid sentinel, returns raw as float64 (passthrough).
func (vc *ValueConverter) Decode(raw int64) float64 {
	if vc.ErrValue.IsInvalid(raw) {
		return float64(raw)
	}
	return float64(raw)/vc.Scale - vc.Offset
}

// Encode converts a decoded float64 back to a raw wire value.
// Mirrors Java BigDecimal.longValue(): TRUNCATION toward zero.
// Do NOT use math.Round here — that would diverge from Java by 1 ULP on
// half-integer products and silently corrupt bytes.
func (vc *ValueConverter) Encode(val float64) int64 {
	longVal := int64(val) // truncation toward zero — matches Java value.longValue()
	if vc.ErrValue.IsInvalid(longVal) {
		return longVal
	}
	result := (val + vc.Offset) * vc.Scale
	return int64(result) // truncation toward zero — matches Java result.longValue()
}

// ---- High Risk (scale >= 1000) ----
// Must pass full-range 0~65535 validation against Java output.

var BatteryVoltageConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 1000, Offset: 0, Precision: 3,
}

var ExtremumVoltageConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 1000, Offset: 0, Precision: 2,
}

var BatteryVoltagesConverter2025 = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 1000, Offset: 0, Precision: 3,
}

var SuperCapVoltageConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 1000, Offset: 0, Precision: 3,
}

var SuperCapExtremumVoltageConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 1000, Offset: 0, Precision: 3,
}

var ConcentrationConverter2025 = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10000, Offset: 0, Precision: 4,
}

// ---- Medium Risk (scale=100 or negative offset) ----
// Must pass 0~50000 critical range validation.

var FuelConsumptionRateConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 100, Offset: 0, Precision: 2,
}

// CurrentConverterChargeElectric converts the 可充电储能装置电压数据 current
// field (ChargeableSubsystemElectric.Current, TLV 0x08).
//
// FIX (simulator audit 2026-08-26): offset was -1000, mirrored from Java
// ChargeableSubsystemElectricCodec.CURRENT_CONVERTER. That sign is wrong:
// decode became raw/10+1000 and encode (v-1000)*10 — mathematically
// self-consistent byte-wise but semantically inverted vs GB/T 32960.3-2016
// (raw = (value+1000)×10). Evidence: golden packet prod_realtime_v2016_01.hex
// carries current raw=0x2900=10496 → +1000 semantics decodes 49.6A (sane),
// -1000 semantics decodes 2049.6A (impossible). Re-encode of golden packets
// stays byte-identical under this fix (offset only flips interpretation).
// Java ChargeableSubsystemElectricCodec.CURRENT_CONVERTER 已同步为 +1000,两侧与国标一致。
var CurrentConverterChargeElectric = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 1000, Precision: 1,
}

// ---- Low Risk (scale <= 10, positive offset) ----
// At least 10+ boundary test cases each.

var SpeedConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 0, Precision: 2,
}

var VoltageConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 0, Precision: 2,
}

var CurrentConverter2016 = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 1000, Precision: 2,
}

var CurrentConverter2025 = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 3000, Precision: 1,
}

var MileageConverter = ValueConverter{
	ErrValue: types.ErrByte4, Scale: 10, Offset: 0, Precision: 1,
}

var ControllerTempConverter = ValueConverter{
	ErrValue: types.ErrByte1, Scale: 1, Offset: 40, Precision: 0,
}

var MotorSpeedConverter2016 = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 1, Offset: 20000, Precision: 0,
}

var MotorSpeedConverter2025 = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 1, Offset: 32000, Precision: 0,
}

var MotorTorqueConverter2016 = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 2000, Precision: 2,
}

var MotorTorqueConverter2025 = ValueConverter{
	ErrValue: types.ErrByte4, Scale: 10, Offset: 20000, Precision: 1,
}

var MotorTempConverter = ValueConverter{
	ErrValue: types.ErrByte1, Scale: 1, Offset: 40, Precision: 0,
}

var ControllerVoltageConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 0, Precision: 2,
}

var ControllerCurrentConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 1000, Precision: 2,
}

var TemperatureConverter = ValueConverter{
	ErrValue: types.ErrByte1, Scale: 1, Offset: 40, Precision: 0,
}

var FuelCellVoltageConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 0, Precision: 2,
}

var FuelCellCurrentConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 0, Precision: 2,
}

var HighestTempHydrogenConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 40, Precision: 2,
}

var HydrogenMaxPressureConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 0, Precision: 2,
}

var ProbeTemperatureConverter = ValueConverter{
	ErrValue: types.ErrByte1, Scale: 1, Offset: 40, Precision: 0,
}

var GasPressureConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 100, Precision: 1,
}

var AirPressureConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 100, Precision: 1,
}

var MaxPressureConverter2025 = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 0, Precision: 0,
}

var TempConverterEngine2025 = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 1, Offset: 40, Precision: 0,
}

var TotalVoltageConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 0, Precision: 1,
}

var TotalCurrentConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 3000, Precision: 1,
}

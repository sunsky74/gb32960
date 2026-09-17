package codec

import (
	"math"

	"github.com/sunsky74/gb32960/types"
)

// ValueConverter 在原始线上值与解码后的 float64 值之间转换。
// 使用与 Java ValueConverter.java 相同的公式:
//
//	Decode: decoded = (raw / scale) - offset   [BigDecimal.divide(scale, precision, HALF_UP)]
//	Encode: raw = (val + offset) * scale       [BigDecimal.multiply, 精确小数]
//
// audit 2026-09-17 (修复 H4):Java 用精确的 BigDecimal 小数做 Encode
// 运算,所以它的 .longValue() 截断只会丢掉调用方提供的超出字段精度的小数位。
// Go 用 float64 计算:来自 Decode 的值是最接近 raw/scale - offset 的
// float64(例如 raw 29、scale 100 解码为 0.28999999999999998),因此乘积
// (val+offset)*scale 可能比预期的原始整数低一个 ULP
// (0.29*100 = 28.999999999999996),截断会返回 28。改用 math.Round
// 可以还原精确的原始整数。Decode 的输出表示小数点后最多 log10(scale)
// 位的十进制数,所以 float64 误差约为 1e-11 个原始单位,
// 远低于 0.5,Round 对 Decode 产出的每个值都精确。
//
// 已知差异(已接受、已记录):对小数位多于字段精度的调用方输入值,
// Java 会截断精确的十进制乘积(例如 2.345*100 = 234.5 -> 234),
// 而 Go 取整(-> 235)。这不会影响解码->编码往返,
// 而往返正是线上兼容性所依赖的不变量。
type ValueConverter struct {
	ErrValue  types.DataErrorValue
	Scale     float64
	Offset    float64
	Precision int // 展示用小数位数
}

// Decode 将原始线上值转换为解码后的 float64。
// 如果 raw 是异常/无效哨兵值,则以 float64 原样返回(直通)。
func (vc *ValueConverter) Decode(raw int64) float64 {
	if vc.ErrValue.IsInvalid(raw) {
		return float64(raw)
	}
	return float64(raw)/vc.Scale - vc.Offset
}

// Encode 将解码后的 float64 转换回原始线上值。
// 取整而非截断:Decode 的输出是二进制近似值,所以
// float64 乘积可能比预期整数低一个 ULP(0.29*100 =
// 28.999999999999996),截断会静默返回错误的字节。
// 完整理由以及调用方输入值带多余小数位时的残余差异,
// 见 ValueConverter 的文档注释。
func (vc *ValueConverter) Encode(val float64) int64 {
	longVal := int64(val) // 向零截断,仅用于哨兵值预检
	if vc.ErrValue.IsInvalid(longVal) {
		return longVal
	}
	result := (val + vc.Offset) * vc.Scale
	// audit 2026-09-17 (H4):原为 int64(result);截断会让 float64 解码值
	// 丢掉一个原始 ULP(例如燃油 raw 29 -> 0.29 -> 28)。
	return int64(math.Round(result))
}

// ---- 高风险(scale >= 1000)----
// 必须通过与 Java 输出对比的全范围 0~65535 校验。

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

// ---- 中风险(scale=100 或负 offset)----
// 必须通过 0~50000 关键范围校验。

var FuelConsumptionRateConverter = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 100, Offset: 0, Precision: 2,
}

// CurrentConverterChargeElectric 转换可充电储能装置电压数据的 current
// 字段(ChargeableSubsystemElectric.Current,TLV 0x08)。
//
// 修复(模拟器审计 2026-08-26):offset 原为 -1000,镜像自 Java
// ChargeableSubsystemElectricCodec.CURRENT_CONVERTER。该符号是错的:
// decode 变成 raw/10+1000,encode 变成 (v-1000)*10,在数学上
// 逐字节自洽,但语义与 GB/T 32960.3-2016 相反
// (raw = (value+1000)×10)。证据:金样报文 prod_realtime_v2016_01.hex
// 中 current raw=0x2900=10496 → +1000 语义解码为 49.6A(合理),
// -1000 语义解码为 2049.6A(不可能)。本次修复下金样报文重编码
// 保持逐字节一致(offset 只反转解释方向)。
//
// audit 2026-09-17:此前"Java 已同步为 +1000"的说法不成立。
// 两个本地 Java 检出(work/senyuan 下的 sirun-connect-gbt32960 及其
// worktree 副本)仍使用 -1000.0。Go 保留 +1000,因为 GB/T 32960-2016
// 表 B.6 定义 raw=(I+1000)×10;本地 Java 副本带的符号是错的,
// 会把该字段解码反。不要把它们镜像回本文件。
var CurrentConverterChargeElectric = ValueConverter{
	ErrValue: types.ErrByte2, Scale: 10, Offset: 1000, Precision: 1,
}

// ---- 低风险(scale <= 10,正 offset)----
// 每项至少 10+ 个边界测试用例。

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

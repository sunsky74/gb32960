package types

// OperatingState 表示车辆运行状态。
type OperatingState byte

const (
	OpStateOn        OperatingState = 0x01
	OpStateOff       OperatingState = 0x02
	OpStateOther     OperatingState = 0x03
	OpStateException OperatingState = 0xFE
	OpStateInvalid   OperatingState = 0xFF
)

// ChargingState 表示车辆充电状态(GB/T 32960-2016 表9 / 表B.4)。
type ChargingState byte

// audit 2026-09-17: M1,原先的取值把 0x03 当作 "complete" 且没有 0x04;
// 已按表 9 修正:0x01 停车充电,0x02 行驶充电,0x03 未充电,0x04 充电完成。
const (
	ChargeStateParking     ChargingState = 0x01 // 停车充电
	ChargeStateDriving     ChargingState = 0x02 // 行驶充电
	ChargeStateNotCharging ChargingState = 0x03 // 未充电
	ChargeStateCompleted   ChargingState = 0x04 // 充电完成
	ChargeStateException   ChargingState = 0xFE // 异常
	ChargeStateInvalid     ChargingState = 0xFF // 无效
)

// OperationMode 表示车辆运行模式。
type OperationMode byte

const (
	OpModeElectric  OperationMode = 0x01
	OpModeHybrid    OperationMode = 0x02
	OpModeFuel      OperationMode = 0x03
	OpModeException OperationMode = 0xFE
	OpModeInvalid   OperationMode = 0xFF
)

// DCState 表示 DC-DC 转换器状态。
type DCState byte

const (
	DCStateOn        DCState = 0x01
	DCStateOff       DCState = 0x02
	DCStateException DCState = 0xFE
	DCStateInvalid   DCState = 0xFF
)

// GearPositionEnum 表示挡位(GB/T 32960-2016 附录A.1)。
// 它占用挡位字节的低四位(位 3..0)。
//
// audit 2026-09-17: H2,原先的常量(GearP=0x01, GearR=0x02,
// GearN=0x03, GearD=0x04, GearOther=0x05)与附录A.1 矛盾;已替换为
// 下面的规范映射。
type GearPositionEnum byte

const (
	GearGap     GearPositionEnum = 0x0 // 空挡
	Gear1       GearPositionEnum = 0x1 // 1挡
	Gear2       GearPositionEnum = 0x2 // 2挡
	Gear3       GearPositionEnum = 0x3 // 3挡
	Gear4       GearPositionEnum = 0x4 // 4挡
	Gear5       GearPositionEnum = 0x5 // 5挡
	Gear6       GearPositionEnum = 0x6 // 6挡
	Gear7       GearPositionEnum = 0x7 // 7挡
	Gear8       GearPositionEnum = 0x8 // 8挡
	Gear9       GearPositionEnum = 0x9 // 9挡
	Gear10      GearPositionEnum = 0xA // 10挡
	Gear11      GearPositionEnum = 0xB // 11挡
	Gear12      GearPositionEnum = 0xC // 12挡
	GearReverse GearPositionEnum = 0xD // 倒挡
	GearAutoD   GearPositionEnum = 0xE // 自动D挡
	GearPark    GearPositionEnum = 0xF // 停车P挡
)

package types

// RealTimeType 表示 V2016 实时数据中的 TLV 类型标志。
type RealTimeType byte

const (
	RealTimeVehicle     RealTimeType = 0x01
	RealTimeMotor       RealTimeType = 0x02
	RealTimeFuelCell    RealTimeType = 0x03
	RealTimeEngine      RealTimeType = 0x04
	RealTimeLocation    RealTimeType = 0x05
	RealTimeExtremum    RealTimeType = 0x06
	RealTimeAlarm       RealTimeType = 0x07
	RealTimeVoltage     RealTimeType = 0x08
	RealTimeTemperature RealTimeType = 0x09
)

// RealTimeV2025Type 表示 V2025 实时数据中的 TLV 类型标志。
// 取值与参考 Java 实现(RealTimeTypeV2025)完全一致。
//
// 重要:早期草稿打乱了这些取值,Java 中 ALARM=0x06(不是 0x07)、
// BATTERY_MIN_PARALLEL_VOLTAGE_VALUE=0x07、BATTERY_TEMP=0x08。V2025 中
// 没有 Extremum / ChargeableVoltage / ChargeableTemperature,这些是
// 凭空杜撰的。见 audit 2026-07-31。
type RealTimeV2025Type byte

const (
	RealTimeV2025Vehicle            RealTimeV2025Type = 0x01 // 整车数据
	RealTimeV2025Motor              RealTimeV2025Type = 0x02 // 驱动电机数据
	RealTimeV2025FuelCellEngine     RealTimeV2025Type = 0x03 // 燃料电池发动机及车载氢系统数据
	RealTimeV2025Engine             RealTimeV2025Type = 0x04 // 发动机数据
	RealTimeV2025Location           RealTimeV2025Type = 0x05 // 车辆位置数据
	RealTimeV2025Alarm              RealTimeV2025Type = 0x06 // 报警数据
	RealTimeV2025MinParallelVoltage RealTimeV2025Type = 0x07 // 动力蓄电池最小并联单元电压数据
	RealTimeV2025BatteryTemp        RealTimeV2025Type = 0x08 // 动力蓄电池温度数据

	RealTimeV2025FuelCellStack    RealTimeV2025Type = 0x30 // 燃料电池电堆数据
	RealTimeV2025SuperCapacitor   RealTimeV2025Type = 0x31 // 超级电容器数据
	RealTimeV2025SuperCapExtremum RealTimeV2025Type = 0x32 // 超级电容器极值数据

	RealTimeV2025CustomStart RealTimeV2025Type = 0x80 // 自定义数据开始标志
	RealTimeV2025Custom      RealTimeV2025Type = 0x80 // 自定义数据标志 (0x80~0xFE 全部映射到这里)
	RealTimeV2025CustomEnd   RealTimeV2025Type = 0xFE // 自定义数据结束标志

	RealTimeV2025Signature RealTimeV2025Type = 0xFF // 签名数据开始标识
)

// RealTimeV2025TypeByCode 对应 Java RealTimeTypeV2025.valueOf():整个
// 0x80~0xFE 范围都映射到 CUSTOM_DATA_FLAG。需要精确自定义字节的调用方
// 应改为直接对原始字节做 switch。
func RealTimeV2025TypeByCode(code byte) (RealTimeV2025Type, bool) {
	u := int(code) & 0xFF
	if u >= 0x80 && u <= 0xFE {
		return RealTimeV2025Custom, true
	}
	switch RealTimeV2025Type(code) {
	case RealTimeV2025Vehicle, RealTimeV2025Motor, RealTimeV2025FuelCellEngine,
		RealTimeV2025Engine, RealTimeV2025Location, RealTimeV2025Alarm,
		RealTimeV2025MinParallelVoltage, RealTimeV2025BatteryTemp,
		RealTimeV2025FuelCellStack, RealTimeV2025SuperCapacitor,
		RealTimeV2025SuperCapExtremum, RealTimeV2025Signature:
		return RealTimeV2025Type(code), true
	}
	return 0, false
}

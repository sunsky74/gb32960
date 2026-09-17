package types

// 通用报警标志 32 位位掩码的报警位位置(GB/T 32960-2016 表18)。
// 位 0..18 已定义;位 19..31 为预留,必须在解码→编码过程中原样保留
// (见 codec/gbt2016/realtime 报警数据编解码器)。
//
// audit 2026-09-17: L6,上一版本止于位 8,且位 7/8 互相颠倒
// (SOC 跳变在位 7,充电过流在位 8);已按表 18 重写。
const (
	AlarmBitTemperatureDifferential         = 0  // 温度差异报警
	AlarmBitBatteryHighTemperature          = 1  // 电池高温报警
	AlarmBitDeviceTypeOverVoltage           = 2  // 车载储能装置类型过压报警
	AlarmBitDeviceTypeUnderVoltage          = 3  // 车载储能装置类型欠压报警
	AlarmBitSOCLow                          = 4  // SOC 过低报警
	AlarmBitMonomerBatteryOverVoltage       = 5  // 单体电池过压报警
	AlarmBitMonomerBatteryUnderVoltage      = 6  // 单体电池欠压报警
	AlarmBitSOCHigh                         = 7  // SOC 过高报警
	AlarmBitSOCJump                         = 8  // SOC 跳变报警
	AlarmBitDeviceTypeDontMatch             = 9  // 车载储能装置类型不匹配报警
	AlarmBitBatteryConsistencyPoor          = 10 // 单体电池一致性差报警
	AlarmBitInsulation                      = 11 // 绝缘报警
	AlarmBitDCTemperature                   = 12 // DC 温度报警
	AlarmBitBrakingSystem                   = 13 // 制动系统报警
	AlarmBitDCStatus                        = 14 // DC 状态报警
	AlarmBitDriveMotorControllerTemperature = 15 // 驱动电机控制器温度报警
	AlarmBitHighPressureInterlock           = 16 // 高压互锁报警
	AlarmBitDriveMotorTemperature           = 17 // 驱动电机温度报警
	AlarmBitDeviceTypeOverFilling           = 18 // 车载储能装置过充报警
)

package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// AlarmData 是 V2016 报警数据子记录(TLV 类型 0x07)。
// 字段名与 Java AlarmData 保持一致。
// 19 个布尔字段与 AlarmBitIdentify 的位 0..18 一一对应;
// 它们作为具名字段存在以便于使用;编解码器既负责从线上位掩码填充它们,
// 又负责在编码时重新打包。
type AlarmData struct {
	MaxAlarmLevel                   int     // 最高报警等级
	AlarmBitIdentify                int64   // 通用报警标志 (原始 32 位线上掩码,符号扩展至 int64)
	TemperatureDifferential         bool    // 温度差异报警 (位 0)
	BatteryHighTemperature          bool    // 电池高温报警 (位 1)
	DeviceTypeOverVoltage           bool    // 车载储能装置类型过压报警 (位 2)
	DeviceTypeUnderVoltage          bool    // 车载储能装置类型欠压报警 (位 3)
	SocLow                          bool    // SOC 过低报警 (位 4)
	MonomerBatteryOverVoltage       bool    // 单体电池过压报警 (位 5)
	MonomerBatteryUnderVoltage      bool    // 单体电池欠压报警 (位 6)
	SocHigh                         bool    // SOC 过高报警 (位 7)
	SocJump                         bool    // SOC 跳变报警 (位 8)
	DeviceTypeDontMatch             bool    // 车载储能装置类型不匹配报警 (位 9)
	BatteryConsistencyPoor          bool    // 单体电池一致性差报警 (位 10)
	Insulation                      bool    // 绝缘报警 (位 11)
	DcTemperature                   bool    // DC 温度报警 (位 12)
	BrakingSystem                   bool    // 制动系统报警 (位 13)
	DcStatus                        bool    // DC 状态报警 (位 14)
	DriveMotorControllerTemperature bool    // 驱动电机控制器温度报警 (位 15)
	HighPressureInterlock           bool    // 高压互锁报警 (位 16)
	DriveMotorTemperature           bool    // 驱动电机温度报警 (位 17)
	DeviceTypeOverFilling           bool    // 车载储能装置过充报警 (位 18)
	BatteryFaultNum                 int     // 可充电储能装置故障总数 N1
	BatteryFaultDatas               []int64 // 可充电储能装置故障代码列表 (长度为 N1)
	MotorFaultNum                   int     // 驱动电机故障总数 N2
	MotorFaultDatas                 []int64 // 驱动电机故障代码列表 (长度为 N2)
	EngineFaultNum                  int     // 发动机故障总数 N3
	EngineFaultDatas                []int64 // 发动机故障代码列表 (长度为 N3)
	OtherFaultNum                   int     // 其他故障总数 N4
	OtherFaultDatas                 []int64 // 其他故障代码列表 (长度为 N4)
}

func (m *AlarmData) Version() api.GBTVersion { return api.V2016 }

func (m *AlarmData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*AlarmData)(nil)).Elem())
}

package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// AlarmData is the V2016 报警数据 sub-record (TLV type 0x07).
// Field names mirror Java AlarmData.
// The 19 Boolean fields correspond 1:1 to bits 0..18 of AlarmBitIdentify —
// they exist as named fields for ergonomic access; the codec is responsible
// for both populating them from the wire bit-mask AND repacking them on encode.
type AlarmData struct {
	MaxAlarmLevel                   int     // 最高报警等级
	AlarmBitIdentify                int64   // 通用报警标志 (raw 32-bit wire mask, sign-extended to int64)
	TemperatureDifferential         bool    // 温度差异报警 (bit 0)
	BatteryHighTemperature          bool    // 电池高温报警 (bit 1)
	DeviceTypeOverVoltage           bool    // 车载储能装置类型过压报警 (bit 2)
	DeviceTypeUnderVoltage          bool    // 车载储能装置类型欠压报警 (bit 3)
	SocLow                          bool    // SOC 过低报警 (bit 4)
	MonomerBatteryOverVoltage       bool    // 单体电池过压报警 (bit 5)
	MonomerBatteryUnderVoltage      bool    // 单体电池欠压报警 (bit 6)
	SocHigh                         bool    // SOC 过高报警 (bit 7)
	SocJump                         bool    // SOC 跳变报警 (bit 8)
	DeviceTypeDontMatch             bool    // 车载储能装置类型不匹配报警 (bit 9)
	BatteryConsistencyPoor          bool    // 单体电池一致性差报警 (bit 10)
	Insulation                      bool    // 绝缘报警 (bit 11)
	DcTemperature                   bool    // DC 温度报警 (bit 12)
	BrakingSystem                   bool    // 制动系统报警 (bit 13)
	DcStatus                        bool    // DC 状态报警 (bit 14)
	DriveMotorControllerTemperature bool    // 驱动电机控制器温度报警 (bit 15)
	HighPressureInterlock           bool    // 高压互锁报警 (bit 16)
	DriveMotorTemperature           bool    // 驱动电机温度报警 (bit 17)
	DeviceTypeOverFilling           bool    // 车载储能装置过充报警 (bit 18)
	BatteryFaultNum                 int     // 可充电储能装置故障总数 N1
	BatteryFaultDatas               []int64 // 可充电储能装置故障代码列表 (length N1)
	MotorFaultNum                   int     // 驱动电机故障总数 N2
	MotorFaultDatas                 []int64 // 驱动电机故障代码列表 (length N2)
	EngineFaultNum                  int     // 发动机故障总数 N3
	EngineFaultDatas                []int64 // 发动机故障代码列表 (length N3)
	OtherFaultNum                   int     // 其他故障总数 N4
	OtherFaultDatas                 []int64 // 其他故障代码列表 (length N4)
}

func (m *AlarmData) Version() api.GBTVersion { return api.V2016 }

func (m *AlarmData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*AlarmData)(nil)).Elem())
}

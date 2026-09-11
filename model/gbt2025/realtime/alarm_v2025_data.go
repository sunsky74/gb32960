package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// AlarmV2025Data is the V2025 alarm data (TLV 0x06).
// V2025 alarm has 28 individual alarm bit fields plus fault lists and
// common alert data. (V2016 TLV 0x06 was Extremum; V2025 0x06 is Alarm.)
type AlarmV2025Data struct {
	MaxAlarmLevel                   int
	AlarmBitIdentify                int64 // 通用报警标志 (原始 32 位掩码; 编码时非 0 原样回写以保留保留位 28..31, 为 0 时从 28 个布尔重建)
	TemperatureDifferential         bool
	BatteryHighTemperature          bool
	DeviceTypeOverVoltage           bool
	DeviceTypeUnderVoltage          bool
	SOCLow                          bool
	MonomerBatteryOverVoltage       bool
	MonomerBatteryUnderVoltage      bool
	SOCHigh                         bool
	SOCJump                         bool
	DeviceTypeDontMatch             bool
	BatteryConsistencyPoor          bool
	Insulation                      bool
	DCTemperature                   bool
	BrakingSystem                   bool
	DCStatus                        bool
	DriveMotorControllerTemperature bool
	HighPressureInterlock           bool
	DriveMotorTemperature           bool
	DeviceTypeOverFilling           bool
	DriveMotorOverSpeed             bool
	DriveMotorOverCurrent           bool
	SuperCapacitorOverTemp          bool
	SuperCapacitorOverVoltage       bool
	DeviceThermalEvent              bool
	HydrogenLeakage                 bool
	HydrogenPressureAbnormal        bool
	HydrogenTemperatureAbnormal     bool
	FuelCellStackOverTemperature    bool

	BatteryFaultNum   int
	BatteryFaultDatas []int64
	MotorFaultNum     int
	MotorFaultDatas   []int64
	EngineFaultNum    int
	EngineFaultDatas  []int64
	OtherFaultNum     int
	OtherFaultDatas   []int64

	CommonAlertNum   int
	CommonAlertDatas []CommonAlertData
}

// CommonAlertData is a single common alert entry: a flag bit sequence number
// and a fault level.
type CommonAlertData struct {
	Seq   int
	Level int
}

func (m *AlarmV2025Data) Version() api.GBTVersion { return api.V2025 }
func (m *AlarmV2025Data) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*AlarmV2025Data)(nil)).Elem())
}

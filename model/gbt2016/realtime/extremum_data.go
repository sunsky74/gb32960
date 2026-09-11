package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// ExtremumData is the V2016 极值数据 sub-record (TLV type 0x06).
// Field names mirror Java ExtremumData.
type ExtremumData struct {
	VoltageMaxSubsystem     int     // 最高电压电池子系统号
	VoltageMaxBattery       int     // 最高电压电池单体代号
	MaxVoltage              float64 // 电池单体电压最高值, V (scale=1000)
	VoltageMinSubsystem     int     // 最低电压电池子系统号
	VoltageMinBattery       int     // 最低电压电池单体代号
	MinVoltage              float64 // 电池单体电压最低值, V (scale=1000)
	TemperatureMaxSubsystem int     // 最高温度子系统号
	TemperatureMaxProbe     int     // 最高温度探针序号
	MaxTemperature          float64 // 最高温度值, °C (scale=1, offset=40)
	TemperatureMinSubsystem int     // 最低温度子系统号
	TemperatureMinProbe     int     // 最低温度探针序号
	MinTemperature          float64 // 最低温度值, °C (scale=1, offset=40)
}

func (m *ExtremumData) Version() api.GBTVersion { return api.V2016 }

func (m *ExtremumData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*ExtremumData)(nil)).Elem())
}

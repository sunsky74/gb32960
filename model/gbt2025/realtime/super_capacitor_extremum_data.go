package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// SuperCapacitorExtremumData 是 V2025 超级电容极值数据(TLV 0x32)。
// V2025 独有类型。
type SuperCapacitorExtremumData struct {
	VoltageMaxSubsystem     int     // 最高电压管理系统号
	VoltageMaxBattery       int     // 最高电压超级电容单体代号
	MaxVoltage              float64 // 超级电容单体电压最高值, scale=0.001
	VoltageMinSubsystem     int     // 最低电压管理系统号
	VoltageMinBattery       int     // 最低电压超级电容单体代号
	MinVoltage              float64 // 超级电容单体电压最低值, scale=0.001
	TemperatureMaxSubsystem int     // 最高温度管理系统号
	TemperatureMaxProbe     int     // 最高温度探针代号
	MaxTemperature          float64 // 最高温度值, offset=40
	TemperatureMinSubsystem int     // 最低温度管理系统号
	TemperatureMinProbe     int     // 最低温度探针代号
	MinTemperature          float64 // 最低温度值, offset=40
}

func (m *SuperCapacitorExtremumData) Version() api.GBTVersion { return api.V2025 }
func (m *SuperCapacitorExtremumData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*SuperCapacitorExtremumData)(nil)).Elem())
}

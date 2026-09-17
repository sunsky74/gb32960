package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// ChargeableSubsystemTemperature 是 V2016 可充电储能装置温度的一条记录
// (TLV 类型 0x09 元素)。字段名与 Java ChargeableSubsystemTemperature 保持一致。
type ChargeableSubsystemTemperature struct {
	SubSystemNumber       int       // 可充电储能子系统号, 1~250
	TemperatureProbeCount int       // 可充电储能温度探针个数 N
	ProbeTemperatures     []float64 // 各探针温度值, °C (scale=1, offset=40), 长度为 N
}

func (m *ChargeableSubsystemTemperature) Version() api.GBTVersion { return api.V2016 }

func (m *ChargeableSubsystemTemperature) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*ChargeableSubsystemTemperature)(nil)).Elem())
}

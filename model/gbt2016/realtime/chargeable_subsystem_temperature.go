package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// ChargeableSubsystemTemperature is one V2016 可充电储能装置温度 entry
// (TLV type 0x09 element). Field names mirror Java ChargeableSubsystemTemperature.
type ChargeableSubsystemTemperature struct {
	SubSystemNumber       int       // 可充电储能子系统号, 1~250
	TemperatureProbeCount int       // 可充电储能温度探针个数 N
	ProbeTemperatures     []float64 // 各探针温度值, °C (scale=1, offset=40) — length N
}

func (m *ChargeableSubsystemTemperature) Version() api.GBTVersion { return api.V2016 }

func (m *ChargeableSubsystemTemperature) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*ChargeableSubsystemTemperature)(nil)).Elem())
}

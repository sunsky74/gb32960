package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// SuperCapacitorData is the V2025 super capacitor data (TLV 0x31).
// V2025-only type.
type SuperCapacitorData struct {
	ManagementSystemNumber int       // 超级电容管理系统号
	TotalVoltage           float64   // V, scale=0.1
	TotalCurrent           float64   // A, offset=3000, scale=0.1
	CapacitorCount         int       // 超级电容单体总数
	CapacitorVoltages      []float64 // 单体电压, scale=0.001
	TemperatureProbeCount  int       // 温度探针总数
	ProbeTemperatures      []float64 // 探针温度, offset=40
}

func (m *SuperCapacitorData) Version() api.GBTVersion { return api.V2025 }
func (m *SuperCapacitorData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*SuperCapacitorData)(nil)).Elem())
}

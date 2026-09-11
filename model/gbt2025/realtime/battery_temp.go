package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// BatteryTemp is the V2025 battery pack temperature info (TLV 0x08).
// Renamed from Java BatteryPackTemperature to match the TLV enum constant name.
type BatteryTemp struct {
	BatteryPackSeq        int       // 动力蓄电池包号
	TemperatureProbeCount int       // 温度探针个数
	ProbeTemperatures     []float64 // 温度值, offset=40
}

func (m *BatteryTemp) Version() api.GBTVersion { return api.V2025 }
func (m *BatteryTemp) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*BatteryTemp)(nil)).Elem())
}

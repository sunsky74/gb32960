package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// MinParallelCellVoltage is the V2025 minimum parallel cell voltage info (TLV 0x07).
// V2025-only type; no V2016 equivalent.
type MinParallelCellVoltage struct {
	BatteryPackSeq   int       // 动力蓄电池包号
	Voltage          float64   // V, scale=0.1
	Current          float64   // A, offset=3000, scale=0.1
	MinParallelUnits int       // 最小并联单元总数
	BatteryVoltages  []float64 // 最小并联单元电压, scale=0.001
}

func (m *MinParallelCellVoltage) Version() api.GBTVersion { return api.V2025 }
func (m *MinParallelCellVoltage) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*MinParallelCellVoltage)(nil)).Elem())
}

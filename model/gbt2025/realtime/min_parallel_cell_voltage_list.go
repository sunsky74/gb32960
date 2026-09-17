package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// MinParallelCellVoltageList 是 V2025 最小并联单元电压列表(TLV 0x07)。
type MinParallelCellVoltageList struct {
	BatteryPackCount int
	Items            []MinParallelCellVoltage
}

func (m *MinParallelCellVoltageList) Version() api.GBTVersion { return api.V2025 }
func (m *MinParallelCellVoltageList) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*MinParallelCellVoltageList)(nil)).Elem())
}

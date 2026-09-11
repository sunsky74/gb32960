package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// FuelCellStackDataList is the V2025 fuel cell stack data list (TLV 0x30).
type FuelCellStackDataList struct {
	StackCount int
	Items      []FuelCellStackData
}

func (m *FuelCellStackDataList) Version() api.GBTVersion { return api.V2025 }
func (m *FuelCellStackDataList) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*FuelCellStackDataList)(nil)).Elem())
}

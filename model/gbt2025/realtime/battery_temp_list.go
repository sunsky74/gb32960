package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// BatteryTempList is the V2025 battery pack temperature data list (TLV 0x08).
// Renamed from Java BatteryPackTemperatureList.
type BatteryTempList struct {
	BatteryPackCount int
	Items            []BatteryTemp
}

func (m *BatteryTempList) Version() api.GBTVersion { return api.V2025 }
func (m *BatteryTempList) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*BatteryTempList)(nil)).Elem())
}

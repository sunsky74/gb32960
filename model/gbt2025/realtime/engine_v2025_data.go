package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// EngineV2025Data is the V2025 engine data (TLV 0x04).
// Compared to V2016 EngineData, V2025 has ONLY crankshaftSpeed (removes state
// and fuelConsumptionRate).
type EngineV2025Data struct {
	CrankshaftSpeed int // rpm, 0~60000
}

func (m *EngineV2025Data) Version() api.GBTVersion { return api.V2025 }
func (m *EngineV2025Data) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*EngineV2025Data)(nil)).Elem())
}

package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// EngineV2025Data 是 V2025 发动机数据(TLV 0x04)。
// 与 V2016 的 EngineData 相比,V2025 仅有 crankshaftSpeed
// (移除了 state 和 fuelConsumptionRate)。
type EngineV2025Data struct {
	CrankshaftSpeed int // rpm, 0~60000
}

func (m *EngineV2025Data) Version() api.GBTVersion { return api.V2025 }
func (m *EngineV2025Data) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*EngineV2025Data)(nil)).Elem())
}

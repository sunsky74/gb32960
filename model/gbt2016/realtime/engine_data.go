package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// EngineData is the V2016 发动机数据 sub-record (TLV type 0x04).
// Field names mirror Java EngineData.
// EngineState is kept as byte because types/ has no enum for it; the codec
// layer translates the raw byte (audit note: types package is fixed scope).
type EngineData struct {
	EngineState         byte    // 发动机状态 (Java enum EngineState; raw byte here)
	CrankshaftSpeed     int     // 曲轴转速, rpm
	FuelConsumptionRate float64 // 燃料消耗率, L/100km (scale=100)
}

func (m *EngineData) Version() api.GBTVersion { return api.V2016 }

func (m *EngineData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*EngineData)(nil)).Elem())
}

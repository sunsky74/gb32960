package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// FuelCellStackData is a single V2025 fuel cell stack record (TLV 0x30).
// V2025-only type.
type FuelCellStackData struct {
	StackSeq               int       // 燃料电池电堆序号
	Voltage                float64   // V, scale=0.1
	Current                float64   // A, scale=0.1
	GasPressure            float64   // kPa, offset=100, scale=0.1
	AirPressure            float64   // kPa, offset=100, scale=0.1
	AirInletTemp           float64   // °C, offset=40
	CoolingWaterProbeCount int       // 冷却水温度探针总数
	CoolingWaterTemps      []float64 // 冷却水出水口温度, offset=40
}

func (m *FuelCellStackData) Version() api.GBTVersion { return api.V2025 }
func (m *FuelCellStackData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*FuelCellStackData)(nil)).Elem())
}

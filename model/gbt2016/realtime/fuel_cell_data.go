package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// FuelCellData is the V2016 燃料电池数据 sub-record (TLV type 0x03).
// Field names mirror Java FuelCellData.
// HighVoltageDCState is kept as byte because types/ has no enum for it;
// the codec layer translates the raw byte (audit note: types package is
// fixed scope for this task).
type FuelCellData struct {
	FuelCellVoltage                      float64   // 燃料电池电压, V (scale=10)
	FuelCellCurrent                      float64   // 燃料电池电流, A (scale=10)
	FuelConsumptionRate                  float64   // 燃料消耗率, kg/100km (scale=100)
	TotalNumberOfFcTp                    int       // 燃料电池温度探针总数 (probe count N)
	ProbeTemperatureValues               []float64 // 探针温度值, °C (scale=1, offset=40) — length N
	HighestTempOfHydrogenSystem          float64   // 氢系统中最高温度, °C (scale=10, offset=40)
	HighestTempProbeCodeOfHydrogenSystem int       // 氢系统中最高温度探针代号
	HighestConOfHydrogen                 int       // 氢气最高浓度, ppm
	HighestHyConSensorCode               int       // 氢气最高浓度传感器代号
	HydrogenMaxPressure                  float64   // 氢气最高压力, MPa (scale=10)
	HydrogenMaxPressureSensorCode        int       // 氢气最高压力传感器代号
	HighVoltageDCState                   byte      // 高压 DC/DC 状态 (Java enum HighVoltageDCState; raw byte here)
}

func (m *FuelCellData) Version() api.GBTVersion { return api.V2016 }

func (m *FuelCellData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*FuelCellData)(nil)).Elem())
}

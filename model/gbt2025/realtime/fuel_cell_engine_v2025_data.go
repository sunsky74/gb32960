package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// FuelCellEngineV2025Data 是 V2025 燃料电池发动机与车载氢系统
// 数据(TLV 0x03)。与 V2016 的 FuelCellData 相比,新增了 fuelPercentage
// 和 dcControllerTemperature,移除了 voltage/current/consumptionRate。
// HighVoltageDCState 映射到 Java 的 HighVoltageDCState:0x01=ON, 0x02=OFF,
// 0xFE=EXCEPTION, 0xFF=INVALID。
type FuelCellEngineV2025Data struct {
	HighestTempOfHydrogenSystem          float64 // °C, offset=40
	HighestTempProbeCodeOfHydrogenSystem int
	HighestConOfHydrogen                 float64 // %
	HighestHyConSensorCode               int
	HydrogenMaxPressure                  float64 // MPa
	HydrogenMaxPressureSensorCode        int
	HighVoltageDCState                   byte    // HighVoltageDCState 枚举
	FuelPercentage                       int     // %
	DCControllerTemperature              float64 // °C, offset=40
}

func (m *FuelCellEngineV2025Data) Version() api.GBTVersion { return api.V2025 }
func (m *FuelCellEngineV2025Data) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*FuelCellEngineV2025Data)(nil)).Elem())
}

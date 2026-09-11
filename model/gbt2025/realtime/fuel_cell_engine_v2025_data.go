package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// FuelCellEngineV2025Data is the V2025 fuel cell engine and onboard hydrogen
// system data (TLV 0x03). Compared to V2016 FuelCellData, ADDS fuelPercentage
// and dcControllerTemperature, and REMOVES voltage/current/consumptionRate.
// HighVoltageDCState maps to Java HighVoltageDCState: 0x01=ON, 0x02=OFF,
// 0xFE=EXCEPTION, 0xFF=INVALID.
type FuelCellEngineV2025Data struct {
	HighestTempOfHydrogenSystem          float64 // °C, offset=40
	HighestTempProbeCodeOfHydrogenSystem int
	HighestConOfHydrogen                 float64 // %
	HighestHyConSensorCode               int
	HydrogenMaxPressure                  float64 // MPa
	HydrogenMaxPressureSensorCode        int
	HighVoltageDCState                   byte    // HighVoltageDCState enum
	FuelPercentage                       int     // %
	DCControllerTemperature              float64 // °C, offset=40
}

func (m *FuelCellEngineV2025Data) Version() api.GBTVersion { return api.V2025 }
func (m *FuelCellEngineV2025Data) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*FuelCellEngineV2025Data)(nil)).Elem())
}

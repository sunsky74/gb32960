package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
	"github.com/sunsky74/gb32960/model/gbt2016/realtime"
	v2025rt "github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// RealTimeV2025Data is the V2025 realtime data report (command 0x02).
// Contains a BeanTime and optional TLV-encoded sub-records.
// V2025 reuses the V2016 VehicleData struct (no separate VehicleDataV2025).
// There is NO Extremum in V2025 (TLV 0x06 is Alarm; audit 2026-07-31).
type RealTimeV2025Data struct {
	_                          GBT2025Body
	BeanTime                   model.BeanTime
	VehicleData                *realtime.VehicleData
	MotorDataList              *v2025rt.MotorDataV2025List
	FuelCellData               *v2025rt.FuelCellEngineV2025Data
	EngineData                 *v2025rt.EngineV2025Data
	LocationData               *v2025rt.LocationV2025Data
	AlarmData                  *v2025rt.AlarmV2025Data
	MinParallelCellVoltages    *v2025rt.MinParallelCellVoltageList
	BatteryPackTemperatures    *v2025rt.BatteryTempList
	FuelCellStackDataList      *v2025rt.FuelCellStackDataList
	SuperCapacitorData         *v2025rt.SuperCapacitorData
	SuperCapacitorExtremumData *v2025rt.SuperCapacitorExtremumData
	VehicleSignature           *v2025rt.VehicleSignature
	CustomData                 []v2025rt.CustomV2025Data
	Items                      []v2025rt.RealTimeV2025Item
}

func (m *RealTimeV2025Data) Version() api.GBTVersion { return api.V2025 }
func (m *RealTimeV2025Data) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*RealTimeV2025Data)(nil)).Elem())
}

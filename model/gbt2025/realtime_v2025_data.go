package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/model/gbt2016/realtime"
	v2025rt "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/modelutil"
)

// RealTimeV2025Data 是 V2025 实时数据上报(命令 0x02)。
// 包含一个 BeanTime 和可选的 TLV 编码子记录。
// V2025 复用 V2016 VehicleData 结构体(没有单独的 VehicleDataV2025)。
// V2025 中没有 Extremum(TLV 0x06 是 Alarm;audit 2026-07-31)。
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

package gbt2016

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/modelutil"
)

// RealTimeData is the V2016 realtime report (command 0x02) and the reissue
// report (command 0x03 — shares the REAL_TIME body). The on-wire body is a
// BeanTime followed by an optional sequence of TLV sub-records. Each sub-record
// is present at most once per frame; any of the pointers below may be nil.
type RealTimeData struct {
	_                                  GBT2016Body
	BeanTime                           model.BeanTime
	VehicleData                        *realtime.VehicleData                        // TLV 0x01
	MotorDataList                      *realtime.MotorDataList                      // TLV 0x02
	FuelCellData                       *realtime.FuelCellData                       // TLV 0x03
	EngineData                         *realtime.EngineData                         // TLV 0x04
	LocationData                       *realtime.LocationData                       // TLV 0x05
	ExtremumData                       *realtime.ExtremumData                       // TLV 0x06
	AlarmData                          *realtime.AlarmData                          // TLV 0x07
	ChargeableSubsystemElectricList    *realtime.ChargeableSubsystemElectricList    // TLV 0x08
	ChargeableSubsystemTemperatureList *realtime.ChargeableSubsystemTemperatureList // TLV 0x09
	CustomData                         map[byte][]byte                              // 用户自定义数据(0x80~0x0FE): key=信息类型标志, value=数据体
}

func (m *RealTimeData) Version() api.GBTVersion { return api.V2016 }

func (m *RealTimeData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*RealTimeData)(nil)).Elem())
}

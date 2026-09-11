package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// ChargeableSubsystemElectric is one V2016 可充电储能装置电压 entry
// (TLV type 0x08 element). Field names mirror Java ChargeableSubsystemElectric.
type ChargeableSubsystemElectric struct {
	ChargeableSubSystemNumber int       // 子系统号, 1~250
	Voltage                   float64   // 电压, V (scale=10)
	Current                   float64   // 电流, A (scale=10, offset=1000)
	BatteryTotalCount         int       // 单体电池总数
	FrameStartBatterySeq      int       // 本帧起始电池序号 (multi-frame split)
	BatteryCount              int       // 本帧单体电池总数 m (1~200)
	BatteryVoltages           []float64 // 单体电池电压, V (scale=1000) — length m
}

func (m *ChargeableSubsystemElectric) Version() api.GBTVersion { return api.V2016 }

func (m *ChargeableSubsystemElectric) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*ChargeableSubsystemElectric)(nil)).Elem())
}

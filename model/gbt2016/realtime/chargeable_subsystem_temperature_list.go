package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// ChargeableSubsystemTemperatureList wraps the V2016 可充电储能装置温度 list
// (TLV type 0x09 body). Mirrors Java ChargeableSubsystemTemperatureList which
// extends MessageList<ChargeableSubsystemTemperature>.
type ChargeableSubsystemTemperatureList struct {
	TemperatureCount int // 温度数据子系统个数 (== len(Items) after decode)
	Items            []ChargeableSubsystemTemperature
}

func (m *ChargeableSubsystemTemperatureList) Version() api.GBTVersion { return api.V2016 }

func (m *ChargeableSubsystemTemperatureList) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*ChargeableSubsystemTemperatureList)(nil)).Elem())
}

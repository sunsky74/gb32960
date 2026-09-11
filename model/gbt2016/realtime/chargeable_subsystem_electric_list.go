package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// ChargeableSubsystemElectricList wraps the V2016 可充电储能装置电压 list
// (TLV type 0x08 body). Mirrors Java ChargeableSubsystemElectricList which
// extends MessageList<ChargeableSubsystemElectric>.
type ChargeableSubsystemElectricList struct {
	ElectricCount int // 电压数据子系统个数 (== len(Items) after decode)
	Items         []ChargeableSubsystemElectric
}

func (m *ChargeableSubsystemElectricList) Version() api.GBTVersion { return api.V2016 }

func (m *ChargeableSubsystemElectricList) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*ChargeableSubsystemElectricList)(nil)).Elem())
}

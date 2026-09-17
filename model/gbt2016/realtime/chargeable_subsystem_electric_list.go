package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// ChargeableSubsystemElectricList 包装 V2016 可充电储能装置电压列表
// (TLV 类型 0x08 消息体)。与 Java ChargeableSubsystemElectricList 一致,
// 后者继承 MessageList<ChargeableSubsystemElectric>。
type ChargeableSubsystemElectricList struct {
	ElectricCount int // 电压数据子系统个数 (解码后 == len(Items))
	Items         []ChargeableSubsystemElectric
}

func (m *ChargeableSubsystemElectricList) Version() api.GBTVersion { return api.V2016 }

func (m *ChargeableSubsystemElectricList) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*ChargeableSubsystemElectricList)(nil)).Elem())
}

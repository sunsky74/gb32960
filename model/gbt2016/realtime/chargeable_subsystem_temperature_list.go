package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// ChargeableSubsystemTemperatureList 包装 V2016 可充电储能装置温度列表
// (TLV 类型 0x09 消息体)。与 Java ChargeableSubsystemTemperatureList 一致,
// 后者继承 MessageList<ChargeableSubsystemTemperature>。
type ChargeableSubsystemTemperatureList struct {
	TemperatureCount int // 温度数据子系统个数 (解码后 == len(Items))
	Items            []ChargeableSubsystemTemperature
}

func (m *ChargeableSubsystemTemperatureList) Version() api.GBTVersion { return api.V2016 }

func (m *ChargeableSubsystemTemperatureList) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*ChargeableSubsystemTemperatureList)(nil)).Elem())
}

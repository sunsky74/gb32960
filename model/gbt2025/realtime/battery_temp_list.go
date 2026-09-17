package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// BatteryTempList 是 V2025 动力蓄电池包温度数据列表(TLV 0x08)。
// 由 Java 的 BatteryPackTemperatureList 重命名而来。
type BatteryTempList struct {
	BatteryPackCount int
	Items            []BatteryTemp
}

func (m *BatteryTempList) Version() api.GBTVersion { return api.V2025 }
func (m *BatteryTempList) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*BatteryTempList)(nil)).Elem())
}

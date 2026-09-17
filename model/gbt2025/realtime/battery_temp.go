package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// BatteryTemp 是 V2025 动力蓄电池包温度信息(TLV 0x08)。
// 由 Java 的 BatteryPackTemperature 重命名而来,以匹配 TLV 枚举常量名。
type BatteryTemp struct {
	BatteryPackSeq        int       // 动力蓄电池包号
	TemperatureProbeCount int       // 温度探针个数
	ProbeTemperatures     []float64 // 温度值, offset=40
}

func (m *BatteryTemp) Version() api.GBTVersion { return api.V2025 }
func (m *BatteryTemp) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*BatteryTemp)(nil)).Elem())
}

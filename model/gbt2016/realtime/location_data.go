package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// LocationData 是 V2016 车辆位置数据子记录(TLV 类型 0x05)。
// 字段名与 Java LocationData 保持一致。
type LocationData struct {
	Valid     bool    // 定位状态: true=有效定位, false=无效定位 (状态字节第 0 位: 0=有效, 1=无效)
	Longitude float64 // 经度, 有符号度数 (线上为 abs×10^6 的 u32, 西经由状态字节第 2 位表示)
	Latitude  float64 // 纬度, 有符号度数 (线上为 abs×10^6 的 u32, 南纬由状态字节第 1 位表示)
}

func (m *LocationData) Version() api.GBTVersion { return api.V2016 }

func (m *LocationData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*LocationData)(nil)).Elem())
}

package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// LocationData is the V2016 车辆位置数据 sub-record (TLV type 0x05).
// Field names mirror Java LocationData.
type LocationData struct {
	Valid     bool    // 定位状态: true=有效定位, false=无效定位 (状态字节 bit0: 0=有效, 1=无效)
	Longitude float64 // 经度, 有符号度数 (线上为 abs×10^6 的 u32, 西经由状态字节 bit2 表示)
	Latitude  float64 // 纬度, 有符号度数 (线上为 abs×10^6 的 u32, 南纬由状态字节 bit1 表示)
}

func (m *LocationData) Version() api.GBTVersion { return api.V2016 }

func (m *LocationData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*LocationData)(nil)).Elem())
}

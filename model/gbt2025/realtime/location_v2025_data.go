package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// LocationV2025Data 是 V2025 车辆定位数据(TLV 0x05)。
// 与 V2016 的 LocationData 相比,新增了 northernFlag、eastFlag、coordinateType
// 和 convertLongitude/convertLatitude 字段。
// CoordinateType 映射到 Java 的 CoordinateType:0x01=WGS84, 0x02=GCJ02, 0x03=OTHER。
type LocationV2025Data struct {
	Valid          bool
	NorthernFlag   bool // true=北纬,false=南纬
	EastFlag       bool // true=东经,false=西经
	CoordinateType byte // CoordinateType 枚举
	// OriginLongitude 原始经度:单位「度」(= 线上原始值 / 10^6,西经为负;
	// fix 2026-09-17:2025.md L320 —— 线上为「度×10^6」的 DWORD,本字段存解码后的度值,
	// 原注释「× 10^6」有误。BYTE4 哨兵值(0xFFFFFFFE/0xFFFFFFFF)原样直通,
	// 不带半球符号。
	OriginLongitude float64
	// OriginLatitude 原始纬度:单位「度」(= 线上原始值 / 10^6,南纬为负;
	// fix 2026-09-17:2025.md L321 —— 线上为「度×10^6」的 DWORD,本字段存解码后的度值,
	// 原注释「× 10^6」有误。BYTE4 哨兵值(0xFFFFFFFE/0xFFFFFFFF)原样直通,
	// 不带半球符号。
	OriginLatitude   float64
	ConvertLongitude float64
	ConvertLatitude  float64
}

func (m *LocationV2025Data) Version() api.GBTVersion { return api.V2025 }
func (m *LocationV2025Data) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*LocationV2025Data)(nil)).Elem())
}

package realtime

import (
	"math"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
	"github.com/sunsky74/gb32960/utils"
)

// LocationV2025Codec 编解码 V2025 车辆位置数据子记录
// (TLV 类型 0x05)。线格式与 Java LocationV2025Codec 一致:
//
//	StatusByte(u8) + CoordinateType(u8)
//	+ OriginLongitude(u32 原始值, ×10^6 并带半球符号)
//	+ OriginLatitude(u32 原始值, ×10^6 并带半球符号)
//
// StatusByte 位提取(Java LocationV2025Codec 第 27-29、42-52 行):
//
//	bit 0 (0x01): 0=有效,        1=无效
//	bit 1 (0x02): 0=北纬,        1=南纬
//	bit 2 (0x04): 0=东经,        1=西经
//
// CoordinateType(Java CoordinateType):0x01=WGS84(应用 WGS84→GCJ02),
// 0x02=GCJ02(直通),0x03=OTHER(直通)。
//
// 当 OriginLongitude/OriginLatitude 为 BYTE4 错误哨兵值时
// (0xFFFFFFFE/0xFFFFFFFF),Java 的 decodeCoordinate 返回 null,
// GCJ02 转换步骤被跳过。我们与此保持一致:此时把
// ConvertLongitude/ConvertLatitude 留为零值。
type LocationV2025Codec struct{}

func init() {
	api.Register[mdl.LocationV2025Data](api.V2025, &LocationV2025Codec{})
}

func (c *LocationV2025Codec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.LocationV2025Data{}

	status := r.ReadUint8()
	m.Valid = (status & 0x01) == 0
	m.NorthernFlag = (status & 0x02) == 0
	m.EastFlag = (status & 0x04) == 0

	m.CoordinateType = r.ReadUint8()

	rawLon := int64(r.ReadUint32())
	rawLat := int64(r.ReadUint32())

	m.OriginLongitude = decodeV2025Coordinate(rawLon, m.EastFlag)
	m.OriginLatitude = decodeV2025Coordinate(rawLat, m.NorthernFlag)

	// 仅当两个原始值都有效时才应用坐标系转换
	// (Java 在 BigDecimal decodeCoordinate 返回 null 时跳过转换)。
	if !types.ErrByte4.IsInvalid(rawLon) && !types.ErrByte4.IsInvalid(rawLat) {
		switch m.CoordinateType {
		case 0x01: // WGS84 → GCJ02
			gcj := utils.WGS84ToGCJ02(m.OriginLongitude, m.OriginLatitude)
			m.ConvertLongitude = gcj.Longitude
			m.ConvertLatitude = gcj.Latitude
		default: // 0x02 GCJ02、0x03 OTHER 或任何未知值 → 直通
			m.ConvertLongitude = m.OriginLongitude
			m.ConvertLatitude = m.OriginLatitude
		}
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *LocationV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.LocationV2025Data)

	var status byte
	if !m.Valid {
		status |= 0x01
	}
	if !m.NorthernFlag {
		status |= 0x02
	}
	if !m.EastFlag {
		status |= 0x04
	}
	w.WriteUint8(status)
	w.WriteUint8(m.CoordinateType)
	w.WriteUint32(uint32(encodeV2025Coordinate(m.OriginLongitude)))
	w.WriteUint32(uint32(encodeV2025Coordinate(m.OriginLatitude)))
	return nil
}

// decodeV2025Coordinate 与 Java LocationV2025Codec.decodeCoordinate 一致:
// 原始值 / 1_000_000(HALF_UP 保留 6 位小数),并根据
// isPositive 施加符号(东/北为正)。错误哨兵值原样直通。
func decodeV2025Coordinate(raw int64, isPositive bool) float64 {
	if types.ErrByte4.IsInvalid(raw) {
		return float64(raw)
	}
	coord := float64(raw) / 1_000_000.0
	if !isPositive {
		coord = -coord
	}
	return coord
}

// encodeV2025Coordinate 与 Java LocationV2025Codec.encodeCoordinate 一致:
// abs(coord) × 1_000_000。Java 的 BigDecimal 运算精确,其
// longValue() 截断因此无损;float64 在坐标量级上带有约 1e-8
// 的表示误差,所以我们改为四舍五入到最近值 —— 对任何小数位 ≤6 的
// 数值(即任何从线格式解码出来的值),这都能恢复出精确的原始整数,
// 保持与 Java 互操作逐字节一致(与 2016 版 encodeLocationCoordinate 相同处理)。
// 哨兵值检查先对截断后的值执行,与 Java 的
// inInvalid(coordinate.longValue()) 顺序一致。
// fix 2026-09-17:2025.md L320-321(表21 经度/纬度 DWORD,度×10^6,
// 精确到百万分之一度)—— 旧实现 int64(coord*1_000_000) 截断会因
// float64 表示误差丢失 1 LSB(如 249→248、16000002→16000001、
// 128000003→128000002),改用 math.Round 后才能与原始整数逐字节互还原。
func encodeV2025Coordinate(coord float64) int64 {
	raw := int64(coord)
	if types.ErrByte4.IsInvalid(raw) {
		return raw
	}
	if coord < 0 {
		coord = -coord
	}
	return int64(math.Round(coord * 1_000_000))
}

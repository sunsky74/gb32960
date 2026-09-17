package realtime

import (
	"math"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// LocationDataCodec 编解码 V2016 车辆位置数据子记录(TLV 类型 0x05)。
// 线格式与 Java LocationDataCodec 一致:
//
//	StatusByte(u8) + Longitude(u32 原始值,度 × 10^6) + Latitude(u32 原始值,度 × 10^6)
//
// StatusByte 位布局(Java LocationStatusBits):
//
//	bit 0 (0x01): 0=有效,     1=无效
//	bit 1 (0x02): 0=北纬,     1=南纬
//	bit 2 (0x04): 0=东经,     1=西经
//
// 模型存储带符号的度:解码把原始 u32 除以 1e6,南/西半球取负;
// 编码取 abs() × 1e6 并根据数值符号重建半球位。BYTE4 错误哨兵值
// (0xFFFFFFFE/0xFFFFFFFF)不缩放直接通过,与 Java
// decodeCoordinate/encodeCoordinate 一致。
type LocationDataCodec struct{}

func init() {
	api.Register[mdl.LocationData](api.V2016, &LocationDataCodec{})
}

func (c *LocationDataCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.LocationData{}

	status := r.ReadUint8()
	m.Valid = (status & 0x01) == 0
	isNorthLatitude := (status & 0x02) == 0
	isEastLongitude := (status & 0x04) == 0

	m.Longitude = decodeLocationCoordinate(int64(r.ReadUint32()), isEastLongitude)
	m.Latitude = decodeLocationCoordinate(int64(r.ReadUint32()), isNorthLatitude)

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *LocationDataCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.LocationData)

	var status byte
	if !m.Valid {
		status |= 0x01
	}
	if m.Latitude < 0 {
		status |= 0x02
	}
	if m.Longitude < 0 {
		status |= 0x04
	}
	w.WriteUint8(status)

	w.WriteUint32(uint32(encodeLocationCoordinate(m.Longitude)))
	w.WriteUint32(uint32(encodeLocationCoordinate(m.Latitude)))
	return nil
}

// decodeLocationCoordinate 与 Java LocationDataCodec.decodeCoordinate 一致:
// raw / 1_000_000 并施加半球符号(东/北 = 正)。
// Java 按 HALF_UP 除以 6 位小数;整数 raw / 1e6 至多有 6 位
// 小数,因此取整是精确的,float64 除法与之等价。
// BYTE4 哨兵值原样通过。
func decodeLocationCoordinate(raw int64, isPositive bool) float64 {
	if types.ErrByte4.IsInvalid(raw) {
		return float64(raw)
	}
	coord := float64(raw) / 1_000_000.0
	if !isPositive {
		coord = -coord
	}
	return coord
}

// encodeLocationCoordinate 与 Java LocationDataCodec.encodeCoordinate 一致:
// abs(coord) × 1_000_000。Java 的 BigDecimal 运算精确,其
// longValue() 截断因此无损;float64 在坐标量级上带有约 1e-8
// 的表示误差,所以我们改为四舍五入到最近值 —— 对任何小数位 ≤6 的
// 数值(即任何从线格式解码出来的值),这都能恢复出精确的原始整数,
// 保持与 Java 互操作逐字节一致。
// 哨兵值检查先对截断后的值执行,与 Java 的
// inInvalid(coordinate.longValue()) 顺序一致。Go 没有 null:零值
// 坐标编码为 0(有效的赤道/本初子午线点)。
func encodeLocationCoordinate(coord float64) int64 {
	if types.ErrByte4.IsInvalid(int64(coord)) {
		return int64(coord)
	}
	if coord < 0 {
		coord = -coord
	}
	return int64(math.Round(coord * 1_000_000))
}

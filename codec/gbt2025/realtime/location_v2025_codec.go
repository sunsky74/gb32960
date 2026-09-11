package realtime

import (
	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
	"github.com/sunsky74/gb32960/utils"
)

// LocationV2025Codec encodes/decodes the V2025 车辆位置数据 sub-record
// (TLV type 0x05). Wire layout mirrors Java LocationV2025Codec:
//
//	StatusByte(u8) + CoordinateType(u8)
//	+ OriginLongitude(u32 raw, ×10^6 with hemisphere sign)
//	+ OriginLatitude(u32 raw, ×10^6 with hemisphere sign)
//
// StatusByte bit extraction (Java LocationV2025Codec lines 27-29, 42-52):
//
//	bit 0 (0x01): 0=valid,        1=invalid
//	bit 1 (0x02): 0=north lat,    1=south lat
//	bit 2 (0x04): 0=east long,    1=west long
//
// CoordinateType (Java CoordinateType): 0x01=WGS84 (apply WGS84→GCJ02),
// 0x02=GCJ02 (passthrough), 0x03=OTHER (passthrough).
//
// When OriginLongitude/OriginLatitude is the BYTE4 error sentinel
// (0xFFFFFFFE/0xFFFFFFFF), Java's decodeCoordinate returns null and the
// GCJ02 conversion step is skipped. We mirror that by leaving
// ConvertLongitude/ConvertLatitude at zero in that case.
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

	// Apply coordinate-system conversion only when both raw values are valid
	// (Java skips conversion when BigDecimal decodeCoordinate returned null).
	if !types.ErrByte4.IsInvalid(rawLon) && !types.ErrByte4.IsInvalid(rawLat) {
		switch m.CoordinateType {
		case 0x01: // WGS84 → GCJ02
			gcj := utils.WGS84ToGCJ02(m.OriginLongitude, m.OriginLatitude)
			m.ConvertLongitude = gcj.Longitude
			m.ConvertLatitude = gcj.Latitude
		default: // 0x02 GCJ02, 0x03 OTHER, or any unknown → passthrough
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

// decodeV2025Coordinate mirrors Java LocationV2025Codec.decodeCoordinate:
// raw / 1_000_000 (HALF_UP to 6 decimals) with sign applied based on
// isPositive (east/north = positive). Error sentinels pass through unchanged.
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

// encodeV2025Coordinate mirrors Java LocationV2025Codec.encodeCoordinate:
// abs(coord) × 1_000_000 truncated toward zero. Error sentinels pass through.
func encodeV2025Coordinate(coord float64) int64 {
	raw := int64(coord)
	if types.ErrByte4.IsInvalid(raw) {
		return raw
	}
	if coord < 0 {
		coord = -coord
	}
	return int64(coord * 1_000_000)
}

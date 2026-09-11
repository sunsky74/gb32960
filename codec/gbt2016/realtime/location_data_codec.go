package realtime

import (
	"math"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// LocationDataCodec encodes/decodes the V2016 车辆位置数据 sub-record (TLV type 0x05).
// Wire layout mirrors Java LocationDataCodec:
//
//	StatusByte(u8) + Longitude(u32 raw, degrees × 10^6) + Latitude(u32 raw, degrees × 10^6)
//
// StatusByte bit layout (Java LocationStatusBits):
//
//	bit 0 (0x01): 0=valid,     1=invalid
//	bit 1 (0x02): 0=north lat, 1=south lat
//	bit 2 (0x04): 0=east long, 1=west long
//
// The model stores signed degrees: decode divides the raw u32 by 1e6 and
// negates for south/west hemispheres; encode takes abs() × 1e6 and rebuilds
// the hemisphere bits from the value sign. BYTE4 error sentinels
// (0xFFFFFFFE/0xFFFFFFFF) pass through unscaled, matching Java
// decodeCoordinate/encodeCoordinate.
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

// decodeLocationCoordinate mirrors Java LocationDataCodec.decodeCoordinate:
// raw / 1_000_000 with the hemisphere sign applied (east/north = positive).
// Java divides with 6-decimal HALF_UP; integer raw / 1e6 has at most 6
// decimals so the rounding is exact and float64 division is equivalent.
// BYTE4 sentinels pass through unchanged.
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

// encodeLocationCoordinate mirrors Java LocationDataCodec.encodeCoordinate:
// abs(coord) × 1_000_000. Java's BigDecimal arithmetic is exact and its
// longValue() truncation is therefore lossless; float64 carries ~1e-8
// representation error at coordinate magnitudes, so we round to nearest
// instead — for any value with ≤6 decimals (i.e. anything decoded from the
// wire) this recovers the exact raw integer, keeping Java interop byte-exact.
// The sentinel check runs on the truncated value first, matching Java's
// inInvalid(coordinate.longValue()) ordering. Go has no null: a zero-value
// coordinate encodes as 0 (valid equator/prime-meridian point).
func encodeLocationCoordinate(coord float64) int64 {
	if types.ErrByte4.IsInvalid(int64(coord)) {
		return int64(coord)
	}
	if coord < 0 {
		coord = -coord
	}
	return int64(math.Round(coord * 1_000_000))
}

// Package realtime contains V2016 realtime sub-record codecs.
package realtime

import (
	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// GearPositionCodec encodes/decodes the V2016 挡位 sub-record packed inside
// VehicleData. One wire byte carries bit 7 档位有效 (V2025-only), bit 5 驱动力,
// bit 4 制动力, and bits 3..0 the gear enum. Mirrors Java GearPositionCodec
// which constructs `new GearPosition(readByte(), false)`.
type GearPositionCodec struct{}

func init() {
	api.Register[mdl.GearPosition](api.V2016, &GearPositionCodec{})
}

// Decode reads 1 byte and derives the GearPosition fields exactly as the Java
// GearPosition(byte, boolean) constructor does:
//
//	effective            = ((origin >>> 7) & 0x01) == 1
//	drivingForceActive   = ((origin >>> 5) & 0x01) == 1
//	brakingTorqueApplied = ((origin >>> 4) & 0x01) == 1
//	gp                   = GearPositionEnum.valueOf(origin & 0x0F)
//
// IsV2025 is always false for the V2016 codec.
func (c *GearPositionCodec) Decode(r api.Reader) (api.Message, error) {
	origin := r.ReadUint8()
	m := &mdl.GearPosition{
		Origin:               origin,
		IsV2025:              false,
		Effective:            (origin >> 7 & 0x01) == 1,
		DrivingForceActive:   (origin >> 5 & 0x01) == 1,
		BrakingTorqueApplied: (origin >> 4 & 0x01) == 1,
		GP:                   types.GearPositionEnum(origin & 0x0F),
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode writes the raw origin byte. Java GearPositionCodec writes
// msg.getOrigin() directly, so derived fields are NOT recomputed on encode.
func (c *GearPositionCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.GearPosition)
	w.WriteUint8(m.Origin)
	return nil
}

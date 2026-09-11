package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
	"github.com/sunsky74/gb32960/types"
)

// GearPosition is the V2016 挡位 sub-record packed inside VehicleData.
// One wire byte encodes: bit 7 档位有效 (V2025-only), bit 5 驱动力, bit 4 制动力,
// bits 3..0 档位 (P/R/N/D/...).
// Mirrors the reference Java implementation (GearPosition).
// The derived Effective/DrivingForceActive/BrakingTorqueApplied/GP fields are
// populated by the VehicleData codec which has access to the origin byte.
type GearPosition struct {
	Origin               byte                   // raw wire byte
	IsV2025              bool                   // 是否是新国标 (false for V2016)
	Effective            bool                   // 档位数据有效标识 (V2025 only)
	DrivingForceActive   bool                   // 有无驱动力标识
	BrakingTorqueApplied bool                   // 有无制动力标识
	GP                   types.GearPositionEnum // 档位 (decoded low nibble)
}

func (m *GearPosition) Version() api.GBTVersion { return api.V2016 }

func (m *GearPosition) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*GearPosition)(nil)).Elem())
}

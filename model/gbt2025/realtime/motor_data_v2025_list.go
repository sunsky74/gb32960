package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// MotorDataV2025List is the V2025 drive motor data list (TLV 0x02).
type MotorDataV2025List struct {
	MotorCount int
	Items      []MotorDataV2025
}

func (m *MotorDataV2025List) Version() api.GBTVersion { return api.V2025 }
func (m *MotorDataV2025List) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*MotorDataV2025List)(nil)).Elem())
}

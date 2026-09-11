package gbt2016

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// VehicleLogin is the V2016 vehicle login request (command 0x01).
// Wire layout: BeanTime(6B) + SerialNum(2B) + ICCID(20B) + Count(1B) + Length(1B) + Count*Length bytes of codes.
// V2016 carries a single shared Length for every subsystem code; V2025 uses per-code lengths.
type VehicleLogin struct {
	_         GBT2016Body
	BeanTime  model.BeanTime
	SerialNum int
	ICCID     string   // 20-byte fixed-length field
	Count     int      // rechargeable subsystem count
	Length    int      // shared code length for every entry (V2016 only)
	Codes     []string // Count entries, each Length bytes
}

func (m *VehicleLogin) Version() api.GBTVersion { return api.V2016 }

func (m *VehicleLogin) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*VehicleLogin)(nil)).Elem())
}

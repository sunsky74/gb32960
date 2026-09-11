package gbt2016

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// VehicleLogout is the V2016 vehicle logout request (command 0x04).
// Wire layout: BeanTime(6B) + SerialNum(2B).
// Also reused by the V2025 decoder for the V2025 0x04 command (Java
// getV2025Body returns the same struct).
type VehicleLogout struct {
	_         GBT2016Body
	BeanTime  model.BeanTime
	SerialNum int
}

func (m *VehicleLogout) Version() api.GBTVersion { return api.V2016 }

func (m *VehicleLogout) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*VehicleLogout)(nil)).Elem())
}

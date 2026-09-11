package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// VehicleLoginV2025 is the V2025 vehicle login request (command 0x01).
// CRITICAL: Lengths is a SLICE (each battery management system has its own
// battery pack count), unlike V2016 VehicleLogin where all codes share a
// single Length field.
type VehicleLoginV2025 struct {
	_         GBT2025Body
	BeanTime  model.BeanTime
	SerialNum int
	ICCID     string // 20 bytes fixed-length
	Count     int    // 电池管理系统数
	Lengths   []int  // one length per battery management system
	Codes     []string
}

func (m *VehicleLoginV2025) Version() api.GBTVersion { return api.V2025 }
func (m *VehicleLoginV2025) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*VehicleLoginV2025)(nil)).Elem())
}

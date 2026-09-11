package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
	"github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// VehicleActivate is the V2025 vehicle activation request (command 0x09).
type VehicleActivate struct {
	_               GBT2025Body
	CollectTime     model.BeanTime
	ChipID          string
	PublicKeyLength int
	PublicKey       []byte
	VIN             string
	Signature       *realtime.VehicleSignature
}

func (m *VehicleActivate) Version() api.GBTVersion { return api.V2025 }
func (m *VehicleActivate) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*VehicleActivate)(nil)).Elem())
}

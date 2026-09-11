package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// VehicleActivateResponse is the V2025 vehicle activation response (command 0x0A).
// ResponseCode maps to Java VehicleActivateResponseEnum:
// 0x00=SUCCESS, 0x01=ACTIVATED, 0x02=VIN_REPEAT.
type VehicleActivateResponse struct {
	_            GBT2025Body
	Success      bool
	ResponseCode byte // VehicleActivateResponseEnum
}

func (m *VehicleActivateResponse) Version() api.GBTVersion { return api.V2025 }
func (m *VehicleActivateResponse) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*VehicleActivateResponse)(nil)).Elem())
}

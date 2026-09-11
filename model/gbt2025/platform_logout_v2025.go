package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// PlatformLogoutV2025 is the V2025 platform logout request (command 0x06).
// V2025 reuses the V2016 PlatformLogout wire format (reference implementation).
type PlatformLogoutV2025 struct {
	_         GBT2025Body
	BeanTime  model.BeanTime
	SerialNum int
}

func (m *PlatformLogoutV2025) Version() api.GBTVersion { return api.V2025 }
func (m *PlatformLogoutV2025) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*PlatformLogoutV2025)(nil)).Elem())
}

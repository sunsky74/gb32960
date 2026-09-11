package gbt2016

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// PlatformLogout is the V2016 platform logout request (command 0x06).
// Wire layout: BeanTime(6B) + SerialNum(2B).
type PlatformLogout struct {
	_         GBT2016Body
	BeanTime  model.BeanTime
	SerialNum int
}

func (m *PlatformLogout) Version() api.GBTVersion { return api.V2016 }

func (m *PlatformLogout) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*PlatformLogout)(nil)).Elem())
}

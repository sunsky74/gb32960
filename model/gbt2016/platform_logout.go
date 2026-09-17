package gbt2016

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// PlatformLogout 是 V2016 平台登出请求(命令 0x06)。
// 线格式布局:BeanTime(6B) + SerialNum(2B)。
type PlatformLogout struct {
	_         GBT2016Body
	BeanTime  model.BeanTime
	SerialNum int
}

func (m *PlatformLogout) Version() api.GBTVersion { return api.V2016 }

func (m *PlatformLogout) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*PlatformLogout)(nil)).Elem())
}

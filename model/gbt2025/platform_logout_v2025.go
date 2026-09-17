package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// PlatformLogoutV2025 是 V2025 平台登出请求(命令 0x06)。
// V2025 复用 V2016 PlatformLogout 的线格式(参考实现)。
type PlatformLogoutV2025 struct {
	_         GBT2025Body
	BeanTime  model.BeanTime
	SerialNum int
}

func (m *PlatformLogoutV2025) Version() api.GBTVersion { return api.V2025 }
func (m *PlatformLogoutV2025) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*PlatformLogoutV2025)(nil)).Elem())
}

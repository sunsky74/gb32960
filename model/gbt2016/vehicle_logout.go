package gbt2016

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// VehicleLogout 是 V2016 车辆登出请求(命令 0x04)。
// 线格式布局:BeanTime(6B) + SerialNum(2B)。
// V2025 解码器处理 V2025 0x04 命令时也会复用它(Java
// getV2025Body 返回同一结构体)。
type VehicleLogout struct {
	_         GBT2016Body
	BeanTime  model.BeanTime
	SerialNum int
}

func (m *VehicleLogout) Version() api.GBTVersion { return api.V2016 }

func (m *VehicleLogout) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*VehicleLogout)(nil)).Elem())
}

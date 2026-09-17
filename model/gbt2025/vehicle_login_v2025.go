package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// VehicleLoginV2025 是 V2025 车辆登入请求(命令 0x01)。
// 关键:Lengths 是一个切片(每个电池管理系统都有自己的
// 电池包计数),不同于 V2016 VehicleLogin 中所有 code 共享
// 单个 Length 字段。
type VehicleLoginV2025 struct {
	_         GBT2025Body
	BeanTime  model.BeanTime
	SerialNum int
	ICCID     string // 20 字节定长
	Count     int    // 电池管理系统数
	Lengths   []int  // 每个电池管理系统一个长度
	Codes     []string
}

func (m *VehicleLoginV2025) Version() api.GBTVersion { return api.V2025 }
func (m *VehicleLoginV2025) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*VehicleLoginV2025)(nil)).Elem())
}

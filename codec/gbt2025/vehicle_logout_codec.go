package gbt2025

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec/gbt2016"
	mdl "github.com/sunsky74/gb32960/model/gbt2016"
)

// V2025 命令 0x04 (车辆登出) 复用 V2016 的消息体布局
// (BeanTime 6B + SerialNum 2B);与 Java getV2025Body(VECHICLE_LOGOUT) 一致。
func init() {
	api.Register[mdl.VehicleLogout](api.V2025, &gbt2016.VehicleLogoutCodec{})
}

package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// VehicleActivateResponse 是 V2025 车辆激活响应(命令 0x0A)。
// Success 对应表 B.4 激活状态:0x01=激活成功,0x02=激活失败;
// 解码时非 0x01 一律为 false(fix 2026-09-17: spec 2025.md L863-865)。
// ResponseCode 对应 Java VehicleActivateResponseEnum:
// 0x00=SUCCESS, 0x01=ACTIVATED, 0x02=VIN_REPEAT。
type VehicleActivateResponse struct {
	_            GBT2025Body
	Success      bool
	ResponseCode byte // VehicleActivateResponseEnum
}

func (m *VehicleActivateResponse) Version() api.GBTVersion { return api.V2025 }
func (m *VehicleActivateResponse) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*VehicleActivateResponse)(nil)).Elem())
}

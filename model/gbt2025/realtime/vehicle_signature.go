package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// VehicleSignature 是 V2025 车辆签名数据(TLV 0xFF)。
// SignatureType 映射到 Java 的 SignatureType:0x01=SM2, 0x02=RSA, 0x03=ECC, 0xFF=OTHER。
type VehicleSignature struct {
	Type     byte   // SignatureType 枚举
	RLength  int    // R值长度
	RValue   []byte // R值
	SLength  int    // S值长度
	SValue   []byte // S值
	SignData []byte // 签名原始数据
}

func (m *VehicleSignature) Version() api.GBTVersion { return api.V2025 }
func (m *VehicleSignature) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*VehicleSignature)(nil)).Elem())
}

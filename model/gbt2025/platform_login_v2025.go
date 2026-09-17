package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// PlatformLoginV2025 是 V2025 平台登入请求(命令 0x05)。
// V2025 复用 V2016 PlatformLogin 的线格式(参考实现)。
// 线格式:BeanTime(6B) + SerialNum(2B) + Username(12B) + Password(20B) + Cipher(1B) = 41 字节。
type PlatformLoginV2025 struct {
	_         GBT2025Body
	BeanTime  model.BeanTime
	SerialNum int
	Username  string // 12 字节定长
	Password  string // 20 字节定长
	Cipher    byte   // 1 字节加密选择符(对应 Java cipherSelect.select(version(), buffer.readByte()))
}

func (m *PlatformLoginV2025) Version() api.GBTVersion { return api.V2025 }
func (m *PlatformLoginV2025) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*PlatformLoginV2025)(nil)).Elem())
}

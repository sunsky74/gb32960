package gbt2016

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// PlatformLogin 是 V2016 平台登入请求(命令 0x05)。
// 线格式布局:BeanTime(6B) + SerialNum(2B) + Username(12B) + Password(20B) + Cipher(1B) = 41 字节。
// Cipher 为 1 字节,与 Java PlatformLoginCodec 第 40 行一致。
type PlatformLogin struct {
	_         GBT2016Body
	BeanTime  model.BeanTime
	SerialNum int
	Username  string // 12 字节定长字段
	Password  string // 20 字节定长字段
	Cipher    byte   // 1 字节加密选择器 (与 Java cipherSelect.select(version(), buffer.readByte()) 一致)
}

func (m *PlatformLogin) Version() api.GBTVersion { return api.V2016 }

func (m *PlatformLogin) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*PlatformLogin)(nil)).Elem())
}

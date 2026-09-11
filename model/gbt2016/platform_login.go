package gbt2016

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// PlatformLogin is the V2016 platform login request (command 0x05).
// Wire layout: BeanTime(6B) + SerialNum(2B) + Username(12B) + Password(20B) + Cipher(1B) = 41 bytes.
// Cipher is 1 byte — matches Java PlatformLoginCodec line 40.
type PlatformLogin struct {
	_         GBT2016Body
	BeanTime  model.BeanTime
	SerialNum int
	Username  string // 12-byte fixed-length field
	Password  string // 20-byte fixed-length field
	Cipher    byte   // 1-byte encryption selector (matches Java cipherSelect.select(version(), buffer.readByte()))
}

func (m *PlatformLogin) Version() api.GBTVersion { return api.V2016 }

func (m *PlatformLogin) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*PlatformLogin)(nil)).Elem())
}

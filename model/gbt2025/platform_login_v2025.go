package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// PlatformLoginV2025 is the V2025 platform login request (command 0x05).
// V2025 reuses the V2016 PlatformLogin wire format (reference implementation).
// Wire format: BeanTime(6B) + SerialNum(2B) + Username(12B) + Password(20B) + Cipher(1B) = 41 bytes.
type PlatformLoginV2025 struct {
	_         GBT2025Body
	BeanTime  model.BeanTime
	SerialNum int
	Username  string // 12 bytes fixed-length
	Password  string // 20 bytes fixed-length
	Cipher    byte   // 1-byte encryption selector (matches Java cipherSelect.select(version(), buffer.readByte()))
}

func (m *PlatformLoginV2025) Version() api.GBTVersion { return api.V2025 }
func (m *PlatformLoginV2025) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*PlatformLoginV2025)(nil)).Elem())
}

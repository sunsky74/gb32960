package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// KeyExchangeData is the V2025 data unit encryption key exchange message (command 0x0B).
// Type maps to Java ExchangeKeyEnum: 0x02=RSA, 0x03=AES, 0x04=SM2, 0x05=SM4, 0x7F=OTHER.
type KeyExchangeData struct {
	_          GBT2025Body
	Type       byte // ExchangeKeyEnum
	Length     int
	Key        []byte
	StartTime  model.BeanTime
	ExpireTime model.BeanTime
}

func (m *KeyExchangeData) Version() api.GBTVersion { return api.V2025 }
func (m *KeyExchangeData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*KeyExchangeData)(nil)).Elem())
}

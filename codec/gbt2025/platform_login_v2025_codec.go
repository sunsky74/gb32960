package gbt2025

import (
	"reflect"
	"strings"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2025"
)

// PlatformLoginV2025Codec encodes/decodes the V2025 platform login request
// (command 0x05). Wire layout mirrors Java PlatformLoginV2025Codec (which
// reuses the V2016 PlatformLogin model):
// BeanTime(6B) + SerialNum(2B) + Username(12B) + Password(20B) + Cipher(1B).
type PlatformLoginV2025Codec struct{}

func init() {
	api.Register[mdl.PlatformLoginV2025](api.V2025, &PlatformLoginV2025Codec{})
}

// Decode reads BeanTime via the registered V2016 BeanTime codec, then
// SerialNum, Username, Password, and the 1-byte Cipher.
func (c *PlatformLoginV2025Codec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.PlatformLoginV2025{}

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return nil, api.ErrCodecNotFound
	}
	bt, err := btCodec.Decode(r)
	if err != nil {
		return nil, err
	}
	m.BeanTime = *bt.(*model.BeanTime)

	m.SerialNum = int(r.ReadUint16())
	m.Username = strings.TrimRight(r.ReadString(12), " ")
	m.Password = strings.TrimRight(r.ReadString(20), " ")
	m.Cipher = r.ReadUint8()

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode writes BeanTime via the registered V2016 BeanTime codec, then
// SerialNum, Username, Password, and the 1-byte Cipher.
func (c *PlatformLoginV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.PlatformLoginV2025)

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return api.ErrCodecNotFound
	}
	if err := btCodec.Encode(w, &m.BeanTime); err != nil {
		return err
	}

	w.WriteUint16(uint16(m.SerialNum))
	w.WriteString(m.Username, 12)
	w.WriteString(m.Password, 20)
	w.WriteUint8(m.Cipher)
	return nil
}

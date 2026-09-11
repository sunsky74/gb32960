package gbt2016

import (
	"reflect"
	"strings"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2016"
)

// PlatformLoginCodec encodes/decodes the V2016 platform login request (command 0x05).
// Wire layout: BeanTime(6B) + SerialNum(2B) + Username(12B) + Password(20B) + Cipher(1B) = 41 bytes.
// Cipher is 1 byte (matches Java PlatformLoginCodec line 40).
type PlatformLoginCodec struct{}

func init() {
	api.Register[mdl.PlatformLogin](api.V2016, &PlatformLoginCodec{})
}

// Decode reads BeanTime via the registered BeanTime codec, then SerialNum,
// Username, Password, and the 1-byte Cipher.
func (c *PlatformLoginCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.PlatformLogin{}

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
		return nil, err
	}
	return m, nil
}

// Encode writes BeanTime via the registered BeanTime codec, then SerialNum,
// Username, Password, and the 1-byte Cipher.
func (c *PlatformLoginCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.PlatformLogin)

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

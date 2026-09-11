package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2025"
)

// KeyExchangeCodec encodes/decodes the V2025 data unit encryption key exchange
// message (command 0x0B). Wire layout mirrors Java KeyExchangeCodec:
//
//	Type(u8 ExchangeKeyEnum) + Length(u16) + Key[Length bytes]
//	+ StartTime(BeanTime 6B) + ExpireTime(BeanTime 6B)
type KeyExchangeCodec struct{}

func init() {
	api.Register[mdl.KeyExchangeData](api.V2025, &KeyExchangeCodec{})
}

// Decode mirrors Java KeyExchangeCodec.decodeBuffer.
func (c *KeyExchangeCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.KeyExchangeData{}

	m.Type = r.ReadUint8()
	m.Length = int(r.ReadUint16())
	if m.Length > 0 {
		m.Key = r.ReadBytes(m.Length)
	}

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return nil, api.ErrCodecNotFound
	}
	startBT, err := btCodec.Decode(r)
	if err != nil {
		return nil, err
	}
	m.StartTime = *startBT.(*model.BeanTime)

	expireBT, err := btCodec.Decode(r)
	if err != nil {
		return nil, err
	}
	m.ExpireTime = *expireBT.(*model.BeanTime)

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode mirrors Java KeyExchangeCodec.encodeBuffer.
func (c *KeyExchangeCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.KeyExchangeData)

	w.WriteUint8(m.Type)
	w.WriteUint16(uint16(m.Length))
	if m.Length > 0 {
		w.WriteBytes(m.Key)
	}

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return api.ErrCodecNotFound
	}
	if err := btCodec.Encode(w, &m.StartTime); err != nil {
		return err
	}
	if err := btCodec.Encode(w, &m.ExpireTime); err != nil {
		return err
	}
	return nil
}

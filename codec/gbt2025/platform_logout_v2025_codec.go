// Package gbt2025 contains GB/T 32960.3-2025 message body codecs.
package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2025"
)

// PlatformLogoutV2025Codec encodes/decodes the V2025 platform logout request
// (command 0x06). Wire layout: BeanTime(6B) + SerialNum(2B).
// BeanTime is shared with V2016; we look it up under V2016 (no V2025 BeanTime
// codec is registered — BeanTime has no version-specific behavior).
type PlatformLogoutV2025Codec struct{}

func init() {
	api.Register[mdl.PlatformLogoutV2025](api.V2025, &PlatformLogoutV2025Codec{})
}

// Decode mirrors Java PlatformLogoutV2025Codec.decodeBuffer.
func (c *PlatformLogoutV2025Codec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.PlatformLogoutV2025{}

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

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode mirrors Java PlatformLogoutV2025Codec.encodeBuffer.
func (c *PlatformLogoutV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.PlatformLogoutV2025)

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return api.ErrCodecNotFound
	}
	if err := btCodec.Encode(w, &m.BeanTime); err != nil {
		return err
	}

	w.WriteUint16(uint16(m.SerialNum))
	return nil
}

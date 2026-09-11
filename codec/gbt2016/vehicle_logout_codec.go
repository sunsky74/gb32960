package gbt2016

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2016"
)

// VehicleLogoutCodec encodes/decodes the V2016 vehicle logout request (command 0x04).
// Wire layout: BeanTime(6B) + SerialNum(2B).
// Also reused by the V2025 decoder for the V2025 0x04 command.
type VehicleLogoutCodec struct{}

func init() {
	api.Register[mdl.VehicleLogout](api.V2016, &VehicleLogoutCodec{})
}

// Decode mirrors Java VehicleLogoutCodec.decodeBuffer.
func (c *VehicleLogoutCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.VehicleLogout{}

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

// Encode mirrors Java VehicleLogoutCodec.encodeBuffer.
func (c *VehicleLogoutCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.VehicleLogout)

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

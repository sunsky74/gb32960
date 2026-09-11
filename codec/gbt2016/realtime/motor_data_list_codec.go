package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
)

// MotorDataListCodec encodes/decodes the V2016 驱动电机 list (TLV type 0x02 body).
// Wire layout: Count(u8) + Count × MotorData. Mirrors Java MotorDataListCodec.
type MotorDataListCodec struct{}

func init() {
	api.Register[mdl.MotorDataList](api.V2016, &MotorDataListCodec{})
}

func (c *MotorDataListCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.MotorDataList{}
	m.Count = int(r.ReadUint8())
	if m.Count > 0 {
		elemCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.MotorData)(nil)).Elem())
		if elemCodec == nil {
			return nil, api.ErrCodecNotFound
		}
		m.Items = make([]mdl.MotorData, m.Count)
		for i := 0; i < m.Count; i++ {
			item, err := elemCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			m.Items[i] = *item.(*mdl.MotorData)
		}
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *MotorDataListCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.MotorDataList)
	w.WriteUint8(byte(m.Count))
	if m.Count > 0 && len(m.Items) > 0 {
		elemCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.MotorData)(nil)).Elem())
		if elemCodec == nil {
			return api.ErrCodecNotFound
		}
		for i := range m.Items {
			if err := elemCodec.Encode(w, &m.Items[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

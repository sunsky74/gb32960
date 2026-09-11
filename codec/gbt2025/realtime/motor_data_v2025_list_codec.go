package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// MotorDataV2025ListCodec encodes/decodes the V2025 驱动电机 list
// (TLV type 0x02 body). Wire layout: MotorCount(u8) + MotorCount × MotorDataV2025.
// Mirrors Java MotorDataV2025ListCodec.
type MotorDataV2025ListCodec struct{}

func init() {
	api.Register[mdl.MotorDataV2025List](api.V2025, &MotorDataV2025ListCodec{})
}

func (c *MotorDataV2025ListCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.MotorDataV2025List{}
	m.MotorCount = int(r.ReadUint8())
	if m.MotorCount > 0 {
		elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.MotorDataV2025)(nil)).Elem())
		if elemCodec == nil {
			return nil, api.ErrCodecNotFound
		}
		m.Items = make([]mdl.MotorDataV2025, m.MotorCount)
		for i := 0; i < m.MotorCount; i++ {
			item, err := elemCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			m.Items[i] = *item.(*mdl.MotorDataV2025)
		}
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *MotorDataV2025ListCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.MotorDataV2025List)
	w.WriteUint8(byte(m.MotorCount))
	if m.MotorCount > 0 && len(m.Items) > 0 {
		elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.MotorDataV2025)(nil)).Elem())
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

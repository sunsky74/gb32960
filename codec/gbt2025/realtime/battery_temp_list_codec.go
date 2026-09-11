package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// BatteryTempListCodec encodes/decodes the V2025 动力蓄电池温度 list
// (TLV type 0x08 body). Wire layout:
//
//	BatteryPackCount(u8) + BatteryPackCount × BatteryTemp
//
// Mirrors Java BatteryPackTemperatureListCodec: a BYTE1 sentinel count
// (0xFE/0xFF) means no entries follow; encode writes the fixed 0xFF byte
// (not the original count) when the count is a sentinel.
type BatteryTempListCodec struct{}

func init() {
	api.Register[mdl.BatteryTempList](api.V2025, &BatteryTempListCodec{})
}

func (c *BatteryTempListCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.BatteryTempList{}
	m.BatteryPackCount = int(r.ReadUint8())
	if !types.ErrByte1.IsInvalid(int64(m.BatteryPackCount)) {
		elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.BatteryTemp)(nil)).Elem())
		if elemCodec == nil {
			return nil, api.ErrCodecNotFound
		}
		m.Items = make([]mdl.BatteryTemp, m.BatteryPackCount)
		for i := 0; i < m.BatteryPackCount; i++ {
			item, err := elemCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			m.Items[i] = *item.(*mdl.BatteryTemp)
		}
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *BatteryTempListCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.BatteryTempList)
	if types.ErrByte1.IsInvalid(int64(m.BatteryPackCount)) {
		w.WriteUint8(byte(types.ErrByte1.Invalid))
		return nil
	}
	w.WriteUint8(byte(m.BatteryPackCount))
	elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.BatteryTemp)(nil)).Elem())
	if elemCodec == nil {
		return api.ErrCodecNotFound
	}
	for i := range m.Items {
		if err := elemCodec.Encode(w, &m.Items[i]); err != nil {
			return err
		}
	}
	return nil
}

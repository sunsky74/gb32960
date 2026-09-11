package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// FuelCellStackDataListCodec encodes/decodes the V2025 燃料电池电堆 list
// (TLV type 0x30 body). Wire layout:
//
//	StackCount(u8) + StackCount × FuelCellStackData
//
// Mirrors Java FuelCellStackDataListCodec: when StackCount is the BYTE1
// sentinel (0xFE/0xFF) no entries follow — decode stops after the count and
// encode writes the count byte as-is without entries.
type FuelCellStackDataListCodec struct{}

func init() {
	api.Register[mdl.FuelCellStackDataList](api.V2025, &FuelCellStackDataListCodec{})
}

func (c *FuelCellStackDataListCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.FuelCellStackDataList{}
	m.StackCount = int(r.ReadUint8())
	if !types.ErrByte1.IsInvalid(int64(m.StackCount)) {
		elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.FuelCellStackData)(nil)).Elem())
		if elemCodec == nil {
			return nil, api.ErrCodecNotFound
		}
		m.Items = make([]mdl.FuelCellStackData, m.StackCount)
		for i := 0; i < m.StackCount; i++ {
			item, err := elemCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			m.Items[i] = *item.(*mdl.FuelCellStackData)
		}
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *FuelCellStackDataListCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.FuelCellStackDataList)
	w.WriteUint8(byte(m.StackCount))
	if !types.ErrByte1.IsInvalid(int64(m.StackCount)) {
		elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.FuelCellStackData)(nil)).Elem())
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

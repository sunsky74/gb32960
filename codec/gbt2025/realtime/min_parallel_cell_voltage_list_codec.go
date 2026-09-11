package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// MinParallelCellVoltageListCodec encodes/decodes the V2025 动力蓄电池最小并联
// 单元电压 list (TLV type 0x07 body). Wire layout:
//
//	BatteryPackCount(u8) + BatteryPackCount × MinParallelCellVoltage
//
// Mirrors Java MinParallelCellVoltageListCodec: a BYTE1 sentinel count
// (0xFE/0xFF) means no entries follow; encode writes the single 0xFF byte
// when the count is a sentinel or the list exceeds 50 packs.
type MinParallelCellVoltageListCodec struct{}

func init() {
	api.Register[mdl.MinParallelCellVoltageList](api.V2025, &MinParallelCellVoltageListCodec{})
}

func (c *MinParallelCellVoltageListCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.MinParallelCellVoltageList{}
	m.BatteryPackCount = int(r.ReadUint8())
	if !types.ErrByte1.IsInvalid(int64(m.BatteryPackCount)) {
		elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.MinParallelCellVoltage)(nil)).Elem())
		if elemCodec == nil {
			return nil, api.ErrCodecNotFound
		}
		m.Items = make([]mdl.MinParallelCellVoltage, m.BatteryPackCount)
		for i := 0; i < m.BatteryPackCount; i++ {
			item, err := elemCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			m.Items[i] = *item.(*mdl.MinParallelCellVoltage)
		}
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *MinParallelCellVoltageListCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.MinParallelCellVoltageList)
	if len(m.Items) > 50 {
		w.WriteUint8(byte(types.ErrByte1.Invalid))
		return nil
	}
	w.WriteUint8(byte(m.BatteryPackCount))
	elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.MinParallelCellVoltage)(nil)).Elem())
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

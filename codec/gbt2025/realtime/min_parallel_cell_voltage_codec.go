package realtime

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// MinParallelCellVoltageCodec encodes/decodes one V2025 动力蓄电池最小并联单元
// 电压 entry (TLV type 0x07 element). Field order and converters mirror Java
// MinParallelCellVoltageCodec exactly:
//
//	BatteryPackSeq(u8)
//	Voltage(u16 VoltageConverter scale=0.1)
//	Current(u16 CurrentConverter2025 offset=3000 scale=0.1)
//	MinParallelUnits(u16)
//	MinParallelUnits × BatteryVoltage(u16 BatteryVoltagesConverter2025 scale=0.001)
//
// When MinParallelUnits is the BYTE2 error sentinel, the per-unit voltage
// list is skipped (Java: if units > 0 && !BATTERY_VOLTAGES_CONVERTER.isInvalid).
type MinParallelCellVoltageCodec struct{}

func init() {
	api.Register[mdl.MinParallelCellVoltage](api.V2025, &MinParallelCellVoltageCodec{})
}

func (c *MinParallelCellVoltageCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.MinParallelCellVoltage{}
	m.BatteryPackSeq = int(r.ReadUint8())
	m.Voltage = codec.VoltageConverter.Decode(int64(r.ReadUint16()))
	m.Current = codec.CurrentConverter2025.Decode(int64(r.ReadUint16()))
	m.MinParallelUnits = int(r.ReadUint16())

	if m.MinParallelUnits > 0 && !types.ErrByte2.IsInvalid(int64(m.MinParallelUnits)) {
		m.BatteryVoltages = make([]float64, m.MinParallelUnits)
		for i := 0; i < m.MinParallelUnits; i++ {
			m.BatteryVoltages[i] = codec.BatteryVoltagesConverter2025.Decode(int64(r.ReadUint16()))
		}
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *MinParallelCellVoltageCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.MinParallelCellVoltage)
	w.WriteUint8(byte(m.BatteryPackSeq))
	w.WriteUint16(uint16(codec.VoltageConverter.Encode(m.Voltage)))
	w.WriteUint16(uint16(codec.CurrentConverter2025.Encode(m.Current)))
	w.WriteUint16(uint16(m.MinParallelUnits))
	if m.MinParallelUnits > 0 && !types.ErrByte2.IsInvalid(int64(m.MinParallelUnits)) {
		for _, v := range m.BatteryVoltages {
			w.WriteUint16(uint16(codec.BatteryVoltagesConverter2025.Encode(v)))
		}
	}
	return nil
}

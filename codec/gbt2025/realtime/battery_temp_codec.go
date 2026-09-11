package realtime

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// BatteryTempCodec encodes/decodes one V2025 动力蓄电池温度 entry
// (TLV type 0x08 element). Renamed from Java BatteryPackTemperature to match
// the TLV enum constant BATTERY_TEMP. Field order and converters mirror Java
// BatteryPackTemperatureCodec exactly:
//
//	BatteryPackSeq(u8)
//	TemperatureProbeCount(u16)
//	TemperatureProbeCount × ProbeTemperature(u8 TemperatureConverter offset=40)
//
// When TemperatureProbeCount is the BYTE2 error sentinel, the probe
// temperature list is skipped (Java early-returns with empty list).
type BatteryTempCodec struct{}

func init() {
	api.Register[mdl.BatteryTemp](api.V2025, &BatteryTempCodec{})
}

func (c *BatteryTempCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.BatteryTemp{}
	m.BatteryPackSeq = int(r.ReadUint8())
	m.TemperatureProbeCount = int(r.ReadUint16())

	if !types.ErrByte2.IsInvalid(int64(m.TemperatureProbeCount)) && m.TemperatureProbeCount > 0 {
		m.ProbeTemperatures = make([]float64, m.TemperatureProbeCount)
		for i := 0; i < m.TemperatureProbeCount; i++ {
			m.ProbeTemperatures[i] = codec.TemperatureConverter.Decode(int64(r.ReadUint8()))
		}
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *BatteryTempCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.BatteryTemp)

	// Java clamps the pack seq to the BYTE1 invalid sentinel when it is a
	// sentinel itself, negative, or beyond the 50-pack protocol maximum.
	if types.ErrByte1.IsInvalid(int64(m.BatteryPackSeq)) || m.BatteryPackSeq < 0 || m.BatteryPackSeq > 50 {
		w.WriteUint8(byte(types.ErrByte1.Invalid))
	} else {
		w.WriteUint8(byte(m.BatteryPackSeq))
	}

	if types.ErrByte2.IsInvalid(int64(m.TemperatureProbeCount)) {
		// Java writes the fixed invalid sentinel (0xFFFF), rewriting 0xFFFE
		w.WriteUint16(uint16(types.ErrByte2.Invalid))
		return nil
	}

	w.WriteUint16(uint16(m.TemperatureProbeCount))
	for _, t := range m.ProbeTemperatures {
		w.WriteUint8(byte(codec.TemperatureConverter.Encode(t)))
	}
	return nil
}

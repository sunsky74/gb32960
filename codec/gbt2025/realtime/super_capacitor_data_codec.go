package realtime

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// SuperCapacitorDataCodec encodes/decodes the V2025 超级电容数据 sub-record
// (TLV type 0x31). Field order and converters mirror Java SuperCapacitorDataCodec:
//
//	ManagementSystemNumber(u8)
//	TotalVoltage(u16 TotalVoltageConverter scale=0.1)
//	TotalCurrent(u16 TotalCurrentConverter offset=3000 scale=0.1)
//	CapacitorCount(u16)
//	CapacitorCount × CapacitorVoltage(u16 SuperCapVoltageConverter scale=0.001)
//	TemperatureProbeCount(u16)
//	TemperatureProbeCount × ProbeTemperature(u8 TemperatureConverter offset=40)
//
// When CapacitorCount/TemperatureProbeCount is the BYTE2 error sentinel, the
// corresponding list is skipped (Java sets Collections.emptyList() and continues).
type SuperCapacitorDataCodec struct{}

func init() {
	api.Register[mdl.SuperCapacitorData](api.V2025, &SuperCapacitorDataCodec{})
}

func (c *SuperCapacitorDataCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.SuperCapacitorData{}
	m.ManagementSystemNumber = int(r.ReadUint8())
	m.TotalVoltage = codec.TotalVoltageConverter.Decode(int64(r.ReadUint16()))
	m.TotalCurrent = codec.TotalCurrentConverter.Decode(int64(r.ReadUint16()))
	m.CapacitorCount = int(r.ReadUint16())

	if !types.ErrByte2.IsInvalid(int64(m.CapacitorCount)) && m.CapacitorCount > 0 {
		m.CapacitorVoltages = make([]float64, m.CapacitorCount)
		for i := 0; i < m.CapacitorCount; i++ {
			m.CapacitorVoltages[i] = codec.SuperCapVoltageConverter.Decode(int64(r.ReadUint16()))
		}
	}

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

func (c *SuperCapacitorDataCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.SuperCapacitorData)
	w.WriteUint8(byte(m.ManagementSystemNumber))
	w.WriteUint16(uint16(codec.TotalVoltageConverter.Encode(m.TotalVoltage)))
	w.WriteUint16(uint16(codec.TotalCurrentConverter.Encode(m.TotalCurrent)))
	w.WriteUint16(uint16(m.CapacitorCount))
	if !types.ErrByte2.IsInvalid(int64(m.CapacitorCount)) && len(m.CapacitorVoltages) > 0 {
		for _, v := range m.CapacitorVoltages {
			w.WriteUint16(uint16(codec.SuperCapVoltageConverter.Encode(v)))
		}
	}
	w.WriteUint16(uint16(m.TemperatureProbeCount))
	if !types.ErrByte2.IsInvalid(int64(m.TemperatureProbeCount)) && len(m.ProbeTemperatures) > 0 {
		for _, t := range m.ProbeTemperatures {
			w.WriteUint8(byte(codec.TemperatureConverter.Encode(t)))
		}
	}
	return nil
}

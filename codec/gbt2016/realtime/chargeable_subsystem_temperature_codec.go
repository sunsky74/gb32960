package realtime

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
)

// ChargeableSubsystemTemperatureCodec encodes/decodes one V2016 可充电储能装置温度
// entry (TLV type 0x09 element). Field order and converters mirror Java
// ChargeableSubsystemTemperatureCodec exactly:
//
//	SubSystemNumber(u8)
//	TemperatureProbeCount(u16)
//	ProbeTemperatures[TemperatureProbeCount × u8 TemperatureConverter]  // scale=1, offset=40
type ChargeableSubsystemTemperatureCodec struct{}

func init() {
	api.Register[mdl.ChargeableSubsystemTemperature](api.V2016, &ChargeableSubsystemTemperatureCodec{})
}

func (c *ChargeableSubsystemTemperatureCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.ChargeableSubsystemTemperature{}
	m.SubSystemNumber = int(r.ReadUint8())
	m.TemperatureProbeCount = int(r.ReadUint16())
	if m.TemperatureProbeCount > 0 {
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

func (c *ChargeableSubsystemTemperatureCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.ChargeableSubsystemTemperature)
	w.WriteUint8(byte(m.SubSystemNumber))
	w.WriteUint16(uint16(m.TemperatureProbeCount))
	for _, t := range m.ProbeTemperatures {
		w.WriteUint8(byte(codec.TemperatureConverter.Encode(t)))
	}
	return nil
}

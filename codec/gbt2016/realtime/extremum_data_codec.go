package realtime

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
)

// ExtremumDataCodec encodes/decodes the V2016 极值数据 sub-record (TLV type 0x06).
// Field order and converters mirror Java ExtremumDataCodec exactly:
//
//	VoltageMaxSubsystem(u8)   VoltageMaxBattery(u8)   MaxVoltage(u16 ExtremumVoltageConverter)
//	VoltageMinSubsystem(u8)   VoltageMinBattery(u8)   MinVoltage(u16 ExtremumVoltageConverter)
//	TemperatureMaxSubsystem(u8) TemperatureMaxProbe(u8) MaxTemperature(u8 TemperatureConverter)
//	TemperatureMinSubsystem(u8) TemperatureMinProbe(u8) MinTemperature(u8 TemperatureConverter)
//
// Temperature raw is u8 with ErrByte1 sentinels (handled by TemperatureConverter).
type ExtremumDataCodec struct{}

func init() {
	api.Register[mdl.ExtremumData](api.V2016, &ExtremumDataCodec{})
}

func (c *ExtremumDataCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.ExtremumData{}
	m.VoltageMaxSubsystem = int(r.ReadUint8())
	m.VoltageMaxBattery = int(r.ReadUint8())
	m.MaxVoltage = codec.ExtremumVoltageConverter.Decode(int64(r.ReadUint16()))

	m.VoltageMinSubsystem = int(r.ReadUint8())
	m.VoltageMinBattery = int(r.ReadUint8())
	m.MinVoltage = codec.ExtremumVoltageConverter.Decode(int64(r.ReadUint16()))

	m.TemperatureMaxSubsystem = int(r.ReadUint8())
	m.TemperatureMaxProbe = int(r.ReadUint8())
	m.MaxTemperature = codec.TemperatureConverter.Decode(int64(r.ReadUint8()))

	m.TemperatureMinSubsystem = int(r.ReadUint8())
	m.TemperatureMinProbe = int(r.ReadUint8())
	m.MinTemperature = codec.TemperatureConverter.Decode(int64(r.ReadUint8()))

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *ExtremumDataCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.ExtremumData)
	w.WriteUint8(byte(m.VoltageMaxSubsystem))
	w.WriteUint8(byte(m.VoltageMaxBattery))
	w.WriteUint16(uint16(codec.ExtremumVoltageConverter.Encode(m.MaxVoltage)))

	w.WriteUint8(byte(m.VoltageMinSubsystem))
	w.WriteUint8(byte(m.VoltageMinBattery))
	w.WriteUint16(uint16(codec.ExtremumVoltageConverter.Encode(m.MinVoltage)))

	w.WriteUint8(byte(m.TemperatureMaxSubsystem))
	w.WriteUint8(byte(m.TemperatureMaxProbe))
	w.WriteUint8(byte(codec.TemperatureConverter.Encode(m.MaxTemperature)))

	w.WriteUint8(byte(m.TemperatureMinSubsystem))
	w.WriteUint8(byte(m.TemperatureMinProbe))
	w.WriteUint8(byte(codec.TemperatureConverter.Encode(m.MinTemperature)))
	return nil
}

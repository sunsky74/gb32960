package realtime

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
)

// MotorDataCodec encodes/decodes one V2016 驱动电机 entry (TLV type 0x02 element).
// Field order and converters mirror Java MotorDataCodec:
//
//	MotorSeq(u8) MotorState(u8 raw byte)
//	ControllerTemperature(u8 ControllerTempConverter)
//	MotorSpeed(u16 MotorSpeedConverter2016)
//	MotorTorque(u16 MotorTorqueConverter2016)
//	MotorTemperature(u8 MotorTempConverter)
//	ControllerVoltage(u16 ControllerVoltageConverter)
//	ControllerCurrent(u16 ControllerCurrentConverter)
type MotorDataCodec struct{}

func init() {
	api.Register[mdl.MotorData](api.V2016, &MotorDataCodec{})
}

func (c *MotorDataCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.MotorData{}
	m.MotorSeq = int(r.ReadUint8())
	m.MotorState = r.ReadUint8()
	m.ControllerTemperature = codec.ControllerTempConverter.Decode(int64(r.ReadUint8()))
	m.MotorSpeed = codec.MotorSpeedConverter2016.Decode(int64(r.ReadUint16()))
	m.MotorTorque = codec.MotorTorqueConverter2016.Decode(int64(r.ReadUint16()))
	m.MotorTemperature = codec.MotorTempConverter.Decode(int64(r.ReadUint8()))
	m.ControllerVoltage = codec.ControllerVoltageConverter.Decode(int64(r.ReadUint16()))
	m.ControllerCurrent = codec.ControllerCurrentConverter.Decode(int64(r.ReadUint16()))

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *MotorDataCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.MotorData)
	w.WriteUint8(byte(m.MotorSeq))
	w.WriteUint8(m.MotorState)
	w.WriteUint8(byte(codec.ControllerTempConverter.Encode(m.ControllerTemperature)))
	w.WriteUint16(uint16(codec.MotorSpeedConverter2016.Encode(m.MotorSpeed)))
	w.WriteUint16(uint16(codec.MotorTorqueConverter2016.Encode(m.MotorTorque)))
	w.WriteUint8(byte(codec.MotorTempConverter.Encode(m.MotorTemperature)))
	w.WriteUint16(uint16(codec.ControllerVoltageConverter.Encode(m.ControllerVoltage)))
	w.WriteUint16(uint16(codec.ControllerCurrentConverter.Encode(m.ControllerCurrent)))
	return nil
}

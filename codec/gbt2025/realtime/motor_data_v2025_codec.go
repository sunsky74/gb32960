package realtime

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// MotorDataV2025Codec encodes/decodes one V2025 驱动电机 entry (TLV type 0x02
// element). Compared to V2016 MotorData, REMOVES ControllerVoltage and
// ControllerCurrent. Uses V2025-specific converters (MotorSpeed offset=32000,
// MotorTorque offset=20000 with ErrByte4).
//
// Field order and converters mirror Java MotorDataV2025Codec exactly:
//
//	MotorSeq(u8) MotorState(u8 raw byte)
//	ControllerTemperature(u8 ControllerTempConverter)
//	MotorSpeed(u16 MotorSpeedConverter2025)
//	MotorTorque(u32 MotorTorqueConverter2025)
//	MotorTemperature(u8 MotorTempConverter)
type MotorDataV2025Codec struct{}

func init() {
	api.Register[mdl.MotorDataV2025](api.V2025, &MotorDataV2025Codec{})
}

func (c *MotorDataV2025Codec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.MotorDataV2025{}
	m.MotorSeq = int(r.ReadUint8())
	m.MotorState = r.ReadUint8()
	m.ControllerTemperature = codec.ControllerTempConverter.Decode(int64(r.ReadUint8()))
	m.MotorSpeed = codec.MotorSpeedConverter2025.Decode(int64(r.ReadUint16()))
	m.MotorTorque = codec.MotorTorqueConverter2025.Decode(int64(r.ReadUint32()))
	m.MotorTemperature = codec.MotorTempConverter.Decode(int64(r.ReadUint8()))

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *MotorDataV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.MotorDataV2025)
	w.WriteUint8(byte(m.MotorSeq))
	w.WriteUint8(m.MotorState)
	w.WriteUint8(byte(codec.ControllerTempConverter.Encode(m.ControllerTemperature)))
	w.WriteUint16(uint16(codec.MotorSpeedConverter2025.Encode(m.MotorSpeed)))
	w.WriteUint32(uint32(codec.MotorTorqueConverter2025.Encode(m.MotorTorque)))
	w.WriteUint8(byte(codec.MotorTempConverter.Encode(m.MotorTemperature)))
	return nil
}

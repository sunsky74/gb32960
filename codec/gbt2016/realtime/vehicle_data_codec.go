package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// VehicleDataCodec 编解码 V2016 整车数据子记录(TLV 类型 0x01)。
// 字段顺序与转换器与 Java VehicleDataCodec 完全一致:
//
//	OperatingState(u8) ChargingState(u8) OperationMode(u8)
//	Speed(u16 SpeedConverter) Mileage(u32 MileageConverter)
//	Voltage(u16 VoltageConverter) Current(u16 CurrentConverter2016)
//	SOC(u8) DC(u8) GearPosition(1B,经注册的编解码器)
//	Insulance(u16) AccelerationValue(u8) BrakePedalCondition(u8)
type VehicleDataCodec struct{}

func init() {
	api.Register[mdl.VehicleData](api.V2016, &VehicleDataCodec{})
}

func (c *VehicleDataCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.VehicleData{}
	m.OperatingState = types.OperatingState(r.ReadUint8())
	m.ChargingState = types.ChargingState(r.ReadUint8())
	m.OperationMode = types.OperationMode(r.ReadUint8())
	m.Speed = codec.SpeedConverter.Decode(int64(r.ReadUint16()))
	m.Mileage = codec.MileageConverter.Decode(int64(r.ReadUint32()))
	m.Voltage = codec.VoltageConverter.Decode(int64(r.ReadUint16()))
	m.Current = codec.CurrentConverter2016.Decode(int64(r.ReadUint16()))
	m.SOC = int(r.ReadUint8())
	m.DC = types.DCState(r.ReadUint8())

	gpCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.GearPosition)(nil)).Elem())
	if gpCodec == nil {
		return nil, api.ErrCodecNotFound
	}
	gp, err := gpCodec.Decode(r)
	if err != nil {
		return nil, err
	}
	m.GearPosition = *gp.(*mdl.GearPosition)

	m.Insulance = int(r.ReadUint16())
	m.AccelerationValue = int(r.ReadUint8())
	m.BrakePedalCondition = int(r.ReadUint8())

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *VehicleDataCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.VehicleData)
	w.WriteUint8(byte(m.OperatingState))
	w.WriteUint8(byte(m.ChargingState))
	w.WriteUint8(byte(m.OperationMode))
	w.WriteUint16(uint16(codec.SpeedConverter.Encode(m.Speed)))
	w.WriteUint32(uint32(codec.MileageConverter.Encode(m.Mileage)))
	w.WriteUint16(uint16(codec.VoltageConverter.Encode(m.Voltage)))
	w.WriteUint16(uint16(codec.CurrentConverter2016.Encode(m.Current)))
	w.WriteUint8(byte(m.SOC))
	w.WriteUint8(byte(m.DC))

	gpCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.GearPosition)(nil)).Elem())
	if gpCodec == nil {
		return api.ErrCodecNotFound
	}
	if err := gpCodec.Encode(w, &m.GearPosition); err != nil {
		return err
	}

	w.WriteUint16(uint16(m.Insulance))
	w.WriteUint8(byte(m.AccelerationValue))
	w.WriteUint8(byte(m.BrakePedalCondition))
	return nil
}

package realtime

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// SuperCapacitorExtremumCodec 编解码 V2025 超级电容极值数据
// 子记录(TLV 类型 0x32)。字段顺序与转换器与 Java
// SuperCapacitorExtremumCodec 完全一致:
//
//	VoltageMaxSubsystem(u8)   VoltageMaxBattery(u16)   MaxVoltage(u16 SuperCapExtremumVoltageConverter scale=0.001)
//	VoltageMinSubsystem(u8)   VoltageMinBattery(u16)   MinVoltage(u16 SuperCapExtremumVoltageConverter scale=0.001)
//	TemperatureMaxSubsystem(u8) TemperatureMaxProbe(u16) MaxTemperature(u8 TemperatureConverter offset=40)
//	TemperatureMinSubsystem(u8) TemperatureMinProbe(u16) MinTemperature(u8 TemperatureConverter offset=40)
//
// 注意:探针/子系统索引是 u8 + u16(不像 V2016
// ExtremumData 那样是 u8 + u8)。与 Java 的字段宽度一致。
type SuperCapacitorExtremumCodec struct{}

func init() {
	api.Register[mdl.SuperCapacitorExtremumData](api.V2025, &SuperCapacitorExtremumCodec{})
}

func (c *SuperCapacitorExtremumCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.SuperCapacitorExtremumData{}
	m.VoltageMaxSubsystem = int(r.ReadUint8())
	m.VoltageMaxBattery = int(r.ReadUint16())
	m.MaxVoltage = codec.SuperCapExtremumVoltageConverter.Decode(int64(r.ReadUint16()))

	m.VoltageMinSubsystem = int(r.ReadUint8())
	m.VoltageMinBattery = int(r.ReadUint16())
	m.MinVoltage = codec.SuperCapExtremumVoltageConverter.Decode(int64(r.ReadUint16()))

	m.TemperatureMaxSubsystem = int(r.ReadUint8())
	m.TemperatureMaxProbe = int(r.ReadUint16())
	m.MaxTemperature = codec.TemperatureConverter.Decode(int64(r.ReadUint8()))

	m.TemperatureMinSubsystem = int(r.ReadUint8())
	m.TemperatureMinProbe = int(r.ReadUint16())
	m.MinTemperature = codec.TemperatureConverter.Decode(int64(r.ReadUint8()))

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *SuperCapacitorExtremumCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.SuperCapacitorExtremumData)
	w.WriteUint8(byte(m.VoltageMaxSubsystem))
	w.WriteUint16(uint16(m.VoltageMaxBattery))
	w.WriteUint16(uint16(codec.SuperCapExtremumVoltageConverter.Encode(m.MaxVoltage)))

	w.WriteUint8(byte(m.VoltageMinSubsystem))
	w.WriteUint16(uint16(m.VoltageMinBattery))
	w.WriteUint16(uint16(codec.SuperCapExtremumVoltageConverter.Encode(m.MinVoltage)))

	w.WriteUint8(byte(m.TemperatureMaxSubsystem))
	w.WriteUint16(uint16(m.TemperatureMaxProbe))
	w.WriteUint8(byte(codec.TemperatureConverter.Encode(m.MaxTemperature)))

	w.WriteUint8(byte(m.TemperatureMinSubsystem))
	w.WriteUint16(uint16(m.TemperatureMinProbe))
	w.WriteUint8(byte(codec.TemperatureConverter.Encode(m.MinTemperature)))
	return nil
}

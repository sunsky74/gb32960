package realtime

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
)

// ChargeableSubsystemElectricCodec encodes/decodes one V2016 可充电储能装置电压
// entry (TLV type 0x08 element). Field order and converters mirror Java
// ChargeableSubsystemElectricCodec exactly:
//
//	ChargeableSubSystemNumber(u8)
//	Voltage(u16 VoltageConverter)            // scale=10
//	Current(u16 CurrentConverterChargeElectric) // scale=10, offset=-1000
//	BatteryTotalCount(u16)
//	FrameStartBatterySeq(u16)
//	BatteryCount(u8)
//	BatteryVoltages[BatteryCount × u16 BatteryVoltageConverter]  // scale=1000
//
// NOTE: task brief's summary "subSystemNo(u8), batteryCount(u16),
// voltages[batteryCount × u16]" omitted the Voltage/Current/TotalCount/
// FrameStartBatterySeq fields. We follow the actual Go model struct and
// Java codec — see model/gbt2016/realtime/chargeable_subsystem_electric.go.
type ChargeableSubsystemElectricCodec struct{}

func init() {
	api.Register[mdl.ChargeableSubsystemElectric](api.V2016, &ChargeableSubsystemElectricCodec{})
}

func (c *ChargeableSubsystemElectricCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.ChargeableSubsystemElectric{}
	m.ChargeableSubSystemNumber = int(r.ReadUint8())
	m.Voltage = codec.VoltageConverter.Decode(int64(r.ReadUint16()))
	m.Current = codec.CurrentConverterChargeElectric.Decode(int64(r.ReadUint16()))
	m.BatteryTotalCount = int(r.ReadUint16())
	m.FrameStartBatterySeq = int(r.ReadUint16())
	m.BatteryCount = int(r.ReadUint8())
	if m.BatteryCount > 0 {
		m.BatteryVoltages = make([]float64, m.BatteryCount)
		for i := 0; i < m.BatteryCount; i++ {
			m.BatteryVoltages[i] = codec.BatteryVoltageConverter.Decode(int64(r.ReadUint16()))
		}
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *ChargeableSubsystemElectricCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.ChargeableSubsystemElectric)
	w.WriteUint8(byte(m.ChargeableSubSystemNumber))
	w.WriteUint16(uint16(codec.VoltageConverter.Encode(m.Voltage)))
	w.WriteUint16(uint16(codec.CurrentConverterChargeElectric.Encode(m.Current)))
	w.WriteUint16(uint16(m.BatteryTotalCount))
	w.WriteUint16(uint16(m.FrameStartBatterySeq))
	w.WriteUint8(byte(m.BatteryCount))
	for _, v := range m.BatteryVoltages {
		w.WriteUint16(uint16(codec.BatteryVoltageConverter.Encode(v)))
	}
	return nil
}

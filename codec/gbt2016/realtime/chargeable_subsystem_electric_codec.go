package realtime

import (
	"fmt"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
)

// ChargeableSubsystemElectricCodec 编解码一个 V2016 可充电储能装置电压
// 条目(TLV 类型 0x08 元素)。字段顺序与转换器与 Java
// ChargeableSubsystemElectricCodec 完全一致:
//
//	ChargeableSubSystemNumber(u8)
//	Voltage(u16 VoltageConverter)            // scale=10
//	Current(u16 CurrentConverterChargeElectric) // scale=10, offset=+1000(国标表B.6;本地 Java 副本仍为 -1000,勿同步)
//	BatteryTotalCount(u16)
//	FrameStartBatterySeq(u16)
//	BatteryCount(u8)
//	BatteryVoltages[BatteryCount × u16 BatteryVoltageConverter]  // scale=1000
//
// 注意:任务简报的摘要 "subSystemNo(u8), batteryCount(u16),
// voltages[batteryCount × u16]" 遗漏了 Voltage/Current/TotalCount/
// FrameStartBatterySeq 字段。我们以实际的 Go 模型结构体与
// Java 编解码器为准 —— 见 model/gbt2016/realtime/chargeable_subsystem_electric.go。
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
	// audit 2026-09-17 (M4):在写出任何内容之前拒绝 BatteryCount/列表不匹配
	// —— 此前线上计数与写出的电压列表可能不一致,
	// 产生畸形的帧。
	if m.BatteryCount != len(m.BatteryVoltages) {
		return fmt.Errorf("gb32960: battery count %d does not match battery voltages length %d", m.BatteryCount, len(m.BatteryVoltages))
	}
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

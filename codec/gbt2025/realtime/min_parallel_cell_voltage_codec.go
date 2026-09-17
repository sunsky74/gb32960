package realtime

import (
	"fmt"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// MinParallelCellVoltageCodec 编解码一条 V2025 动力蓄电池最小并联单元
// 电压条目(TLV 类型 0x07 元素)。字段顺序与转换器与 Java
// MinParallelCellVoltageCodec 完全一致:
//
//	BatteryPackSeq(u8)
//	Voltage(u16 VoltageConverter scale=0.1)
//	Current(u16 CurrentConverter2025 offset=3000 scale=0.1)
//	MinParallelUnits(u16)
//	MinParallelUnits × BatteryVoltage(u16 BatteryVoltagesConverter2025 scale=0.001)
//
// 当 MinParallelUnits 为 BYTE2 错误哨兵值时,逐单元电压
// 列表被跳过(Java: if units > 0 && !BATTERY_VOLTAGES_CONVERTER.isInvalid)。
type MinParallelCellVoltageCodec struct{}

func init() {
	api.Register[mdl.MinParallelCellVoltage](api.V2025, &MinParallelCellVoltageCodec{})
}

func (c *MinParallelCellVoltageCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.MinParallelCellVoltage{}
	m.BatteryPackSeq = int(r.ReadUint8())
	m.Voltage = codec.VoltageConverter.Decode(int64(r.ReadUint16()))
	m.Current = codec.CurrentConverter2025.Decode(int64(r.ReadUint16()))
	m.MinParallelUnits = int(r.ReadUint16())

	if m.MinParallelUnits > 0 && !types.ErrByte2.IsInvalid(int64(m.MinParallelUnits)) {
		m.BatteryVoltages = make([]float64, m.MinParallelUnits)
		for i := 0; i < m.MinParallelUnits; i++ {
			m.BatteryVoltages[i] = codec.BatteryVoltagesConverter2025.Decode(int64(r.ReadUint16()))
		}
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *MinParallelCellVoltageCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.MinParallelCellVoltage)
	// fix 2026-09-17: 表12 L210-211 —— 最小并联单元总数哨兵值
	// (0xFFFE 异常/0xFFFF 无效)不携带任何逐单元电压,列表必须为空;
	// 普通计数必须与逐单元电压个数完全一致,否则拒绝编码
	// (此前哨兵值会静默丢弃非空电压、普通计数不校验长度)。
	if types.ErrByte2.IsInvalid(int64(m.MinParallelUnits)) {
		if len(m.BatteryVoltages) != 0 {
			return fmt.Errorf("gb32960: min parallel units %d is a sentinel but %d battery voltages present", m.MinParallelUnits, len(m.BatteryVoltages))
		}
	} else if m.MinParallelUnits != len(m.BatteryVoltages) {
		return fmt.Errorf("gb32960: min parallel units %d does not match battery voltages length %d", m.MinParallelUnits, len(m.BatteryVoltages))
	}
	w.WriteUint8(byte(m.BatteryPackSeq))
	w.WriteUint16(uint16(codec.VoltageConverter.Encode(m.Voltage)))
	w.WriteUint16(uint16(codec.CurrentConverter2025.Encode(m.Current)))
	w.WriteUint16(uint16(m.MinParallelUnits))
	for _, v := range m.BatteryVoltages {
		w.WriteUint16(uint16(codec.BatteryVoltagesConverter2025.Encode(v)))
	}
	return nil
}

package realtime

import (
	"fmt"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// SuperCapacitorDataCodec 编解码 V2025 超级电容数据子记录
// (TLV 类型 0x31)。字段顺序与转换器与 Java SuperCapacitorDataCodec 一致:
//
//	ManagementSystemNumber(u8)
//	TotalVoltage(u16 TotalVoltageConverter scale=0.1)
//	TotalCurrent(u16 TotalCurrentConverter offset=3000 scale=0.1)
//	CapacitorCount(u16)
//	CapacitorCount × CapacitorVoltage(u16 SuperCapVoltageConverter scale=0.001)
//	TemperatureProbeCount(u16)
//	TemperatureProbeCount × ProbeTemperature(u8 TemperatureConverter offset=40)
//
// 当 CapacitorCount/TemperatureProbeCount 为 BYTE2 错误哨兵值时,
// 对应列表被跳过(Java 设为 Collections.emptyList() 并继续)。
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
	// fix 2026-09-17: GB/T 32960.3-2025 表25(L407/L409) —— 超级电容单体总数 M
	// 与温度探针总数 N 均为 WORD:哨兵值(0xFFFE 异常 / 0xFFFF 无效)不携带
	// 对应列表,普通计数必须与列表长度一致。旧代码在计数不匹配时静默截断/超写。
	// 校验先于写出。
	if types.ErrByte2.IsInvalid(int64(m.CapacitorCount)) {
		if len(m.CapacitorVoltages) != 0 {
			return fmt.Errorf("gb32960: super capacitor count %d is a sentinel but %d voltages present", m.CapacitorCount, len(m.CapacitorVoltages))
		}
	} else if m.CapacitorCount != len(m.CapacitorVoltages) {
		return fmt.Errorf("gb32960: super capacitor count %d does not match voltages length %d", m.CapacitorCount, len(m.CapacitorVoltages))
	}
	if types.ErrByte2.IsInvalid(int64(m.TemperatureProbeCount)) {
		if len(m.ProbeTemperatures) != 0 {
			return fmt.Errorf("gb32960: super capacitor probe count %d is a sentinel but %d temps present", m.TemperatureProbeCount, len(m.ProbeTemperatures))
		}
	} else if m.TemperatureProbeCount != len(m.ProbeTemperatures) {
		return fmt.Errorf("gb32960: super capacitor probe count %d does not match temps length %d", m.TemperatureProbeCount, len(m.ProbeTemperatures))
	}

	w.WriteUint8(byte(m.ManagementSystemNumber))
	w.WriteUint16(uint16(codec.TotalVoltageConverter.Encode(m.TotalVoltage)))
	w.WriteUint16(uint16(codec.TotalCurrentConverter.Encode(m.TotalCurrent)))
	w.WriteUint16(uint16(m.CapacitorCount))
	for _, v := range m.CapacitorVoltages {
		w.WriteUint16(uint16(codec.SuperCapVoltageConverter.Encode(v)))
	}
	w.WriteUint16(uint16(m.TemperatureProbeCount))
	for _, t := range m.ProbeTemperatures {
		w.WriteUint8(byte(codec.TemperatureConverter.Encode(t)))
	}
	return nil
}

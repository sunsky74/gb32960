package realtime

import (
	"fmt"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// ChargeableSubsystemTemperatureCodec 编解码一个 V2016 可充电储能装置温度
// 条目(TLV 类型 0x09 元素)。字段顺序与转换器与 Java
// ChargeableSubsystemTemperatureCodec 完全一致:
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
	// audit 2026-09-17 (M2):WORD 计数哨兵值(0xFFFE 异常 / 0xFFFF 无效,
	// GB/T 32960-2016 表 12/表 B.8)不携带任何探针数据单元 —— 在哨兵值之后
	// 读取 N 字节会下溢,并使之后每个字段错位。
	if m.TemperatureProbeCount > 0 && !types.ErrByte2.IsInvalid(int64(m.TemperatureProbeCount)) {
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
	// audit 2026-09-17 (M2, M4):哨兵计数不携带任何探针字节
	//(列表必须为空);普通计数必须与列表长度完全匹配。
	// 两条路径上计数都按原样写出。
	if types.ErrByte2.IsInvalid(int64(m.TemperatureProbeCount)) {
		if len(m.ProbeTemperatures) != 0 {
			return fmt.Errorf("gb32960: temperature probe count %d is a sentinel but %d probe values present", m.TemperatureProbeCount, len(m.ProbeTemperatures))
		}
	} else if m.TemperatureProbeCount != len(m.ProbeTemperatures) {
		return fmt.Errorf("gb32960: temperature probe count %d does not match probe values length %d", m.TemperatureProbeCount, len(m.ProbeTemperatures))
	}
	w.WriteUint8(byte(m.SubSystemNumber))
	w.WriteUint16(uint16(m.TemperatureProbeCount))
	for _, t := range m.ProbeTemperatures {
		w.WriteUint8(byte(codec.TemperatureConverter.Encode(t)))
	}
	return nil
}

package realtime

import (
	"fmt"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// BatteryTempCodec 编解码一条 V2025 动力蓄电池温度条目
// (TLV 类型 0x08 元素)。从 Java BatteryPackTemperature 改名而来,以匹配
// TLV 枚举常量 BATTERY_TEMP。字段顺序与转换器与 Java
// BatteryPackTemperatureCodec 完全一致:
//
//	BatteryPackSeq(u8)
//	TemperatureProbeCount(u16)
//	TemperatureProbeCount × ProbeTemperature(u8 TemperatureConverter offset=40)
//
// 当 TemperatureProbeCount 为 BYTE2 错误哨兵值时,探针
// 温度列表被跳过(Java 提前返回空列表)。
type BatteryTempCodec struct{}

func init() {
	api.Register[mdl.BatteryTemp](api.V2025, &BatteryTempCodec{})
}

func (c *BatteryTempCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.BatteryTemp{}
	m.BatteryPackSeq = int(r.ReadUint8())
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

func (c *BatteryTempCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.BatteryTemp)

	// fix 2026-09-17: 表14 L228-229 —— 探针个数哨兵值(0xFFFE 异常/0xFFFF 无效)
	// 不携带任何温度单元,必须原样写出而不是改写成 0xFFFF;
	// 普通计数必须与温度值个数完全一致,否则拒绝编码。
	if types.ErrByte2.IsInvalid(int64(m.TemperatureProbeCount)) {
		if len(m.ProbeTemperatures) != 0 {
			return fmt.Errorf("gb32960: temperature probe count %d is a sentinel but %d probe values present", m.TemperatureProbeCount, len(m.ProbeTemperatures))
		}
	} else if m.TemperatureProbeCount != len(m.ProbeTemperatures) {
		return fmt.Errorf("gb32960: temperature probe count %d does not match probe values length %d", m.TemperatureProbeCount, len(m.ProbeTemperatures))
	}

	// fix 2026-09-17: 表14 L227 —— 包号 0xFE/0xFF 哨兵值原样写出;
	// 仅对 1~50 有效范围之外的非哨兵值(0、负数、51~253)保留防御性钳制。
	if types.ErrByte1.IsInvalid(int64(m.BatteryPackSeq)) {
		w.WriteUint8(byte(m.BatteryPackSeq))
	} else if m.BatteryPackSeq >= 1 && m.BatteryPackSeq <= 50 {
		w.WriteUint8(byte(m.BatteryPackSeq))
	} else {
		w.WriteUint8(byte(types.ErrByte1.Invalid))
	}

	w.WriteUint16(uint16(m.TemperatureProbeCount))
	for _, t := range m.ProbeTemperatures {
		w.WriteUint8(byte(codec.TemperatureConverter.Encode(t)))
	}
	return nil
}

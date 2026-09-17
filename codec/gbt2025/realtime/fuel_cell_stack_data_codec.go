package realtime

import (
	"fmt"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// FuelCellStackDataCodec 编解码一条 V2025 燃料电池电堆条目
// (TLV 类型 0x30 元素)。字段顺序与转换器与 Java
// FuelCellStackDataCodec 完全一致:
//
//	StackSeq(u8)
//	Voltage(u16 FuelCellVoltageConverter scale=0.1)
//	Current(u16 FuelCellCurrentConverter scale=0.1)
//	GasPressure(u16 GasPressureConverter offset=100 scale=0.1)
//	AirPressure(u16 AirPressureConverter offset=100 scale=0.1)
//	AirInletTemp(u8 ControllerTempConverter offset=40)
//	CoolingWaterProbeCount(u16)
//	CoolingWaterProbeCount × CoolingWaterTemp(u8 TemperatureConverter offset=40)
//
// 当 CoolingWaterProbeCount 为 BYTE2 错误哨兵值时,逐探针
// 列表被跳过(Java 提前返回空列表)。
type FuelCellStackDataCodec struct{}

func init() {
	api.Register[mdl.FuelCellStackData](api.V2025, &FuelCellStackDataCodec{})
}

func (c *FuelCellStackDataCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.FuelCellStackData{}
	m.StackSeq = int(r.ReadUint8())
	m.Voltage = codec.FuelCellVoltageConverter.Decode(int64(r.ReadUint16()))
	m.Current = codec.FuelCellCurrentConverter.Decode(int64(r.ReadUint16()))
	m.GasPressure = codec.GasPressureConverter.Decode(int64(r.ReadUint16()))
	m.AirPressure = codec.AirPressureConverter.Decode(int64(r.ReadUint16()))
	m.AirInletTemp = codec.ControllerTempConverter.Decode(int64(r.ReadUint8()))
	m.CoolingWaterProbeCount = int(r.ReadUint16())

	if !types.ErrByte2.IsInvalid(int64(m.CoolingWaterProbeCount)) && m.CoolingWaterProbeCount > 0 {
		m.CoolingWaterTemps = make([]float64, m.CoolingWaterProbeCount)
		for i := 0; i < m.CoolingWaterProbeCount; i++ {
			m.CoolingWaterTemps[i] = codec.TemperatureConverter.Decode(int64(r.ReadUint8()))
		}
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *FuelCellStackDataCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.FuelCellStackData)
	// fix 2026-09-17: GB/T 32960.3-2025 表19(L299) —— 冷却水出水口温度探针总数
	// WORD 计数哨兵值(0xFFFE 异常 / 0xFFFF 无效)不携带温度列表;普通计数必须
	// 与列表长度完全匹配。旧代码在哨兵携带温度时静默丢弃、在普通计数不匹配时
	// 静默截断/超写。校验先于写出(镜像 2016 FuelCellDataCodec 的 M2/M4 校验)。
	if types.ErrByte2.IsInvalid(int64(m.CoolingWaterProbeCount)) {
		if len(m.CoolingWaterTemps) != 0 {
			return fmt.Errorf("gb32960: fuel cell stack probe count %d is a sentinel but %d temps present", m.CoolingWaterProbeCount, len(m.CoolingWaterTemps))
		}
	} else if m.CoolingWaterProbeCount != len(m.CoolingWaterTemps) {
		return fmt.Errorf("gb32960: fuel cell stack probe count %d does not match temps length %d", m.CoolingWaterProbeCount, len(m.CoolingWaterTemps))
	}
	w.WriteUint8(byte(m.StackSeq))
	w.WriteUint16(uint16(codec.FuelCellVoltageConverter.Encode(m.Voltage)))
	w.WriteUint16(uint16(codec.FuelCellCurrentConverter.Encode(m.Current)))
	w.WriteUint16(uint16(codec.GasPressureConverter.Encode(m.GasPressure)))
	w.WriteUint16(uint16(codec.AirPressureConverter.Encode(m.AirPressure)))
	w.WriteUint8(byte(codec.ControllerTempConverter.Encode(m.AirInletTemp)))
	w.WriteUint16(uint16(m.CoolingWaterProbeCount))
	for _, t := range m.CoolingWaterTemps {
		w.WriteUint8(byte(codec.TemperatureConverter.Encode(t)))
	}
	return nil
}

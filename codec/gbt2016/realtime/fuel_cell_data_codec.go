package realtime

import (
	"fmt"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// FuelCellDataCodec 编解码 V2016 燃料电池数据子记录
// (TLV 类型 0x03)。字段顺序与转换器与 Java FuelCellDataCodec 一致:
//
//	FuelCellVoltage(u16 FuelCellVoltageConverter)
//	FuelCellCurrent(u16 FuelCellCurrentConverter)
//	FuelConsumptionRate(u16 FuelConsumptionRateConverter)
//	TotalNumberOfFcTp(u16)
//	ProbeTemperatureValues[N × u8 ProbeTemperatureConverter]
//	HighestTempOfHydrogenSystem(u16 HighestTempHydrogenConverter)
//	HighestTempProbeCodeOfHydrogenSystem(u8)
//	HighestConOfHydrogen(u16)
//	HighestHyConSensorCode(u8)
//	HydrogenMaxPressure(u16 HydrogenMaxPressureConverter)
//	HydrogenMaxPressureSensorCode(u8)
//	HighVoltageDCState(u8 原始字节)
type FuelCellDataCodec struct{}

func init() {
	api.Register[mdl.FuelCellData](api.V2016, &FuelCellDataCodec{})
}

func (c *FuelCellDataCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.FuelCellData{}
	m.FuelCellVoltage = codec.FuelCellVoltageConverter.Decode(int64(r.ReadUint16()))
	m.FuelCellCurrent = codec.FuelCellCurrentConverter.Decode(int64(r.ReadUint16()))
	m.FuelConsumptionRate = codec.FuelConsumptionRateConverter.Decode(int64(r.ReadUint16()))

	m.TotalNumberOfFcTp = int(r.ReadUint16())
	// audit 2026-09-17 (M2):WORD 计数哨兵值(0xFFFE 异常 / 0xFFFF 无效,
	// GB/T 32960-2016 表 12/表 B.8)不携带任何探针数据单元 —— 在哨兵值之后
	// 读取 N 字节会下溢,并使之后每个字段错位。
	if m.TotalNumberOfFcTp > 0 && !types.ErrByte2.IsInvalid(int64(m.TotalNumberOfFcTp)) {
		m.ProbeTemperatureValues = make([]float64, m.TotalNumberOfFcTp)
		for i := 0; i < m.TotalNumberOfFcTp; i++ {
			m.ProbeTemperatureValues[i] = codec.ProbeTemperatureConverter.Decode(int64(r.ReadUint8()))
		}
	}

	m.HighestTempOfHydrogenSystem = codec.HighestTempHydrogenConverter.Decode(int64(r.ReadUint16()))
	m.HighestTempProbeCodeOfHydrogenSystem = int(r.ReadUint8())
	m.HighestConOfHydrogen = int(r.ReadUint16())
	m.HighestHyConSensorCode = int(r.ReadUint8())
	m.HydrogenMaxPressure = codec.HydrogenMaxPressureConverter.Decode(int64(r.ReadUint16()))
	m.HydrogenMaxPressureSensorCode = int(r.ReadUint8())
	m.HighVoltageDCState = r.ReadUint8()

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *FuelCellDataCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.FuelCellData)
	// audit 2026-09-17 (M2, M4):哨兵计数不携带任何探针字节
	//(列表必须为空);普通计数必须与列表长度完全匹配。
	// 两条路径上计数都按原样写出。
	if types.ErrByte2.IsInvalid(int64(m.TotalNumberOfFcTp)) {
		if len(m.ProbeTemperatureValues) != 0 {
			return fmt.Errorf("gb32960: fuel cell probe count %d is a sentinel but %d probe values present", m.TotalNumberOfFcTp, len(m.ProbeTemperatureValues))
		}
	} else if m.TotalNumberOfFcTp != len(m.ProbeTemperatureValues) {
		return fmt.Errorf("gb32960: fuel cell probe count %d does not match probe values length %d", m.TotalNumberOfFcTp, len(m.ProbeTemperatureValues))
	}
	w.WriteUint16(uint16(codec.FuelCellVoltageConverter.Encode(m.FuelCellVoltage)))
	w.WriteUint16(uint16(codec.FuelCellCurrentConverter.Encode(m.FuelCellCurrent)))
	w.WriteUint16(uint16(codec.FuelConsumptionRateConverter.Encode(m.FuelConsumptionRate)))

	w.WriteUint16(uint16(m.TotalNumberOfFcTp))
	if len(m.ProbeTemperatureValues) > 0 {
		for _, t := range m.ProbeTemperatureValues {
			w.WriteUint8(byte(codec.ProbeTemperatureConverter.Encode(t)))
		}
	}

	w.WriteUint16(uint16(codec.HighestTempHydrogenConverter.Encode(m.HighestTempOfHydrogenSystem)))
	w.WriteUint8(byte(m.HighestTempProbeCodeOfHydrogenSystem))
	w.WriteUint16(uint16(m.HighestConOfHydrogen))
	w.WriteUint8(byte(m.HighestHyConSensorCode))
	w.WriteUint16(uint16(codec.HydrogenMaxPressureConverter.Encode(m.HydrogenMaxPressure)))
	w.WriteUint8(byte(m.HydrogenMaxPressureSensorCode))
	w.WriteUint8(m.HighVoltageDCState)
	return nil
}

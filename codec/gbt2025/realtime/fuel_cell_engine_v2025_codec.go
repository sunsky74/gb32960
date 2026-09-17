package realtime

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// FuelCellEngineV2025Codec 编解码 V2025 燃料电池发动机及车载氢系统
// 数据子记录(TLV 类型 0x03)。与 V2016 FuelCellData 相比,新增了
// FuelPercentage 与 DCControllerTemperature,并移除了电压/电流/
// consumptionRate/探针列表。
//
// 字段顺序与转换器与 Java FuelCellEngineV2025Codec 完全一致:
//
//	HighestTempOfHydrogenSystem(u16 TempConverterEngine2025)
//	HighestTempProbeCodeOfHydrogenSystem(u8)
//	HighestConOfHydrogen(u16 ConcentrationConverter2025)
//	HighestHyConSensorCode(u8)
//	HydrogenMaxPressure(u16 MaxPressureConverter2025)
//	HydrogenMaxPressureSensorCode(u8)
//	HighVoltageDCState(u8 原始字节)
//	FuelPercentage(u8)
//	DCControllerTemperature(u8 ControllerTempConverter)
type FuelCellEngineV2025Codec struct{}

func init() {
	api.Register[mdl.FuelCellEngineV2025Data](api.V2025, &FuelCellEngineV2025Codec{})
}

func (c *FuelCellEngineV2025Codec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.FuelCellEngineV2025Data{}
	m.HighestTempOfHydrogenSystem = codec.TempConverterEngine2025.Decode(int64(r.ReadUint16()))
	m.HighestTempProbeCodeOfHydrogenSystem = int(r.ReadUint8())
	m.HighestConOfHydrogen = codec.ConcentrationConverter2025.Decode(int64(r.ReadUint16()))
	m.HighestHyConSensorCode = int(r.ReadUint8())
	m.HydrogenMaxPressure = codec.MaxPressureConverter2025.Decode(int64(r.ReadUint16()))
	m.HydrogenMaxPressureSensorCode = int(r.ReadUint8())
	m.HighVoltageDCState = r.ReadUint8()
	m.FuelPercentage = int(r.ReadUint8())
	m.DCControllerTemperature = codec.ControllerTempConverter.Decode(int64(r.ReadUint8()))

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *FuelCellEngineV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.FuelCellEngineV2025Data)
	w.WriteUint16(uint16(codec.TempConverterEngine2025.Encode(m.HighestTempOfHydrogenSystem)))
	w.WriteUint8(byte(m.HighestTempProbeCodeOfHydrogenSystem))
	w.WriteUint16(uint16(codec.ConcentrationConverter2025.Encode(m.HighestConOfHydrogen)))
	w.WriteUint8(byte(m.HighestHyConSensorCode))
	w.WriteUint16(uint16(codec.MaxPressureConverter2025.Encode(m.HydrogenMaxPressure)))
	w.WriteUint8(byte(m.HydrogenMaxPressureSensorCode))
	w.WriteUint8(m.HighVoltageDCState)
	w.WriteUint8(byte(m.FuelPercentage))
	w.WriteUint8(byte(codec.ControllerTempConverter.Encode(m.DCControllerTemperature)))
	return nil
}

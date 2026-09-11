package realtime

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// FuelCellStackDataCodec encodes/decodes one V2025 燃料电池电堆 entry
// (TLV type 0x30 element). Field order and converters mirror Java
// FuelCellStackDataCodec exactly:
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
// When CoolingWaterProbeCount is the BYTE2 error sentinel, the per-probe list
// is skipped (Java early-returns with empty list).
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
	w.WriteUint8(byte(m.StackSeq))
	w.WriteUint16(uint16(codec.FuelCellVoltageConverter.Encode(m.Voltage)))
	w.WriteUint16(uint16(codec.FuelCellCurrentConverter.Encode(m.Current)))
	w.WriteUint16(uint16(codec.GasPressureConverter.Encode(m.GasPressure)))
	w.WriteUint16(uint16(codec.AirPressureConverter.Encode(m.AirPressure)))
	w.WriteUint8(byte(codec.ControllerTempConverter.Encode(m.AirInletTemp)))
	w.WriteUint16(uint16(m.CoolingWaterProbeCount))

	if types.ErrByte2.IsInvalid(int64(m.CoolingWaterProbeCount)) {
		return nil
	}
	for _, t := range m.CoolingWaterTemps {
		w.WriteUint8(byte(codec.TemperatureConverter.Encode(t)))
	}
	return nil
}

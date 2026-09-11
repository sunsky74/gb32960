package realtime

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
)

// FuelCellDataCodec encodes/decodes the V2016 燃料电池数据 sub-record
// (TLV type 0x03). Field order and converters mirror Java FuelCellDataCodec:
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
//	HighVoltageDCState(u8 raw byte)
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
	if m.TotalNumberOfFcTp > 0 {
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

package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
)

// ChargeableSubsystemTemperatureListCodec encodes/decodes the V2016 可充电储能装置温度
// list (TLV type 0x09 body). Wire layout: TemperatureCount(u8) + TemperatureCount ×
// ChargeableSubsystemTemperature entries.
//
// DEVIATION FROM JAVA (reported): same count-width inconsistency as the
// electric list — Java reads 1 byte, writes 2 bytes. Go keeps both sides u8
// so roundtrip is byte-stable. Matches the Plan 2 list pattern.
type ChargeableSubsystemTemperatureListCodec struct{}

func init() {
	api.Register[mdl.ChargeableSubsystemTemperatureList](api.V2016, &ChargeableSubsystemTemperatureListCodec{})
}

func (c *ChargeableSubsystemTemperatureListCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.ChargeableSubsystemTemperatureList{}
	m.TemperatureCount = int(r.ReadUint8())
	if m.TemperatureCount > 0 {
		elemCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.ChargeableSubsystemTemperature)(nil)).Elem())
		if elemCodec == nil {
			return nil, api.ErrCodecNotFound
		}
		m.Items = make([]mdl.ChargeableSubsystemTemperature, m.TemperatureCount)
		for i := 0; i < m.TemperatureCount; i++ {
			item, err := elemCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			m.Items[i] = *item.(*mdl.ChargeableSubsystemTemperature)
		}
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *ChargeableSubsystemTemperatureListCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.ChargeableSubsystemTemperatureList)
	w.WriteUint8(byte(m.TemperatureCount))
	if m.TemperatureCount > 0 && len(m.Items) > 0 {
		elemCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.ChargeableSubsystemTemperature)(nil)).Elem())
		if elemCodec == nil {
			return api.ErrCodecNotFound
		}
		for i := range m.Items {
			if err := elemCodec.Encode(w, &m.Items[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

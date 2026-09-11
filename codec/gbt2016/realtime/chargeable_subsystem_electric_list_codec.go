package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
)

// ChargeableSubsystemElectricListCodec encodes/decodes the V2016 可充电储能装置电压
// list (TLV type 0x08 body). Wire layout: ElectricCount(u8) + ElectricCount ×
// ChargeableSubsystemElectric entries.
//
// DEVIATION FROM JAVA (reported): Java decodes the count as readByteAsInt
// (1 byte) but encodes it as writeShort (2 bytes) — a known Java inconsistency
// that breaks Java's own encode(decode(x)) byte stability. The Go port keeps
// both sides u8 so roundtrip is byte-stable. This matches the Plan 2 list
// pattern (see Plan 2 Part C4 MotorDataListCodec sample) which reads and
// writes the count as a single byte.
type ChargeableSubsystemElectricListCodec struct{}

func init() {
	api.Register[mdl.ChargeableSubsystemElectricList](api.V2016, &ChargeableSubsystemElectricListCodec{})
}

func (c *ChargeableSubsystemElectricListCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.ChargeableSubsystemElectricList{}
	m.ElectricCount = int(r.ReadUint8())
	if m.ElectricCount > 0 {
		elemCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.ChargeableSubsystemElectric)(nil)).Elem())
		if elemCodec == nil {
			return nil, api.ErrCodecNotFound
		}
		m.Items = make([]mdl.ChargeableSubsystemElectric, m.ElectricCount)
		for i := 0; i < m.ElectricCount; i++ {
			item, err := elemCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			m.Items[i] = *item.(*mdl.ChargeableSubsystemElectric)
		}
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *ChargeableSubsystemElectricListCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.ChargeableSubsystemElectricList)
	w.WriteUint8(byte(m.ElectricCount))
	if m.ElectricCount > 0 && len(m.Items) > 0 {
		elemCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.ChargeableSubsystemElectric)(nil)).Elem())
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

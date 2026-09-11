package gbt2025

import (
	"reflect"
	"strings"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2025"
)

// VehicleLoginV2025Codec encodes/decodes the V2025 vehicle login request
// (command 0x01). Wire layout mirrors Java VehicleLoginV2025Codec:
//
//	BeanTime(6B) + SerialNum(u16) + ICCID(20B) + Count(u8)
//	+ Count × Length[u8]  (one per battery management system)
//	+ sum(Lengths) × Code[24B]
//
// CRITICAL: every code is 24 bytes — NOT Length[i] bytes per code (audit
// 2026-07-31). The total code count is sum(Lengths), not Count.
type VehicleLoginV2025Codec struct{}

func init() {
	api.Register[mdl.VehicleLoginV2025](api.V2025, &VehicleLoginV2025Codec{})
}

// Decode mirrors Java VehicleLoginV2025Codec.decodeBuffer (lines 25-53).
func (c *VehicleLoginV2025Codec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.VehicleLoginV2025{}

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return nil, api.ErrCodecNotFound
	}
	bt, err := btCodec.Decode(r)
	if err != nil {
		return nil, err
	}
	m.BeanTime = *bt.(*model.BeanTime)

	m.SerialNum = int(r.ReadUint16())
	m.ICCID = strings.TrimRight(r.ReadString(20), " ")
	m.Count = int(r.ReadUint8())

	if m.Count > 0 {
		m.Lengths = make([]int, m.Count)
		for i := 0; i < m.Count; i++ {
			m.Lengths[i] = int(r.ReadUint8())
		}

		sum := 0
		for _, l := range m.Lengths {
			sum += l
		}

		if sum > 0 {
			m.Codes = make([]string, sum)
			for i := 0; i < sum; i++ {
				m.Codes[i] = strings.TrimRight(r.ReadString(24), " ")
			}
		}
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode mirrors Java VehicleLoginV2025Codec.encodeBuffer (lines 57-71).
func (c *VehicleLoginV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.VehicleLoginV2025)

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return api.ErrCodecNotFound
	}
	if err := btCodec.Encode(w, &m.BeanTime); err != nil {
		return err
	}

	w.WriteUint16(uint16(m.SerialNum))
	w.WriteString(m.ICCID, 20)
	w.WriteUint8(byte(m.Count))

	if m.Count > 0 {
		if len(m.Lengths) < m.Count {
			return api.ErrBufferUnderflow
		}
		for i := 0; i < m.Count; i++ {
			w.WriteUint8(byte(m.Lengths[i]))
		}
		for _, code := range m.Codes {
			w.WriteString(code, 24)
		}
	}
	return nil
}

package gbt2016

import (
	"reflect"
	"strings"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2016"
)

// VehicleLoginCodec encodes/decodes the V2016 vehicle login request (command 0x01).
// Wire layout: BeanTime(6B) + SerialNum(2B) + ICCID(20B) + Count(1B) + Length(1B)
// + Count*Length bytes of subsystem codes.
type VehicleLoginCodec struct{}

func init() {
	api.Register[mdl.VehicleLogin](api.V2016, &VehicleLoginCodec{})
}

// Decode mirrors Java VehicleLoginCodec.decodeBuffer: BeanTime via registered
// codec, then SerialNum(u16), ICCID(20B trimmed), Count(u8), Length(u8), and
// Count fixed-length code strings (each Length bytes, right-trimmed).
func (c *VehicleLoginCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.VehicleLogin{}

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
	m.Length = int(r.ReadUint8())

	if m.Count > 0 && m.Length > 0 {
		m.Codes = make([]string, m.Count)
		for i := 0; i < m.Count; i++ {
			m.Codes[i] = strings.TrimRight(r.ReadString(m.Length), " ")
		}
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode mirrors Java VehicleLoginCodec.encodeBuffer: BeanTime, SerialNum(u16),
// ICCID padded to 20B, Count(u8), Length(u8), then each code emitted raw
// (NOT re-padded — matches Java which writes code.getBytes(UTF_8) directly).
func (c *VehicleLoginCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.VehicleLogin)

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
	w.WriteUint8(byte(m.Length))

	if m.Count > 0 && m.Length > 0 && len(m.Codes) > 0 {
		for _, code := range m.Codes {
			// Java writes the raw code bytes; if a code is shorter than Length,
			// the reader's right-trim still recovers it on decode. To stay
			// roundtrip-stable for the typical case (codes already Length
			// bytes), we right-pad with spaces to Length.
			w.WriteString(code, m.Length)
		}
	}
	return nil
}

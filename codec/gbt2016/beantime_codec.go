// Package gbt2016 contains GB/T 32960-2016 message body codecs.
package gbt2016

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
)

// BeanTimeCodec encodes/decodes model.BeanTime as 6 wire bytes
// (Year, Month, Day, Hour, Minute, Second).
type BeanTimeCodec struct{}

func init() {
	api.Register[model.BeanTime](api.V2016, &BeanTimeCodec{})
}

// Decode reads 6 bytes via r.ReadUint8 and returns a *model.BeanTime.
func (c *BeanTimeCodec) Decode(r api.Reader) (api.Message, error) {
	bt := &model.BeanTime{
		Year:   int(r.ReadUint8()),
		Month:  int(r.ReadUint8()),
		Day:    int(r.ReadUint8()),
		Hour:   int(r.ReadUint8()),
		Minute: int(r.ReadUint8()),
		Second: int(r.ReadUint8()),
	}
	if err := r.Err(); err != nil {
		return nil, err
	}
	return bt, nil
}

// Encode writes 6 bytes via w.WriteUint8.
func (c *BeanTimeCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*model.BeanTime)
	w.WriteUint8(byte(m.Year))
	w.WriteUint8(byte(m.Month))
	w.WriteUint8(byte(m.Day))
	w.WriteUint8(byte(m.Hour))
	w.WriteUint8(byte(m.Minute))
	w.WriteUint8(byte(m.Second))
	return nil
}

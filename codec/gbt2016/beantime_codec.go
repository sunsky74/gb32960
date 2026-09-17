// Package gbt2016 包含 GB/T 32960-2016 消息体编解码器。
package gbt2016

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
)

// BeanTimeCodec 将 model.BeanTime 编码/解码为 6 个线上字节
// (Year, Month, Day, Hour, Minute, Second)。
type BeanTimeCodec struct{}

func init() {
	api.Register[model.BeanTime](api.V2016, &BeanTimeCodec{})
}

// Decode 通过 r.ReadUint8 读取 6 字节,返回 *model.BeanTime。
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

// Encode 通过 w.WriteUint8 写入 6 字节。
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

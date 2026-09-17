package gbt2025

import (
	"reflect"
	"strings"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2025"
)

// PlatformLoginV2025Codec 编码/解码 V2025 平台登入请求(命令 0x05)。
// 线格式与 Java PlatformLoginV2025Codec 一致(后者复用了 V2016 的
// PlatformLogin 模型):
// BeanTime(6B) + SerialNum(2B) + Username(12B) + Password(20B) + Cipher(1B).
type PlatformLoginV2025Codec struct{}

func init() {
	api.Register[mdl.PlatformLoginV2025](api.V2025, &PlatformLoginV2025Codec{})
}

// Decode 通过注册的 V2016 BeanTime 编解码器读取 BeanTime,然后依次读取
// SerialNum、Username、Password 和 1 字节的 Cipher。
func (c *PlatformLoginV2025Codec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.PlatformLoginV2025{}

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
	m.Username = strings.TrimRight(r.ReadString(12), " ")
	m.Password = strings.TrimRight(r.ReadString(20), " ")
	m.Cipher = r.ReadUint8()

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode 通过注册的 V2016 BeanTime 编解码器写入 BeanTime,然后依次写入
// SerialNum、Username、Password 和 1 字节的 Cipher。
func (c *PlatformLoginV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.PlatformLoginV2025)

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return api.ErrCodecNotFound
	}
	if err := btCodec.Encode(w, &m.BeanTime); err != nil {
		return err
	}

	w.WriteUint16(uint16(m.SerialNum))
	w.WriteString(m.Username, 12)
	w.WriteString(m.Password, 20)
	w.WriteUint8(m.Cipher)
	return nil
}

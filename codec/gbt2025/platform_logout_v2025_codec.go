// Package gbt2025 包含 GB/T 32960.3-2025 消息体编解码器。
package gbt2025

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2025"
)

// PlatformLogoutV2025Codec 编码/解码 V2025 平台登出请求(命令 0x06)。
// 线格式:BeanTime(6B) + SerialNum(2B)。
// BeanTime 与 V2016 共用;我们按 V2016 查找(未注册 V2025 的 BeanTime
// 编解码器:BeanTime 没有版本相关的行为)。
type PlatformLogoutV2025Codec struct{}

func init() {
	api.Register[mdl.PlatformLogoutV2025](api.V2025, &PlatformLogoutV2025Codec{})
}

// Decode 与 Java PlatformLogoutV2025Codec.decodeBuffer 一致。
func (c *PlatformLogoutV2025Codec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.PlatformLogoutV2025{}

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

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode 与 Java PlatformLogoutV2025Codec.encodeBuffer 一致。
func (c *PlatformLogoutV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.PlatformLogoutV2025)

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return api.ErrCodecNotFound
	}
	if err := btCodec.Encode(w, &m.BeanTime); err != nil {
		return err
	}

	w.WriteUint16(uint16(m.SerialNum))
	return nil
}

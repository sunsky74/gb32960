package gbt2016

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2016"
)

// PlatformLogoutCodec 编解码 V2016 平台登出请求(命令 0x06)。
// 线格式布局:BeanTime(6B) + SerialNum(2B)。
type PlatformLogoutCodec struct{}

func init() {
	api.Register[mdl.PlatformLogout](api.V2016, &PlatformLogoutCodec{})
}

// Decode 镜像 Java PlatformLogoutCodec.decodeBuffer。
func (c *PlatformLogoutCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.PlatformLogout{}

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

// Encode 镜像 Java PlatformLogoutCodec.encodeBuffer。
func (c *PlatformLogoutCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.PlatformLogout)

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

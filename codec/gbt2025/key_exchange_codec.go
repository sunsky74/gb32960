package gbt2025

import (
	"fmt"
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2025"
)

// KeyExchangeCodec 编码/解码 V2025 数据单元加密密钥交换消息(命令 0x0B)。
// 线格式与 Java KeyExchangeCodec 一致:
//
//	Type(u8 ExchangeKeyEnum) + Length(u16) + Key[Length bytes]
//	+ StartTime(BeanTime 6B) + ExpireTime(BeanTime 6B)
type KeyExchangeCodec struct{}

func init() {
	api.Register[mdl.KeyExchangeData](api.V2025, &KeyExchangeCodec{})
}

// Decode 与 Java KeyExchangeCodec.decodeBuffer 一致。
func (c *KeyExchangeCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.KeyExchangeData{}

	m.Type = r.ReadUint8()
	m.Length = int(r.ReadUint16())
	if m.Length > 0 {
		m.Key = r.ReadBytes(m.Length)
	}

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return nil, api.ErrCodecNotFound
	}
	startBT, err := btCodec.Decode(r)
	if err != nil {
		return nil, err
	}
	m.StartTime = *startBT.(*model.BeanTime)

	expireBT, err := btCodec.Decode(r)
	if err != nil {
		return nil, err
	}
	m.ExpireTime = *expireBT.(*model.BeanTime)

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode 与 Java KeyExchangeCodec.encodeBuffer 一致。
func (c *KeyExchangeCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.KeyExchangeData)

	// fix 2026-09-17: spec 2025.md L486-488(表31)—— 密钥长度 N(2B)必须
	// 等于实际密钥字节数;长度不符时写出的密钥字节数与声明不一致,会使
	// 启用/失效时间整体错位。写入任何内容之前先校验。
	if len(m.Key) != m.Length {
		return fmt.Errorf("gb32960: key exchange key length %d does not match declared length %d", len(m.Key), m.Length)
	}

	w.WriteUint8(m.Type)
	w.WriteUint16(uint16(m.Length))
	if m.Length > 0 {
		w.WriteBytes(m.Key)
	}

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return api.ErrCodecNotFound
	}
	if err := btCodec.Encode(w, &m.StartTime); err != nil {
		return err
	}
	if err := btCodec.Encode(w, &m.ExpireTime); err != nil {
		return err
	}
	return nil
}

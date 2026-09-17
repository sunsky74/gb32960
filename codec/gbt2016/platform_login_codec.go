package gbt2016

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2016"
)

// PlatformLoginCodec 编解码 V2016 平台登入请求(命令 0x05)。
// 线格式布局:BeanTime(6B) + SerialNum(2B) + Username(12B) + Password(20B) + Cipher(1B) = 41 字节。
// Cipher 为 1 字节(与 Java PlatformLoginCodec 第 40 行一致)。
type PlatformLoginCodec struct{}

func init() {
	api.Register[mdl.PlatformLogin](api.V2016, &PlatformLoginCodec{})
}

// Decode 先经由已注册的 BeanTime 编解码器读取 BeanTime,然后读取
// SerialNum、Username、Password 和 1 字节的 Cipher。
func (c *PlatformLoginCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.PlatformLogin{}

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
		return nil, err
	}
	return m, nil
}

// Encode 先经由已注册的 BeanTime 编解码器写入 BeanTime,然后写入
// SerialNum、Username、Password 和 1 字节的 Cipher。
func (c *PlatformLoginCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.PlatformLogin)

	// fix 2026-09-17: GB/T 32960.3-2016 表 21(L371/L372)—— 平台用户名为 12 字节、
	// 平台密码为 20 字节定长 STRING,超长会被 WriteString 静默截断,必须在
	// 写出任何字节之前拒绝(密码不落错误信息,只报长度)。
	if len(m.Username) > 12 {
		return fmt.Errorf("gb32960: platform login username %q longer than 12 bytes", m.Username)
	}
	if len(m.Password) > 20 {
		return fmt.Errorf("gb32960: platform login password length %d longer than 20 bytes", len(m.Password))
	}

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

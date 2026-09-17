package gbt2016

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2016"
)

// VehicleLoginCodec 编解码 V2016 车辆登入请求(命令 0x01)。
// 线格式布局:BeanTime(6B) + SerialNum(2B) + ICCID(20B) + Count(1B) + Length(1B)
// + Count*Length 字节的子系统编码。
type VehicleLoginCodec struct{}

func init() {
	api.Register[mdl.VehicleLogin](api.V2016, &VehicleLoginCodec{})
}

// Decode 镜像 Java VehicleLoginCodec.decodeBuffer:先经由已注册的
// 编解码器解码 BeanTime,然后读取 SerialNum(u16)、ICCID(20B,右侧去空格)、
// Count(u8)、Length(u8),以及 Count 个定长编码字符串(每个 Length 字节,右侧去空格)。
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

// Encode 镜像 Java VehicleLoginCodec.encodeBuffer:BeanTime、SerialNum(u16)、
// 补齐到 20B 的 ICCID、Count(u8)、Length(u8),然后逐个原样输出编码
// (不重新补齐,与直接写 code.getBytes(UTF_8) 的 Java 一致)。
func (c *VehicleLoginCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.VehicleLogin)

	// fix 2026-09-17: GB/T 32960.3-2016 表 6(L109)—— ICCID 为 20 字节 STRING
	// 定长字段,超长会被 WriteString 静默截断成错误的卡号,必须拒绝。
	if len(m.ICCID) > 20 {
		return fmt.Errorf("gb32960: login ICCID %q longer than 20 bytes", m.ICCID)
	}

	// audit 2026-09-17 (M4):写入任何内容之前先校验。表 6 定义
	// m=0 为 "不上传该编码":Count 可以大于 0,而线上为 0 个编码字节
	// (真实生产报文 prod_login_v2016_01.hex 编码了 Count=1、m=0),
	// 因此只要 Length==0,Codes 就必须为空。只有 Length>0 时
	// 才要求 Count 等于 len(Codes);任何编码都不得超过 Length,
	// 否则 WriteString 会静默截断超长编码。
	if m.Length == 0 {
		if len(m.Codes) != 0 {
			return fmt.Errorf("gb32960: login length 0 means no codes uploaded, but %d codes present", len(m.Codes))
		}
	} else if m.Count != len(m.Codes) {
		return fmt.Errorf("gb32960: login count %d does not match codes length %d", m.Count, len(m.Codes))
	}
	for i, code := range m.Codes {
		if len(code) > m.Length {
			return fmt.Errorf("gb32960: login code[%d] %q longer than length %d", i, code, m.Length)
		}
	}

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
			// Java 直接写入编码的原始字节;如果编码短于 Length,
			// 读取方的右侧去空格在解码时仍能还原它。为了在典型情况
			// (编码本就等于 Length 字节)下保持往返稳定,
			// 我们用空格右补齐到 Length。
			w.WriteString(code, m.Length)
		}
	}
	return nil
}

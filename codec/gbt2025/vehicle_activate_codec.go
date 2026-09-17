package gbt2025

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2025"
	v2025rt "github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// VehicleActivateCodec 编码/解码 V2025 车辆激活请求(命令 0x09)。
// 线格式与 Java VehicleActivateCodec 一致:
//
//	CollectTime(BeanTime 6B) + ChipID(16B) + PublicKeyLength(u16)
//	+ PublicKey[PublicKeyLength bytes] + VIN(17B) + VehicleSignature(sub-codec)
//
// 末尾的 VehicleSignature 通过注册的 V2025 编解码器解码
// (见 codec/gbt2025/realtime/vehicle_signature_codec.go)。消费方必须
// 空白导入该包,查找才能成功。
type VehicleActivateCodec struct{}

func init() {
	api.Register[mdl.VehicleActivate](api.V2025, &VehicleActivateCodec{})
}

// Decode 与 Java VehicleActivateCodec.decodeBuffer 一致。
func (c *VehicleActivateCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.VehicleActivate{}

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return nil, api.ErrCodecNotFound
	}
	bt, err := btCodec.Decode(r)
	if err != nil {
		return nil, err
	}
	m.CollectTime = *bt.(*model.BeanTime)

	m.ChipID = strings.TrimRight(r.ReadString(16), " ")
	m.PublicKeyLength = int(r.ReadUint16())
	if m.PublicKeyLength > 0 {
		m.PublicKey = r.ReadBytes(m.PublicKeyLength)
	}
	m.VIN = strings.TrimRight(r.ReadString(17), " ")

	// SignData = 数据采集时间首字节起至 VIN 末字节(含)的全部字节,即签名字段
	// 之前的所有内容(国标 2025 表 B.3: 签名信息紧接 VIN 之后开始)。快照须在
	// 签名子解码推进 reader 之前取。
	consumed := r.Consumed()
	signData := append([]byte(nil), consumed...)

	sigCodec := api.GetCodec(api.V2025, reflect.TypeOf((*v2025rt.VehicleSignature)(nil)).Elem())
	if sigCodec == nil {
		return nil, api.ErrCodecNotFound
	}
	sig, err := sigCodec.Decode(r)
	if err != nil {
		return nil, err
	}
	m.Signature = sig.(*v2025rt.VehicleSignature)
	m.Signature.SignData = signData

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode 与 Java VehicleActivateCodec.encodeBuffer 一致。
//
// Java 通过 writeString(StringUtils.trim(...), 16) 写入 ChipID(固定 16B),
// 但通过 writeString(msg.getVin()) 写入 VIN(变长)。为使与解码侧(按 17B
// 读取 VIN)的往返保持逐字节稳定,我们将 VIN 写成固定 17 字节的字符串。
func (c *VehicleActivateCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.VehicleActivate)

	// fix 2026-09-17: spec 2025.md L847-854(表B.3)—— 芯片ID 16 字节
	// (不足由空格补齐);公钥长度 N(2B)必须等于实际公钥字节数;VIN 17 字节。
	// 芯片ID/VIN 超长会被 WriteString 静默截断;公钥长度不符会使
	// VIN/签名整体错位。写入任何内容之前先校验。
	if len(m.ChipID) > 16 {
		return fmt.Errorf("gb32960: activate chip ID %q longer than 16 bytes", m.ChipID)
	}
	if len(m.PublicKey) != m.PublicKeyLength {
		return fmt.Errorf("gb32960: activate public key length %d does not match declared length %d", len(m.PublicKey), m.PublicKeyLength)
	}
	if len(m.VIN) > 17 {
		return fmt.Errorf("gb32960: activate VIN %q longer than 17 bytes", m.VIN)
	}

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return api.ErrCodecNotFound
	}
	if err := btCodec.Encode(w, &m.CollectTime); err != nil {
		return err
	}

	w.WriteString(strings.TrimSpace(m.ChipID), 16)
	w.WriteUint16(uint16(m.PublicKeyLength))
	if m.PublicKeyLength > 0 {
		w.WriteBytes(m.PublicKey)
	}
	w.WriteString(m.VIN, 17)

	if m.Signature != nil {
		sigCodec := api.GetCodec(api.V2025, reflect.TypeOf((*v2025rt.VehicleSignature)(nil)).Elem())
		if sigCodec == nil {
			return api.ErrCodecNotFound
		}
		if err := sigCodec.Encode(w, m.Signature); err != nil {
			return err
		}
	}
	return nil
}

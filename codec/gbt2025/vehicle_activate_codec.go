package gbt2025

import (
	"reflect"
	"strings"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2025"
	v2025rt "github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// VehicleActivateCodec encodes/decodes the V2025 vehicle activation request
// (command 0x09). Wire layout mirrors Java VehicleActivateCodec:
//
//	CollectTime(BeanTime 6B) + ChipID(16B) + PublicKeyLength(u16)
//	+ PublicKey[PublicKeyLength bytes] + VIN(17B) + VehicleSignature(sub-codec)
//
// The trailing VehicleSignature is decoded via the registered V2025 codec
// (see codec/gbt2025/realtime/vehicle_signature_codec.go). Consumers must
// blank-import that package for the lookup to succeed.
type VehicleActivateCodec struct{}

func init() {
	api.Register[mdl.VehicleActivate](api.V2025, &VehicleActivateCodec{})
}

// Decode mirrors Java VehicleActivateCodec.decodeBuffer.
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

// Encode mirrors Java VehicleActivateCodec.encodeBuffer.
//
// Java writes ChipID via writeString(StringUtils.trim(...), 16) (fixed 16B)
// but VIN via writeString(msg.getVin()) (variable length). To keep roundtrip
// byte-stable with the decode side (which reads VIN as 17B), we write VIN
// as a fixed 17-byte string.
func (c *VehicleActivateCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.VehicleActivate)

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

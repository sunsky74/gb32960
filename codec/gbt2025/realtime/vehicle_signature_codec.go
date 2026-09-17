package realtime

import (
	"fmt"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// VehicleSignatureCodec 编解码 V2025 车辆签名子记录
// (TLV 类型 0xFF)。字段顺序与 Java VehicleSignatureCodec 完全一致:
//
//	Type(u8 SignatureType) + RLength(u16) + RValue[RLength bytes]
//	+ SLength(u16) + SValue[SLength bytes]
//
// SignData 不属于线格式的一部分;Java 从父级
// 缓冲区(SIGNATURE_MARK TLV 标志之前的字节)设置它。上层
// (RealTimeDataV2025Codec、VehicleActivateCodec)填充 SignData;编解码器
// 本身只处理自己的 5 个字段。
type VehicleSignatureCodec struct{}

func init() {
	api.Register[mdl.VehicleSignature](api.V2025, &VehicleSignatureCodec{})
}

func (c *VehicleSignatureCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.VehicleSignature{}
	m.Type = r.ReadUint8()
	m.RLength = int(r.ReadUint16())
	if m.RLength > 0 {
		m.RValue = r.ReadBytes(m.RLength)
	}
	m.SLength = int(r.ReadUint16())
	if m.SLength > 0 {
		m.SValue = r.ReadBytes(m.SLength)
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *VehicleSignatureCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.VehicleSignature)

	// fix 2026-09-17: spec 2025.md L142-146(表8)—— 签名R值长度/S值长度
	// (各 2B)必须等于实际字节数;长度不符时 S 值长度/值会随 R 的偏差
	// 整体错位。写入任何内容之前先校验。
	if len(m.RValue) != m.RLength {
		return fmt.Errorf("gb32960: signature R value length %d does not match declared RLength %d", len(m.RValue), m.RLength)
	}
	if len(m.SValue) != m.SLength {
		return fmt.Errorf("gb32960: signature S value length %d does not match declared SLength %d", len(m.SValue), m.SLength)
	}

	w.WriteUint8(m.Type)
	w.WriteUint16(uint16(m.RLength))
	if m.RLength > 0 {
		w.WriteBytes(m.RValue)
	}
	w.WriteUint16(uint16(m.SLength))
	if m.SLength > 0 {
		w.WriteBytes(m.SValue)
	}
	return nil
}

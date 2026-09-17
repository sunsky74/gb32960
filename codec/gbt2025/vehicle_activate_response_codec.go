package gbt2025

import (
	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025"
)

// VehicleActivateResponseCodec 编码/解码 V2025 车辆激活响应(命令 0x0A)。
// 线格式:激活状态(u8:0x01=激活成功,0x02=激活失败) + 信息 ResponseCode(u8)。
//
// 与 Plan 2 示例的偏差:Plan 2 Part D 表把 responseCode 列为 u16,
// 但 Java VehicleActivateResponseCodec 读写的是单个字节
// (VehicleActivateResponseEnum.status 是 byte)。Go 模型
// (VehicleActivateResponse.ResponseCode)也是 byte 类型。我们遵循
// 实际 API 形态(Java + Go 模型):responseCode 为 1 字节。
type VehicleActivateResponseCodec struct{}

func init() {
	api.Register[mdl.VehicleActivateResponse](api.V2025, &VehicleActivateResponseCodec{})
}

// Decode 与 Java VehicleActivateResponseCodec.decodeBuffer 一致:
// readByteAsBool + VehicleActivateResponseEnum.valueOf(readByte)。
//
// fix 2026-09-17: spec 2025.md L863-865(表B.4)—— 激活状态 1B:
// 0x01=激活成功,0x02=激活失败;非 0x01 一律视为失败
// (旧实现按 !=0 判定,会把 0x02 失败误判为成功)。
func (c *VehicleActivateResponseCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.VehicleActivateResponse{}
	m.Success = r.ReadUint8() == 0x01
	m.ResponseCode = r.ReadUint8()

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode 与 Java VehicleActivateResponseCodec.encodeBuffer 一致:
// writeByte(success) + writeByte(responseCode.status)。
//
// fix 2026-09-17: spec 2025.md L863-865(表B.4)—— 成功写 0x01、失败写 0x02;
// 旧实现在失败时写 0x00,规范未定义该值。
func (c *VehicleActivateResponseCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.VehicleActivateResponse)
	if m.Success {
		w.WriteUint8(0x01)
	} else {
		w.WriteUint8(0x02)
	}
	w.WriteUint8(m.ResponseCode)
	return nil
}

package realtime

import (
	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// CustomV2025DataCodec 编解码 V2025 自定义数据条目
// (TLV 类型 0x80~0xFE)。字段顺序与 Java CustomV2025DataCodec 一致:
//
//	Decode: Length(u16) + Data[Length bytes]   (CustomKey 不在线格式上;由调用方设置)
//	Encode: CustomKey(u8) + Length(u16) + Data[Length bytes]
//
// 编码/解码的不对称是有意为之:解码时父级 TLV
// 分发器已经消费了 customKey 字节(它就是 TLV 标志),
// 因此本编解码器只读取 Length + Data。编码时本编解码器
// 先写入自己的 customKey 字节(分发器不会为
// CustomV2025Data 另写一个标志;与 Java RealTimeDataV2025Codec.encodePayload
// 一致,后者把类型字节写在 CustomV2025Data 编解码器自身内部)。
//
// 当 Length 为 BYTE2 错误哨兵值时,Data 主体被跳过
// (Java: if (!DataErrorValue.BYTE2.inInvalid(length)))。
type CustomV2025DataCodec struct{}

func init() {
	api.Register[mdl.CustomV2025Data](api.V2025, &CustomV2025DataCodec{})
}

func (c *CustomV2025DataCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.CustomV2025Data{}
	m.Length = int(r.ReadUint16())
	if !types.ErrByte2.IsInvalid(int64(m.Length)) && m.Length > 0 {
		m.Data = r.ReadBytes(m.Length)
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *CustomV2025DataCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.CustomV2025Data)
	w.WriteUint8(m.CustomKey)
	w.WriteUint16(uint16(m.Length))
	if m.Length > 0 && len(m.Data) > 0 {
		w.WriteBytes(m.Data)
	}
	return nil
}

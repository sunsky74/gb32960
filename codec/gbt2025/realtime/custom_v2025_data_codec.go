package realtime

import (
	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// CustomV2025DataCodec encodes/decodes the V2025 自定义数据 entry
// (TLV types 0x80~0xFE). Field order mirrors Java CustomV2025DataCodec:
//
//	Decode: Length(u16) + Data[Length bytes]   (CustomKey NOT on wire — set by caller)
//	Encode: CustomKey(u8) + Length(u16) + Data[Length bytes]
//
// The encode/decode asymmetry is intentional: on decode the parent TLV
// dispatcher has already consumed the customKey byte (it IS the TLV flag),
// so this codec reads only Length + Data. On encode the codec writes its
// own customKey byte first (the dispatcher does not write a separate flag
// for CustomV2025Data — matches Java RealTimeDataV2025Codec.encodePayload
// which writes the type byte inside the CustomV2025Data codec itself).
//
// When Length is the BYTE2 error sentinel, the Data body is skipped
// (Java: if (!DataErrorValue.BYTE2.inInvalid(length))).
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

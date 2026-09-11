package realtime

import (
	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// VehicleSignatureCodec encodes/decodes the V2025 车辆签名 sub-record
// (TLV type 0xFF). Field order mirrors Java VehicleSignatureCodec exactly:
//
//	Type(u8 SignatureType) + RLength(u16) + RValue[RLength bytes]
//	+ SLength(u16) + SValue[SLength bytes]
//
// SignData is NOT part of the wire format — Java sets it from the parent
// buffer (the bytes preceding the SIGNATURE_MARK TLV flag). Higher layers
// (RealTimeDataV2025Codec, VehicleActivateCodec) populate SignData; the codec
// itself only handles its own 5 fields.
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

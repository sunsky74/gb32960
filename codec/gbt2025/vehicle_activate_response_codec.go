package gbt2025

import (
	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025"
)

// VehicleActivateResponseCodec encodes/decodes the V2025 vehicle activation
// response (command 0x0A). Wire layout: Success(u8→bool) + ResponseCode(u8).
//
// DEVIATION FROM PLAN 2 SAMPLE: Plan 2 Part D table lists responseCode as
// u16, but Java VehicleActivateResponseCodec reads/writes a single byte
// (VehicleActivateResponseEnum.status is a byte). The Go model
// (VehicleActivateResponse.ResponseCode) is also byte-typed. We follow the
// actual API surface (Java + Go model): 1 byte for responseCode.
type VehicleActivateResponseCodec struct{}

func init() {
	api.Register[mdl.VehicleActivateResponse](api.V2025, &VehicleActivateResponseCodec{})
}

// Decode mirrors Java VehicleActivateResponseCodec.decodeBuffer:
// readByteAsBool + VehicleActivateResponseEnum.valueOf(readByte).
func (c *VehicleActivateResponseCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.VehicleActivateResponse{}
	m.Success = r.ReadUint8() != 0
	m.ResponseCode = r.ReadUint8()

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode mirrors Java VehicleActivateResponseCodec.encodeBuffer:
// writeByte(success) + writeByte(responseCode.status).
func (c *VehicleActivateResponseCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.VehicleActivateResponse)
	if m.Success {
		w.WriteUint8(1)
	} else {
		w.WriteUint8(0)
	}
	w.WriteUint8(m.ResponseCode)
	return nil
}

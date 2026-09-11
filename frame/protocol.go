// Package frame contains the protocol-frame type (ProtocolMessage) and the
// command-to-message-type dispatch (payloadType). It lives in a separate
// package to break the circular import: ProtocolMessage references gbt2016
// and gbt2025 types, while gbt2016/gbt2025 reference model (for BeanTime).
// Placing the frame type here allows it to import all model subpackages
// without creating a cycle back to model.
package frame

import (
	"errors"
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/model/gbt2016"
	"github.com/sunsky74/gb32960/model/gbt2025"
	"github.com/sunsky74/gb32960/types"
	"github.com/sunsky74/gb32960/utils"
)

// ProtocolMessage represents a complete GB/T 32960 protocol frame.
type ProtocolMessage struct {
	Version       api.GBTVersion
	RequestType   any // *types.CommandV2016 or *types.CommandV2025
	ResponseType  types.ResponseType
	VIN           string
	Encryption    types.EncryptionType
	PayloadLength int
	RawBytes      []byte
	Payload       model.MessageBody
	CheckCode     byte
}

// PayloadType returns the Go struct type for a (version, command) pair,
// mirroring the reference Java implementation (getV2016Body / getV2025Body).
// Returns nil for commands with no decodable body (Heartbeat, ClockCorrect,
// passthrough config/control).
func PayloadType(v api.GBTVersion, cmdCode byte) reflect.Type {
	switch v {
	case api.V2016:
		switch cmdCode {
		case 0x01:
			return reflect.TypeOf((*gbt2016.VehicleLogin)(nil)).Elem()
		case 0x02, 0x03:
			return reflect.TypeOf((*gbt2016.RealTimeData)(nil)).Elem()
		case 0x04:
			return reflect.TypeOf((*gbt2016.VehicleLogout)(nil)).Elem()
		}
	case api.V2025:
		switch cmdCode {
		case 0x01:
			return reflect.TypeOf((*gbt2025.VehicleLoginV2025)(nil)).Elem()
		case 0x02, 0x03:
			return reflect.TypeOf((*gbt2025.RealTimeV2025Data)(nil)).Elem()
		case 0x04:
			return reflect.TypeOf((*gbt2016.VehicleLogout)(nil)).Elem()
		case 0x05:
			return reflect.TypeOf((*gbt2016.PlatformLogin)(nil)).Elem()
		case 0x06:
			return reflect.TypeOf((*gbt2016.PlatformLogout)(nil)).Elem()
		case 0x09:
			return reflect.TypeOf((*gbt2025.VehicleActivate)(nil)).Elem()
		case 0x0A:
			return reflect.TypeOf((*gbt2025.VehicleActivateResponse)(nil)).Elem()
		case 0x0B:
			return reflect.TypeOf((*gbt2025.KeyExchangeData)(nil)).Elem()
		}
	}
	return nil
}

// CommandCode extracts the wire command byte from RequestType.
func CommandCode(rt any) (byte, bool) {
	switch c := rt.(type) {
	case *types.CommandV2016:
		return c.Code, true
	case *types.CommandV2025:
		return c.Code, true
	}
	return 0, false
}

// Bytes encodes the entire protocol frame to bytes.
func (m *ProtocolMessage) Bytes() ([]byte, error) {
	if m.Payload == nil {
		return nil, errors.New("gb32960: Payload is nil, cannot encode protocol frame")
	}
	w := utils.NewByteWriter()
	w.WriteString(types.Header(m.Version), 2)
	code, ok := CommandCode(m.RequestType)
	if !ok {
		return nil, errors.New("gb32960: RequestType is nil or unknown command type")
	}
	w.WriteUint8(code)
	w.WriteUint8(m.ResponseType.Code())
	w.WriteString(m.VIN, 17)
	w.WriteUint8(byte(m.Encryption))

	payload, err := m.Payload.Bytes()
	if err != nil {
		return nil, err
	}

	if m.Encryption != types.EncryptionNone {
		return nil, api.ErrEncryptionNotSupported
	}

	w.WriteUint16(uint16(len(payload)))
	w.WriteBytes(payload)

	bccRange := w.Bytes()[2:]
	w.WriteUint8(utils.CalcBCC(bccRange))

	return w.Bytes(), nil
}

// DecodePayload decodes RawBytes into the Payload field.
func (m *ProtocolMessage) DecodePayload() error {
	if m.RawBytes == nil {
		return api.ErrBufferUnderflow
	}
	code, ok := CommandCode(m.RequestType)
	if !ok {
		return errors.New("gb32960: RequestType is nil or unknown command type")
	}

	msgType := PayloadType(m.Version, code)
	if msgType == nil {
		return nil
	}

	codec := api.GetCodec(m.Version, msgType)
	if codec == nil {
		return api.ErrCodecNotFound
	}

	r := utils.NewByteReader(m.RawBytes)
	decoded, err := codec.Decode(r)
	if err != nil {
		return err
	}

	var ok2 bool
	m.Payload, ok2 = decoded.(model.MessageBody)
	if !ok2 {
		return errors.New("gb32960: decoded message does not implement MessageBody")
	}
	return nil
}

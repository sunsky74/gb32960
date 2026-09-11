package codec

import (
	"errors"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/frame"
	"github.com/sunsky74/gb32960/types"
	"github.com/sunsky74/gb32960/utils"
)

// ProtocolCodec is the top-level frame codec.
// Decode: validates BCC, parses header fields, decrypts payload (0x01 only).
// Encode: constructs header, encrypts payload, appends BCC.
var ProtocolCodec = &protocolMessageCodec{}

type protocolMessageCodec struct{}

// Decode parses a complete GB/T 32960 protocol frame from r.
//
// Wire layout: Header(2B) | Cmd(1B) | Response(1B) | VIN(17B) | Encryption(1B)
//
//	| PayloadLen(2B) | Payload(N B) | BCC(1B)
//
// BCC covers bytes [2:] of the frame (everything after the 2-byte header).
// Only EncryptionNone (0x01, pass-through) is supported; any other mode
// returns api.ErrEncryptionNotSupported.
func (c *protocolMessageCodec) Decode(r api.Reader) (api.Message, error) {
	m := &frame.ProtocolMessage{}

	// 1. Read version header (2 bytes).
	header := r.ReadUint16()
	m.Version = types.GBTVersionByHeader(header)

	// 2. Read remaining bytes (command through BCC, inclusive).
	bodyWithBCC := r.ReadBytes(r.Remaining())

	// 3. Validate BCC: XOR of all bytes except the last must equal the last byte.
	// An empty body means either the header underflowed or no body at all —
	// surface as ErrBufferUnderflow (tolerant of r.Err() per api.Reader contract).
	if len(bodyWithBCC) == 0 {
		return nil, api.ErrBufferUnderflow
	}
	bodyBytes := bodyWithBCC[:len(bodyWithBCC)-1]
	expectedBCC := utils.CalcBCC(bodyBytes)
	actualBCC := bodyWithBCC[len(bodyWithBCC)-1]
	if expectedBCC != actualBCC {
		return nil, api.ErrBCCMismatch
	}
	m.CheckCode = actualBCC

	// 4. Parse protocol header fields from bodyBytes.
	bodyReader := utils.NewByteReader(bodyBytes)
	cmd := bodyReader.ReadUint8()
	// Dispatch by version: V2025 has its own command table; V2016 is the default.
	// (audit 2026-07-31: earlier draft hardcoded CommandV2016ByCode, which would
	// mis-classify V2025 frames. Mirror the reference implementation's switch.)
	switch m.Version {
	case api.V2025:
		m.RequestType = types.CommandV2025ByCode(cmd)
	default:
		m.RequestType = types.CommandV2016ByCode(cmd)
	}
	m.ResponseType = types.ResponseByCode(bodyReader.ReadUint8())
	m.VIN = bodyReader.ReadString(17)
	m.Encryption = types.EncryptionType(bodyReader.ReadUint8())
	m.PayloadLength = int(bodyReader.ReadUint16())

	// 5. Read payload and reject non-pass-through encryption.
	// For 0x01 (EncryptionNone) the wire bytes ARE the plaintext payload.
	encrypted := bodyReader.ReadBytes(m.PayloadLength)
	if m.Encryption != types.EncryptionNone {
		return nil, api.ErrEncryptionNotSupported
	}
	m.RawBytes = encrypted // 0x01: encrypted == plain

	// Surface any underflow recorded while parsing the body fields
	// (matches the pattern in codec/gbt2016/platform_login_codec.go).
	if err := bodyReader.Err(); err != nil {
		return nil, err
	}
	return m, nil
}

// Encode writes a complete GB/T 32960 protocol frame for msg into w.
//
// DEVIATION FROM PLAN 2 SAMPLE (reported): Plan 2 Part B's sample calls
// `w.Bytes()[2:]` to compute the BCC range, but the api.Writer interface
// (see api/api.go) has no Bytes() method — only the concrete *utils.ByteWriter
// does. To stay faithful to the frame layout while working with any
// api.Writer, we mirror frame.ProtocolMessage.Bytes() into a local
// *utils.ByteWriter (which does expose Bytes()), compute BCC over its [2:]
// slice, then emit the assembled frame via w.WriteBytes. The resulting wire
// bytes are byte-for-byte identical to frame.ProtocolMessage.Bytes().
func (c *protocolMessageCodec) Encode(w api.Writer, msg api.Message) error {
	m, ok := msg.(*frame.ProtocolMessage)
	if !ok {
		return errors.New("gb32960: message is not *frame.ProtocolMessage")
	}
	if m.Payload == nil {
		return errors.New("gb32960: Payload is nil, cannot encode protocol frame")
	}

	code, ok := frame.CommandCode(m.RequestType)
	if !ok {
		return errors.New("gb32960: RequestType is nil or unknown command type")
	}

	// Build into a local writer so the BCC range ([2:]) is observable.
	// This mirrors frame.ProtocolMessage.Bytes() exactly.
	local := utils.NewByteWriter()
	local.WriteString(types.Header(m.Version), 2)
	local.WriteUint8(code)
	local.WriteUint8(m.ResponseType.Code())
	local.WriteString(m.VIN, 17)
	local.WriteUint8(byte(m.Encryption))

	payload, err := m.Payload.Bytes()
	if err != nil {
		return err
	}
	if m.Encryption != types.EncryptionNone {
		return api.ErrEncryptionNotSupported
	}

	local.WriteUint16(uint16(len(payload)))
	local.WriteBytes(payload)

	// BCC: XOR of everything from byte[2] (after the 2-byte header) to end of payload.
	bccRange := local.Bytes()[2:]
	local.WriteUint8(utils.CalcBCC(bccRange))

	// Emit the assembled frame into the caller-provided writer.
	w.WriteBytes(local.Bytes())
	return nil
}

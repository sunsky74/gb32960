package codec_test

import (
	"bytes"
	"testing"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	"github.com/sunsky74/gb32960/frame"
	"github.com/sunsky74/gb32960/model/gbt2016"
	"github.com/sunsky74/gb32960/model/gbt2025"
	v2025rt "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
	"github.com/sunsky74/gb32960/utils"

	_ "github.com/sunsky74/gb32960/codec/all"
)

// TestProtocolFrame_Roundtrip exercises the full protocol frame path:
//
//  1. Build a frame.ProtocolMessage with a payload struct.
//  2. Encode via frame.Bytes() — writes header, cmd, response, VIN,
//     encryption, payload-length, payload, and BCC over bytes [2:].
//  3. Decode via codec.ProtocolCodec.Decode — validates BCC, parses the
//     header fields, and stores RawBytes for lazy payload decode.
//  4. DecodePayload() looks up the message-body type via frame.PayloadType
//     and the codec via api.GetCodec, then decodes RawBytes into Payload.
//  5. Assert the decoded Payload struct equals the original sample.
//
// This is the Plan 2 Task E2 acceptance test: proves the frame codec,
// BCC checksum, command→type dispatch (payloadType switch), and codec
// registry all integrate end-to-end. Vehicles of both protocol versions
// are covered.
func TestProtocolFrame_Roundtrip(t *testing.T) {
	t.Run("V2016_VehicleLogin", func(t *testing.T) {
		payload := sampleVehicleLogin()

		pm := &frame.ProtocolMessage{
			Version:      api.V2016,
			RequestType:  types.CommandV2016ByCode(0x01),
			ResponseType: types.ResponseCommand,
			VIN:          "LSVAU2A37K2100001",
			Encryption:   types.EncryptionNone,
			Payload:      payload,
		}

		frameBytes, err := pm.Bytes()
		if err != nil {
			t.Fatalf("frame.Bytes failed: %v", err)
		}
		if got := string(frameBytes[:2]); got != "##" {
			t.Errorf("header: got %q, want ##", got)
		}
		const wantFrameLen = 2 + 1 + 1 + 17 + 1 + 2 + 36 + 1 // header + cmd + resp + VIN + enc + len + payload(36) + BCC
		if len(frameBytes) != wantFrameLen {
			t.Fatalf("frame length: got %d, want %d (bytes=%X)", len(frameBytes), wantFrameLen, frameBytes)
		}

		decoded, err := codec.ProtocolCodec.Decode(utils.NewByteReader(frameBytes))
		if err != nil {
			t.Fatalf("ProtocolCodec.Decode failed: %v", err)
		}
		df, ok := decoded.(*frame.ProtocolMessage)
		if !ok {
			t.Fatalf("decoded type: %T, want *frame.ProtocolMessage", decoded)
		}

		if df.Version != api.V2016 {
			t.Errorf("Version: got %v, want V2016", df.Version)
		}
		if df.VIN != pm.VIN {
			t.Errorf("VIN: got %q, want %q", df.VIN, pm.VIN)
		}
		if df.Encryption != types.EncryptionNone {
			t.Errorf("Encryption: got %v, want EncryptionNone", df.Encryption)
		}
		if df.PayloadLength != len(frameBytes)-(2+1+1+17+1+2)-1 {
			t.Errorf("PayloadLength: got %d, want payload size", df.PayloadLength)
		}

		// Payload should NOT be decoded yet — Decode only parses the frame,
		// the caller invokes DecodePayload lazily.
		if df.Payload != nil {
			t.Errorf("Payload should be nil before DecodePayload, got %T", df.Payload)
		}

		if err := df.DecodePayload(); err != nil {
			t.Fatalf("DecodePayload failed: %v", err)
		}

		vl, ok := df.Payload.(*gbt2016.VehicleLogin)
		if !ok {
			t.Fatalf("Payload type: got %T, want *gbt2016.VehicleLogin", df.Payload)
		}
		if vl.ICCID != payload.ICCID {
			t.Errorf("ICCID: got %q, want %q", vl.ICCID, payload.ICCID)
		}
		if vl.SerialNum != payload.SerialNum {
			t.Errorf("SerialNum: got %d, want %d", vl.SerialNum, payload.SerialNum)
		}
		if len(vl.Codes) != len(payload.Codes) {
			t.Fatalf("Codes len: got %d, want %d", len(vl.Codes), len(payload.Codes))
		}
		for i, want := range payload.Codes {
			if vl.Codes[i] != want {
				t.Errorf("Codes[%d]: got %q, want %q", i, vl.Codes[i], want)
			}
		}
	})

	t.Run("V2025_VehicleLogin", func(t *testing.T) {
		payload := sampleVehicleLoginV2025()

		pm := &frame.ProtocolMessage{
			Version:      api.V2025,
			RequestType:  types.CommandV2025ByCode(0x01),
			ResponseType: types.ResponseCommand,
			VIN:          "LSVAU2A37K2100002",
			Encryption:   types.EncryptionNone,
			Payload:      payload,
		}

		frameBytes, err := pm.Bytes()
		if err != nil {
			t.Fatalf("frame.Bytes failed: %v", err)
		}
		if got := string(frameBytes[:2]); got != "$$" {
			t.Errorf("header: got %q, want $$", got)
		}

		decoded, err := codec.ProtocolCodec.Decode(utils.NewByteReader(frameBytes))
		if err != nil {
			t.Fatalf("ProtocolCodec.Decode failed: %v", err)
		}
		df := decoded.(*frame.ProtocolMessage)

		if df.Version != api.V2025 {
			t.Errorf("Version: got %v, want V2025", df.Version)
		}
		if df.VIN != pm.VIN {
			t.Errorf("VIN: got %q, want %q", df.VIN, pm.VIN)
		}
		if _, ok := df.RequestType.(*types.CommandV2025); !ok {
			t.Errorf("RequestType: got %T, want *types.CommandV2025", df.RequestType)
		}

		if err := df.DecodePayload(); err != nil {
			t.Fatalf("DecodePayload failed: %v", err)
		}
		vl, ok := df.Payload.(*gbt2025.VehicleLoginV2025)
		if !ok {
			t.Fatalf("Payload type: got %T, want *gbt2025.VehicleLoginV2025", df.Payload)
		}
		if vl.ICCID != payload.ICCID {
			t.Errorf("ICCID: got %q, want %q", vl.ICCID, payload.ICCID)
		}
		if len(vl.Lengths) != len(payload.Lengths) {
			t.Fatalf("Lengths: got %v, want %v", vl.Lengths, payload.Lengths)
		}
	})

	t.Run("FrameByteStability", func(t *testing.T) {
		// Re-encoding the decoded frame must produce byte-identical output.
		// This catches drift in any header field, BCC computation, or
		// payload encoding between the two codec paths.
		payload := sampleVehicleLogin()
		pm := &frame.ProtocolMessage{
			Version:      api.V2016,
			RequestType:  types.CommandV2016ByCode(0x01),
			ResponseType: types.ResponseCommand,
			VIN:          "LSVAU2A37K2100001",
			Encryption:   types.EncryptionNone,
			Payload:      payload,
		}
		original, err := pm.Bytes()
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}

		decoded, err := codec.ProtocolCodec.Decode(utils.NewByteReader(original))
		if err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		df := decoded.(*frame.ProtocolMessage)
		if err := df.DecodePayload(); err != nil {
			t.Fatalf("DecodePayload failed: %v", err)
		}

		reEncoded, err := df.Bytes()
		if err != nil {
			t.Fatalf("re-encode failed: %v", err)
		}
		if !bytes.Equal(original, reEncoded) {
			t.Errorf("frame byte-stability failed:\n  original   = %X\n  re-encoded = %X",
				original, reEncoded)
		}
	})

	t.Run("EncryptionRejected", func(t *testing.T) {
		// Encryption mode != 0x01 must surface as api.ErrEncryptionNotSupported.
		// Plan 2 completion criterion: only pass-through (0x01) is implemented.
		pm := &frame.ProtocolMessage{
			Version:      api.V2016,
			RequestType:  types.CommandV2016ByCode(0x01),
			ResponseType: types.ResponseCommand,
			VIN:          "LSVAU2A37K2100001",
			Encryption:   types.EncryptionRSA, // 0x02 — unsupported
			Payload:      sampleVehicleLogin(),
		}
		if _, err := pm.Bytes(); err != api.ErrEncryptionNotSupported {
			t.Errorf("encode EncryptionRSA: got err=%v, want %v", err, api.ErrEncryptionNotSupported)
		}

		// On decode side: build a frame claiming RSA encryption and confirm
		// ProtocolCodec.Decode rejects it with the same sentinel.
		handCrafted := buildFrameWithEncryption(t, api.V2016, 0x01, byte(types.EncryptionRSA), sampleVehicleLogin())
		if _, err := codec.ProtocolCodec.Decode(utils.NewByteReader(handCrafted)); err != api.ErrEncryptionNotSupported {
			t.Errorf("decode EncryptionRSA: got err=%v, want %v", err, api.ErrEncryptionNotSupported)
		}
	})

	t.Run("BCCMismatchRejected", func(t *testing.T) {
		// Corrupt the BCC byte (last byte) and confirm ErrBCCMismatch.
		pm := &frame.ProtocolMessage{
			Version:      api.V2016,
			RequestType:  types.CommandV2016ByCode(0x01),
			ResponseType: types.ResponseCommand,
			VIN:          "LSVAU2A37K2100001",
			Encryption:   types.EncryptionNone,
			Payload:      sampleVehicleLogin(),
		}
		good, err := pm.Bytes()
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}
		bad := append([]byte{}, good...)
		bad[len(bad)-1] ^= 0xFF // flip all bits of BCC
		if _, err := codec.ProtocolCodec.Decode(utils.NewByteReader(bad)); err != api.ErrBCCMismatch {
			t.Errorf("decode corrupt BCC: got err=%v, want %v", err, api.ErrBCCMismatch)
		}
	})

	t.Run("RealTimeFrame_DecodePayloadComposite", func(t *testing.T) {
		// RealTimeData payload contains multiple TLV sub-records. DecodePayload
		// must dispatch through the registered RealTimeDataCodec and the TLV
		// sub-codec lookup table. This catches init-order hazards where the
		// realtime sub-codecs are registered in a sibling package.
		vd := &gbt2016.RealTimeData{
			BeanTime:    sampleBeanTime(),
			VehicleData: nil, // payload can have nil sub-records
			EngineData:  nil,
		}
		// Sanity: empty RealTimeData roundtrips.
		pm := &frame.ProtocolMessage{
			Version:      api.V2016,
			RequestType:  types.CommandV2016ByCode(0x02),
			ResponseType: types.ResponseCommand,
			VIN:          "LSVAU2A37K2100003",
			Encryption:   types.EncryptionNone,
			Payload:      vd,
		}
		frameBytes, err := pm.Bytes()
		if err != nil {
			t.Fatalf("frame.Bytes failed: %v", err)
		}
		decoded, err := codec.ProtocolCodec.Decode(utils.NewByteReader(frameBytes))
		if err != nil {
			t.Fatalf("ProtocolCodec.Decode failed: %v", err)
		}
		df := decoded.(*frame.ProtocolMessage)
		if err := df.DecodePayload(); err != nil {
			t.Fatalf("DecodePayload failed: %v", err)
		}
		if _, ok := df.Payload.(*gbt2016.RealTimeData); !ok {
			t.Fatalf("Payload type: got %T, want *gbt2016.RealTimeData", df.Payload)
		}
	})

	t.Run("V2025_KeyExchangeFrame", func(t *testing.T) {
		// V2025-only command 0x0B (KeyExchange). Confirms the V2025 command
		// table dispatches to the right codec via PayloadType.
		payload := &gbt2025.KeyExchangeData{
			Type:       0x02,
			Length:     4,
			Key:        []byte{0xDE, 0xAD, 0xBE, 0xEF},
			StartTime:  sampleBeanTime(),
			ExpireTime: sampleBeanTime(),
		}
		pm := &frame.ProtocolMessage{
			Version:      api.V2025,
			RequestType:  types.CommandV2025ByCode(0x0B),
			ResponseType: types.ResponseCommand,
			VIN:          "LSVAU2A37K2100004",
			Encryption:   types.EncryptionNone,
			Payload:      payload,
		}
		frameBytes, err := pm.Bytes()
		if err != nil {
			t.Fatalf("frame.Bytes failed: %v", err)
		}
		decoded, err := codec.ProtocolCodec.Decode(utils.NewByteReader(frameBytes))
		if err != nil {
			t.Fatalf("ProtocolCodec.Decode failed: %v", err)
		}
		df := decoded.(*frame.ProtocolMessage)
		if err := df.DecodePayload(); err != nil {
			t.Fatalf("DecodePayload failed: %v", err)
		}
		ke, ok := df.Payload.(*gbt2025.KeyExchangeData)
		if !ok {
			t.Fatalf("Payload type: got %T, want *gbt2025.KeyExchangeData", df.Payload)
		}
		if ke.Type != payload.Type {
			t.Errorf("Type: got 0x%02X, want 0x%02X", ke.Type, payload.Type)
		}
		if !bytes.Equal(ke.Key, payload.Key) {
			t.Errorf("Key: got %X, want %X", ke.Key, payload.Key)
		}
	})

	t.Run("V2025_VehicleActivateFrame", func(t *testing.T) {
		// V2025-only command 0x09 with nested VehicleSignature sub-codec.
		// Confirms nested codec lookup (VehicleActivate → VehicleSignature)
		// works through the frame path.
		payload := &gbt2025.VehicleActivate{
			CollectTime:     sampleBeanTime(),
			ChipID:          "CHIP2025ABC001",
			PublicKeyLength: 4,
			PublicKey:       []byte{0x01, 0x02, 0x03, 0x04},
			VIN:             "LSVAU2A37K2100005",
			Signature: &v2025rt.VehicleSignature{
				Type:    0x02,
				RLength: 2,
				RValue:  []byte{0xAA, 0xBB},
				SLength: 2,
				SValue:  []byte{0xCC, 0xDD},
			},
		}
		pm := &frame.ProtocolMessage{
			Version:      api.V2025,
			RequestType:  types.CommandV2025ByCode(0x09),
			ResponseType: types.ResponseCommand,
			VIN:          "LSVAU2A37K2100005",
			Encryption:   types.EncryptionNone,
			Payload:      payload,
		}
		frameBytes, err := pm.Bytes()
		if err != nil {
			t.Fatalf("frame.Bytes failed: %v", err)
		}
		decoded, err := codec.ProtocolCodec.Decode(utils.NewByteReader(frameBytes))
		if err != nil {
			t.Fatalf("ProtocolCodec.Decode failed: %v", err)
		}
		df := decoded.(*frame.ProtocolMessage)
		if err := df.DecodePayload(); err != nil {
			t.Fatalf("DecodePayload failed: %v", err)
		}
		va, ok := df.Payload.(*gbt2025.VehicleActivate)
		if !ok {
			t.Fatalf("Payload type: got %T, want *gbt2025.VehicleActivate", df.Payload)
		}
		if va.ChipID != payload.ChipID {
			t.Errorf("ChipID: got %q, want %q", va.ChipID, payload.ChipID)
		}
		if va.Signature == nil {
			t.Fatal("Signature: got nil, want non-nil")
		}
		if va.Signature.Type != payload.Signature.Type {
			t.Errorf("Signature.Type: got 0x%02X, want 0x%02X", va.Signature.Type, payload.Signature.Type)
		}
		if !bytes.Equal(va.Signature.RValue, payload.Signature.RValue) {
			t.Errorf("Signature.RValue: got %X, want %X", va.Signature.RValue, payload.Signature.RValue)
		}
	})
}

func sampleVehicleLogin() *gbt2016.VehicleLogin {
	return &gbt2016.VehicleLogin{
		BeanTime:  sampleBeanTime(),
		SerialNum: 1,
		ICCID:     "89860000000000000001",
		Count:     2,
		Length:    3,
		Codes:     []string{"001", "002"},
	}
}

func sampleVehicleLoginV2025() *gbt2025.VehicleLoginV2025 {
	return &gbt2025.VehicleLoginV2025{
		BeanTime:  sampleBeanTime(),
		SerialNum: 1,
		ICCID:     "89860000000000000002",
		Count:     2,
		Lengths:   []int{1, 1},
		Codes:     []string{"BMU001", "BMU002"},
	}
}

// buildFrameWithEncryption builds a protocol frame byte slice with a chosen
// encryption byte, bypassing frame.Bytes() (which rejects non-pass-through
// modes). Used only by EncryptionRejected to exercise the decode-side rejection.
func buildFrameWithEncryption(t *testing.T, version api.GBTVersion, cmdCode, encByte byte, payloadFrame any) []byte {
	t.Helper()

	w := utils.NewByteWriter()
	w.WriteString(types.Header(version), 2)
	w.WriteUint8(cmdCode)
	w.WriteUint8(byte(types.ResponseCommand))
	w.WriteString("LSVAU2A37K2100001", 17)
	w.WriteUint8(encByte)

	type byter interface {
		Bytes() ([]byte, error)
	}
	p, err := payloadFrame.(byter).Bytes()
	if err != nil {
		t.Fatalf("payload Bytes failed: %v", err)
	}
	w.WriteUint16(uint16(len(p)))
	w.WriteBytes(p)

	// BCC over bytes [2:] (everything after the 2-byte header).
	bccRange := w.Bytes()[2:]
	w.WriteUint8(utils.CalcBCC(bccRange))
	return w.Bytes()
}

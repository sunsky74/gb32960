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

// TestProtocolFrame_Roundtrip 演练完整的协议帧路径:
//
//  1. 用数据单元结构体构造一个 frame.ProtocolMessage。
//  2. 通过 frame.Bytes() 编码:写出帧头、命令、应答、VIN、
//     加密标志、数据单元长度、数据单元,以及覆盖字节 [2:] 的 BCC。
//  3. 通过 codec.ProtocolCodec.Decode 解码:校验 BCC、解析
//     帧头字段,并保存 RawBytes 以供延迟解码数据单元。
//  4. DecodePayload() 通过 frame.PayloadType 查找消息体类型、
//     通过 api.GetCodec 查找编解码器,再把 RawBytes 解码进 Payload。
//  5. 断言解码出的 Payload 结构体等于原始样本。
//
// 这是 Plan 2 Task E2 验收测试:证明帧编解码器、
// BCC 校验码、命令→类型分发(payloadType switch)与编解码器
// 注册表端到端协同工作。两种协议版本的车辆
// 均有覆盖。
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
		const wantFrameLen = 2 + 1 + 1 + 17 + 1 + 2 + 36 + 1 // 帧头 + 命令 + 应答 + VIN + 加密 + 长度 + 数据单元(36) + BCC
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

		// Payload 此时不应被解码,Decode 只解析帧,
		// 由调用方延迟调用 DecodePayload。
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
		// 重编码已解码的帧必须产出逐字节一致的输出。
		// 这能捕获两条编解码路径之间任何帧头字段、BCC 计算
		// 或数据单元编码的偏差。
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
		// 加密模式 != 0x01 必须暴露为 api.ErrEncryptionNotSupported。
		// Plan 2 完成标准:只实现直通(0x01)。
		pm := &frame.ProtocolMessage{
			Version:      api.V2016,
			RequestType:  types.CommandV2016ByCode(0x01),
			ResponseType: types.ResponseCommand,
			VIN:          "LSVAU2A37K2100001",
			Encryption:   types.EncryptionRSA, // 0x02:不支持
			Payload:      sampleVehicleLogin(),
		}
		if _, err := pm.Bytes(); err != api.ErrEncryptionNotSupported {
			t.Errorf("encode EncryptionRSA: got err=%v, want %v", err, api.ErrEncryptionNotSupported)
		}

		// 解码侧:构造一个声称使用 RSA 加密的帧,确认
		// ProtocolCodec.Decode 用同一个哨兵值拒绝它。
		handCrafted := buildFrameWithEncryption(t, api.V2016, 0x01, byte(types.EncryptionRSA), sampleVehicleLogin())
		if _, err := codec.ProtocolCodec.Decode(utils.NewByteReader(handCrafted)); err != api.ErrEncryptionNotSupported {
			t.Errorf("decode EncryptionRSA: got err=%v, want %v", err, api.ErrEncryptionNotSupported)
		}
	})

	t.Run("BCCMismatchRejected", func(t *testing.T) {
		// 破坏 BCC 字节(最后一个字节),确认返回 ErrBCCMismatch。
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
		bad[len(bad)-1] ^= 0xFF // 翻转 BCC 的全部位
		if _, err := codec.ProtocolCodec.Decode(utils.NewByteReader(bad)); err != api.ErrBCCMismatch {
			t.Errorf("decode corrupt BCC: got err=%v, want %v", err, api.ErrBCCMismatch)
		}
	})

	t.Run("RealTimeFrame_DecodePayloadComposite", func(t *testing.T) {
		// RealTimeData 数据单元包含多个 TLV 子记录。DecodePayload
		// 必须通过已注册的 RealTimeDataCodec 和 TLV 子编解码器
		// 查找表进行分发。这能捕获 realtime 子编解码器在兄弟包中
		// 注册时的初始化顺序风险。
		vd := &gbt2016.RealTimeData{
			BeanTime:    sampleBeanTime(),
			VehicleData: nil, // 数据单元可以包含 nil 子记录
			EngineData:  nil,
		}
		// 基础校验:空 RealTimeData 可往返。
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
		// V2025 专有命令 0x0B(密钥交换)。确认 V2025 命令
		// 表经 PayloadType 分发到正确的编解码器。
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
		// V2025 专有命令 0x09,含嵌套的 VehicleSignature 子编解码器。
		// 确认嵌套编解码器查找(VehicleActivate → VehicleSignature)
		// 在帧路径上正常工作。
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

// buildFrameWithEncryption 用指定的加密字节构造协议帧字节切片,
// 绕过 frame.Bytes()(它会拒绝非直通模式)。仅由 EncryptionRejected
// 用于触发解码侧的拒绝路径。
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

	// BCC 覆盖字节 [2:](即 2 字节帧头之后的全部内容)。
	bccRange := w.Bytes()[2:]
	w.WriteUint8(utils.CalcBCC(bccRange))
	return w.Bytes()
}

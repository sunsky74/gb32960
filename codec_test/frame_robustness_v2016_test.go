package codec_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	_ "github.com/sunsky74/gb32960/codec/all"
	"github.com/sunsky74/gb32960/frame"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2016"
	mdl2025 "github.com/sunsky74/gb32960/model/gbt2025"
	"github.com/sunsky74/gb32960/types"
	"github.com/sunsky74/gb32960/utils"
)

// testVIN 是帧健壮性测试共用的 17 字节 VIN。
const testVIN = "LSVAU2A37K2100001"

// buildRobustnessFrame 组装一个带显式声明数据单元长度的原始帧,
// 绕过 frame.Bytes(),以便构造畸形的长度字段。
// BCC 覆盖字节 [2:](即 2 字节帧头之后的全部内容)。
func buildRobustnessFrame(header string, cmd, response byte, declaredLen int, body []byte) []byte {
	w := utils.NewByteWriter()
	w.WriteString(header, 2)
	w.WriteUint8(cmd)
	w.WriteUint8(response)
	w.WriteString(testVIN, 17)
	w.WriteUint8(byte(types.EncryptionNone))
	w.WriteUint16(uint16(declaredLen))
	w.WriteBytes(body)
	w.WriteUint8(utils.CalcBCC(w.Bytes()[2:]))
	return w.Bytes()
}

// decodePayloadNoPanic 在 recover 保护下调用 DecodePayload:C1 修复
// 要求对未知命令字节返回错误,而绝不能 panic。
func decodePayloadNoPanic(t *testing.T, pm *frame.ProtocolMessage) (err error) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DecodePayload panicked for RequestType %#v: %v", pm.RequestType, r)
		}
	}()
	return pm.DecodePayload()
}

// TestFrame_UnknownCommandByte_NoPanic 覆盖 audit C1:字节 0x00/0xFF 映射为
// 类型化的 nil *types.CommandV2016 / *types.CommandV2025;Decode 必须成功,
// 而 DecodePayload 必须返回错误,而不是解引用 nil 命令。
func TestFrame_UnknownCommandByte_NoPanic(t *testing.T) {
	cases := []struct {
		name   string
		header string
		cmd    byte
	}{
		{"V2016_0x00", types.HeaderV2016, 0x00},
		{"V2016_0xFF", types.HeaderV2016, 0xFF},
		{"V2025_0xFF", types.HeaderV2025, 0xFF},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := buildRobustnessFrame(tc.header, tc.cmd, byte(types.ResponseCommand), 0, nil)
			decoded, err := codec.ProtocolCodec.Decode(utils.NewByteReader(raw))
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			pm, ok := decoded.(*frame.ProtocolMessage)
			if !ok {
				t.Fatalf("decoded type: %T, want *frame.ProtocolMessage", decoded)
			}
			if code, ok := frame.CommandCode(pm.RequestType); ok {
				t.Fatalf("CommandCode: got (0x%02X, true) for unknown byte, want (0, false)", code)
			}
			if err := decodePayloadNoPanic(t, pm); err == nil {
				t.Fatalf("DecodePayload: got nil error for unknown command byte 0x%02X, want error", tc.cmd)
			}
		})
	}
}

// TestFrame_EmptyBodyHeartbeatClock 覆盖 audit H3:按 GB/T 32960 附录 B,
// 0x07/0x08 携带 0 字节数据单元,必须能以 nil Payload 编码,
// 而带数据单元的命令(0x01)使用 nil Payload 必须被拒绝。
func TestFrame_EmptyBodyHeartbeatClock(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  byte
	}{
		{"heartbeat_0x07", 0x07},
		{"clock_correct_0x08", 0x08},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pm := &frame.ProtocolMessage{
				Version:      api.V2016,
				RequestType:  types.CommandV2016ByCode(tc.cmd),
				ResponseType: types.ResponseCommand,
				VIN:          testVIN,
				Encryption:   types.EncryptionNone,
				Payload:      nil,
			}

			raw, err := pm.Bytes()
			if err != nil {
				t.Fatalf("Bytes: %v", err)
			}
			// 帧头2 + 命令1 + 应答1 + VIN17 + 加密1 + 长度2 + 数据单元0 + BCC1
			const wantLen = 2 + 1 + 1 + 17 + 1 + 2 + 0 + 1
			if len(raw) != wantLen {
				t.Fatalf("frame length: got %d, want %d (bytes=%X)", len(raw), wantLen, raw)
			}

			decoded, err := codec.ProtocolCodec.Decode(utils.NewByteReader(raw))
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			df, ok := decoded.(*frame.ProtocolMessage)
			if !ok {
				t.Fatalf("decoded type: %T, want *frame.ProtocolMessage", decoded)
			}
			if df.PayloadLength != 0 {
				t.Errorf("PayloadLength: got %d, want 0", df.PayloadLength)
			}
			if df.Version != api.V2016 {
				t.Errorf("Version: got %v, want V2016", df.Version)
			}
			gotCmd, ok := frame.CommandCode(df.RequestType)
			if !ok || gotCmd != tc.cmd {
				t.Errorf("CommandCode: got (0x%02X, %v), want (0x%02X, true)", gotCmd, ok, tc.cmd)
			}
			if df.Payload != nil {
				t.Errorf("Payload should be nil before DecodePayload, got %T", df.Payload)
			}
			// 无数据单元的命令解码得到 nil Payload,且不报错。
			if err := df.DecodePayload(); err != nil {
				t.Errorf("DecodePayload on body-less command: %v", err)
			}

			// codec.Encode 路径必须表现一致(audit H3)。
			w := utils.NewByteWriter()
			if err := codec.ProtocolCodec.Encode(w, pm); err != nil {
				t.Fatalf("ProtocolCodec.Encode: %v", err)
			}
			if got := len(w.Bytes()); got != wantLen {
				t.Fatalf("Encode frame length: got %d, want %d", got, wantLen)
			}
		})
	}

	t.Run("nil_payload_requires_body_0x01", func(t *testing.T) {
		pm := &frame.ProtocolMessage{
			Version:      api.V2016,
			RequestType:  types.CommandV2016ByCode(0x01),
			ResponseType: types.ResponseCommand,
			VIN:          testVIN,
			Encryption:   types.EncryptionNone,
			Payload:      nil,
		}
		if _, err := pm.Bytes(); err == nil {
			t.Fatal("Bytes: got nil error for nil Payload on command 0x01, want error")
		}
		w := utils.NewByteWriter()
		if err := codec.ProtocolCodec.Encode(w, pm); err == nil {
			t.Fatal("ProtocolCodec.Encode: got nil error for nil Payload on command 0x01, want error")
		}
	})
}

// TestFrame_V2016PlatformLoginLogoutDispatch 覆盖 audit H1(a):0x05/0x06 必须
// 端到端分发到已注册的 gbt2016 平台登入/登出编解码器。
func TestFrame_V2016PlatformLoginLogoutDispatch(t *testing.T) {
	t.Run("0x05_platform_login", func(t *testing.T) {
		want := &mdl.PlatformLogin{
			BeanTime:  model.BeanTime{Year: 26, Month: 9, Day: 17, Hour: 12, Minute: 0, Second: 0},
			SerialNum: 42,
			Username:  "user01",
			Password:  "pass01",
			Cipher:    0x01,
		}
		pm := &frame.ProtocolMessage{
			Version:      api.V2016,
			RequestType:  types.CommandV2016ByCode(0x05),
			ResponseType: types.ResponseCommand,
			VIN:          testVIN,
			Encryption:   types.EncryptionNone,
			Payload:      want,
		}

		raw, err := pm.Bytes()
		if err != nil {
			t.Fatalf("Bytes: %v", err)
		}
		// 帧头2 + 命令1 + 应答1 + VIN17 + 加密1 + 长度2 + 数据单元41 + BCC1
		const wantLen = 2 + 1 + 1 + 17 + 1 + 2 + 41 + 1
		if len(raw) != wantLen {
			t.Fatalf("frame length: got %d, want %d", len(raw), wantLen)
		}

		decoded, err := codec.ProtocolCodec.Decode(utils.NewByteReader(raw))
		if err != nil {
			t.Fatalf("Decode: %v", err)
		}
		df, ok := decoded.(*frame.ProtocolMessage)
		if !ok {
			t.Fatalf("decoded type: %T, want *frame.ProtocolMessage", decoded)
		}
		if df.PayloadLength != 41 {
			t.Errorf("PayloadLength: got %d, want 41", df.PayloadLength)
		}
		if err := df.DecodePayload(); err != nil {
			t.Fatalf("DecodePayload: %v", err)
		}
		got, ok := df.Payload.(*mdl.PlatformLogin)
		if !ok {
			t.Fatalf("Payload type: got %T, want *gbt2016.PlatformLogin", df.Payload)
		}
		assertInt(t, "SerialNum", got.SerialNum, want.SerialNum)
		if got.BeanTime != want.BeanTime {
			t.Errorf("BeanTime: got %v, want %v", got.BeanTime, want.BeanTime)
		}
		if got.Username != want.Username {
			t.Errorf("Username: got %q, want %q", got.Username, want.Username)
		}
		if got.Password != want.Password {
			t.Errorf("Password: got %q, want %q", got.Password, want.Password)
		}
		if got.Cipher != want.Cipher {
			t.Errorf("Cipher: got 0x%02X, want 0x%02X", got.Cipher, want.Cipher)
		}
	})

	t.Run("0x06_platform_logout", func(t *testing.T) {
		want := &mdl.PlatformLogout{
			BeanTime:  model.BeanTime{Year: 26, Month: 9, Day: 17, Hour: 12, Minute: 30, Second: 5},
			SerialNum: 43,
		}
		pm := &frame.ProtocolMessage{
			Version:      api.V2016,
			RequestType:  types.CommandV2016ByCode(0x06),
			ResponseType: types.ResponseCommand,
			VIN:          testVIN,
			Encryption:   types.EncryptionNone,
			Payload:      want,
		}

		raw, err := pm.Bytes()
		if err != nil {
			t.Fatalf("Bytes: %v", err)
		}
		// 帧头2 + 命令1 + 应答1 + VIN17 + 加密1 + 长度2 + 数据单元8 + BCC1
		const wantLen = 2 + 1 + 1 + 17 + 1 + 2 + 8 + 1
		if len(raw) != wantLen {
			t.Fatalf("frame length: got %d, want %d", len(raw), wantLen)
		}

		decoded, err := codec.ProtocolCodec.Decode(utils.NewByteReader(raw))
		if err != nil {
			t.Fatalf("Decode: %v", err)
		}
		df := decoded.(*frame.ProtocolMessage)
		if err := df.DecodePayload(); err != nil {
			t.Fatalf("DecodePayload: %v", err)
		}
		got, ok := df.Payload.(*mdl.PlatformLogout)
		if !ok {
			t.Fatalf("Payload type: got %T, want *gbt2016.PlatformLogout", df.Payload)
		}
		assertInt(t, "SerialNum", got.SerialNum, want.SerialNum)
		if got.BeanTime != want.BeanTime {
			t.Errorf("BeanTime: got %v, want %v", got.BeanTime, want.BeanTime)
		}
	})
}

// TestFrame_V2025DispatchAligned 覆盖 audit H1(b)/(c):V2025 的 0x04/0x05/0x06
// 必须映射到其编解码器确实已在 V2025 下注册的类型
// (0x04 = 与 V2016 共用的车辆登出编解码器,0x05/0x06 = V2025 平台编解码器)。
func TestFrame_V2025DispatchAligned(t *testing.T) {
	cases := []struct {
		cmd  byte
		want reflect.Type
	}{
		{0x04, reflect.TypeOf((*mdl.VehicleLogout)(nil)).Elem()},
		{0x05, reflect.TypeOf((*mdl2025.PlatformLoginV2025)(nil)).Elem()},
		{0x06, reflect.TypeOf((*mdl2025.PlatformLogoutV2025)(nil)).Elem()},
	}
	for _, tc := range cases {
		got := frame.PayloadType(api.V2025, tc.cmd)
		if got == nil {
			t.Fatalf("PayloadType(V2025, 0x%02X) = nil, want %v", tc.cmd, tc.want)
		}
		if got != tc.want {
			t.Errorf("PayloadType(V2025, 0x%02X) = %v, want %v", tc.cmd, got, tc.want)
		}
		if c := api.GetCodec(api.V2025, got); c == nil {
			t.Errorf("GetCodec(V2025, %v) = nil for cmd 0x%02X, want registered codec", got, tc.cmd)
		}
	}
}

// TestFrame_ResponseUnknownPreserved 覆盖 audit M7:未定义的应答字节
// 保留其原始值,便于调用方识别,且 decode->encode 保持字节忠实,
// 已定义的编码则保持不变。
func TestFrame_ResponseUnknownPreserved(t *testing.T) {
	for _, code := range []byte{0x00, 0x08, 0x42, 0xFD, 0xFF} {
		got := types.ResponseByCode(code)
		if got.Code() != code {
			t.Errorf("ResponseByCode(0x%02X).Code() = 0x%02X, want 0x%02X", code, got.Code(), code)
		}
		if got == types.ResponseCommand {
			t.Errorf("ResponseByCode(0x%02X) = ResponseCommand, want raw byte preserved", code)
		}
	}

	known := []struct {
		code byte
		want types.ResponseType
	}{
		{0x01, types.ResponseSuccess},
		{0x02, types.ResponseFailed},
		{0x03, types.ResponseVINDup},
		{0x04, types.ResponseVINNotExist},
		{0x05, types.ResponseSignErr},
		{0x06, types.ResponseStructureErr},
		{0x07, types.ResponseDecodeErr},
		{0xFE, types.ResponseCommand},
	}
	for _, tc := range known {
		if got := types.ResponseByCode(tc.code); got != tc.want {
			t.Errorf("ResponseByCode(0x%02X) = %v, want %v", tc.code, got, tc.want)
		}
	}
}

// oversizedBody 是一个最小化的 model.MessageBody,用于触发
// 65531 字节数据单元长度守卫,而无需手工构造巨型帧。
type oversizedBody struct{ size int }

func (b oversizedBody) Version() api.GBTVersion { return api.V2016 }
func (b oversizedBody) Bytes() ([]byte, error)  { return make([]byte, b.size), nil }

// TestFrame_LengthMismatch 覆盖 audit L2:声明长度与实际长度不一致,
// 以及长度超出规范范围 0~65531,在解码和编码两条路径上
// 都必须被拒绝。
func TestFrame_LengthMismatch(t *testing.T) {
	t.Run("decode_trailing_bytes", func(t *testing.T) {
		// 声明长度为 0,但 BCC 之前多了一个垃圾数据字节。
		raw := buildRobustnessFrame(types.HeaderV2016, 0x07, byte(types.ResponseCommand), 0, []byte{0xAA})
		if _, err := codec.ProtocolCodec.Decode(utils.NewByteReader(raw)); !errors.Is(err, api.ErrLengthMismatch) {
			t.Fatalf("Decode: got %v, want api.ErrLengthMismatch", err)
		}
	})

	t.Run("decode_declared_over_range", func(t *testing.T) {
		raw := buildRobustnessFrame(types.HeaderV2016, 0x07, byte(types.ResponseCommand), 65532, nil)
		if _, err := codec.ProtocolCodec.Decode(utils.NewByteReader(raw)); !errors.Is(err, api.ErrLengthMismatch) {
			t.Fatalf("Decode: got %v, want api.ErrLengthMismatch", err)
		}
	})

	t.Run("encode_payload_over_range", func(t *testing.T) {
		pm := &frame.ProtocolMessage{
			Version:      api.V2016,
			RequestType:  types.CommandV2016ByCode(0x01),
			ResponseType: types.ResponseCommand,
			VIN:          testVIN,
			Encryption:   types.EncryptionNone,
			Payload:      oversizedBody{size: 65532},
		}
		if _, err := pm.Bytes(); !errors.Is(err, api.ErrLengthMismatch) {
			t.Fatalf("Bytes: got %v, want api.ErrLengthMismatch", err)
		}
		w := utils.NewByteWriter()
		if err := codec.ProtocolCodec.Encode(w, pm); !errors.Is(err, api.ErrLengthMismatch) {
			t.Fatalf("ProtocolCodec.Encode: got %v, want api.ErrLengthMismatch", err)
		}
	})

	t.Run("encode_payload_at_limit_ok", func(t *testing.T) {
		pm := &frame.ProtocolMessage{
			Version:      api.V2016,
			RequestType:  types.CommandV2016ByCode(0x01),
			ResponseType: types.ResponseCommand,
			VIN:          testVIN,
			Encryption:   types.EncryptionNone,
			Payload:      oversizedBody{size: 65531},
		}
		raw, err := pm.Bytes()
		if err != nil {
			t.Fatalf("Bytes at 65531-byte payload: %v", err)
		}
		const wantLen = 2 + 1 + 1 + 17 + 1 + 2 + 65531 + 1
		if len(raw) != wantLen {
			t.Fatalf("frame length: got %d, want %d", len(raw), wantLen)
		}
	})
}

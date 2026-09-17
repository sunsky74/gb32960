package codec_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	_ "github.com/sunsky74/gb32960/codec/all"
	"github.com/sunsky74/gb32960/frame"
	"github.com/sunsky74/gb32960/model/gbt2016"
	"github.com/sunsky74/gb32960/types"
	"github.com/sunsky74/gb32960/utils"
)

// dataUnitOffset 是数据单元在帧中的起始偏移:
// 起始符2 + 命令1 + 应答1 + VIN17 + 加密1 + 长度2。
const dataUnitOffset = 2 + 1 + 1 + 17 + 1 + 2

// TestFrameFixWP_CommandCodeFidelity 锁定工作包 B(命令码保真 + 数据单元透传):
// 范围内预留/平台交换自定义命令(如 0x50、0xC5)经 Decode 后,命令字节
// 必须保留实际线码(而不是范围基值),数据单元必须以原始字节透传,
// Decode -> Bytes()/Encode 必须逐字节一致。
//
// 规范依据:
//   - 2016.md 表3 L55-65:0x09~0x7F 上行数据系统预留、0x83~0xBF 下行
//     数据系统预留、0xC0~0xFE 平台交换自定义数据;
//   - 2025.md 表3 L49-62:0x0C~0x7F 上行数据系统预留、0x83~0xBF 下行
//     数据系统预留、0xC0~0xFE 平台交换自定义数据。
func TestFrameFixWP_CommandCodeFidelity(t *testing.T) {
	cases := []struct {
		name   string
		header string
		cmd    byte
		body   []byte
	}{
		// 2025:上行预留范围 0x0C~0x7F 中的非基值线码。
		{"V2025_reserved_0x50", types.HeaderV2025, 0x50, []byte{0xA1, 0xB2, 0xC3, 0xD4}},
		// 2025:平台交换自定义数据 0xC0~0xFE。
		{"V2025_custom_0xC5", types.HeaderV2025, 0xC5, []byte{0x01, 0x02, 0x03}},
		// 2025:下行预留范围 0x83~0xBF。
		{"V2025_downlink_reserved_0x90", types.HeaderV2025, 0x90, []byte{0x55}},
		// 2016:上行预留范围 0x09~0x7F 中的非基值线码。
		{"V2016_reserved_0x50", types.HeaderV2016, 0x50, []byte{0xDE, 0xAD, 0xBE, 0xEF}},
		// 2016:平台交换自定义数据 0xC0~0xFE。
		{"V2016_custom_0xC5", types.HeaderV2016, 0xC5, []byte{0x10, 0x20, 0x30}},
		// 2016:下行预留范围 0x83~0xBF。
		{"V2016_downlink_reserved_0xA0", types.HeaderV2016, 0xA0, []byte{0x77, 0x88}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := buildRobustnessFrame(tc.header, tc.cmd, byte(types.ResponseCommand), len(tc.body), tc.body)
			if raw[2] != tc.cmd {
				t.Fatalf("test setup: raw command byte = 0x%02X, want 0x%02X", raw[2], tc.cmd)
			}

			decoded, err := codec.ProtocolCodec.Decode(utils.NewByteReader(raw))
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			pm, ok := decoded.(*frame.ProtocolMessage)
			if !ok {
				t.Fatalf("decoded type: %T, want *frame.ProtocolMessage", decoded)
			}

			// 命令单元保真:CommandCode 必须返回实际线码,而不是范围基值
			// (修复前 V2016 0x50 会变成 0x09、0xC5 会变成 0xC0)。
			gotCmd, ok := frame.CommandCode(pm.RequestType)
			if !ok || gotCmd != tc.cmd {
				t.Fatalf("CommandCode: got (0x%02X, %v), want (0x%02X, true)", gotCmd, ok, tc.cmd)
			}

			// 无注册消息体类型的命令:DecodePayload 必须是 no-op,
			// Payload 保持 nil。
			if err := pm.DecodePayload(); err != nil {
				t.Fatalf("DecodePayload: %v", err)
			}
			if pm.Payload != nil {
				t.Fatalf("Payload: got %T, want nil for code without registered body type", pm.Payload)
			}

			// frame.ProtocolMessage.Bytes():数据单元原始透传 + 命令字节保真。
			out, err := pm.Bytes()
			if err != nil {
				t.Fatalf("ProtocolMessage.Bytes: %v", err)
			}
			if !bytes.Equal(raw, out) {
				t.Fatalf("Bytes round-trip not byte-identical:\n  raw = %X\n  out = %X", raw, out)
			}
			if !bytes.Equal(out[dataUnitOffset:dataUnitOffset+len(tc.body)], tc.body) {
				t.Errorf("data unit not preserved: got %X, want %X",
					out[dataUnitOffset:], tc.body)
			}

			// codec.ProtocolCodec.Encode:与 Bytes() 保持逐字节一致。
			w := utils.NewByteWriter()
			if err := codec.ProtocolCodec.Encode(w, pm); err != nil {
				t.Fatalf("ProtocolCodec.Encode: %v", err)
			}
			if !bytes.Equal(raw, w.Bytes()) {
				t.Errorf("Encode round-trip not byte-identical:\n  raw = %X\n  enc = %X",
					raw, w.Bytes())
			}
		})
	}

	// ByCode 复制语义直接断言:范围项返回值携带实际线码(修复前为范围基值)。
	t.Run("ByCode_returns_actual_code", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			got  byte
			want byte
		}{
			{"V2016_range_0x50", types.CommandV2016ByCode(0x50).Code, 0x50},
			{"V2016_custom_0xC5", types.CommandV2016ByCode(0xC5).Code, 0xC5},
			{"V2025_range_0x50", types.CommandV2025ByCode(0x50).Code, 0x50},
			{"V2025_custom_0xC5", types.CommandV2025ByCode(0xC5).Code, 0xC5},
		} {
			if tc.got != tc.want {
				t.Errorf("%s: Code = 0x%02X, want 0x%02X", tc.name, tc.got, tc.want)
			}
		}
		// 多次调用必须稳定:证明静态范围表未被首次调用改写。
		if got := types.CommandV2016ByCode(0x50).Code; got != 0x50 {
			t.Errorf("second V2016ByCode(0x50).Code = 0x%02X, want 0x50", got)
		}
		if got := types.CommandV2025ByCode(0xC5).Code; got != 0xC5 {
			t.Errorf("second V2025ByCode(0xC5).Code = 0x%02X, want 0xC5", got)
		}
	})
}

// TestFrameFixWP_InvalidStartFlagRejected 锁定工作包 B(起始符校验):
// 起始符不是 "##"/"$$" 的帧必须返回 api.ErrInvalidHeader;
// 不足 2 字节的截断输入保持既有的 ErrBufferUnderflow 语义。
//
// 规范依据:2016.md 表2 L31(同表B.1 L614)与 2025.md 表2 L32 ——
// 起始符固定为 "##"(0x23,0x23)或 "$$"(0x24,0x24)。
func TestFrameFixWP_InvalidStartFlagRejected(t *testing.T) {
	t.Run("header_XY", func(t *testing.T) {
		// 0x58 0x59("XY")既不是 "##" 也不是 "$$"。
		raw := buildRobustnessFrame("XY", 0x01, byte(types.ResponseCommand), 0, nil)
		if _, err := codec.ProtocolCodec.Decode(utils.NewByteReader(raw)); !errors.Is(err, api.ErrInvalidHeader) {
			t.Fatalf("Decode: got err=%v, want api.ErrInvalidHeader", err)
		}
	})

	t.Run("header_zero", func(t *testing.T) {
		raw := buildRobustnessFrame("\x00\x00", 0x07, byte(types.ResponseCommand), 0, nil)
		if _, err := codec.ProtocolCodec.Decode(utils.NewByteReader(raw)); !errors.Is(err, api.ErrInvalidHeader) {
			t.Fatalf("Decode: got err=%v, want api.ErrInvalidHeader", err)
		}
	})

	t.Run("truncated_header_underflow", func(t *testing.T) {
		// 只有 1 个字节("##" 的残片):读取下溢,必须保持 ErrBufferUnderflow,
		// 不能被改报为帧头错误。
		if _, err := codec.ProtocolCodec.Decode(utils.NewByteReader([]byte{0x23})); !errors.Is(err, api.ErrBufferUnderflow) {
			t.Fatalf("Decode: got err=%v, want api.ErrBufferUnderflow", err)
		}
	})
}

// TestFrameFixWP_HeartbeatEmptyBodyUnchanged 锁定工作包 B(透传不改变既有行为):
// 0x07 心跳携带 0 字节数据单元,RawBytes 为空,Decode -> Bytes()/Encode
// 必须保持逐字节一致,不得凭空注入数据单元。
func TestFrameFixWP_HeartbeatEmptyBodyUnchanged(t *testing.T) {
	raw := buildRobustnessFrame(types.HeaderV2016, 0x07, byte(types.ResponseCommand), 0, nil)
	decoded, err := codec.ProtocolCodec.Decode(utils.NewByteReader(raw))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	pm, ok := decoded.(*frame.ProtocolMessage)
	if !ok {
		t.Fatalf("decoded type: %T, want *frame.ProtocolMessage", decoded)
	}
	if pm.PayloadLength != 0 || len(pm.RawBytes) != 0 {
		t.Fatalf("heartbeat: PayloadLength=%d, len(RawBytes)=%d, want 0/0",
			pm.PayloadLength, len(pm.RawBytes))
	}
	if err := pm.DecodePayload(); err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}

	out, err := pm.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if !bytes.Equal(raw, out) {
		t.Errorf("heartbeat round-trip not byte-identical:\n  raw = %X\n  out = %X", raw, out)
	}

	w := utils.NewByteWriter()
	if err := codec.ProtocolCodec.Encode(w, pm); err != nil {
		t.Fatalf("ProtocolCodec.Encode: %v", err)
	}
	if !bytes.Equal(raw, w.Bytes()) {
		t.Errorf("heartbeat Encode not byte-identical:\n  raw = %X\n  enc = %X", raw, w.Bytes())
	}
}

// TestFrameFixWP_KnownCommandUnchanged 锁定工作包 B(既有行为不变):
// 0x01 车辆登入等已注册消息体类型的命令仍走 Payload.Bytes() 路径,
// RawBytes 不参与编码;Decode -> DecodePayload -> Bytes 逐字节一致,
// 且"PayloadType != nil 时 Payload 必须非 nil"的校验不被透传分支绕过。
func TestFrameFixWP_KnownCommandUnchanged(t *testing.T) {
	want := sampleVehicleLogin()
	pm := &frame.ProtocolMessage{
		Version:      api.V2016,
		RequestType:  types.CommandV2016ByCode(0x01),
		ResponseType: types.ResponseCommand,
		VIN:          testVIN,
		Encryption:   types.EncryptionNone,
		Payload:      want,
	}
	original, err := pm.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	decoded, err := codec.ProtocolCodec.Decode(utils.NewByteReader(original))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	df, ok := decoded.(*frame.ProtocolMessage)
	if !ok {
		t.Fatalf("decoded type: %T, want *frame.ProtocolMessage", decoded)
	}
	if got, ok := frame.CommandCode(df.RequestType); !ok || got != 0x01 {
		t.Fatalf("CommandCode: got (0x%02X, %v), want (0x01, true)", got, ok)
	}
	if err := df.DecodePayload(); err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}
	vl, ok := df.Payload.(*gbt2016.VehicleLogin)
	if !ok {
		t.Fatalf("Payload type: got %T, want *gbt2016.VehicleLogin", df.Payload)
	}
	if vl.ICCID != want.ICCID {
		t.Errorf("ICCID: got %q, want %q", vl.ICCID, want.ICCID)
	}

	// 带 Payload 的命令:RawBytes 不参与,重编码仍逐字节一致。
	out, err := df.Bytes()
	if err != nil {
		t.Fatalf("re-encode: %v", err)
	}
	if !bytes.Equal(original, out) {
		t.Errorf("known-command round-trip not byte-identical:\n  original = %X\n  out = %X",
			original, out)
	}

	// 已注册消息体但 Payload 尚为 nil(懒解码状态)时,即使 RawBytes 非空,
	// 也必须继续报错 —— 透传分支不得绕过既有的消息体必填校验。
	t.Run("payload_required_when_type_registered", func(t *testing.T) {
		raw := buildRobustnessFrame(types.HeaderV2016, 0x01, byte(types.ResponseCommand), 4,
			[]byte{0x01, 0x02, 0x03, 0x04})
		msg, err := codec.ProtocolCodec.Decode(utils.NewByteReader(raw))
		if err != nil {
			t.Fatalf("Decode: %v", err)
		}
		lazyPM, ok := msg.(*frame.ProtocolMessage)
		if !ok {
			t.Fatalf("decoded type: %T, want *frame.ProtocolMessage", msg)
		}
		if len(lazyPM.RawBytes) == 0 {
			t.Fatal("test setup: RawBytes should be non-empty")
		}
		if _, err := lazyPM.Bytes(); err == nil {
			t.Fatal("Bytes: got nil error for nil Payload on registered 0x01, want error")
		}
		w := utils.NewByteWriter()
		if err := codec.ProtocolCodec.Encode(w, lazyPM); err == nil {
			t.Fatal("ProtocolCodec.Encode: got nil error for nil Payload on registered 0x01, want error")
		}
	})
}

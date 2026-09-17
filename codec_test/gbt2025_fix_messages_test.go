package codec_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/model/gbt2025"
	v2025rt "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/utils"

	// 聚合导入:触发所有编解码器的 init() 注册。
	_ "github.com/sunsky74/gb32960/codec/all"
)

// 本文件覆盖 2026-09-17 审计的 WP-E 修复:
//  1. 激活应答 激活状态 语义:0x01=激活成功,0x02=激活失败(表B.4);
//  2. 登入(V2025)/密钥交换/车辆签名/车辆激活 四个编解码器的编码侧
//     长度一致性校验(长度不符必须报错,而不是静默写出错位帧)。
//
// 合法样本的逐字节往返稳定性由 roundtrip_test.go 继续覆盖,此处只补一条
// 自包含的编码→解码→重编码校验,并固守 WP-E 的行为边界。

// fix25BeanTime 返回本文件使用的固定时间,不依赖其他测试文件的辅助函数。
func fix25BeanTime() model.BeanTime {
	return model.BeanTime{Year: 26, Month: 9, Day: 17, Hour: 10, Minute: 30, Second: 0}
}

// fix25Codec 查找 msg 对应协议版本已注册的编解码器。
func fix25Codec(t *testing.T, version api.GBTVersion, msg api.Message) api.Codecer {
	t.Helper()
	tp := reflect.TypeOf(msg)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	cdc := api.GetCodec(version, tp)
	if cdc == nil {
		t.Fatalf("codec not found for %v", tp)
	}
	return cdc
}

// fix25Encode 编码 msg,同时返回写出的字节(即使出错也返回,便于断言
// 校验失败时没有写出部分帧)。
func fix25Encode(t *testing.T, version api.GBTVersion, msg api.Message) ([]byte, error) {
	t.Helper()
	w := utils.NewByteWriter()
	err := fix25Codec(t, version, msg).Encode(w, msg)
	return w.Bytes(), err
}

// fix25Decode 按 msg 类型解码 wire。
func fix25Decode(t *testing.T, version api.GBTVersion, msg api.Message, wire []byte) api.Message {
	t.Helper()
	decoded, err := fix25Codec(t, version, msg).Decode(utils.NewByteReader(wire))
	if err != nil {
		t.Fatalf("decode: %v (wire=%X)", err, wire)
	}
	return decoded
}

// fix25ValidSignature 返回长度自洽的签名子记录,供激活样本复用。
func fix25ValidSignature() *v2025rt.VehicleSignature {
	return &v2025rt.VehicleSignature{
		Type:    0x02,
		RLength: 2,
		RValue:  []byte{0xAA, 0xBB},
		SLength: 2,
		SValue:  []byte{0xCC, 0xDD},
	}
}

// TestActivateResponseSemantics 固守表B.4 激活状态语义:
// 解码 0x01→Success=true、非 0x01(0x02/0x00)→false;
// 编码 true→0x01、false→0x02,且信息 ResponseCode 原样保留。
func TestActivateResponseSemantics(t *testing.T) {
	t.Run("decode/0x01激活成功", func(t *testing.T) {
		got := fix25Decode(t, api.V2025, &gbt2025.VehicleActivateResponse{}, []byte{0x01, 0x00}).(*gbt2025.VehicleActivateResponse)
		if !got.Success {
			t.Errorf("Success: got false, want true (0x01=激活成功)")
		}
		if got.ResponseCode != 0x00 {
			t.Errorf("ResponseCode: got 0x%02X, want 0x00", got.ResponseCode)
		}
	})

	t.Run("decode/0x02激活失败", func(t *testing.T) {
		got := fix25Decode(t, api.V2025, &gbt2025.VehicleActivateResponse{}, []byte{0x02, 0x00}).(*gbt2025.VehicleActivateResponse)
		if got.Success {
			t.Errorf("Success: got true, want false (0x02=激活失败)")
		}
	})

	t.Run("decode/0x00未定义值按失败", func(t *testing.T) {
		got := fix25Decode(t, api.V2025, &gbt2025.VehicleActivateResponse{}, []byte{0x00, 0x00}).(*gbt2025.VehicleActivateResponse)
		if got.Success {
			t.Errorf("Success: got true, want false (0x00 未被表B.4 定义)")
		}
	})

	t.Run("decode/信息字段原样保留", func(t *testing.T) {
		got := fix25Decode(t, api.V2025, &gbt2025.VehicleActivateResponse{}, []byte{0x02, 0x02}).(*gbt2025.VehicleActivateResponse)
		if got.Success {
			t.Errorf("Success: got true, want false")
		}
		if got.ResponseCode != 0x02 {
			t.Errorf("ResponseCode: got 0x%02X, want 0x02(VIN重复)", got.ResponseCode)
		}
	})

	t.Run("encode/true写0x01", func(t *testing.T) {
		wire, err := fix25Encode(t, api.V2025, &gbt2025.VehicleActivateResponse{Success: true, ResponseCode: 0x00})
		if err != nil {
			t.Fatalf("encode: %v", err)
		}
		if want := []byte{0x01, 0x00}; !bytes.Equal(wire, want) {
			t.Errorf("wire: got %X, want %X", wire, want)
		}
	})

	t.Run("encode/false写0x02", func(t *testing.T) {
		wire, err := fix25Encode(t, api.V2025, &gbt2025.VehicleActivateResponse{Success: false, ResponseCode: 0x02})
		if err != nil {
			t.Fatalf("encode: %v", err)
		}
		if want := []byte{0x02, 0x02}; !bytes.Equal(wire, want) {
			t.Errorf("wire: got %X, want %X", wire, want)
		}
	})
}

// TestEncodeLengthConsistencyRejectsMismatch 固守四个编解码器的编码侧
// 长度一致性校验:长度不符必须返回错误,且校验发生在任何写入之前
// (失败时线上字节数为 0,不会留下部分帧)。
func TestEncodeLengthConsistencyRejectsMismatch(t *testing.T) {
	cases := []struct {
		name string
		msg  api.Message
	}{
		{
			"login/Lengths与Count不符",
			&gbt2025.VehicleLoginV2025{
				BeanTime: fix25BeanTime(), SerialNum: 1, ICCID: "89860000000000000002",
				Count: 2, Lengths: []int{1}, Codes: []string{"A"},
			},
		},
		{
			"login/Codes与sum(Lengths)不符",
			&gbt2025.VehicleLoginV2025{
				BeanTime: fix25BeanTime(), SerialNum: 1, ICCID: "89860000000000000002",
				Count: 2, Lengths: []int{1, 1}, Codes: []string{"A"},
			},
		},
		{
			"login/编码超过24字节",
			&gbt2025.VehicleLoginV2025{
				BeanTime: fix25BeanTime(), SerialNum: 1, ICCID: "89860000000000000002",
				Count: 1, Lengths: []int{1}, Codes: []string{strings.Repeat("X", 25)},
			},
		},
		{
			"keyExchange/密钥长度与声明不符",
			&gbt2025.KeyExchangeData{
				Type: 0x02, Length: 8, Key: []byte{0x01, 0x02},
				StartTime: fix25BeanTime(), ExpireTime: fix25BeanTime(),
			},
		},
		{
			"signature/R值长度与声明不符",
			&v2025rt.VehicleSignature{
				Type: 0x02, RLength: 3, RValue: []byte{0xAA, 0xBB},
				SLength: 2, SValue: []byte{0xCC, 0xDD},
			},
		},
		{
			"signature/S值长度与声明不符",
			&v2025rt.VehicleSignature{
				Type: 0x02, RLength: 2, RValue: []byte{0xAA, 0xBB},
				SLength: 3, SValue: []byte{0xCC, 0xDD},
			},
		},
		{
			"activate/公钥长度与声明不符",
			&gbt2025.VehicleActivate{
				CollectTime: fix25BeanTime(), ChipID: "CHIP2025ABC001",
				PublicKeyLength: 8, PublicKey: []byte{0x01, 0x02},
				VIN: "LSVAU2A37K2100001", Signature: fix25ValidSignature(),
			},
		},
		{
			"activate/芯片ID超过16字节",
			&gbt2025.VehicleActivate{
				CollectTime: fix25BeanTime(), ChipID: strings.Repeat("C", 17),
				PublicKeyLength: 0, PublicKey: nil,
				VIN: "LSVAU2A37K2100001", Signature: fix25ValidSignature(),
			},
		},
		{
			"activate/VIN超过17字节",
			&gbt2025.VehicleActivate{
				CollectTime: fix25BeanTime(), ChipID: "CHIP2025ABC001",
				PublicKeyLength: 0, PublicKey: nil,
				VIN: "LSVAU2A37K21000018", Signature: fix25ValidSignature(),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := fix25Encode(t, api.V2025, tc.msg)
			if err == nil {
				t.Fatalf("encode: got nil error, want length-consistency error (wire=%X)", wire)
			}
			if len(wire) != 0 {
				t.Errorf("encode: wrote %d bytes before failing, want 0 (no partial frame)", len(wire))
			}
		})
	}
}

// TestEncodeValidSamplesStayByteExact 确认新增校验不改变合法样本的
// 编码结果:编码→解码→重编码逐字节一致。
func TestEncodeValidSamplesStayByteExact(t *testing.T) {
	cases := []struct {
		name string
		msg  api.Message
	}{
		{"activateResponse", &gbt2025.VehicleActivateResponse{Success: true, ResponseCode: 0x00}},
		{"loginV2025", &gbt2025.VehicleLoginV2025{
			BeanTime: fix25BeanTime(), SerialNum: 5, ICCID: "89860000000000000002",
			Count: 2, Lengths: []int{1, 1}, Codes: []string{"BMU001", "BMU002"},
		}},
		{"keyExchange", &gbt2025.KeyExchangeData{
			Type: 0x02, Length: 4, Key: []byte{0xDE, 0xAD, 0xBE, 0xEF},
			StartTime: fix25BeanTime(), ExpireTime: fix25BeanTime(),
		}},
		{"vehicleSignature", fix25ValidSignature()},
		{"vehicleActivate", &gbt2025.VehicleActivate{
			CollectTime: fix25BeanTime(), ChipID: "CHIP2025ABC001",
			PublicKeyLength: 4, PublicKey: []byte{0x01, 0x02, 0x03, 0x04},
			VIN: "LSVAU2A37K2100001", Signature: fix25ValidSignature(),
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			first, err := fix25Encode(t, api.V2025, tc.msg)
			if err != nil {
				t.Fatalf("first encode: %v", err)
			}
			decoded := fix25Decode(t, api.V2025, tc.msg, first)
			second, err := fix25Encode(t, api.V2025, decoded)
			if err != nil {
				t.Fatalf("re-encode: %v", err)
			}
			if !bytes.Equal(first, second) {
				t.Errorf("re-encode differs:\n  first  = %X\n  second = %X", first, second)
			}
		})
	}
}

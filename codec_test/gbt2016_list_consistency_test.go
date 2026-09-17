package codec_test

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	"github.com/sunsky74/gb32960/frame"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/model/gbt2016"
	mdlrt "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/utils"

	// 聚合导入:触发所有编解码器的 init() 注册,使
	// api.GetCodec 查找(经由各模型的 Bytes() 方法)能够成功。
	_ "github.com/sunsky74/gb32960/codec/all"
)

// TestListEncodeCountMismatchErrors 锁定 audit 2026-09-17 (M4):所有线格式为
// "计数 + 列表"的 V2016 编解码器都必须拒绝计数与列表长度不一致的编码,
// 而不是写出畸形帧(旧代码在计数为 0 或小于列表长度时会
// 静默丢弃条目)。错误必须在写入任何字节之前抛出。
func TestListEncodeCountMismatchErrors(t *testing.T) {
	cases := []struct {
		name string
		mb   model.MessageBody
	}{
		{"MotorDataList", &mdlrt.MotorDataList{Count: 2, Items: []mdlrt.MotorData{{MotorSeq: 1}}}},
		{"ChargeableSubsystemElectricList", &mdlrt.ChargeableSubsystemElectricList{
			ElectricCount: 2,
			Items:         []mdlrt.ChargeableSubsystemElectric{{ChargeableSubSystemNumber: 1}},
		}},
		{"ChargeableSubsystemTemperatureList", &mdlrt.ChargeableSubsystemTemperatureList{
			TemperatureCount: 2,
			Items:            []mdlrt.ChargeableSubsystemTemperature{{SubSystemNumber: 1}},
		}},
		{"ChargeableSubsystemElectric", &mdlrt.ChargeableSubsystemElectric{BatteryCount: 2, BatteryVoltages: []float64{3.5}}},
		{"FuelCellData", &mdlrt.FuelCellData{TotalNumberOfFcTp: 2, ProbeTemperatureValues: []float64{25.0}}},
		{"VehicleLogin", &gbt2016.VehicleLogin{Count: 2, Length: 10, Codes: []string{"A"}}},
		{"VehicleLogin_CodeLongerThanLength", &gbt2016.VehicleLogin{Count: 1, Length: 2, Codes: []string{"TOOLONG"}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if raw, err := tc.mb.Bytes(); err == nil {
				t.Errorf("Encode succeeded (% X), want count/list mismatch error", raw)
			}
		})
	}
}

// TestListEncodeConsistentPasses 是 M4 校验的正向对照:
// 与列表长度一致的计数必须正常编码,且线格式长度为
// 预期值(1 字节计数 + 一条 12 字节电机条目)。
func TestListEncodeConsistentPasses(t *testing.T) {
	raw, err := (&mdlrt.MotorDataList{Count: 1, Items: []mdlrt.MotorData{{MotorSeq: 1}}}).Bytes()
	if err != nil {
		t.Fatalf("consistent encode failed: %v", err)
	}
	if want := 1 + 12; len(raw) != want {
		t.Errorf("encoded length: got %d, want %d", len(raw), want)
	}
}

// TestVehicleLoginLengthZeroNoCodes 锁定 audit 2026-09-17 (M4) 的后续修复:
// GB/T 32960-2016 表6 将 m=0 定义为 "不上传该编码",Count 可以大于 0 而
// 线格式上编码字节为零(真实生产报文 prod_login_v2016_01.hex
// 就编码了 Count=1、m=0)。早先的 Count==len(Codes) 检查错误地拒绝了
// 这一合法编码;只有 Length 非零时才要求两者一致。
func TestVehicleLoginLengthZeroNoCodes(t *testing.T) {
	// 合法的生产组合:Count=1、m=0,线格式上无编码。
	raw, err := (&gbt2016.VehicleLogin{Count: 1, Length: 0, Codes: nil}).Bytes()
	if err != nil {
		t.Fatalf("Count=1, Length=0, Codes=nil: encode failed: %v", err)
	}
	if want := 6 + 2 + 20 + 1 + 1; len(raw) != want {
		t.Errorf("encoded length: got %d, want %d", len(raw), want)
	}

	// Length=0 但存在 Codes 无法在线格式上表示:拒绝。
	if raw, err := (&gbt2016.VehicleLogin{Count: 1, Length: 0, Codes: []string{"A"}}).Bytes(); err == nil {
		t.Errorf("Count=1, Length=0, Codes=[A]: Encode succeeded (% X), want error", raw)
	}

	// 正向对照:Count=0 且 Length 非零、无编码时仍能编码。
	if _, err := (&gbt2016.VehicleLogin{Count: 0, Length: 10}).Bytes(); err != nil {
		t.Errorf("Count=0, Length=10: encode failed: %v", err)
	}
}

// TestGoldenV2016ReencodeByteExact 锁定端到端不变量:每个 V2016 金样报文
// 都必须逐字节一致地通过 decode -> DecodePayload -> Bytes。
// 它是 M4 登入回归(合法的 Count>0/m=0 生产报文在重编码时被拒绝)
// 以及其他 2016 修复的回归防护网。
func TestGoldenV2016ReencodeByteExact(t *testing.T) {
	for _, name := range []string{
		"prod_login_v2016_01.hex",
		"prod_logout_v2016_01.hex",
		"prod_realtime_v2016_01.hex",
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("..", "golden", "layer_a", name)
			hexData, err := os.ReadFile(path)
			if err != nil {
				t.Skipf("golden file not found: %s", path)
				return
			}
			raw, err := utils.HexToBytes(strings.TrimSpace(string(hexData)))
			if err != nil {
				t.Fatalf("hex decode failed: %v", err)
			}

			msg, err := codec.ProtocolCodec.Decode(utils.NewByteReader(raw))
			if err != nil {
				t.Fatalf("ProtocolCodec.Decode: %v", err)
			}
			pm := msg.(*frame.ProtocolMessage)
			if err := pm.DecodePayload(); err != nil {
				t.Fatalf("DecodePayload: %v", err)
			}

			out, err := pm.Bytes()
			if err != nil {
				t.Fatalf("re-encode failed: %v", err)
			}
			if !bytes.Equal(out, raw) {
				t.Errorf("re-encode not byte-exact: got %d bytes, want %d bytes", len(out), len(raw))
			}
		})
	}
}

// TestFuelCellProbeCountSentinelDecode 锁定 audit 2026-09-17 (M2):探针计数为
// WORD 哨兵值(0xFFFE 异常 / 0xFFFF 无效, GB/T 32960-2016 表 12/表 B.8)时
// 表示"无探针数据单元"。Decode 不得读取 N 个探针字节,否则会下溢
// 并导致计数之后的每个字段错位。重编码必须按哨兵值原样写出,
// 不带任何探针字节。
func TestFuelCellProbeCountSentinelDecode(t *testing.T) {
	t.Run("FuelCellData", func(t *testing.T) {
		payload := []byte{
			0x00, 0x00, // 燃料电池电压
			0x00, 0x00, // 燃料电池电流
			0x00, 0x00, // 燃料消耗率
			0xFF, 0xFF, // 温度探针总数 = 0xFFFF (无效)
			0x00, 0x64, // 氢系统中最高温度
			0x2A,       // 氢系统中最高温度探针代号
			0x01, 0x90, // 氢气最高浓度
			0x01,       // 氢气最高浓度传感器代号
			0x00, 0xC8, // 氢气最高压力
			0x02, // 氢气最高压力传感器代号
			0x03, // 高压 DC/DC 状态
		}

		fc := mustDecode(t, payload, api.V2016, reflect.TypeOf(mdlrt.FuelCellData{})).(*mdlrt.FuelCellData)
		if fc.TotalNumberOfFcTp != 0xFFFF {
			t.Errorf("TotalNumberOfFcTp: got %d, want 65535", fc.TotalNumberOfFcTp)
		}
		if len(fc.ProbeTemperatureValues) != 0 {
			t.Errorf("ProbeTemperatureValues: got %v, want empty list", fc.ProbeTemperatureValues)
		}
		// 计数之后的字段解码时不得错位。
		if fc.HighestTempProbeCodeOfHydrogenSystem != 0x2A {
			t.Errorf("HighestTempProbeCodeOfHydrogenSystem: got %d, want 42", fc.HighestTempProbeCodeOfHydrogenSystem)
		}
		if fc.HighestConOfHydrogen != 0x0190 {
			t.Errorf("HighestConOfHydrogen: got %d, want 400", fc.HighestConOfHydrogen)
		}
		if fc.HighestHyConSensorCode != 0x01 {
			t.Errorf("HighestHyConSensorCode: got %d, want 1", fc.HighestHyConSensorCode)
		}
		if fc.HydrogenMaxPressureSensorCode != 0x02 {
			t.Errorf("HydrogenMaxPressureSensorCode: got %d, want 2", fc.HydrogenMaxPressureSensorCode)
		}
		if fc.HighVoltageDCState != 0x03 {
			t.Errorf("HighVoltageDCState: got 0x%02X, want 0x03", fc.HighVoltageDCState)
		}

		re := mustBytes(t, fc)
		if !bytes.Equal(re, payload) {
			t.Errorf("sentinel re-encode:\n got % X\nwant % X", re, payload)
		}
	})

	t.Run("ChargeableSubsystemTemperature", func(t *testing.T) {
		payload := []byte{0x01, 0xFF, 0xFE} // 子系统号=1, 探针数=0xFFFE (异常)

		ct := mustDecode(t, payload, api.V2016, reflect.TypeOf(mdlrt.ChargeableSubsystemTemperature{})).(*mdlrt.ChargeableSubsystemTemperature)
		if ct.TemperatureProbeCount != 0xFFFE {
			t.Errorf("TemperatureProbeCount: got %d, want 65534", ct.TemperatureProbeCount)
		}
		if len(ct.ProbeTemperatures) != 0 {
			t.Errorf("ProbeTemperatures: got %v, want empty list", ct.ProbeTemperatures)
		}

		re := mustBytes(t, ct)
		if !bytes.Equal(re, payload) {
			t.Errorf("sentinel re-encode:\n got % X\nwant % X", re, payload)
		}
	})
}

// TestCustomFlagOutOfRangeEncode 锁定 audit 2026-09-17 (L1):GB/T 32960-2016
// 表8 自定义数据标志为 0x80~0xFE;任何其他映射键(例如 0xFF,解码时会作为
// 未知 TLV 类型被拒绝)都必须导致编码失败。
func TestCustomFlagOutOfRangeEncode(t *testing.T) {
	bad := &gbt2016.RealTimeData{
		BeanTime:   sampleBeanTime(),
		CustomData: map[byte][]byte{0xFF: {0x01, 0x02}},
	}
	if raw, err := bad.Bytes(); err == nil {
		t.Errorf("0xFF custom flag: Encode succeeded (% X), want out-of-range error", raw)
	}

	good := &gbt2016.RealTimeData{
		BeanTime:   sampleBeanTime(),
		CustomData: map[byte][]byte{0x80: {0xAA}},
	}
	if _, err := good.Bytes(); err != nil {
		t.Errorf("0x80 custom flag: Encode failed: %v", err)
	}
}

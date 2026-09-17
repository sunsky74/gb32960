package codec_test

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/sunsky74/gb32960/codec"
)

type converterExpected struct {
	name    string
	raw     int64
	decoded float64
	encoded int64
}

func loadExpected(t *testing.T, path string) map[string][]converterExpected {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Skipf("golden file not found: %s (run Java GenerateConverterExpected first)", path)
		return nil
	}
	defer f.Close()

	result := make(map[string][]converterExpected)
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 1024*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		parts := strings.Split(sc.Text(), "\t")
		if len(parts) != 4 {
			continue
		}
		raw, _ := strconv.ParseInt(parts[1], 10, 64)
		dec, _ := strconv.ParseFloat(parts[2], 64)
		enc, _ := strconv.ParseInt(parts[3], 10, 64)
		result[parts[0]] = append(result[parts[0]], converterExpected{
			name: parts[0], raw: raw, decoded: dec, encoded: enc,
		})
	}
	return result
}

func TestValueConverter_FullRange_JavaGolden(t *testing.T) {
	data := loadExpected(t, "../golden/layer_b/converter_expected.tsv")
	if data == nil {
		return // 已跳过:尚无 Java 金样数据
	}

	converters := map[string]struct {
		vc   codec.ValueConverter
		name string
	}{
		"BatteryVoltage":          {codec.BatteryVoltageConverter, "BatteryVoltage"},
		"ExtremumVoltage":         {codec.ExtremumVoltageConverter, "ExtremumVoltage"},
		"BatteryVoltages2025":     {codec.BatteryVoltagesConverter2025, "BatteryVoltages2025"},
		"SuperCapVoltage":         {codec.SuperCapVoltageConverter, "SuperCapVoltage"},
		"SuperCapExtremumVoltage": {codec.SuperCapExtremumVoltageConverter, "SuperCapExtremumVoltage"},
		"Concentration2025":       {codec.ConcentrationConverter2025, "Concentration2025"},
	}

	for name, c := range converters {
		t.Run(name, func(t *testing.T) {
			entries := data[name]
			if len(entries) == 0 {
				t.Skipf("no %s entries in golden data", name)
			}
			failures := 0
			for _, e := range entries {
				gotDec := c.vc.Decode(e.raw)
				if math.Abs(gotDec-e.decoded) > 0.001 {
					if failures < 5 {
						t.Errorf("Decode(%d): got %f, java %f", e.raw, gotDec, e.decoded)
					}
					failures++
				}
				gotEnc := c.vc.Encode(e.decoded)
				if gotEnc != e.encoded {
					if failures < 5 {
						t.Errorf("Encode(%f): got %d, java %d (raw=%d)", e.decoded, gotEnc, e.encoded, e.raw)
					}
					failures++
				}
			}
			if failures > 0 {
				t.Errorf("%s: %d failures out of %d cases", name, failures, len(entries))
			} else {
				t.Logf("%s: %d cases PASS vs Java golden", name, len(entries))
			}
		})
	}
}

func TestValueConverter_FullRange_SelfValidation(t *testing.T) {
	// 对全部 6 个高风险转换器,在 0~65535 范围内校验往返。
	// audit 2026-09-17 (H4):这些现在是硬断言:每个合法的原始
	// 值都必须逐字节一致地通过 Decode -> Encode。不匹配时打印前 10 个
	// 样本,超过合理上限后中止,以避免日志泛滥。
	converters := map[string]codec.ValueConverter{
		"BatteryVoltage":          codec.BatteryVoltageConverter,
		"ExtremumVoltage":         codec.ExtremumVoltageConverter,
		"BatteryVoltages2025":     codec.BatteryVoltagesConverter2025,
		"SuperCapVoltage":         codec.SuperCapVoltageConverter,
		"SuperCapExtremumVoltage": codec.SuperCapExtremumVoltageConverter,
		"Concentration2025":       codec.ConcentrationConverter2025,
	}

	for name, vc := range converters {
		t.Run(name, func(t *testing.T) {
			failures := 0
			for raw := int64(0); raw <= 65535; raw++ {
				decoded := vc.Decode(raw)
				if vc.ErrValue.IsInvalid(raw) {
					// 异常/无效:解码结果应与原始值完全一致(直通)
					if decoded != float64(raw) {
						t.Errorf("%s Decode(%d): passthrough got %f, want %f", name, raw, decoded, float64(raw))
						failures++
						if failures >= 10 {
							t.Fatalf("too many failures")
						}
					}
					continue
				}
				// 正常值:重编码并校验逐字节一致的往返
				reEncoded := vc.Encode(decoded)
				if reEncoded != raw {
					if failures < 10 {
						t.Errorf("%s: raw %d -> decode %v -> encode %d", name, raw, decoded, reEncoded)
					}
					failures++
					if failures > 100 {
						t.Fatalf("%s: >100 roundtrip failures (first 10 printed); aborting", name)
					}
				}
			}
			if failures == 0 {
				t.Logf("%s: full range 0-65535 roundtrip PASS (65536 raw values)", name)
			}
		})
	}
}

func TestValueConverter_FullRange_MediumRisk(t *testing.T) {
	// 在 0..50000 关键区间校验中风险转换器。
	// audit 2026-09-17 (H4):硬断言:不匹配时打印前 10 个
	// 样本,超过合理上限后中止。
	cases := []struct {
		name string
		vc   codec.ValueConverter
	}{
		{"FuelConsumptionRate", codec.FuelConsumptionRateConverter},
		{"ChargeElectric", codec.CurrentConverterChargeElectric},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			failures := 0
			for raw := int64(0); raw <= 50000; raw++ {
				if c.vc.ErrValue.IsInvalid(raw) {
					continue
				}
				decoded := c.vc.Decode(raw)
				if reEncoded := c.vc.Encode(decoded); reEncoded != raw {
					if failures < 10 {
						t.Errorf("%s: raw %d -> decode %v -> encode %d", c.name, raw, decoded, reEncoded)
					}
					failures++
					if failures > 100 {
						t.Fatalf("%s: >100 roundtrip failures (first 10 printed); aborting", c.name)
					}
				}
			}
			if failures == 0 {
				t.Logf("%s: 0-50000 roundtrip PASS (50001 raw values)", c.name)
			}
		})
	}
}

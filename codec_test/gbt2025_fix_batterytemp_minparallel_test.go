package codec_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/sunsky74/gb32960/api"
	v2025rt "github.com/sunsky74/gb32960/model/gbt2025/realtime"

	// 聚合导入:触发所有编解码器的 init() 注册,使
	// api.GetCodec 查找(经由各模型的 Bytes() 方法)能够成功。
	_ "github.com/sunsky74/gb32960/codec/all"
)

// TestV2025BatteryTempEntrySentinelWire 锁定 fix 2026-09-17 (WP-D) 表14
// (L227-229)的条目级哨兵线格式:包号 0xFE/0xFF 与探针个数 0xFFFE/0xFFFF
// 必须原样直通而不是被改写,哨兵计数不得携带任何温度单元;
// decode→re-encode 必须逐字节稳定。
func TestV2025BatteryTempEntrySentinelWire(t *testing.T) {
	tempType := reflect.TypeOf(v2025rt.BatteryTemp{})
	cases := []struct {
		name string
		in   *v2025rt.BatteryTemp
		want []byte
	}{
		{"SeqError", &v2025rt.BatteryTemp{BatteryPackSeq: 0xFE}, []byte{0xFE, 0x00, 0x00}},
		{"SeqInvalid", &v2025rt.BatteryTemp{BatteryPackSeq: 0xFF}, []byte{0xFF, 0x00, 0x00}},
		{"CountErrorNoTemps", &v2025rt.BatteryTemp{BatteryPackSeq: 1, TemperatureProbeCount: 0xFFFE}, []byte{0x01, 0xFF, 0xFE}},
		{"CountInvalidNoTemps", &v2025rt.BatteryTemp{BatteryPackSeq: 1, TemperatureProbeCount: 0xFFFF}, []byte{0x01, 0xFF, 0xFF}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := mustBytes(t, tc.in); !bytes.Equal(got, tc.want) {
				t.Fatalf("encode: got % X, want % X", got, tc.want)
			}
			m := mustDecode(t, tc.want, api.V2025, tempType).(*v2025rt.BatteryTemp)
			if re := mustBytes(t, m); !bytes.Equal(re, tc.want) {
				t.Errorf("decode→re-encode: got % X, want % X", re, tc.want)
			}
		})
	}

	// 解码哨兵计数时不得读取温度字节:seq=0xFE、count=0xFFFE 之后无错位
	bt := mustDecode(t, []byte{0xFE, 0xFF, 0xFE}, api.V2025, tempType).(*v2025rt.BatteryTemp)
	if bt.BatteryPackSeq != 0xFE || bt.TemperatureProbeCount != 0xFFFE || len(bt.ProbeTemperatures) != 0 {
		t.Errorf("decode sentinels: got seq=%d count=%d temps=%v", bt.BatteryPackSeq, bt.TemperatureProbeCount, bt.ProbeTemperatures)
	}
}

// TestV2025BatteryTempEntryEncodeRejectsMismatch 锁定 fix 2026-09-17 (WP-D):
// 探针计数哨兵值携带非空温度列表,或普通计数与温度个数不一致时,
// 编码必须报错而不是写出丢数据/畸形帧。
func TestV2025BatteryTempEntryEncodeRejectsMismatch(t *testing.T) {
	cases := []struct {
		name string
		in   *v2025rt.BatteryTemp
	}{
		{"SentinelWithTemps", &v2025rt.BatteryTemp{BatteryPackSeq: 1, TemperatureProbeCount: 0xFFFE, ProbeTemperatures: []float64{25.0}}},
		{"InvalidWithTemps", &v2025rt.BatteryTemp{BatteryPackSeq: 1, TemperatureProbeCount: 0xFFFF, ProbeTemperatures: []float64{25.0}}},
		{"CountTooSmall", &v2025rt.BatteryTemp{BatteryPackSeq: 1, TemperatureProbeCount: 1, ProbeTemperatures: []float64{25.0, 26.0}}},
		{"CountTooLarge", &v2025rt.BatteryTemp{BatteryPackSeq: 1, TemperatureProbeCount: 2, ProbeTemperatures: []float64{25.0}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if raw, err := tc.in.Bytes(); err == nil {
				t.Errorf("Encode succeeded (% X), want error", raw)
			}
		})
	}

	// 正向对照:计数与温度个数一致时正常编码并往返
	assertRoundtrip(t, &v2025rt.BatteryTemp{BatteryPackSeq: 1, TemperatureProbeCount: 2, ProbeTemperatures: []float64{25.0, 26.0}})
}

// TestV2025BatteryTempListSentinelWire 锁定 fix 2026-09-17 (WP-D) 表13
// (L220):列表计数哨兵 0xFE 必须原样写出(旧实现改写为 0xFF),
// 哨兵之后不写任何条目;普通计数必须与 Items 长度一致。
func TestV2025BatteryTempListSentinelWire(t *testing.T) {
	listType := reflect.TypeOf(v2025rt.BatteryTempList{})

	if got := mustBytes(t, &v2025rt.BatteryTempList{BatteryPackCount: 0xFE}); !bytes.Equal(got, []byte{0xFE}) {
		t.Errorf("0xFE encode: got % X, want FE", got)
	}
	if got := mustBytes(t, &v2025rt.BatteryTempList{BatteryPackCount: 0xFF}); !bytes.Equal(got, []byte{0xFF}) {
		t.Errorf("0xFF encode: got % X, want FF", got)
	}
	m := mustDecode(t, []byte{0xFE}, api.V2025, listType).(*v2025rt.BatteryTempList)
	if m.BatteryPackCount != 0xFE || len(m.Items) != 0 {
		t.Errorf("decode: got count=%d items=%d, want 254/0", m.BatteryPackCount, len(m.Items))
	}
	if re := mustBytes(t, m); !bytes.Equal(re, []byte{0xFE}) {
		t.Errorf("re-encode: got % X, want FE (byte-exact)", re)
	}

	// 哨兵计数 + 非空 Items 无法在线格式上表示:拒绝。
	if raw, err := (&v2025rt.BatteryTempList{BatteryPackCount: 0xFE, Items: []v2025rt.BatteryTemp{{BatteryPackSeq: 1}}}).Bytes(); err == nil {
		t.Errorf("sentinel+items: Encode succeeded (% X), want error", raw)
	}
	// 普通计数与 Items 长度不一致:拒绝(此前会静默漏写条目)。
	if raw, err := (&v2025rt.BatteryTempList{BatteryPackCount: 2, Items: []v2025rt.BatteryTemp{{BatteryPackSeq: 1}}}).Bytes(); err == nil {
		t.Errorf("count mismatch: Encode succeeded (% X), want error", raw)
	}
	// 正向对照:计数一致时正常编码并往返
	assertRoundtrip(t, &v2025rt.BatteryTempList{BatteryPackCount: 1, Items: []v2025rt.BatteryTemp{
		{BatteryPackSeq: 1, TemperatureProbeCount: 1, ProbeTemperatures: []float64{25.0}},
	}})
}

// TestV2025MinParallelListSentinelAndOverflow 锁定 fix 2026-09-17 (WP-D)
// 表11(L200):哨兵计数原样写出且不得跟条目字节;>50 包防御性折叠为
// 单个 0xFF(保留旧行为);普通计数必须与 Items 长度一致。
func TestV2025MinParallelListSentinelAndOverflow(t *testing.T) {
	listType := reflect.TypeOf(v2025rt.MinParallelCellVoltageList{})

	if got := mustBytes(t, &v2025rt.MinParallelCellVoltageList{BatteryPackCount: 0xFE}); !bytes.Equal(got, []byte{0xFE}) {
		t.Errorf("0xFE encode: got % X, want FE", got)
	}
	m := mustDecode(t, []byte{0xFE}, api.V2025, listType).(*v2025rt.MinParallelCellVoltageList)
	if m.BatteryPackCount != 0xFE || len(m.Items) != 0 {
		t.Errorf("decode: got count=%d items=%d, want 254/0", m.BatteryPackCount, len(m.Items))
	}
	if re := mustBytes(t, m); !bytes.Equal(re, []byte{0xFE}) {
		t.Errorf("re-encode: got % X, want FE (byte-exact)", re)
	}

	// >50 包仍折叠为单个 0xFF(test-locked 行为保留)
	over := &v2025rt.MinParallelCellVoltageList{BatteryPackCount: 51, Items: make([]v2025rt.MinParallelCellVoltage, 51)}
	if got := mustBytes(t, over); !bytes.Equal(got, []byte{0xFF}) {
		t.Errorf(">50 encode: got % X, want single FF", got)
	}

	// 哨兵计数 + 非空 Items:拒绝(旧实现在哨兵字节后继续写条目 → 畸形帧)
	if raw, err := (&v2025rt.MinParallelCellVoltageList{BatteryPackCount: 0xFF, Items: []v2025rt.MinParallelCellVoltage{{BatteryPackSeq: 1}}}).Bytes(); err == nil {
		t.Errorf("sentinel+items: Encode succeeded (% X), want error", raw)
	}
	// 普通计数与 Items 长度不一致:拒绝
	if raw, err := (&v2025rt.MinParallelCellVoltageList{BatteryPackCount: 2, Items: []v2025rt.MinParallelCellVoltage{{BatteryPackSeq: 1}}}).Bytes(); err == nil {
		t.Errorf("count mismatch: Encode succeeded (% X), want error", raw)
	}
	// 正向对照:计数一致时正常编码并往返
	assertRoundtrip(t, &v2025rt.MinParallelCellVoltageList{BatteryPackCount: 2, Items: []v2025rt.MinParallelCellVoltage{
		{BatteryPackSeq: 1, Voltage: 250.0, Current: 80.0, MinParallelUnits: 1, BatteryVoltages: []float64{3.5}},
		{BatteryPackSeq: 2, Voltage: 251.0, Current: 81.0, MinParallelUnits: 2, BatteryVoltages: []float64{3.5, 3.6}},
	}})
}

// TestV2025MinParallelEntrySentinelWire 锁定 fix 2026-09-17 (WP-D) 表12
// (L210-211):最小并联单元总数哨兵值(0xFFFE 异常/0xFFFF 无效)
// 必须原样写出且电压列表必须为空;普通计数必须与电压个数一致,
// 否则拒绝编码(此前哨兵+非空电压会静默丢数据)。
func TestV2025MinParallelEntrySentinelWire(t *testing.T) {
	elemType := reflect.TypeOf(v2025rt.MinParallelCellVoltage{})

	// 哨兵计数 0xFFFE:线格式为 seq(1)+电压(2)+电流(2)+计数(2),原样保留
	raw := mustBytes(t, &v2025rt.MinParallelCellVoltage{BatteryPackSeq: 1, MinParallelUnits: 0xFFFE})
	if len(raw) != 7 || raw[5] != 0xFF || raw[6] != 0xFE {
		t.Errorf("sentinel units not preserved: got % X", raw)
	}
	m := mustDecode(t, raw, api.V2025, elemType).(*v2025rt.MinParallelCellVoltage)
	if m.MinParallelUnits != 0xFFFE || len(m.BatteryVoltages) != 0 {
		t.Errorf("decode sentinel: got units=%d volts=%d, want 65534/0", m.MinParallelUnits, len(m.BatteryVoltages))
	}
	if re := mustBytes(t, m); !bytes.Equal(re, raw) {
		t.Errorf("roundtrip: got % X want % X", re, raw)
	}

	// 哨兵计数 + 非空电压:拒绝而不是丢数据
	if out, err := (&v2025rt.MinParallelCellVoltage{BatteryPackSeq: 1, MinParallelUnits: 0xFFFE, BatteryVoltages: []float64{3.5}}).Bytes(); err == nil {
		t.Errorf("sentinel+voltages: Encode succeeded (% X), want error", out)
	}
	// 普通计数与电压个数不一致:拒绝
	if out, err := (&v2025rt.MinParallelCellVoltage{BatteryPackSeq: 1, MinParallelUnits: 2, BatteryVoltages: []float64{3.5}}).Bytes(); err == nil {
		t.Errorf("count mismatch: Encode succeeded (% X), want error", out)
	}
	// 正向对照:计数一致 → 编码长度 = 1+2+2+2 + 2×N
	good := &v2025rt.MinParallelCellVoltage{BatteryPackSeq: 1, Voltage: 250.0, Current: 80.0, MinParallelUnits: 2, BatteryVoltages: []float64{3.5, 3.6}}
	raw = mustBytes(t, good)
	if want := 1 + 2 + 2 + 2 + 2*2; len(raw) != want {
		t.Errorf("encoded length: got %d, want %d", len(raw), want)
	}
	assertRoundtrip(t, good)
}

package codec_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/model/gbt2016"
	mdlrt "github.com/sunsky74/gb32960/model/gbt2016/realtime"

	// 聚合导入:触发所有编解码器的 init() 注册(与同包其他测试一致)。
	_ "github.com/sunsky74/gb32960/codec/all"
)

// TestChargeableSubsystemListCountSentinelDecode 锁定 fix 2026-09-17 (WP-A):
// GB/T 32960.3-2016 表 B.5(L759)/表 B.7(L782) 规定可充电储能子系统个数为
// 1 字节,有效值 1~250,"0xFE"表示异常 / "0xFF"表示无效。哨兵帧不携带任何
// 子系统数据单元,Decode 不得按计数读取条目(旧代码对 254/255 会读取
// 254/255 个条目 → ErrBufferUnderflow)。重编码必须按哨兵值原样写出。
func TestChargeableSubsystemListCountSentinelDecode(t *testing.T) {
	cases := []struct {
		name     string
		typ      reflect.Type
		raw      []byte
		wantCnt  int
		getCount func(api.Message) int
		getItems func(api.Message) int
	}{
		{
			name:    "Electric_0xFE",
			typ:     reflect.TypeOf(mdlrt.ChargeableSubsystemElectricList{}),
			raw:     []byte{0xFE},
			wantCnt: 254,
			getCount: func(m api.Message) int {
				return m.(*mdlrt.ChargeableSubsystemElectricList).ElectricCount
			},
			getItems: func(m api.Message) int {
				return len(m.(*mdlrt.ChargeableSubsystemElectricList).Items)
			},
		},
		{
			name:    "Electric_0xFF",
			typ:     reflect.TypeOf(mdlrt.ChargeableSubsystemElectricList{}),
			raw:     []byte{0xFF},
			wantCnt: 255,
			getCount: func(m api.Message) int {
				return m.(*mdlrt.ChargeableSubsystemElectricList).ElectricCount
			},
			getItems: func(m api.Message) int {
				return len(m.(*mdlrt.ChargeableSubsystemElectricList).Items)
			},
		},
		{
			name:    "Temperature_0xFE",
			typ:     reflect.TypeOf(mdlrt.ChargeableSubsystemTemperatureList{}),
			raw:     []byte{0xFE},
			wantCnt: 254,
			getCount: func(m api.Message) int {
				return m.(*mdlrt.ChargeableSubsystemTemperatureList).TemperatureCount
			},
			getItems: func(m api.Message) int {
				return len(m.(*mdlrt.ChargeableSubsystemTemperatureList).Items)
			},
		},
		{
			name:    "Temperature_0xFF",
			typ:     reflect.TypeOf(mdlrt.ChargeableSubsystemTemperatureList{}),
			raw:     []byte{0xFF},
			wantCnt: 255,
			getCount: func(m api.Message) int {
				return m.(*mdlrt.ChargeableSubsystemTemperatureList).TemperatureCount
			},
			getItems: func(m api.Message) int {
				return len(m.(*mdlrt.ChargeableSubsystemTemperatureList).Items)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := mustDecode(t, tc.raw, api.V2016, tc.typ)
			if got := tc.getCount(m); got != tc.wantCnt {
				t.Errorf("count: got %d, want %d", got, tc.wantCnt)
			}
			if got := tc.getItems(m); got != 0 {
				t.Errorf("items: got %d items, want empty list", got)
			}

			// 哨兵重编码:逐字节一致(不追加任何条目字节)。
			re := mustBytes(t, m.(model.MessageBody))
			if !bytes.Equal(re, tc.raw) {
				t.Errorf("sentinel re-encode:\n got % X\nwant % X", re, tc.raw)
			}
		})
	}
}

// TestChargeableSubsystemListCountSentinelEncode 锁定 fix 2026-09-17 (WP-A)
// 编码侧:哨兵计数 + 空列表必须按哨兵字节原样写出;哨兵计数 + 非空列表
// 无法在线格式上表示,必须在写出任何字节之前报错。
func TestChargeableSubsystemListCountSentinelEncode(t *testing.T) {
	ok := []struct {
		name string
		mb   model.MessageBody
		want []byte
	}{
		{"Electric_0xFE", &mdlrt.ChargeableSubsystemElectricList{ElectricCount: 254}, []byte{0xFE}},
		{"Electric_0xFF", &mdlrt.ChargeableSubsystemElectricList{ElectricCount: 255}, []byte{0xFF}},
		{"Temperature_0xFE", &mdlrt.ChargeableSubsystemTemperatureList{TemperatureCount: 254}, []byte{0xFE}},
		{"Temperature_0xFF", &mdlrt.ChargeableSubsystemTemperatureList{TemperatureCount: 255}, []byte{0xFF}},
	}
	for _, tc := range ok {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := tc.mb.Bytes()
			if err != nil {
				t.Fatalf("sentinel encode failed: %v", err)
			}
			if !bytes.Equal(raw, tc.want) {
				t.Errorf("sentinel encode:\n got % X\nwant % X", raw, tc.want)
			}
		})
	}

	t.Run("SentinelWithItemsErrors", func(t *testing.T) {
		bad := []model.MessageBody{
			&mdlrt.ChargeableSubsystemElectricList{
				ElectricCount: 255,
				Items:         []mdlrt.ChargeableSubsystemElectric{{ChargeableSubSystemNumber: 1}},
			},
			&mdlrt.ChargeableSubsystemTemperatureList{
				TemperatureCount: 254,
				Items:            []mdlrt.ChargeableSubsystemTemperature{{SubSystemNumber: 1}},
			},
		}
		for _, mb := range bad {
			if raw, err := mb.Bytes(); err == nil {
				t.Errorf("%T: Encode succeeded (% X), want sentinel-with-items error", mb, raw)
			}
		}
	})
}

// TestChargeableSubsystemListCountMismatchStillErrors 锁定 audit 2026-09-17 (M4)
// 的既有行为不被哨兵修复放宽:普通计数与列表长度不一致时 Encode 必须报错。
func TestChargeableSubsystemListCountMismatchStillErrors(t *testing.T) {
	cases := []struct {
		name string
		mb   model.MessageBody
	}{
		{"Electric", &mdlrt.ChargeableSubsystemElectricList{
			ElectricCount: 2,
			Items:         []mdlrt.ChargeableSubsystemElectric{{ChargeableSubSystemNumber: 1}},
		}},
		{"Temperature", &mdlrt.ChargeableSubsystemTemperatureList{
			TemperatureCount: 2,
			Items:            []mdlrt.ChargeableSubsystemTemperature{{SubSystemNumber: 1}},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if raw, err := tc.mb.Bytes(); err == nil {
				t.Errorf("Encode succeeded (% X), want count/list mismatch error", raw)
			}
		})
	}
}

// TestChargeableSubsystemListNormalCountRoundTrip 是哨兵修复的正向对照:
// 普通计数(带条目)的解码路径不受影响,结构 → 编码 → 解码 → 再编码
// 逐字节稳定。
func TestChargeableSubsystemListNormalCountRoundTrip(t *testing.T) {
	t.Run("Electric", func(t *testing.T) {
		src := &mdlrt.ChargeableSubsystemElectricList{
			ElectricCount: 1,
			Items:         []mdlrt.ChargeableSubsystemElectric{{ChargeableSubSystemNumber: 7}},
		}
		raw := mustBytes(t, src)

		dec := mustDecode(t, raw, api.V2016, reflect.TypeOf(mdlrt.ChargeableSubsystemElectricList{})).(*mdlrt.ChargeableSubsystemElectricList)
		if dec.ElectricCount != 1 || len(dec.Items) != 1 {
			t.Fatalf("decoded count/items: got %d/%d, want 1/1", dec.ElectricCount, len(dec.Items))
		}
		if dec.Items[0].ChargeableSubSystemNumber != 7 {
			t.Errorf("ChargeableSubSystemNumber: got %d, want 7", dec.Items[0].ChargeableSubSystemNumber)
		}
		if re := mustBytes(t, dec); !bytes.Equal(re, raw) {
			t.Errorf("normal-count re-encode not byte-exact:\n got % X\nwant % X", re, raw)
		}
	})

	t.Run("Temperature", func(t *testing.T) {
		src := &mdlrt.ChargeableSubsystemTemperatureList{
			TemperatureCount: 1,
			Items:            []mdlrt.ChargeableSubsystemTemperature{{SubSystemNumber: 9}},
		}
		raw := mustBytes(t, src)

		dec := mustDecode(t, raw, api.V2016, reflect.TypeOf(mdlrt.ChargeableSubsystemTemperatureList{})).(*mdlrt.ChargeableSubsystemTemperatureList)
		if dec.TemperatureCount != 1 || len(dec.Items) != 1 {
			t.Fatalf("decoded count/items: got %d/%d, want 1/1", dec.TemperatureCount, len(dec.Items))
		}
		if dec.Items[0].SubSystemNumber != 9 {
			t.Errorf("SubSystemNumber: got %d, want 9", dec.Items[0].SubSystemNumber)
		}
		if re := mustBytes(t, dec); !bytes.Equal(re, raw) {
			t.Errorf("normal-count re-encode not byte-exact:\n got % X\nwant % X", re, raw)
		}
	})
}

// TestLoginFieldLengthGuards 锁定 fix 2026-09-17 (WP-A) 的编码侧长度校验:
// 定长 STRING 字段超长会被 WriteString 静默截断,必须在写出任何字节之前报错;
// 合法长度(含定长上限)仍必须正常编码,线格式长度符合表 6/表 21。
func TestLoginFieldLengthGuards(t *testing.T) {
	t.Run("VehicleLoginICCID", func(t *testing.T) {
		// 表 6(L109):ICCID 为 20 字节 STRING;21 字节必须拒绝。
		bad := &gbt2016.VehicleLogin{BeanTime: sampleBeanTime(), ICCID: strings.Repeat("A", 21)}
		if raw, err := bad.Bytes(); err == nil {
			t.Errorf("ICCID 21 bytes: Encode succeeded (% X), want error", raw)
		}

		// 正向对照:20 字节合法;线格式 = 6(时间)+2(流水号)+20(ICCID)+1(Count)+1(Length)。
		good := &gbt2016.VehicleLogin{BeanTime: sampleBeanTime(), ICCID: strings.Repeat("A", 20)}
		raw, err := good.Bytes()
		if err != nil {
			t.Fatalf("ICCID 20 bytes: encode failed: %v", err)
		}
		if want := 6 + 2 + 20 + 1 + 1; len(raw) != want {
			t.Errorf("ICCID 20 bytes: encoded length got %d, want %d", len(raw), want)
		}
	})

	t.Run("PlatformLoginUsername", func(t *testing.T) {
		// 表 21(L371):平台用户名为 12 字节 STRING;13 字节必须拒绝。
		bad := &gbt2016.PlatformLogin{
			BeanTime: sampleBeanTime(),
			Username: strings.Repeat("U", 13),
			Password: strings.Repeat("P", 20),
		}
		if raw, err := bad.Bytes(); err == nil {
			t.Errorf("username 13 bytes: Encode succeeded (% X), want error", raw)
		}
	})

	t.Run("PlatformLoginPassword", func(t *testing.T) {
		// 表 21(L372):平台密码为 20 字节 STRING;21 字节必须拒绝。
		bad := &gbt2016.PlatformLogin{
			BeanTime: sampleBeanTime(),
			Username: strings.Repeat("U", 12),
			Password: strings.Repeat("P", 21),
		}
		if raw, err := bad.Bytes(); err == nil {
			t.Errorf("password 21 bytes: Encode succeeded (% X), want error", raw)
		}
	})

	t.Run("ValidLengthsPass", func(t *testing.T) {
		// 正向对照:12 字节用户名 + 20 字节密码;
		// 线格式 = 6(时间)+2(流水号)+12(用户名)+20(密码)+1(加密规则) = 41 字节。
		good := &gbt2016.PlatformLogin{
			BeanTime: sampleBeanTime(),
			Username: strings.Repeat("U", 12),
			Password: strings.Repeat("P", 20),
			Cipher:   0x01,
		}
		raw, err := good.Bytes()
		if err != nil {
			t.Fatalf("valid lengths: encode failed: %v", err)
		}
		if want := 6 + 2 + 12 + 20 + 1; len(raw) != want {
			t.Errorf("valid lengths: encoded length got %d, want %d", len(raw), want)
		}
	})
}

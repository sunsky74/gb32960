package codec_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	v2025rt "github.com/sunsky74/gb32960/model/gbt2025/realtime"

	// 聚合导入:触发所有编解码器的 init() 注册(与同包其他测试一致)。
	_ "github.com/sunsky74/gb32960/codec/all"
)

// TestV2025MotorListCountSentinelAndValidation 锁定 fix 2026-09-17 (WP-F):
// GB/T 32960.3-2025 表15(L238) —— 驱动电机个数为 BYTE1,有效值 1~253,
// “0xFE”表示异常、“0xFF”表示无效。哨兵帧不携带任何电机条目:
//   - Decode 不得按哨兵计数读取 254/255 个条目(旧代码 → ErrBufferUnderflow);
//   - Encode 哨兵计数必须原样写出且列表为空,普通计数必须等于列表长度。
func TestV2025MotorListCountSentinelAndValidation(t *testing.T) {
	motorType := reflect.TypeOf(v2025rt.MotorDataV2025List{})

	t.Run("DecodeSentinel", func(t *testing.T) {
		cases := []struct {
			raw     []byte
			wantCnt int
		}{
			{[]byte{0xFE}, 254},
			{[]byte{0xFF}, 255},
		}
		for _, tc := range cases {
			m := mustDecode(t, tc.raw, api.V2025, motorType).(*v2025rt.MotorDataV2025List)
			if m.MotorCount != tc.wantCnt {
				t.Errorf("raw % X: count got %d, want %d", tc.raw, m.MotorCount, tc.wantCnt)
			}
			if len(m.Items) != 0 {
				t.Errorf("raw % X: items got %d, want empty list", tc.raw, len(m.Items))
			}
		}
	})

	t.Run("EncodeSentinelEmpty", func(t *testing.T) {
		cases := []struct {
			count int
			want  []byte
		}{
			{254, []byte{0xFE}},
			{255, []byte{0xFF}},
		}
		for _, tc := range cases {
			raw, err := (&v2025rt.MotorDataV2025List{MotorCount: tc.count}).Bytes()
			if err != nil {
				t.Fatalf("count %d: sentinel encode failed: %v", tc.count, err)
			}
			if !bytes.Equal(raw, tc.want) {
				t.Errorf("count %d:\n got % X\nwant % X", tc.count, raw, tc.want)
			}
		}
	})

	t.Run("EncodeRejectsMalformed", func(t *testing.T) {
		bad := []model.MessageBody{
			// 哨兵计数 + 条目:线格式无法表示。
			&v2025rt.MotorDataV2025List{MotorCount: 0xFE, Items: []v2025rt.MotorDataV2025{{MotorSeq: 1}}},
			&v2025rt.MotorDataV2025List{MotorCount: 0xFF, Items: []v2025rt.MotorDataV2025{{MotorSeq: 1}}},
			// 普通计数 != 列表长度。
			&v2025rt.MotorDataV2025List{MotorCount: 2, Items: []v2025rt.MotorDataV2025{{MotorSeq: 1}}},
			&v2025rt.MotorDataV2025List{MotorCount: 1},
			// 0 计数 + 条目(旧代码静默丢弃条目)。
			&v2025rt.MotorDataV2025List{MotorCount: 0, Items: []v2025rt.MotorDataV2025{{MotorSeq: 1}}},
		}
		for _, mb := range bad {
			if raw, err := mb.Bytes(); err == nil {
				t.Errorf("%T: Encode succeeded (% X), want count/list mismatch error", mb, raw)
			}
		}
	})

	t.Run("NormalCountRoundTrip", func(t *testing.T) {
		src := &v2025rt.MotorDataV2025List{
			MotorCount: 2,
			Items: []v2025rt.MotorDataV2025{
				{MotorSeq: 1, MotorState: 0x01},
				{MotorSeq: 2, MotorState: 0x02},
			},
		}
		raw := mustBytes(t, src)
		dec := mustDecode(t, raw, api.V2025, motorType).(*v2025rt.MotorDataV2025List)
		if dec.MotorCount != 2 || len(dec.Items) != 2 {
			t.Fatalf("decoded count/items: got %d/%d, want 2/2", dec.MotorCount, len(dec.Items))
		}
		if re := mustBytes(t, dec); !bytes.Equal(re, raw) {
			t.Errorf("normal-count re-encode not byte-exact:\n got % X\nwant % X", re, raw)
		}
	})
}

// TestV2025FuelCellStackCountValidation 锁定 fix 2026-09-17 (WP-F):
//   - 表18(L281):燃料电池电堆个数为 BYTE1,有效范围 1~253,
//     0xFE 异常 / 0xFF 无效;哨兵计数后不携带电堆条目;
//   - 表19(L299~L300):冷却水出水口温度探针总数为 WORD,有效范围
//     0~65531,0xFFFE 异常 / 0xFFFF 无效;哨兵计数后不携带温度列表。
func TestV2025FuelCellStackCountValidation(t *testing.T) {
	stackListType := reflect.TypeOf(v2025rt.FuelCellStackDataList{})
	stackType := reflect.TypeOf(v2025rt.FuelCellStackData{})

	t.Run("ListDecodeSentinel", func(t *testing.T) {
		cases := []struct {
			raw     []byte
			wantCnt int
		}{
			{[]byte{0xFE}, 254},
			{[]byte{0xFF}, 255},
		}
		for _, tc := range cases {
			m := mustDecode(t, tc.raw, api.V2025, stackListType).(*v2025rt.FuelCellStackDataList)
			if m.StackCount != tc.wantCnt {
				t.Errorf("raw % X: count got %d, want %d", tc.raw, m.StackCount, tc.wantCnt)
			}
			if len(m.Items) != 0 {
				t.Errorf("raw % X: items got %d, want empty list", tc.raw, len(m.Items))
			}
		}
	})

	t.Run("ListEncodeSentinelRaw", func(t *testing.T) {
		// roundtrip_test.go:690-696 已锁定 0xFE 直通;此处补齐 0xFF 与解码重编码。
		for _, want := range [][]byte{{0xFE}, {0xFF}} {
			raw, err := (&v2025rt.FuelCellStackDataList{StackCount: int(want[0])}).Bytes()
			if err != nil {
				t.Fatalf("sentinel % X: encode failed: %v", want, err)
			}
			if !bytes.Equal(raw, want) {
				t.Errorf("sentinel % X:\n got % X\nwant % X", want, raw, want)
			}
			dec := mustDecode(t, want, api.V2025, stackListType).(*v2025rt.FuelCellStackDataList)
			if re := mustBytes(t, dec); !bytes.Equal(re, want) {
				t.Errorf("sentinel % X: re-encode got % X, want byte-exact", want, re)
			}
		}
	})

	t.Run("ListEncodeRejectsMalformed", func(t *testing.T) {
		bad := []model.MessageBody{
			// 哨兵计数 + 条目(旧代码写出条目,不可解析)。
			&v2025rt.FuelCellStackDataList{StackCount: 0xFE, Items: []v2025rt.FuelCellStackData{{StackSeq: 1}}},
			&v2025rt.FuelCellStackDataList{StackCount: 0xFF, Items: []v2025rt.FuelCellStackData{{StackSeq: 1}}},
			// 普通计数 != 列表长度。
			&v2025rt.FuelCellStackDataList{StackCount: 2, Items: []v2025rt.FuelCellStackData{{StackSeq: 1}}},
			&v2025rt.FuelCellStackDataList{StackCount: 1},
		}
		for _, mb := range bad {
			if raw, err := mb.Bytes(); err == nil {
				t.Errorf("%T: Encode succeeded (% X), want count/list mismatch error", mb, raw)
			}
		}
	})

	t.Run("ElementProbeSentinelRaw", func(t *testing.T) {
		// 探针总数哨兵:计数原样写出,不携带温度字节。
		// 线格式 = 序号(1)+电压(2)+电流(2)+氢气压力(2)+空气压力(2)+
		// 空气温度(1)+探针总数(2)= 12 字节。
		for _, sentinel := range []int{0xFFFE, 0xFFFF} {
			raw, err := (&v2025rt.FuelCellStackData{CoolingWaterProbeCount: sentinel}).Bytes()
			if err != nil {
				t.Fatalf("sentinel %#X: encode failed: %v", sentinel, err)
			}
			if len(raw) != 12 {
				t.Fatalf("sentinel %#X: encoded length got %d, want 12", sentinel, len(raw))
			}
			want := []byte{byte(sentinel >> 8), byte(sentinel)}
			if !bytes.Equal(raw[10:12], want) {
				t.Errorf("sentinel %#X: count bytes got % X, want % X", sentinel, raw[10:12], want)
			}
		}
	})

	t.Run("ElementRejectsMalformed", func(t *testing.T) {
		bad := []model.MessageBody{
			// 哨兵计数 + 温度列表(旧代码静默丢弃温度)。
			&v2025rt.FuelCellStackData{CoolingWaterProbeCount: 0xFFFE, CoolingWaterTemps: []float64{25.0}},
			&v2025rt.FuelCellStackData{CoolingWaterProbeCount: 0xFFFF, CoolingWaterTemps: []float64{25.0}},
			// 普通计数 != 列表长度。
			&v2025rt.FuelCellStackData{CoolingWaterProbeCount: 2, CoolingWaterTemps: []float64{25.0}},
			&v2025rt.FuelCellStackData{CoolingWaterProbeCount: 1},
		}
		for _, mb := range bad {
			if raw, err := mb.Bytes(); err == nil {
				t.Errorf("%T %+v: Encode succeeded (% X), want count/list mismatch error", mb, mb, raw)
			}
		}
	})

	t.Run("ElementNormalCountRoundTrip", func(t *testing.T) {
		src := &v2025rt.FuelCellStackData{
			StackSeq:               1,
			Voltage:                250.0,
			Current:                80.0,
			GasPressure:            50.0,
			AirPressure:            40.0,
			AirInletTemp:           35.0,
			CoolingWaterProbeCount: 2,
			CoolingWaterTemps:      []float64{45.0, 47.0},
		}
		raw := mustBytes(t, src)
		dec := mustDecode(t, raw, api.V2025, stackType).(*v2025rt.FuelCellStackData)
		if dec.CoolingWaterProbeCount != 2 || len(dec.CoolingWaterTemps) != 2 {
			t.Fatalf("decoded count/temps: got %d/%d, want 2/2", dec.CoolingWaterProbeCount, len(dec.CoolingWaterTemps))
		}
		if re := mustBytes(t, dec); !bytes.Equal(re, raw) {
			t.Errorf("normal-count re-encode not byte-exact:\n got % X\nwant % X", re, raw)
		}
	})
}

// TestV2025AlarmCountValidation 锁定 fix 2026-09-17 (WP-F):
// GB/T 32960.3-2025 表23(L342~L356) —— 可充电储能装置/驱动电机/发动机/其他
// 故障总数 N1~N4 及通用报警故障总数 N5 均为 BYTE1(有效值 0~253,
// 0xFE 异常 / 0xFF 无效):普通计数必须等于代码列表长度;0 或哨兵计数
// 不携带任何列表条目。解码侧原本已正确跳过哨兵,此处锁定编码侧。
func TestV2025AlarmCountValidation(t *testing.T) {
	alarmType := reflect.TypeOf(v2025rt.AlarmV2025Data{})

	t.Run("SentinelCountsRoundTrip", func(t *testing.T) {
		src := &v2025rt.AlarmV2025Data{
			MaxAlarmLevel:   1,
			BatteryFaultNum: 0xFE,
			MotorFaultNum:   0xFF,
			EngineFaultNum:  0xFE,
			OtherFaultNum:   0xFF,
			CommonAlertNum:  0xFE,
		}
		raw := mustBytes(t, src)
		// 线格式:等级(1)+掩码(4)+5 个计数(1) = 10 字节。
		if len(raw) != 10 {
			t.Fatalf("sentinel frame length got %d, want 10", len(raw))
		}
		if want := []byte{0xFE, 0xFF, 0xFE, 0xFF, 0xFE}; !bytes.Equal(raw[5:10], want) {
			t.Errorf("sentinel counts:\n got % X\nwant % X", raw[5:10], want)
		}

		dec := mustDecode(t, raw, api.V2025, alarmType).(*v2025rt.AlarmV2025Data)
		if dec.BatteryFaultNum != 0xFE || dec.MotorFaultNum != 0xFF ||
			dec.EngineFaultNum != 0xFE || dec.OtherFaultNum != 0xFF || dec.CommonAlertNum != 0xFE {
			t.Errorf("decoded counts: got %d/%d/%d/%d/%d, want 254/255/254/255/254",
				dec.BatteryFaultNum, dec.MotorFaultNum, dec.EngineFaultNum, dec.OtherFaultNum, dec.CommonAlertNum)
		}
		if n := len(dec.BatteryFaultDatas) + len(dec.MotorFaultDatas) + len(dec.EngineFaultDatas) +
			len(dec.OtherFaultDatas) + len(dec.CommonAlertDatas); n != 0 {
			t.Errorf("decoded sentinel lists should be empty, got %d entries", n)
		}
		if re := mustBytes(t, dec); !bytes.Equal(re, raw) {
			t.Errorf("sentinel re-encode not byte-exact:\n got % X\nwant % X", re, raw)
		}
	})

	t.Run("EncodeRejectsMalformed", func(t *testing.T) {
		bad := []model.MessageBody{
			// N1: 普通计数 != 列表长度 / 哨兵 + 代码。
			&v2025rt.AlarmV2025Data{BatteryFaultNum: 2, BatteryFaultDatas: []int64{0x100}},
			&v2025rt.AlarmV2025Data{BatteryFaultNum: 0xFE, BatteryFaultDatas: []int64{0x100}},
			// N2: 普通计数 >0 但列表为空。
			&v2025rt.AlarmV2025Data{MotorFaultNum: 1},
			// N3: 0 计数 + 代码。
			&v2025rt.AlarmV2025Data{EngineFaultNum: 0, EngineFaultDatas: []int64{0x300}},
			// N4: 无效哨兵 + 代码。
			&v2025rt.AlarmV2025Data{OtherFaultNum: 0xFF, OtherFaultDatas: []int64{0x400}},
			// N5: 普通计数 != 列表长度 / 0 或哨兵 + 条目。
			&v2025rt.AlarmV2025Data{CommonAlertNum: 2, CommonAlertDatas: []v2025rt.CommonAlertData{{Seq: 1, Level: 2}}},
			&v2025rt.AlarmV2025Data{CommonAlertNum: 0, CommonAlertDatas: []v2025rt.CommonAlertData{{Seq: 1, Level: 2}}},
			&v2025rt.AlarmV2025Data{CommonAlertNum: 0xFE, CommonAlertDatas: []v2025rt.CommonAlertData{{Seq: 1, Level: 2}}},
		}
		for _, mb := range bad {
			if raw, err := mb.Bytes(); err == nil {
				t.Errorf("%T %+v: Encode succeeded (% X), want count/list mismatch error", mb, mb, raw)
			}
		}
	})

	t.Run("NormalCountsRoundTrip", func(t *testing.T) {
		src := &v2025rt.AlarmV2025Data{
			MaxAlarmLevel:     3,
			BatteryFaultNum:   1,
			BatteryFaultDatas: []int64{0x1000},
			MotorFaultNum:     1,
			MotorFaultDatas:   []int64{0x2001},
			EngineFaultNum:    1,
			EngineFaultDatas:  []int64{0x3001},
			OtherFaultNum:     2,
			OtherFaultDatas:   []int64{0x4001, 0x4002},
			CommonAlertNum:    2,
			CommonAlertDatas:  []v2025rt.CommonAlertData{{Seq: 1, Level: 2}, {Seq: 5, Level: 1}},
		}
		raw := mustBytes(t, src)
		dec := mustDecode(t, raw, api.V2025, alarmType).(*v2025rt.AlarmV2025Data)
		if dec.BatteryFaultNum != 1 || len(dec.BatteryFaultDatas) != 1 ||
			dec.MotorFaultNum != 1 || len(dec.MotorFaultDatas) != 1 ||
			dec.EngineFaultNum != 1 || len(dec.EngineFaultDatas) != 1 ||
			dec.OtherFaultNum != 2 || len(dec.OtherFaultDatas) != 2 ||
			dec.CommonAlertNum != 2 || len(dec.CommonAlertDatas) != 2 {
			t.Fatalf("decoded counts/lists drifted: %+v", dec)
		}
		if re := mustBytes(t, dec); !bytes.Equal(re, raw) {
			t.Errorf("normal-count re-encode not byte-exact:\n got % X\nwant % X", re, raw)
		}
	})
}

// TestV2025SuperCapacitorCountValidation 锁定 fix 2026-09-17 (WP-F):
// GB/T 32960.3-2025 表25(L407~L410) —— 超级电容单体总数 M 与温度探针
// 总数 N 均为 WORD(有效范围 1~65531,0xFFFE 异常 / 0xFFFF 无效):
// 哨兵计数不携带对应列表;普通计数必须等于列表长度。
func TestV2025SuperCapacitorCountValidation(t *testing.T) {
	capType := reflect.TypeOf(v2025rt.SuperCapacitorData{})

	t.Run("SentinelCountsRaw", func(t *testing.T) {
		raw, err := (&v2025rt.SuperCapacitorData{
			CapacitorCount:        0xFFFE,
			TemperatureProbeCount: 0xFFFF,
		}).Bytes()
		if err != nil {
			t.Fatalf("sentinel encode failed: %v", err)
		}
		// 线格式 = 管理系统号(1)+总电压(2)+总电流(2)+单体总数(2)+探针总数(2)。
		if len(raw) != 9 {
			t.Fatalf("sentinel frame length got %d, want 9", len(raw))
		}
		if !bytes.Equal(raw[5:7], []byte{0xFF, 0xFE}) {
			t.Errorf("capacitor count bytes got % X, want FF FE", raw[5:7])
		}
		if !bytes.Equal(raw[7:9], []byte{0xFF, 0xFF}) {
			t.Errorf("probe count bytes got % X, want FF FF", raw[7:9])
		}

		dec := mustDecode(t, raw, api.V2025, capType).(*v2025rt.SuperCapacitorData)
		if dec.CapacitorCount != 0xFFFE || len(dec.CapacitorVoltages) != 0 {
			t.Errorf("decoded capacitor count/list: got %d/%d, want 65534/0", dec.CapacitorCount, len(dec.CapacitorVoltages))
		}
		if dec.TemperatureProbeCount != 0xFFFF || len(dec.ProbeTemperatures) != 0 {
			t.Errorf("decoded probe count/list: got %d/%d, want 65535/0", dec.TemperatureProbeCount, len(dec.ProbeTemperatures))
		}
		if re := mustBytes(t, dec); !bytes.Equal(re, raw) {
			t.Errorf("sentinel re-encode not byte-exact:\n got % X\nwant % X", re, raw)
		}
	})

	t.Run("EncodeRejectsMalformed", func(t *testing.T) {
		bad := []model.MessageBody{
			// M 哨兵 + 电压列表 / N 哨兵 + 温度列表(旧代码静默丢弃列表)。
			&v2025rt.SuperCapacitorData{CapacitorCount: 0xFFFE, CapacitorVoltages: []float64{2.7}},
			&v2025rt.SuperCapacitorData{CapacitorCount: 0xFFFF, CapacitorVoltages: []float64{2.7}},
			&v2025rt.SuperCapacitorData{TemperatureProbeCount: 0xFFFE, ProbeTemperatures: []float64{25.0}},
			&v2025rt.SuperCapacitorData{TemperatureProbeCount: 0xFFFF, ProbeTemperatures: []float64{25.0}},
			// 普通计数 != 列表长度(含计数 >0 列表为空)。
			&v2025rt.SuperCapacitorData{CapacitorCount: 2, CapacitorVoltages: []float64{2.7}},
			&v2025rt.SuperCapacitorData{CapacitorCount: 1},
			&v2025rt.SuperCapacitorData{TemperatureProbeCount: 2, ProbeTemperatures: []float64{25.0}},
			&v2025rt.SuperCapacitorData{TemperatureProbeCount: 1},
		}
		for _, mb := range bad {
			if raw, err := mb.Bytes(); err == nil {
				t.Errorf("%T %+v: Encode succeeded (% X), want count/list mismatch error", mb, mb, raw)
			}
		}
	})

	t.Run("NormalCountsRoundTrip", func(t *testing.T) {
		src := &v2025rt.SuperCapacitorData{
			ManagementSystemNumber: 1,
			TotalVoltage:           48.0,
			TotalCurrent:           -15.0,
			CapacitorCount:         2,
			CapacitorVoltages:      []float64{2.7, 2.71},
			TemperatureProbeCount:  1,
			ProbeTemperatures:      []float64{25.0},
		}
		raw := mustBytes(t, src)
		dec := mustDecode(t, raw, api.V2025, capType).(*v2025rt.SuperCapacitorData)
		if dec.CapacitorCount != 2 || len(dec.CapacitorVoltages) != 2 {
			t.Fatalf("decoded capacitor count/list: got %d/%d, want 2/2", dec.CapacitorCount, len(dec.CapacitorVoltages))
		}
		if dec.TemperatureProbeCount != 1 || len(dec.ProbeTemperatures) != 1 {
			t.Fatalf("decoded probe count/list: got %d/%d, want 1/1", dec.TemperatureProbeCount, len(dec.ProbeTemperatures))
		}
		if re := mustBytes(t, dec); !bytes.Equal(re, raw) {
			t.Errorf("normal-count re-encode not byte-exact:\n got % X\nwant % X", re, raw)
		}
	})
}

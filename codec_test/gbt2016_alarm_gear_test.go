package codec_test

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"

	"github.com/sunsky74/gb32960/api"
	mdlrt16 "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
	"github.com/sunsky74/gb32960/utils"

	// 聚合导入:触发所有编解码器的 init() 注册,使下面的
	// api.GetCodec 查找能够成功。
	_ "github.com/sunsky74/gb32960/codec/all"
)

// TestGearPositionEnum_MatchesGB2016 锁定 附录A.1 低四位挡位映射
// (audit 2026-09-17: H2)。
func TestGearPositionEnum_MatchesGB2016(t *testing.T) {
	cases := []struct {
		name string
		got  types.GearPositionEnum
		want byte
	}{
		{"GearGap", types.GearGap, 0x0},
		{"Gear1", types.Gear1, 0x1},
		{"Gear2", types.Gear2, 0x2},
		{"Gear3", types.Gear3, 0x3},
		{"Gear4", types.Gear4, 0x4},
		{"Gear5", types.Gear5, 0x5},
		{"Gear6", types.Gear6, 0x6},
		{"Gear7", types.Gear7, 0x7},
		{"Gear8", types.Gear8, 0x8},
		{"Gear9", types.Gear9, 0x9},
		{"Gear10", types.Gear10, 0xA},
		{"Gear11", types.Gear11, 0xB},
		{"Gear12", types.Gear12, 0xC},
		{"GearReverse", types.GearReverse, 0xD},
		{"GearAutoD", types.GearAutoD, 0xE},
		{"GearPark", types.GearPark, 0xF},
	}
	for _, c := range cases {
		if byte(c.got) != c.want {
			t.Errorf("%s = 0x%X, want 0x%X", c.name, byte(c.got), c.want)
		}
	}
}

// TestGearPositionCodec_DecodesA1Bits 用线格式字节驱动注册的 V2016 挡位编解码器,
// 校验低四位与标志位的提取(audit 2026-09-17: H2)。
func TestGearPositionCodec_DecodesA1Bits(t *testing.T) {
	gpType := reflect.TypeOf(mdlrt16.GearPosition{})
	decode := func(t *testing.T, b byte) *mdlrt16.GearPosition {
		t.Helper()
		m, err := api.GetCodec(api.V2016, gpType).Decode(utils.NewByteReader([]byte{b}))
		if err != nil {
			t.Fatalf("decode 0x%02X: %v", b, err)
		}
		return m.(*mdlrt16.GearPosition)
	}

	t.Run("0x1E autoD+braking", func(t *testing.T) {
		m := decode(t, 0x1E) // bit4 制动力=1, bit5 驱动力=0, 低四位=0xE
		assertInt(t, "GP", int(m.GP), int(types.GearAutoD))
		if !m.BrakingTorqueApplied {
			t.Errorf("braking: got false, want true (bit 4 set in 0x1E)")
		}
		if m.DrivingForceActive {
			t.Errorf("driving force: got true, want false (bit 5 clear in 0x1E)")
		}
	})

	t.Run("0x0F park", func(t *testing.T) {
		m := decode(t, 0x0F) // 低四位=0xF; bit4 制动力=0
		assertInt(t, "GP", int(m.GP), int(types.GearPark))
		if m.BrakingTorqueApplied {
			t.Errorf("braking: got true, want false (bit 4 clear in 0x0F)")
		}
	})

	t.Run("0x1F park+braking", func(t *testing.T) {
		m := decode(t, 0x1F) // 低四位=0xF,且 bit4 制动力=1
		assertInt(t, "GP", int(m.GP), int(types.GearPark))
		if !m.BrakingTorqueApplied {
			t.Errorf("braking: got false, want true (bit 4 set in 0x1F)")
		}
	})

	t.Run("0x00 gap", func(t *testing.T) {
		m := decode(t, 0x00)
		assertInt(t, "GP", int(m.GP), int(types.GearGap))
		if m.BrakingTorqueApplied || m.DrivingForceActive {
			t.Errorf("flags: got braking=%v driving=%v, want both false", m.BrakingTorqueApplied, m.DrivingForceActive)
		}
	})
}

// TestChargingStateConstants 锁定 表9 / 表B.4 充电状态取值
// (audit 2026-09-17: M1)。
func TestChargingStateConstants(t *testing.T) {
	cases := []struct {
		name string
		got  types.ChargingState
		want byte
	}{
		{"ChargeStateParking", types.ChargeStateParking, 0x01},
		{"ChargeStateDriving", types.ChargeStateDriving, 0x02},
		{"ChargeStateNotCharging", types.ChargeStateNotCharging, 0x03},
		{"ChargeStateCompleted", types.ChargeStateCompleted, 0x04},
		{"ChargeStateException", types.ChargeStateException, 0xFE},
		{"ChargeStateInvalid", types.ChargeStateInvalid, 0xFF},
	}
	for _, c := range cases {
		if byte(c.got) != c.want {
			t.Errorf("%s = 0x%02X, want 0x%02X", c.name, byte(c.got), c.want)
		}
	}
}

// alarmPayload 组装一个最小报警数据数据单元:level(u8)、mask(u32 BE),
// 然后是四个计数字节 N1..N4,不带故障条目(用于后续没有列表的
// 哨兵值与保留位场景)。
func alarmPayload(mask uint32, counts [4]byte) []byte {
	w := utils.NewByteWriter()
	w.WriteUint8(0x00) // 最高报警等级
	w.WriteUint32(mask)
	for _, c := range counts {
		w.WriteUint8(c)
	}
	return w.Bytes()
}

// TestAlarmCountSentinel_Decode 锁定 表17 N1..N4 哨兵值约定:
// 计数为 0xFE/0xFF 时不携带故障条目,正常计数后跟相应数量的
// u32 编码,重编码哨兵值保持字节稳定
// (audit 2026-09-17: M2)。
func TestAlarmCountSentinel_Decode(t *testing.T) {
	alarmType := reflect.TypeOf(mdlrt16.AlarmData{})
	codec := api.GetCodec(api.V2016, alarmType)
	decode := func(t *testing.T, raw []byte) *mdlrt16.AlarmData {
		t.Helper()
		m, err := codec.Decode(utils.NewByteReader(raw))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		return m.(*mdlrt16.AlarmData)
	}

	for _, sentinel := range []byte{0xFF, 0xFE} {
		t.Run(fmt.Sprintf("N1=0x%02X", sentinel), func(t *testing.T) {
			raw := alarmPayload(0, [4]byte{sentinel, 0x00, 0x00, 0x00})
			if len(raw) != 9 {
				t.Fatalf("payload length = %d, want 9", len(raw))
			}
			m := decode(t, raw)
			assertInt(t, "BatteryFaultNum", m.BatteryFaultNum, int(sentinel))
			assertInt(t, "len(BatteryFaultDatas)", len(m.BatteryFaultDatas), 0)

			w := utils.NewByteWriter()
			if err := codec.Encode(w, m); err != nil {
				t.Fatalf("re-encode: %v", err)
			}
			if got := w.Bytes(); len(got) != len(raw) || got[5] != sentinel {
				t.Errorf("re-encode drift: got % X, want count 0x%02X and no entries (raw % X)", got, sentinel, raw)
			}
		})
	}

	t.Run("N1=0x02 with two codes", func(t *testing.T) {
		w := utils.NewByteWriter()
		w.WriteUint8(0x00)        // 最高报警等级
		w.WriteUint32(0x00000000) // 通用报警标志
		w.WriteUint8(0x02)        // N1
		w.WriteUint32(0x00001234)
		w.WriteUint32(0x0000ABCD)
		w.WriteUint8(0x00) // N2
		w.WriteUint8(0x00) // N3
		w.WriteUint8(0x00) // N4

		m := decode(t, w.Bytes())
		assertInt(t, "BatteryFaultNum", m.BatteryFaultNum, 2)
		assertInt(t, "len(BatteryFaultDatas)", len(m.BatteryFaultDatas), 2)
		assertInt(t, "code[0]", int(m.BatteryFaultDatas[0]), 0x1234)
		assertInt(t, "code[1]", int(m.BatteryFaultDatas[1]), 0xABCD)
	})
}

// TestAlarmReservedBitsRoundTrip 锁定保留位的保留性:通用报警标志的
// bit 19..31 必须完整通过 decode → encode,而已定义位仍来自
// 各布尔字段(audit 2026-09-17: L4)。
func TestAlarmReservedBitsRoundTrip(t *testing.T) {
	alarmType := reflect.TypeOf(mdlrt16.AlarmData{})
	codec := api.GetCodec(api.V2016, alarmType)

	raw := alarmPayload(0x00180000, [4]byte{0x00, 0x00, 0x00, 0x00}) // bit 19+20
	decoded, err := codec.Decode(utils.NewByteReader(raw))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	m := decoded.(*mdlrt16.AlarmData)
	if m.AlarmBitIdentify != 0x00180000 {
		t.Fatalf("decoded AlarmBitIdentify = 0x%08X, want 0x00180000", m.AlarmBitIdentify)
	}

	reEncodeMask := func(t *testing.T) uint32 {
		t.Helper()
		w := utils.NewByteWriter()
		if err := codec.Encode(w, m); err != nil {
			t.Fatalf("encode: %v", err)
		}
		return binary.BigEndian.Uint32(w.Bytes()[1:5])
	}

	if got := reEncodeMask(t); got != 0x00180000 {
		t.Errorf("reserved bits lost: mask = 0x%08X, want 0x00180000", got)
	}

	m.TemperatureDifferential = true
	if got := reEncodeMask(t); got != 0x00180001 {
		t.Errorf("defined bit not merged with reserved bits: mask = 0x%08X, want 0x00180001", got)
	}
}

// TestAlarmEncodeCountMismatchErrors 锁定编码时的计数/列表一致性
// (audit 2026-09-17: M4)。
func TestAlarmEncodeCountMismatchErrors(t *testing.T) {
	alarmType := reflect.TypeOf(mdlrt16.AlarmData{})
	codec := api.GetCodec(api.V2016, alarmType)

	t.Run("mismatch", func(t *testing.T) {
		cases := []struct {
			name string
			msg  *mdlrt16.AlarmData
		}{
			{"battery count 2 vs 1 entry", &mdlrt16.AlarmData{BatteryFaultNum: 2, BatteryFaultDatas: []int64{1}}},
			{"motor count 3 vs 0 entries", &mdlrt16.AlarmData{MotorFaultNum: 3}},
			{"engine sentinel 0xFF with entry", &mdlrt16.AlarmData{EngineFaultNum: 0xFF, EngineFaultDatas: []int64{7}}},
			{"other sentinel 0xFE with entry", &mdlrt16.AlarmData{OtherFaultNum: 0xFE, OtherFaultDatas: []int64{9}}},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				if err := codec.Encode(utils.NewByteWriter(), c.msg); err == nil {
					t.Errorf("Encode = nil error, want count/list mismatch error")
				}
			})
		}
	})

	t.Run("consistent", func(t *testing.T) {
		m := &mdlrt16.AlarmData{
			BatteryFaultNum: 2, BatteryFaultDatas: []int64{0x11, 0x22},
			MotorFaultNum: 1, MotorFaultDatas: []int64{0x33},
			EngineFaultNum: 0,
			OtherFaultNum:  0xFF,
		}
		if err := codec.Encode(utils.NewByteWriter(), m); err != nil {
			t.Errorf("Encode = %v, want nil", err)
		}
	})
}

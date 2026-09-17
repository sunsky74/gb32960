package codec_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/sunsky74/gb32960/api"
	mdlrt16 "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	mdl2025 "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
	"github.com/sunsky74/gb32960/utils"

	// 聚合导入:触发所有编解码器的 init() 注册,使下面的
	// api.GetCodec 查找能够成功。
	_ "github.com/sunsky74/gb32960/codec/all"
)

// TestLocationV2025_EncodeIsByteExact 锁定 2025.md L320-321(表21):
// 经度/纬度线格式为「以度为单位的经度值乘以 10^6,精确到百万分之一度」的
// DWORD,解码→编码必须逐字节还原。
// 缺陷回归值 249/16000002/128000003 在旧的 int64 截断实现下会各丢失 1 LSB
// (249→248、16000002→16000001、128000003→128000002)。
func TestLocationV2025_EncodeIsByteExact(t *testing.T) {
	locationCodec := api.GetCodec(api.V2025, reflect.TypeOf(mdl2025.LocationV2025Data{}))
	if locationCodec == nil {
		t.Fatal("V2025 LocationV2025Data 编解码器未注册(检查 codec/all 空白导入)")
	}

	// 构造线上字节(StatusByte + CoordinateType=0x01 WGS84 + 经度 u32 + 纬度 u32),
	// 解码后再编码,要求输出与输入逐字节一致,并返回解码模型。
	roundtrip := func(t *testing.T, status byte, rawLon, rawLat uint32) *mdl2025.LocationV2025Data {
		t.Helper()
		w := utils.NewByteWriter()
		w.WriteUint8(status)
		w.WriteUint8(0x01)
		w.WriteUint32(rawLon)
		w.WriteUint32(rawLat)
		in := w.Bytes()

		msg, err := locationCodec.Decode(utils.NewByteReader(in))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		out := utils.NewByteWriter()
		if err := locationCodec.Encode(out, msg); err != nil {
			t.Fatalf("encode: %v", err)
		}
		if !bytes.Equal(in, out.Bytes()) {
			t.Errorf("字节不一致(2025.md L320-321 要求精确到百万分之一度):\n  线上输入 = % X\n  编码输出 = % X", in, out.Bytes())
		}
		return msg.(*mdl2025.LocationV2025Data)
	}

	// 前三个为审计实证的截断丢 1 LSB 的原始值;后两个为既有往返用例值。
	regressionRaws := []uint32{249, 16000002, 128000003, 116407400, 39904200}

	t.Run("经度原始值", func(t *testing.T) {
		for _, raw := range regressionRaws {
			roundtrip(t, 0x00, raw, 39904200)
		}
	})

	t.Run("纬度原始值", func(t *testing.T) {
		for _, raw := range regressionRaws {
			roundtrip(t, 0x00, 116407400, raw)
		}
	})

	t.Run("西经南纬(status=0x06)得负值且可还原", func(t *testing.T) {
		m := roundtrip(t, 0x06, 128000003, 39904200)
		if m.EastFlag || m.NorthernFlag {
			t.Errorf("status=0x06 应解出西经且南纬,got east=%v north=%v", m.EastFlag, m.NorthernFlag)
		}
		assertFloat(t, "西经度", m.OriginLongitude, -128.000003, 1e-9)
		assertFloat(t, "南纬度", m.OriginLatitude, -39.9042, 1e-9)
	})

	t.Run("哨兵 0xFFFFFFFE 原样直通", func(t *testing.T) {
		m := roundtrip(t, 0x00, 0xFFFFFFFE, 39904200)
		if m.OriginLongitude != float64(0xFFFFFFFE) {
			t.Errorf("经度哨兵: got %v, want %v(原样直通,不做缩放/取负)", m.OriginLongitude, float64(0xFFFFFFFE))
		}
	})
}

// TestGearPositionV2025_ViaVehicleDataIsV2025 锁定 2025.md L505(附录A.1 表A.1)
// 与 Java VehicleDataV2025Codec.java L51-52:V2025 整车数据中的挡位字节
// 必须由 V2025 挡位编解码器解析(IsV2025=true),bit7=1 表示挡位无效。
func TestGearPositionV2025_ViaVehicleDataIsV2025(t *testing.T) {
	gpType := reflect.TypeOf(mdlrt16.GearPosition{})
	v16Codec := api.GetCodec(api.V2016, gpType)
	v25Codec := api.GetCodec(api.V2025, gpType)
	if v16Codec == nil || v25Codec == nil {
		t.Fatal("挡位编解码器必须同时注册 V2016 与 V2025")
	}
	if reflect.TypeOf(v16Codec) == reflect.TypeOf(v25Codec) {
		t.Fatalf("V2025 必须使用独立的 GearPositionV2025Codec,当前共用 %T", v25Codec)
	}

	vehCodec := api.GetCodec(api.V2025, reflect.TypeOf(mdlrt16.VehicleData{}))
	if vehCodec == nil {
		t.Fatal("V2025 VehicleData 编解码器未注册")
	}

	decodeGear := func(t *testing.T, gear byte) *mdlrt16.GearPosition {
		t.Helper()
		msg, err := vehCodec.Decode(utils.NewByteReader(wpC2025VehicleData(gear)))
		if err != nil {
			t.Fatalf("V2025 整车数据 decode: %v", err)
		}
		return &msg.(*mdlrt16.VehicleData).GearPosition
	}

	t.Run("0x80 挡位无效(bit7=1)", func(t *testing.T) {
		gp := decodeGear(t, 0x80)
		if !gp.IsV2025 {
			t.Errorf("IsV2025: got false, want true(必须走 V2025 挡位编解码器)")
		}
		if gp.Effective {
			t.Errorf("Effective: got true, want false(bit7=1 表示挡位无效)")
		}
		if gp.DrivingForceActive || gp.BrakingTorqueApplied {
			t.Errorf("驱动力/制动力: got %v/%v, want false/false", gp.DrivingForceActive, gp.BrakingTorqueApplied)
		}
		assertInt(t, "GP", int(gp.GP), int(types.GearGap))
		assertInt(t, "Origin", int(gp.Origin), 0x80)
	})

	t.Run("0x01 挡位有效(bit7=0)", func(t *testing.T) {
		gp := decodeGear(t, 0x01)
		if !gp.IsV2025 {
			t.Errorf("IsV2025: got false, want true")
		}
		if !gp.Effective {
			t.Errorf("Effective: got false, want true(bit7=0 表示挡位有效)")
		}
		assertInt(t, "GP", int(gp.GP), int(types.Gear1))
	})

	t.Run("0x1E 制动力位不受影响", func(t *testing.T) {
		gp := decodeGear(t, 0x1E)
		if !gp.BrakingTorqueApplied || gp.DrivingForceActive {
			t.Errorf("制动力/驱动力: got %v/%v, want true/false", gp.BrakingTorqueApplied, gp.DrivingForceActive)
		}
		if !gp.Effective {
			t.Errorf("Effective: got false, want true(bit7=0)")
		}
		assertInt(t, "GP", int(gp.GP), int(types.GearAutoD))
	})

	t.Run("0x2D 驱动力位不受影响", func(t *testing.T) {
		gp := decodeGear(t, 0x2D)
		if !gp.DrivingForceActive || gp.BrakingTorqueApplied {
			t.Errorf("驱动力/制动力: got %v/%v, want true/false", gp.DrivingForceActive, gp.BrakingTorqueApplied)
		}
		assertInt(t, "GP", int(gp.GP), int(types.GearReverse))
	})

	t.Run("整车数据编解码逐字节还原(挡位仍写 Origin 原始字节)", func(t *testing.T) {
		in := wpC2025VehicleData(0x80)
		msg, err := vehCodec.Decode(utils.NewByteReader(in))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		out := utils.NewByteWriter()
		if err := vehCodec.Encode(out, msg); err != nil {
			t.Fatalf("encode: %v", err)
		}
		if !bytes.Equal(in, out.Bytes()) {
			t.Errorf("字节不一致:\n  线上输入 = % X\n  编码输出 = % X", in, out.Bytes())
		}
	})
}

// TestGearPositionV2016_ViaVehicleDataIsV2016 确认 2016 整车数据路径仍走
// V2016 挡位编解码器(IsV2025=false);2016 帧 bit7 为预留位(规范要求恒 0)。
func TestGearPositionV2016_ViaVehicleDataIsV2016(t *testing.T) {
	vehCodec := api.GetCodec(api.V2016, reflect.TypeOf(mdlrt16.VehicleData{}))
	if vehCodec == nil {
		t.Fatal("V2016 VehicleData 编解码器未注册")
	}

	decodeGear := func(t *testing.T, gear byte) *mdlrt16.GearPosition {
		t.Helper()
		msg, err := vehCodec.Decode(utils.NewByteReader(wpC2016VehicleData(gear)))
		if err != nil {
			t.Fatalf("V2016 整车数据 decode: %v", err)
		}
		return &msg.(*mdlrt16.VehicleData).GearPosition
	}

	t.Run("0x01", func(t *testing.T) {
		gp := decodeGear(t, 0x01)
		if gp.IsV2025 {
			t.Errorf("IsV2025: got true, want false(V2016 路径)")
		}
		if !gp.Effective {
			t.Errorf("Effective: got false, want true(2016 预留位恒 0,解出有效)")
		}
		assertInt(t, "GP", int(gp.GP), int(types.Gear1))
	})

	t.Run("0x80", func(t *testing.T) {
		gp := decodeGear(t, 0x80)
		if gp.IsV2025 {
			t.Errorf("IsV2025: got true, want false(V2016 路径)")
		}
		assertInt(t, "Origin", int(gp.Origin), 0x80)
	})
}

// wpC2025VehicleData 构造一段 V2025 整车数据子记录(18 字节,字段顺序见
// codec/gbt2025/realtime/vehicle_data_v2025_codec.go),gear 为挡位原始字节;
// 其余字段取可逐字节还原的零值。
func wpC2025VehicleData(gear byte) []byte {
	w := utils.NewByteWriter()
	w.WriteUint8(0x01) // OperatingState
	w.WriteUint8(0x01) // ChargingState
	w.WriteUint8(0x01) // OperationMode
	w.WriteUint16(0)   // Speed
	w.WriteUint32(0)   // Mileage
	w.WriteUint16(0)   // Voltage
	w.WriteUint16(0)   // Current
	w.WriteUint8(0)    // SOC
	w.WriteUint8(0x01) // DC
	w.WriteUint8(gear) // GearPosition
	w.WriteUint16(0)   // Insulance
	return w.Bytes()
}

// wpC2016VehicleData 构造一段 V2016 整车数据子记录(20 字节,字段顺序见
// codec/gbt2016/realtime/vehicle_data_codec.go),gear 为挡位原始字节。
func wpC2016VehicleData(gear byte) []byte {
	w := utils.NewByteWriter()
	w.WriteUint8(0x01) // OperatingState
	w.WriteUint8(0x01) // ChargingState
	w.WriteUint8(0x01) // OperationMode
	w.WriteUint16(0)   // Speed
	w.WriteUint32(0)   // Mileage
	w.WriteUint16(0)   // Voltage
	w.WriteUint16(0)   // Current
	w.WriteUint8(0)    // SOC
	w.WriteUint8(0x01) // DC
	w.WriteUint8(gear) // GearPosition
	w.WriteUint16(0)   // Insulance
	w.WriteUint8(0)    // AccelerationValue
	w.WriteUint8(0)    // BrakePedalCondition
	return w.Bytes()
}

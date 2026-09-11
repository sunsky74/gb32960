package codec_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/model/gbt2016"
	mdlrt16 "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/model/gbt2025"
	v2025rt "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/utils"

	// Aggregator import: triggers all codec init() registrations so api.GetCodec
	// lookups below succeed for every V2016 + V2025 message body type.
	_ "github.com/sunsky74/gb32960/codec/all"
)

// TestRoundtrip exercises the encode→decode→re-encode byte-stability contract
// for every V2016 + V2025 message body type. For each sample:
//
//  1. Encode via model.MessageBody.Bytes() (delegates to the registered codec).
//  2. Decode via api.GetCodec(...).Decode(ByteReader).
//  3. Re-encode the decoded message.
//  4. Compare original vs re-encoded — must be byte-equal.
//
// This is the Plan 2 Task E1 acceptance test: any drift in codec field order,
// converter scale/offset, or list-count width surfaces here as a byte mismatch.
// All previously existing codec_test/*_roundtrip_test.go files continue to
// run alongside this suite; this file broadens coverage to every message type
// instead of relying on per-type ad-hoc tests.
func TestRoundtrip(t *testing.T) {
	t.Run("V2016", func(t *testing.T) {
		// ---- Top-level message bodies ----
		t.Run("VehicleLogin", func(t *testing.T) {
			assertRoundtrip(t, &gbt2016.VehicleLogin{
				BeanTime:  sampleBeanTime(),
				SerialNum: 1,
				ICCID:     "89860000000000000001",
				Count:     2,
				Length:    3,
				Codes:     []string{"001", "002"},
			})
		})
		t.Run("VehicleLogout", func(t *testing.T) {
			assertRoundtrip(t, &gbt2016.VehicleLogout{
				BeanTime:  sampleBeanTime(),
				SerialNum: 42,
			})
		})
		t.Run("PlatformLogin", func(t *testing.T) {
			assertRoundtrip(t, &gbt2016.PlatformLogin{
				BeanTime:  sampleBeanTime(),
				SerialNum: 7,
				Username:  "admin",
				Password:  "secret123",
				Cipher:    0x01,
			})
		})
		t.Run("PlatformLogout", func(t *testing.T) {
			assertRoundtrip(t, &gbt2016.PlatformLogout{
				BeanTime:  sampleBeanTime(),
				SerialNum: 9,
			})
		})

		// ---- Realtime sub-records (each tested standalone) ----
		t.Run("VehicleData", func(t *testing.T) {
			assertRoundtrip(t, &mdlrt16.VehicleData{
				OperatingState:      0x01,
				ChargingState:       0x02,
				OperationMode:       0x01,
				Speed:               35.5,
				Mileage:             12345.6,
				Voltage:             400.2,
				Current:             -50.5,
				SOC:                 78,
				DC:                  0x01,
				GearPosition:        mdlrt16.GearPosition{Origin: 0x31, GP: 0x01, DrivingForceActive: true},
				Insulance:           500,
				AccelerationValue:   45,
				BrakePedalCondition: 12,
			})
		})
		t.Run("MotorData", func(t *testing.T) {
			assertRoundtrip(t, &mdlrt16.MotorData{
				MotorSeq:              1,
				MotorState:            0x01,
				ControllerTemperature: 45.0,
				MotorSpeed:            3000.0,
				MotorTorque:           150.5,
				MotorTemperature:      60.0,
				ControllerVoltage:     400.0,
				ControllerCurrent:     100.0,
			})
		})
		t.Run("MotorDataList", func(t *testing.T) {
			assertRoundtrip(t, &mdlrt16.MotorDataList{
				Count: 2,
				Items: []mdlrt16.MotorData{
					{MotorSeq: 1, MotorState: 0x01, MotorSpeed: 3000.0, MotorTorque: 150.5},
					{MotorSeq: 2, MotorState: 0x02, MotorSpeed: 4500.0, MotorTorque: 220.0},
				},
			})
		})
		t.Run("FuelCellData", func(t *testing.T) {
			assertRoundtrip(t, &mdlrt16.FuelCellData{
				FuelCellVoltage:                      350.0,
				FuelCellCurrent:                      12.5,
				FuelConsumptionRate:                  1.23,
				TotalNumberOfFcTp:                    3,
				ProbeTemperatureValues:               []float64{25.0, 30.0, 35.0},
				HighestTempOfHydrogenSystem:          60.0,
				HighestTempProbeCodeOfHydrogenSystem: 2,
				HighestConOfHydrogen:                 5000,
				HighestHyConSensorCode:               1,
				HydrogenMaxPressure:                  35.0,
				HydrogenMaxPressureSensorCode:        3,
				HighVoltageDCState:                   0x01,
			})
		})
		t.Run("EngineData", func(t *testing.T) {
			assertRoundtrip(t, &mdlrt16.EngineData{
				EngineState:         0x01,
				CrankshaftSpeed:     3000,
				FuelConsumptionRate: 5.67,
			})
		})
		t.Run("LocationData", func(t *testing.T) {
			// Model stores signed degrees (Java parity: ÷1e6 + hemisphere bits)
			assertRoundtrip(t, &mdlrt16.LocationData{
				Valid:     true,
				Longitude: 116.4074,
				Latitude:  39.9042,
			})
			// South/west hemispheres exercise the sign bits (status 0x06)
			assertRoundtrip(t, &mdlrt16.LocationData{
				Valid:     true,
				Longitude: -116.4074,
				Latitude:  -39.9042,
			})
			// BYTE4 error sentinel passes through unscaled
			assertRoundtrip(t, &mdlrt16.LocationData{
				Valid:     false,
				Longitude: 4294967294,
				Latitude:  4294967295,
			})
		})
		t.Run("ExtremumData", func(t *testing.T) {
			assertRoundtrip(t, &mdlrt16.ExtremumData{
				VoltageMaxSubsystem:     1,
				VoltageMaxBattery:       5,
				MaxVoltage:              3.65,
				VoltageMinSubsystem:     2,
				VoltageMinBattery:       3,
				MinVoltage:              3.45,
				TemperatureMaxSubsystem: 1,
				TemperatureMaxProbe:     7,
				MaxTemperature:          45.0,
				TemperatureMinSubsystem: 1,
				TemperatureMinProbe:     9,
				MinTemperature:          25.0,
			})
		})
		t.Run("AlarmData", func(t *testing.T) {
			// AlarmData booleans are the source of truth on encode (codec
			// repacks them into AlarmBitIdentify). Set booleans, NOT the mask.
			assertRoundtrip(t, &mdlrt16.AlarmData{
				MaxAlarmLevel:             2,
				TemperatureDifferential:   true,
				BatteryHighTemperature:    false,
				DeviceTypeOverVoltage:     true,
				SocLow:                    true,
				MonomerBatteryOverVoltage: false,
				SocHigh:                   true,
				Insulation:                true,
				DcStatus:                  true,
				HighPressureInterlock:     true,
				BatteryFaultNum:           2,
				BatteryFaultDatas:         []int64{0x100, 0x200},
				MotorFaultNum:             1,
				MotorFaultDatas:           []int64{0x300},
				EngineFaultNum:            0,
				OtherFaultNum:             1,
				OtherFaultDatas:           []int64{0x400},
			})
		})
		t.Run("ChargeableSubsystemElectric", func(t *testing.T) {
			assertRoundtrip(t, &mdlrt16.ChargeableSubsystemElectric{
				ChargeableSubSystemNumber: 1,
				Voltage:                   350.0,
				Current:                   -45.0,
				BatteryTotalCount:         100,
				FrameStartBatterySeq:      1,
				BatteryCount:              3,
				BatteryVoltages:           []float64{3.45, 3.50, 3.55},
			})
		})
		t.Run("ChargeableSubsystemElectricList", func(t *testing.T) {
			assertRoundtrip(t, &mdlrt16.ChargeableSubsystemElectricList{
				ElectricCount: 2,
				Items: []mdlrt16.ChargeableSubsystemElectric{
					{ChargeableSubSystemNumber: 1, Voltage: 350.0, Current: -45.0, BatteryTotalCount: 96, FrameStartBatterySeq: 1, BatteryCount: 2, BatteryVoltages: []float64{3.45, 3.50}},
					{ChargeableSubSystemNumber: 2, Voltage: 352.0, Current: 12.5, BatteryTotalCount: 96, FrameStartBatterySeq: 1, BatteryCount: 1, BatteryVoltages: []float64{3.55}},
				},
			})
		})
		t.Run("ChargeableSubsystemTemperature", func(t *testing.T) {
			assertRoundtrip(t, &mdlrt16.ChargeableSubsystemTemperature{
				SubSystemNumber:       1,
				TemperatureProbeCount: 3,
				ProbeTemperatures:     []float64{25.0, 28.0, 32.0},
			})
		})
		t.Run("ChargeableSubsystemTemperatureList", func(t *testing.T) {
			assertRoundtrip(t, &mdlrt16.ChargeableSubsystemTemperatureList{
				TemperatureCount: 2,
				Items: []mdlrt16.ChargeableSubsystemTemperature{
					{SubSystemNumber: 1, TemperatureProbeCount: 2, ProbeTemperatures: []float64{25.0, 26.0}},
					{SubSystemNumber: 2, TemperatureProbeCount: 3, ProbeTemperatures: []float64{27.0, 28.0, 29.0}},
				},
			})
		})
		t.Run("GearPosition", func(t *testing.T) {
			// Origin is the source of truth on encode — derived fields are
			// recomputed on decode but never re-encoded.
			assertRoundtrip(t, &mdlrt16.GearPosition{Origin: 0x31})
		})
		// model.BeanTime has no Bytes() method (it is not a MessageBody); its
		// registered codec is exercised indirectly by every wrapper struct
		// that embeds a BeanTime field (VehicleLogin, KeyExchangeData, ...).

		// ---- RealTimeData composite (multiple sub-records in one frame) ----
		t.Run("RealTimeData_Composite", func(t *testing.T) {
			assertRoundtrip(t, &gbt2016.RealTimeData{
				BeanTime:      sampleBeanTime(),
				VehicleData:   &mdlrt16.VehicleData{OperatingState: 0x01, ChargingState: 0x02, OperationMode: 0x01, Speed: 30.0, Mileage: 1000.0, Voltage: 400.0, Current: -10.0, SOC: 75, DC: 0x01, GearPosition: mdlrt16.GearPosition{Origin: 0x31}, Insulance: 400, AccelerationValue: 30, BrakePedalCondition: 5},
				MotorDataList: &mdlrt16.MotorDataList{Count: 1, Items: []mdlrt16.MotorData{{MotorSeq: 1, MotorState: 0x01, MotorSpeed: 2000.0, MotorTorque: 100.0}}},
				EngineData:    &mdlrt16.EngineData{EngineState: 0x02, CrankshaftSpeed: 1500, FuelConsumptionRate: 4.5},
				LocationData:  &mdlrt16.LocationData{Valid: true, Longitude: 116.4074, Latitude: 39.9042},
				ExtremumData:  &mdlrt16.ExtremumData{VoltageMaxSubsystem: 1, VoltageMaxBattery: 1, MaxVoltage: 3.65, VoltageMinSubsystem: 1, VoltageMinBattery: 2, MinVoltage: 3.40, TemperatureMaxSubsystem: 1, TemperatureMaxProbe: 1, MaxTemperature: 35.0, TemperatureMinSubsystem: 1, TemperatureMinProbe: 2, MinTemperature: 25.0},
				AlarmData:     &mdlrt16.AlarmData{MaxAlarmLevel: 1, BatteryFaultNum: 1, BatteryFaultDatas: []int64{0x500}},
				ChargeableSubsystemElectricList: &mdlrt16.ChargeableSubsystemElectricList{
					ElectricCount: 1,
					Items: []mdlrt16.ChargeableSubsystemElectric{
						{ChargeableSubSystemNumber: 1, Voltage: 350.0, Current: -10.0, BatteryTotalCount: 96, FrameStartBatterySeq: 1, BatteryCount: 2, BatteryVoltages: []float64{3.45, 3.50}},
					},
				},
				ChargeableSubsystemTemperatureList: &mdlrt16.ChargeableSubsystemTemperatureList{
					TemperatureCount: 1,
					Items: []mdlrt16.ChargeableSubsystemTemperature{
						{SubSystemNumber: 1, TemperatureProbeCount: 2, ProbeTemperatures: []float64{25.0, 26.0}},
					},
				},
			})
		})
	})

	t.Run("V2025", func(t *testing.T) {
		// ---- Top-level message bodies ----
		t.Run("VehicleLoginV2025", func(t *testing.T) {
			// Per-code length slice + 24-byte codes. sum(Lengths)=2 codes.
			assertRoundtrip(t, &gbt2025.VehicleLoginV2025{
				BeanTime:  sampleBeanTime(),
				SerialNum: 5,
				ICCID:     "89860000000000000002",
				Count:     2,
				Lengths:   []int{1, 1},
				Codes: []string{
					"BMU001", "BMU002",
				},
			})
		})
		t.Run("PlatformLoginV2025", func(t *testing.T) {
			assertRoundtrip(t, &gbt2025.PlatformLoginV2025{
				BeanTime:  sampleBeanTime(),
				SerialNum: 1,
				Username:  "platform",
				Password:  "v2025secret",
				Cipher:    0x01,
			})
		})
		t.Run("PlatformLogoutV2025", func(t *testing.T) {
			assertRoundtrip(t, &gbt2025.PlatformLogoutV2025{
				BeanTime:  sampleBeanTime(),
				SerialNum: 1,
			})
		})
		t.Run("VehicleActivateResponse", func(t *testing.T) {
			assertRoundtrip(t, &gbt2025.VehicleActivateResponse{
				Success:      true,
				ResponseCode: 0x00,
			})
		})
		t.Run("KeyExchangeData", func(t *testing.T) {
			assertRoundtrip(t, &gbt2025.KeyExchangeData{
				Type:       0x02, // RSA
				Length:     4,
				Key:        []byte{0xDE, 0xAD, 0xBE, 0xEF},
				StartTime:  sampleBeanTime(),
				ExpireTime: model.BeanTime{Year: 26, Month: 12, Day: 31, Hour: 23, Minute: 59, Second: 59},
			})
		})
		t.Run("VehicleActivate", func(t *testing.T) {
			// Signature and PublicKey bytes are arbitrary here; roundtrip
			// stability only depends on length-prefixed byte preservation.
			assertRoundtrip(t, &gbt2025.VehicleActivate{
				CollectTime:     sampleBeanTime(),
				ChipID:          "CHIP2025ABC001",
				PublicKeyLength: 4,
				PublicKey:       []byte{0x01, 0x02, 0x03, 0x04},
				VIN:             "LSVAU2A37K2100001",
				Signature: &v2025rt.VehicleSignature{
					Type:    0x02, // RSA
					RLength: 2,
					RValue:  []byte{0xAA, 0xBB},
					SLength: 2,
					SValue:  []byte{0xCC, 0xDD},
				},
			})
		})

		// ---- V2025 realtime sub-records (standalone) ----
		t.Run("VehicleData_V2025Codec", func(t *testing.T) {
			// V2025 reuses the V2016 VehicleData struct TYPE but registers a
			// distinct V2025 codec under api.V2025 (see
			// codec/gbt2025/realtime/vehicle_data_v2025_codec.go): uses
			// CurrentConverter2025 (offset=3000, not 1000) and OMITS
			// AccelerationValue / BrakePedalCondition. Bypass the generic
			// helper (which would route via Version()=V2016) and exercise
			// the V2025 codec path directly.
			m := &mdlrt16.VehicleData{
				OperatingState: 0x01,
				ChargingState:  0x02,
				OperationMode:  0x01,
				Speed:          35.5,
				Mileage:        12345.6,
				Voltage:        400.2,
				Current:        -50.5,
				SOC:            78,
				DC:             0x01,
				GearPosition:   mdlrt16.GearPosition{Origin: 0x31},
				Insulance:      500,
				// AccelerationValue / BrakePedalCondition deliberately omitted —
				// V2025 codec does not encode them.
			}
			codec := api.GetCodec(api.V2025, reflect.TypeOf(m).Elem())
			if codec == nil {
				t.Fatalf("V2025 codec not found for VehicleData")
			}
			w := utils.NewByteWriter()
			if err := codec.Encode(w, m); err != nil {
				t.Fatalf("encode failed: %v", err)
			}
			encoded := w.Bytes()

			decoded, err := codec.Decode(utils.NewByteReader(encoded))
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			d := decoded.(*mdlrt16.VehicleData)

			w2 := utils.NewByteWriter()
			if err := codec.Encode(w2, d); err != nil {
				t.Fatalf("re-encode failed: %v", err)
			}
			reEncoded := w2.Bytes()
			if !bytes.Equal(encoded, reEncoded) {
				t.Errorf("byte-stability failed for V2025 VehicleData:\n  original   = %X (%d bytes)\n  re-encoded = %X (%d bytes)",
					encoded, len(encoded), reEncoded, len(reEncoded))
			}
			t.Logf("OK realtime.VehicleData (V2025 codec, %d bytes)", len(encoded))
		})
		t.Run("MotorDataV2025", func(t *testing.T) {
			assertRoundtrip(t, &v2025rt.MotorDataV2025{
				MotorSeq:              1,
				MotorState:            0x01,
				ControllerTemperature: 45.0,
				MotorSpeed:            3000.0,
				MotorTorque:           150.5,
				MotorTemperature:      60.0,
			})
		})
		t.Run("MotorDataV2025List", func(t *testing.T) {
			assertRoundtrip(t, &v2025rt.MotorDataV2025List{
				MotorCount: 2,
				Items: []v2025rt.MotorDataV2025{
					{MotorSeq: 1, MotorState: 0x01, ControllerTemperature: 45.0, MotorSpeed: 3000.0, MotorTorque: 150.5, MotorTemperature: 60.0},
					{MotorSeq: 2, MotorState: 0x02, ControllerTemperature: 50.0, MotorSpeed: 3500.0, MotorTorque: 180.0, MotorTemperature: 65.0},
				},
			})
		})
		t.Run("FuelCellEngineV2025Data", func(t *testing.T) {
			// HighestConOfHydrogen uses ConcentrationConverter2025 (scale=10000):
			// decoded range is 0..6.5535 (representing % volume, e.g. 1.5% = 15000 ppm).
			assertRoundtrip(t, &v2025rt.FuelCellEngineV2025Data{
				HighestTempOfHydrogenSystem:          65.0,
				HighestTempProbeCodeOfHydrogenSystem: 2,
				HighestConOfHydrogen:                 1.5,
				HighestHyConSensorCode:               1,
				HydrogenMaxPressure:                  70.0,
				HydrogenMaxPressureSensorCode:        3,
				HighVoltageDCState:                   0x01,
				FuelPercentage:                       80,
				DCControllerTemperature:              55.0,
			})
		})
		t.Run("EngineV2025Data", func(t *testing.T) {
			assertRoundtrip(t, &v2025rt.EngineV2025Data{
				CrankshaftSpeed: 3000,
			})
		})
		t.Run("LocationV2025Data", func(t *testing.T) {
			// East+north + GCJ02 passthrough (coordType=0x02). Only Origin
			// values are encoded; Convert* fields are derived on decode and
			// not re-encoded, so they cannot break byte stability.
			assertRoundtrip(t, &v2025rt.LocationV2025Data{
				Valid:           true,
				NorthernFlag:    true,
				EastFlag:        true,
				CoordinateType:  0x02,
				OriginLongitude: 116.4074,
				OriginLatitude:  39.9042,
			})
		})
		t.Run("AlarmV2025Data", func(t *testing.T) {
			// 28 boolean fields are source-of-truth on encode (repacked).
			assertRoundtrip(t, &v2025rt.AlarmV2025Data{
				MaxAlarmLevel:                3,
				TemperatureDifferential:      true,
				BatteryHighTemperature:       true,
				DeviceTypeOverVoltage:        false,
				SOCLow:                       true,
				Insulation:                   true,
				DCStatus:                     true,
				HighPressureInterlock:        true,
				DriveMotorOverSpeed:          true,
				HydrogenLeakage:              true,
				FuelCellStackOverTemperature: true,
				BatteryFaultNum:              1,
				BatteryFaultDatas:            []int64{0x1000},
				MotorFaultNum:                0,
				EngineFaultNum:               0,
				OtherFaultNum:                1,
				OtherFaultDatas:              []int64{0x2000},
				CommonAlertNum:               2,
				CommonAlertDatas: []v2025rt.CommonAlertData{
					{Seq: 1, Level: 2},
					{Seq: 5, Level: 1},
				},
			})
		})
		t.Run("MinParallelCellVoltage", func(t *testing.T) {
			assertRoundtrip(t, &v2025rt.MinParallelCellVoltage{
				BatteryPackSeq:   1,
				Voltage:          350.0,
				Current:          -45.0,
				MinParallelUnits: 3,
				BatteryVoltages:  []float64{3.65, 3.66, 3.67},
			})
		})
		t.Run("MinParallelCellVoltageList", func(t *testing.T) {
			assertRoundtrip(t, &v2025rt.MinParallelCellVoltageList{
				BatteryPackCount: 2,
				Items: []v2025rt.MinParallelCellVoltage{
					{BatteryPackSeq: 1, Voltage: 350.0, Current: -45.0, MinParallelUnits: 2, BatteryVoltages: []float64{3.65, 3.66}},
					{BatteryPackSeq: 2, Voltage: 352.0, Current: 12.5, MinParallelUnits: 1, BatteryVoltages: []float64{3.67}},
				},
			})
			// BYTE1 sentinel count: no entries follow (Java parity)
			assertRoundtrip(t, &v2025rt.MinParallelCellVoltageList{BatteryPackCount: 0xFF})
		})
		t.Run("BatteryTemp", func(t *testing.T) {
			assertRoundtrip(t, &v2025rt.BatteryTemp{
				BatteryPackSeq:        1,
				TemperatureProbeCount: 3,
				ProbeTemperatures:     []float64{25.0, 28.0, 32.0},
			})
		})
		t.Run("BatteryTempList", func(t *testing.T) {
			assertRoundtrip(t, &v2025rt.BatteryTempList{
				BatteryPackCount: 2,
				Items: []v2025rt.BatteryTemp{
					{BatteryPackSeq: 1, TemperatureProbeCount: 2, ProbeTemperatures: []float64{25.0, 26.0}},
					{BatteryPackSeq: 2, TemperatureProbeCount: 1, ProbeTemperatures: []float64{27.0}},
				},
			})
			// BYTE1 sentinel count: encode rewrites to fixed 0xFF (Java parity)
			assertRoundtrip(t, &v2025rt.BatteryTempList{BatteryPackCount: 0xFE})
		})
		t.Run("FuelCellStackData", func(t *testing.T) {
			assertRoundtrip(t, &v2025rt.FuelCellStackData{
				StackSeq:               1,
				Voltage:                250.0,
				Current:                80.0,
				GasPressure:            50.0,
				AirPressure:            40.0,
				AirInletTemp:           35.0,
				CoolingWaterProbeCount: 2,
				CoolingWaterTemps:      []float64{45.0, 47.0},
			})
		})
		t.Run("FuelCellStackDataList", func(t *testing.T) {
			assertRoundtrip(t, &v2025rt.FuelCellStackDataList{
				StackCount: 2,
				Items: []v2025rt.FuelCellStackData{
					{StackSeq: 1, Voltage: 250.0, Current: 80.0, GasPressure: 50.0, AirPressure: 40.0, AirInletTemp: 35.0, CoolingWaterProbeCount: 1, CoolingWaterTemps: []float64{45.0}},
					{StackSeq: 2, Voltage: 255.0, Current: 82.0, GasPressure: 52.0, AirPressure: 42.0, AirInletTemp: 36.0, CoolingWaterProbeCount: 2, CoolingWaterTemps: []float64{46.0, 47.0}},
				},
			})
			// BYTE1 sentinel count: count written as-is, no entries (Java parity)
			assertRoundtrip(t, &v2025rt.FuelCellStackDataList{StackCount: 0xFE})
		})
		t.Run("SuperCapacitorData", func(t *testing.T) {
			assertRoundtrip(t, &v2025rt.SuperCapacitorData{
				ManagementSystemNumber: 1,
				TotalVoltage:           48.0,
				TotalCurrent:           -15.0,
				CapacitorCount:         3,
				CapacitorVoltages:      []float64{2.7, 2.71, 2.72},
				TemperatureProbeCount:  2,
				ProbeTemperatures:      []float64{25.0, 26.0},
			})
		})
		t.Run("SuperCapacitorExtremumData", func(t *testing.T) {
			assertRoundtrip(t, &v2025rt.SuperCapacitorExtremumData{
				VoltageMaxSubsystem:     1,
				VoltageMaxBattery:       10,
				MaxVoltage:              2.85,
				VoltageMinSubsystem:     1,
				VoltageMinBattery:       5,
				MinVoltage:              2.55,
				TemperatureMaxSubsystem: 1,
				TemperatureMaxProbe:     2,
				MaxTemperature:          45.0,
				TemperatureMinSubsystem: 1,
				TemperatureMinProbe:     3,
				MinTemperature:          20.0,
			})
		})
		t.Run("VehicleSignature", func(t *testing.T) {
			// Standalone VehicleSignature: codec reads/writes Type + R + S.
			// SignData is NOT on wire — only populated by parent codec.
			assertRoundtrip(t, &v2025rt.VehicleSignature{
				Type:    0x02, // RSA
				RLength: 4,
				RValue:  []byte{0x11, 0x22, 0x33, 0x44},
				SLength: 4,
				SValue:  []byte{0x55, 0x66, 0x77, 0x88},
			})
		})

		// ---- RealTimeV2025Data composite (multiple sub-records) ----
		t.Run("RealTimeV2025Data_Composite", func(t *testing.T) {
			assertRoundtrip(t, &gbt2025.RealTimeV2025Data{
				BeanTime:      sampleBeanTime(),
				VehicleData:   &mdlrt16.VehicleData{OperatingState: 0x01, ChargingState: 0x02, OperationMode: 0x01, Speed: 30.0, Mileage: 1000.0, Voltage: 400.0, Current: -10.0, SOC: 75, DC: 0x01, GearPosition: mdlrt16.GearPosition{Origin: 0x31}, Insulance: 400, AccelerationValue: 30, BrakePedalCondition: 5},
				MotorDataList: &v2025rt.MotorDataV2025List{MotorCount: 1, Items: []v2025rt.MotorDataV2025{{MotorSeq: 1, MotorState: 0x01, ControllerTemperature: 45.0, MotorSpeed: 3000.0, MotorTorque: 150.5, MotorTemperature: 60.0}}},
				EngineData:    &v2025rt.EngineV2025Data{CrankshaftSpeed: 1500},
				LocationData:  &v2025rt.LocationV2025Data{Valid: true, NorthernFlag: true, EastFlag: true, CoordinateType: 0x02, OriginLongitude: 116.4074, OriginLatitude: 39.9042},
				AlarmData:     &v2025rt.AlarmV2025Data{MaxAlarmLevel: 1, BatteryFaultNum: 1, BatteryFaultDatas: []int64{0x500}, CommonAlertNum: 1, CommonAlertDatas: []v2025rt.CommonAlertData{{Seq: 1, Level: 2}}},
				MinParallelCellVoltages: &v2025rt.MinParallelCellVoltageList{
					BatteryPackCount: 1,
					Items: []v2025rt.MinParallelCellVoltage{
						{BatteryPackSeq: 1, Voltage: 350.0, Current: -45.0, MinParallelUnits: 2, BatteryVoltages: []float64{3.65, 3.66}},
					},
				},
				BatteryPackTemperatures: &v2025rt.BatteryTempList{
					BatteryPackCount: 1,
					Items: []v2025rt.BatteryTemp{
						{BatteryPackSeq: 1, TemperatureProbeCount: 2, ProbeTemperatures: []float64{25.0, 26.0}},
					},
				},
				FuelCellStackDataList: &v2025rt.FuelCellStackDataList{
					StackCount: 1,
					Items: []v2025rt.FuelCellStackData{
						{StackSeq: 1, Voltage: 250.0, Current: 80.0, GasPressure: 50.0, AirPressure: 40.0, AirInletTemp: 35.0, CoolingWaterProbeCount: 1, CoolingWaterTemps: []float64{45.0}},
					},
				},
				SuperCapacitorData:         &v2025rt.SuperCapacitorData{ManagementSystemNumber: 1, TotalVoltage: 48.0, TotalCurrent: -15.0, CapacitorCount: 2, CapacitorVoltages: []float64{2.7, 2.71}, TemperatureProbeCount: 1, ProbeTemperatures: []float64{25.0}},
				SuperCapacitorExtremumData: &v2025rt.SuperCapacitorExtremumData{VoltageMaxSubsystem: 1, VoltageMaxBattery: 1, MaxVoltage: 2.85, VoltageMinSubsystem: 1, VoltageMinBattery: 2, MinVoltage: 2.55, TemperatureMaxSubsystem: 1, TemperatureMaxProbe: 1, MaxTemperature: 35.0, TemperatureMinSubsystem: 1, TemperatureMinProbe: 2, MinTemperature: 25.0},
				VehicleSignature:           &v2025rt.VehicleSignature{Type: 0x02, RLength: 2, RValue: []byte{0xAA, 0xBB}, SLength: 2, SValue: []byte{0xCC, 0xDD}},
			})
		})

		// ---- CustomV2025Data roundtrip via parent (dispatcher) ----
		// The CustomV2025Data codec has an encode/decode asymmetry: Encode
		// writes CustomKey+Length+Data, but Decode reads only Length+Data
		// (the dispatcher has already consumed the CustomKey byte as the TLV
		// flag). So a standalone Bytes()→codec.Decode round trip does not
		// apply — verify through the parent RealTimeV2025Data path instead.
		t.Run("RealTimeV2025Data_CustomItem", func(t *testing.T) {
			assertRoundtrip(t, &gbt2025.RealTimeV2025Data{
				BeanTime:   sampleBeanTime(),
				CustomData: []v2025rt.CustomV2025Data{{CustomKey: 0x80, Length: 3, Data: []byte{0xAA, 0xBB, 0xCC}}},
			})
		})
	})
}

// assertRoundtrip encodes mb via its registered codec, decodes the result,
// re-encodes, and asserts byte-identical output. Any drift in field order,
// converter scale/offset, or list-count width surfaces as a byte mismatch.
func assertRoundtrip(t *testing.T, mb model.MessageBody) {
	t.Helper()

	original, err := mb.Bytes()
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	version := mb.Version()
	rt := reflect.TypeOf(mb)
	if rt.Kind() != reflect.Ptr {
		t.Fatalf("assertRoundtrip expects a pointer, got %T", mb)
	}
	structType := rt.Elem()
	codec := api.GetCodec(version, structType)
	if codec == nil {
		t.Fatalf("codec not found for %s (version %v)", structType.String(), version)
	}

	decoded, err := codec.Decode(utils.NewByteReader(original))
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	decodedMB, ok := decoded.(model.MessageBody)
	if !ok {
		t.Fatalf("decoded value %T does not implement model.MessageBody", decoded)
	}

	reEncoded, err := decodedMB.Bytes()
	if err != nil {
		t.Fatalf("re-encode failed: %v", err)
	}

	if !bytes.Equal(original, reEncoded) {
		t.Errorf("byte-stability failed for %s:\n  original   = %X (%d bytes)\n  re-encoded = %X (%d bytes)",
			structType.String(), original, len(original), reEncoded, len(reEncoded))
	}

	// Sanity: decoded struct type must match the input struct type — guards
	// against a codec that silently returns the wrong concrete type.
	decodedStructType := reflect.TypeOf(decodedMB).Elem()
	if decodedStructType != structType {
		t.Errorf("decoded type drift: input %s, decoded %s", structType.String(), decodedStructType.String())
	}

	// Emit a coverage line on success so the test output makes the breadth
	// of types covered visible without -v flag gymnastics.
	t.Logf("OK %s (version=%v, %d bytes)", structType.String(), version, len(original))
}

func sampleBeanTime() model.BeanTime {
	return model.BeanTime{Year: 26, Month: 7, Day: 31, Hour: 10, Minute: 30, Second: 0}
}

func mustBytes(t *testing.T, mb model.MessageBody) []byte {
	t.Helper()
	b, err := mb.Bytes()
	if err != nil {
		t.Fatalf("encode %T: %v", mb, err)
	}
	return b
}

func mustDecode(t *testing.T, raw []byte, version api.GBTVersion, wantType reflect.Type) api.Message {
	t.Helper()
	m, err := api.GetCodec(version, wantType).Decode(utils.NewByteReader(raw))
	if err != nil {
		t.Fatalf("decode %s: %v", wantType.String(), err)
	}
	return m
}

// TestV2025ListCountSentinels locks the exact wire bytes for the BYTE1
// sentinel-count contract of the three V2025 list codecs, mirroring Java:
// decode never reads entries after a sentinel count; encode behaves
// per-codec (stack writes the count as-is, min-parallel caps >50 at a single
// 0xFF, battery-temp rewrites any sentinel to the fixed 0xFF byte).
func TestV2025ListCountSentinels(t *testing.T) {
	stackType := reflect.TypeOf(v2025rt.FuelCellStackDataList{})
	minParallelType := reflect.TypeOf(v2025rt.MinParallelCellVoltageList{})
	tempType := reflect.TypeOf(v2025rt.BatteryTempList{})

	t.Run("FuelCellStackDataList", func(t *testing.T) {
		if got := mustBytes(t, &v2025rt.FuelCellStackDataList{StackCount: 0xFE}); !bytes.Equal(got, []byte{0xFE}) {
			t.Errorf("sentinel encode: got % X, want FE (count written as-is)", got)
		}
		m := mustDecode(t, []byte{0xFF}, api.V2025, stackType).(*v2025rt.FuelCellStackDataList)
		if m.StackCount != 0xFF || len(m.Items) != 0 {
			t.Errorf("sentinel decode: got count=%d items=%d, want 255/0", m.StackCount, len(m.Items))
		}
	})

	t.Run("MinParallelCellVoltageList", func(t *testing.T) {
		if got := mustBytes(t, &v2025rt.MinParallelCellVoltageList{BatteryPackCount: 0xFF}); !bytes.Equal(got, []byte{0xFF}) {
			t.Errorf("sentinel encode: got % X, want FF", got)
		}
		over := &v2025rt.MinParallelCellVoltageList{BatteryPackCount: 51, Items: make([]v2025rt.MinParallelCellVoltage, 51)}
		if got := mustBytes(t, over); !bytes.Equal(got, []byte{0xFF}) {
			t.Errorf(">50 encode: got %d bytes, want single FF", len(got))
		}
		m := mustDecode(t, []byte{0xFE}, api.V2025, minParallelType).(*v2025rt.MinParallelCellVoltageList)
		if m.BatteryPackCount != 0xFE || len(m.Items) != 0 {
			t.Errorf("sentinel decode: got count=%d items=%d, want 254/0", m.BatteryPackCount, len(m.Items))
		}
	})

	t.Run("BatteryTempList", func(t *testing.T) {
		// Java writes DataErrorValue.BYTE1.getInvalid() (0xFF) for ANY
		// sentinel count — 0xFE is rewritten, not passed through.
		if got := mustBytes(t, &v2025rt.BatteryTempList{BatteryPackCount: 0xFE}); !bytes.Equal(got, []byte{0xFF}) {
			t.Errorf("sentinel encode: got % X, want fixed FF", got)
		}
		m := mustDecode(t, []byte{0xFF}, api.V2025, tempType).(*v2025rt.BatteryTempList)
		if m.BatteryPackCount != 0xFF || len(m.Items) != 0 {
			t.Errorf("sentinel decode: got count=%d items=%d, want 255/0", m.BatteryPackCount, len(m.Items))
		}
	})
}

// TestV2025BatteryTempEntryClamps locks the entry-level encode clamps of
// BatteryTemp, mirroring Java BatteryPackTemperatureCodec.encodeBuffer:
// pack seq is rewritten to 0xFF when sentinel/negative/>50, and a BYTE2
// sentinel probe count is rewritten to the fixed 0xFFFF with temps dropped.
func TestV2025BatteryTempEntryClamps(t *testing.T) {
	// seq 51 (>50) → 0xFF; remaining fields still written (25.0°C → 25+40 = 0x41)
	raw := mustBytes(t, &v2025rt.BatteryTemp{BatteryPackSeq: 51, TemperatureProbeCount: 1, ProbeTemperatures: []float64{25.0}})
	if !bytes.Equal(raw, []byte{0xFF, 0x00, 0x01, 0x41}) {
		t.Errorf("seq>50: got % X", raw)
	}

	// seq sentinel → 0xFF; zero count still written as u16
	raw = mustBytes(t, &v2025rt.BatteryTemp{BatteryPackSeq: 0xFE})
	if !bytes.Equal(raw, []byte{0xFF, 0x00, 0x00}) {
		t.Errorf("seq sentinel: got % X", raw)
	}

	// probe count 0xFFFE → fixed 0xFFFF rewrite, temps dropped
	raw = mustBytes(t, &v2025rt.BatteryTemp{BatteryPackSeq: 1, TemperatureProbeCount: 0xFFFE, ProbeTemperatures: []float64{25.0}})
	if !bytes.Equal(raw, []byte{0x01, 0xFF, 0xFF}) {
		t.Errorf("count sentinel rewrite: got % X", raw)
	}
}

// TestV2025AlarmMaskPreservation locks the AlarmV2025 encode mask rule:
// a non-zero AlarmBitIdentify is written back as-is (reserved bits 28..31
// survive the roundtrip, mirroring Java's non-null branch); a zero mask is
// rebuilt from the 28 booleans (Java's null branch).
func TestV2025AlarmMaskPreservation(t *testing.T) {
	src := &v2025rt.AlarmV2025Data{MaxAlarmLevel: 1, AlarmBitIdentify: 0x40000001, TemperatureDifferential: true}
	raw := mustBytes(t, src)
	if !bytes.Equal(raw[1:5], []byte{0x40, 0x00, 0x00, 0x01}) {
		t.Errorf("stored mask not preserved: got % X", raw[1:5])
	}
	m := mustDecode(t, raw, api.V2025, reflect.TypeOf(v2025rt.AlarmV2025Data{})).(*v2025rt.AlarmV2025Data)
	if m.AlarmBitIdentify != 0x40000001 || !m.TemperatureDifferential {
		t.Errorf("decode: mask=%#X bit0=%v", m.AlarmBitIdentify, m.TemperatureDifferential)
	}
	if re := mustBytes(t, m); !bytes.Equal(re, raw) {
		t.Errorf("roundtrip: got % X want % X", re, raw)
	}

	// zero mask + booleans → rebuilt (bit0 | bit24 = 0x01000001)
	rb := mustBytes(t, &v2025rt.AlarmV2025Data{MaxAlarmLevel: 1, TemperatureDifferential: true, HydrogenLeakage: true})
	if !bytes.Equal(rb[1:5], []byte{0x01, 0x00, 0x00, 0x01}) {
		t.Errorf("rebuilt mask: got % X", rb[1:5])
	}
}

// TestV2025RealtimeItemsPreserveOrder verifies the items encode path: a
// frame whose TLV order differs from the canonical field order (alarm before
// vehicle), a repeated type (two motor lists), custom data, and a trailing
// signature all survive decode→re-encode byte-for-byte, and the signature's
// SignData is refreshed to cover everything written before the 0xFF flag.
func TestV2025RealtimeItemsPreserveOrder(t *testing.T) {
	bt := sampleBeanTime()
	alarm := &v2025rt.AlarmV2025Data{MaxAlarmLevel: 1}
	vehicle := &mdlrt16.VehicleData{SOC: 50}
	motor1 := &v2025rt.MotorDataV2025List{MotorCount: 1, Items: []v2025rt.MotorDataV2025{{MotorSeq: 1, MotorState: 0x01}}}
	motor2 := &v2025rt.MotorDataV2025List{MotorCount: 1, Items: []v2025rt.MotorDataV2025{{MotorSeq: 2, MotorState: 0x02}}}
	custom := &v2025rt.CustomV2025Data{CustomKey: 0x80, Length: 1, Data: []byte{0xAA}}
	sig := &v2025rt.VehicleSignature{Type: 0x02, RLength: 2, RValue: []byte{0xAA, 0xBB}, SLength: 2, SValue: []byte{0xCC, 0xDD}}

	buf := utils.NewByteWriter()
	if err := api.GetCodec(api.V2016, reflect.TypeOf(model.BeanTime{})).Encode(buf, &bt); err != nil {
		t.Fatalf("beantime: %v", err)
	}
	appendTLV := func(flag byte, body api.Message) {
		t.Helper()
		codec := api.GetCodec(api.V2025, reflect.TypeOf(body).Elem())
		if codec == nil {
			t.Fatalf("no V2025 codec for %T", body)
		}
		buf.WriteUint8(flag)
		if err := codec.Encode(buf, body); err != nil {
			t.Fatalf("encode %T: %v", body, err)
		}
	}
	appendTLV(0x06, alarm) // alarm BEFORE vehicle — non-canonical order
	appendTLV(0x01, vehicle)
	appendTLV(0x02, motor1)
	appendTLV(0x02, motor2) // repeated type
	if err := api.GetCodec(api.V2025, reflect.TypeOf(v2025rt.CustomV2025Data{})).Encode(buf, custom); err != nil {
		t.Fatalf("custom: %v", err)
	}
	appendTLV(0xFF, sig)
	wire := buf.Bytes()

	decoded := mustDecode(t, wire, api.V2025, reflect.TypeOf(gbt2025.RealTimeV2025Data{})).(*gbt2025.RealTimeV2025Data)
	wantOrder := []byte{0x06, 0x01, 0x02, 0x02, 0x80, 0xFF}
	if len(decoded.Items) != len(wantOrder) {
		t.Fatalf("items: got %d, want %d", len(decoded.Items), len(wantOrder))
	}
	for i, want := range wantOrder {
		if decoded.Items[i].Type != want {
			t.Errorf("items[%d].Type: got %#02X, want %#02X", i, decoded.Items[i].Type, want)
		}
	}
	if decoded.MotorDataList == nil || decoded.MotorDataList.Items[0].MotorSeq != 2 {
		t.Errorf("typed motor field should hold the LAST motor list, got %+v", decoded.MotorDataList)
	}

	// signature TLV = 1 flag + 1 type + 2 rlen + 2 r + 2 slen + 2 s = 10 bytes;
	// decode-side fill covers everything written before the 0xFF flag byte
	wantSignData := wire[:len(wire)-10]
	if !bytes.Equal(decoded.VehicleSignature.SignData, wantSignData) {
		t.Errorf("decode-side SignData mismatch:\n got % X\nwant % X", decoded.VehicleSignature.SignData, wantSignData)
	}

	if re := mustBytes(t, decoded); !bytes.Equal(re, wire) {
		t.Errorf("items re-encode not byte-faithful:\n got % X\nwant % X", re, wire)
	}
}

// TestVehicleActivateDecodeSignData locks the decode-side SignData fill of
// VehicleActivate: 签名信息紧接 VIN 之后开始(国标 2025 表 B.3),被签数据覆盖
// 数据采集时间首字节至 VIN 末字节(含),即签名字段之前的全部 45 字节。
func TestVehicleActivateDecodeSignData(t *testing.T) {
	bt := sampleBeanTime()
	buf := utils.NewByteWriter()
	if err := api.GetCodec(api.V2016, reflect.TypeOf(model.BeanTime{})).Encode(buf, &bt); err != nil {
		t.Fatalf("beantime: %v", err)
	}
	buf.WriteString("CHIP20250731001", 16)
	buf.WriteUint16(4)
	buf.WriteBytes([]byte{0x0A, 0x0B, 0x0C, 0x0D})
	buf.WriteString("VIN1234567890AB", 17)
	sig := &v2025rt.VehicleSignature{Type: 0x02, RLength: 1, RValue: []byte{0xEE}, SLength: 1, SValue: []byte{0xFF}}
	sigCodec := api.GetCodec(api.V2025, reflect.TypeOf(v2025rt.VehicleSignature{}))
	if sigCodec == nil {
		t.Fatal("signature codec not found")
	}
	if err := sigCodec.Encode(buf, sig); err != nil {
		t.Fatalf("signature encode: %v", err)
	}
	wire := buf.Bytes()

	decoded := mustDecode(t, wire, api.V2025, reflect.TypeOf(gbt2025.VehicleActivate{})).(*gbt2025.VehicleActivate)

	// 6B time + 16B chip + 2B keyLen + 4B key + 17B VIN = 45 bytes before the
	// signature; SignData covers all of them (VIN 末字节含).
	want := wire[:45]
	if !bytes.Equal(decoded.Signature.SignData, want) {
		t.Errorf("SignData mismatch:\n got % X\nwant % X", decoded.Signature.SignData, want)
	}
}

// TestV2016RealtimeCustomData locks the 2016 user-custom-data handling:
// TLV flag 0x80~0xFE is followed by WORD length + BYTE[N] body (国标 2016
// 表 8/表 19) and stored raw in RealTimeData.CustomData keyed by the flag;
// re-encode writes them back byte-faithfully (keys emitted in sorted order
// for determinism).
func TestV2016RealtimeCustomData(t *testing.T) {
	m := &gbt2016.RealTimeData{
		BeanTime: sampleBeanTime(),
		CustomData: map[byte][]byte{
			0x80: {0xAA, 0xBB},
			0x81: {0xCC},
		},
	}
	raw := mustBytes(t, m)
	decoded := mustDecode(t, raw, api.V2016, reflect.TypeOf(gbt2016.RealTimeData{})).(*gbt2016.RealTimeData)
	if !bytes.Equal(decoded.CustomData[0x80], []byte{0xAA, 0xBB}) || !bytes.Equal(decoded.CustomData[0x81], []byte{0xCC}) {
		t.Fatalf("custom data: % X", decoded.CustomData)
	}
	if re := mustBytes(t, decoded); !bytes.Equal(re, raw) {
		t.Errorf("roundtrip:\n got % X\nwant % X", re, raw)
	}
}

package codec_test

import (
	"math"
	"os"
	"strings"
	"testing"

	"github.com/sunsky74/gb32960/codec"
	_ "github.com/sunsky74/gb32960/codec/all"
	"github.com/sunsky74/gb32960/frame"
	mdl "github.com/sunsky74/gb32960/model/gbt2016"
	"github.com/sunsky74/gb32960/utils"
)

func TestLayerA_RealWorldPackets(t *testing.T) {
	tests := []struct {
		file   string
		expect realtimeExpect
	}{
		{
			file: "../golden/layer_a/prod_realtime_v2016_01.hex",
			expect: realtimeExpect{
				version: 2016, vin: "H3V21AA24RZ016158",
				speed: 4.80, mileage: 93266.0, voltage: 341.70, current: 2.50, soc: 50,
				longitude: 114.366778, latitude: 30.591045,
				motorCount: 1, electricListCount: 1, batteryTotalCount: 104,
				tempListCount: 1, probeCount: 8,
				maxVoltage: 3.28, minVoltage: 3.28, maxTemp: 29.0, minTemp: 27.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			hexData, err := os.ReadFile(tt.file)
			if err != nil {
				t.Skipf("golden file not found: %s", tt.file)
				return
			}
			hexStr := strings.TrimSpace(string(hexData))
			data, err := utils.HexToBytes(hexStr)
			if err != nil {
				t.Fatalf("hex decode failed: %v", err)
			}
			t.Logf("Packet size: %d bytes", len(data))

			r := utils.NewByteReader(data)
			msg, err := codec.ProtocolCodec.Decode(r)
			if err != nil {
				t.Fatalf("ProtocolCodec.Decode: %v", err)
			}
			pm := msg.(*frame.ProtocolMessage)

			if int(pm.Version) != tt.expect.version {
				t.Errorf("version: got %d, want %d", pm.Version, tt.expect.version)
			}
			if pm.VIN != tt.expect.vin {
				t.Errorf("vin: got %q, want %q", pm.VIN, tt.expect.vin)
			}
			t.Logf("Frame: v=%d, VIN=%s, enc=%d, len=%d, BCC=0x%02X",
				pm.Version, pm.VIN, pm.Encryption, pm.PayloadLength, pm.CheckCode)

			if err := pm.DecodePayload(); err != nil {
				t.Fatalf("DecodePayload: %v", err)
			}
			rt := pm.Payload.(*mdl.RealTimeData)

			// VehicleData
			if rt.VehicleData == nil {
				t.Fatal("VehicleData is nil")
			}
			vd := rt.VehicleData
			assertFloat(t, "speed", vd.Speed, tt.expect.speed, 0.05)
			assertFloat(t, "mileage", vd.Mileage, tt.expect.mileage, 1.0)
			assertFloat(t, "voltage", vd.Voltage, tt.expect.voltage, 0.5)
			assertFloat(t, "current", vd.Current, tt.expect.current, 0.1)
			assertInt(t, "soc", vd.SOC, tt.expect.soc)
			t.Logf("VehicleData: speed=%.2f, mileage=%.0f, voltage=%.2f, current=%.2f, soc=%d",
				vd.Speed, vd.Mileage, vd.Voltage, vd.Current, vd.SOC)

			// LocationData (model stores signed degrees; codec applies ÷1e6
			// and hemisphere sign, mirroring Java LocationDataCodec)
			if rt.LocationData == nil {
				t.Fatal("LocationData is nil")
			}
			assertFloat(t, "longitude", rt.LocationData.Longitude, tt.expect.longitude, 0.0001)
			assertFloat(t, "latitude", rt.LocationData.Latitude, tt.expect.latitude, 0.0001)
			// Note: Valid=true means positioning IS valid (status bit0 == 0)
			t.Logf("LocationData: lon=%.6f, lat=%.6f, valid=%v",
				rt.LocationData.Longitude, rt.LocationData.Latitude, rt.LocationData.Valid)

			// MotorDataList
			if rt.MotorDataList == nil {
				t.Fatal("MotorDataList is nil")
			}
			assertInt(t, "motorCount", rt.MotorDataList.Count, tt.expect.motorCount)
			assertInt(t, "motorItems", len(rt.MotorDataList.Items), tt.expect.motorCount)
			t.Logf("MotorDataList: count=%d", rt.MotorDataList.Count)

			// ChargeableSubsystemElectricList
			if rt.ChargeableSubsystemElectricList == nil {
				t.Fatal("ChargeableSubsystemElectricList is nil")
			}
			el := rt.ChargeableSubsystemElectricList
			assertInt(t, "elecCount", el.ElectricCount, tt.expect.electricListCount)
			if len(el.Items) > 0 {
				e0 := &el.Items[0]
				assertInt(t, "batteryTotal", e0.BatteryTotalCount, tt.expect.batteryTotalCount)
				assertInt(t, "batVoltagesLen", len(e0.BatteryVoltages), tt.expect.batteryTotalCount)
				t.Logf("Electric: subSys=%d, totalBat=%d, inFrame=%d, v0=%.3f, vLast=%.3f",
					e0.ChargeableSubSystemNumber, e0.BatteryTotalCount,
					e0.BatteryCount, e0.BatteryVoltages[0], e0.BatteryVoltages[e0.BatteryCount-1])
			}

			// ChargeableSubsystemTemperatureList
			if rt.ChargeableSubsystemTemperatureList == nil {
				t.Fatal("ChargeableSubsystemTemperatureList is nil")
			}
			tl := rt.ChargeableSubsystemTemperatureList
			assertInt(t, "tempCount", tl.TemperatureCount, tt.expect.tempListCount)
			if len(tl.Items) > 0 {
				t0 := &tl.Items[0]
				assertInt(t, "probeCount", t0.TemperatureProbeCount, tt.expect.probeCount)
				t.Logf("Temp: subSys=%d, probes=%d, t0=%.0f, tLast=%.0f",
					t0.SubSystemNumber, t0.TemperatureProbeCount,
					t0.ProbeTemperatures[0], t0.ProbeTemperatures[t0.TemperatureProbeCount-1])
			}

			// ExtremumData
			if rt.ExtremumData == nil {
				t.Fatal("ExtremumData is nil")
			}
			ex := rt.ExtremumData
			assertFloat(t, "maxVoltage", ex.MaxVoltage, tt.expect.maxVoltage, 0.01)
			assertFloat(t, "minVoltage", ex.MinVoltage, tt.expect.minVoltage, 0.01)
			assertFloat(t, "maxTemp", ex.MaxTemperature, tt.expect.maxTemp, 0.5)
			assertFloat(t, "minTemp", ex.MinTemperature, tt.expect.minTemp, 0.5)
			t.Logf("Extremum: maxV=%.2f(B%d.%d), minV=%.2f(B%d.%d), maxT=%.1f(P%d), minT=%.1f(P%d)",
				ex.MaxVoltage, ex.VoltageMaxSubsystem, ex.VoltageMaxBattery,
				ex.MinVoltage, ex.VoltageMinSubsystem, ex.VoltageMinBattery,
				ex.MaxTemperature, ex.TemperatureMaxProbe,
				ex.MinTemperature, ex.TemperatureMinProbe)

			t.Log("✅ Layer A: PASS — real production packet decoded, all semantic fields match Java")
		})
	}
}

type realtimeExpect struct {
	version                          int
	vin                              string
	speed, mileage, voltage, current float64
	soc                              int
	longitude, latitude              float64
	motorCount                       int
	electricListCount                int
	batteryTotalCount                int
	tempListCount                    int
	probeCount                       int
	maxVoltage, minVoltage           float64
	maxTemp, minTemp                 float64
}

func assertFloat(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Errorf("%s: got %.6f, want %.6f (diff=%.6f)", name, got, want, math.Abs(got-want))
	}
}

func assertInt(t *testing.T, name string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %d, want %d", name, got, want)
	}
}

// TestLayerA_V2025 validates a V2025 reissue production packet against known Java output.
func TestLayerA_V2025(t *testing.T) {
	file := "../golden/layer_a/prod_reissue_v2025_01.hex"
	hexData, err := os.ReadFile(file)
	if err != nil {
		t.Skipf("golden file not found: %s", file)
		return
	}
	data, _ := utils.HexToBytes(strings.TrimSpace(string(hexData)))
	r := utils.NewByteReader(data)
	msg, err := codec.ProtocolCodec.Decode(r)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	pm := msg.(*frame.ProtocolMessage)

	assertInt(t, "version", int(pm.Version), 2025)
	if pm.VIN != "H3V21BA25SZ007433" {
		t.Errorf("vin: got %q", pm.VIN)
	}
	t.Logf("Frame: v=%d, VIN=%s, len=%d, BCC=0x%02X",
		pm.Version, pm.VIN, pm.PayloadLength, pm.CheckCode)

	if err := pm.DecodePayload(); err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}

	t.Logf("Payload type: %T", pm.Payload)
	t.Log("✅ Layer A V2025: production reissue packet decoded successfully")
}

// TestLayerA_VehicleLogin validates V2016 vehicle login production packet.
func TestLayerA_VehicleLogin(t *testing.T) {
	file := "../golden/layer_a/prod_login_v2016_01.hex"
	hexData, err := os.ReadFile(file)
	if err != nil {
		t.Skipf("golden file not found: %s", file)
		return
	}
	data, _ := utils.HexToBytes(strings.TrimSpace(string(hexData)))
	r := utils.NewByteReader(data)
	msg, err := codec.ProtocolCodec.Decode(r)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	pm := msg.(*frame.ProtocolMessage)

	assertInt(t, "version", int(pm.Version), 2016)
	if pm.VIN != "H3V21BA29SZ003109" {
		t.Errorf("vin: got %q", pm.VIN)
	}
	t.Logf("Frame: v=%d, VIN=%s, len=%d, BCC=0x%02X",
		pm.Version, pm.VIN, pm.PayloadLength, pm.CheckCode)

	if err := pm.DecodePayload(); err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}
	t.Logf("Payload type: %T", pm.Payload)
	t.Logf("Payload: %+v", pm.Payload)
	t.Log("✅ Layer A: V2016 vehicle login packet decoded successfully")
}

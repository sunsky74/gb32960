package codec_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	"github.com/sunsky74/gb32960/frame"
	"github.com/sunsky74/gb32960/utils"

	_ "github.com/sunsky74/gb32960/codec/all"
)

// TestLayerC_GoldenRegression validates that Go decodes Java-encoded golden
// files and re-encodes them byte-for-byte identical.
//
// For each .hex file in golden/layer_c/, the test:
//  1. Reads the Java-encoded hex
//  2. Decodes the full protocol frame
//  3. Auto-decodes the payload via DecodePayload()
//  4. Re-encodes the frame via frame.ProtocolMessage.Bytes()
//  5. Compares: Go re-encoded bytes must match Java original bytes
//
// When golden files are absent, the test skips gracefully.
func TestLayerC_GoldenRegression(t *testing.T) {
	goldenDir := "golden/layer_c"
	files, err := filepath.Glob(filepath.Join(goldenDir, "*.hex"))
	if err != nil || len(files) == 0 {
		t.Skipf("no golden files found in %s — run Java GenerateAllGoldenFiles first", goldenDir)
	}

	passed, failed := 0, 0
	for _, hexFile := range files {
		name := strings.TrimSuffix(filepath.Base(hexFile), ".hex")
		t.Run(name, func(t *testing.T) {
			// 1. Read Java-encoded hex
			javaHex, err := os.ReadFile(hexFile)
			if err != nil {
				t.Fatalf("read golden file: %v", err)
			}
			javaBytes, err := utils.HexToBytes(strings.TrimSpace(string(javaHex)))
			if err != nil {
				t.Fatalf("hex decode: %v", err)
			}

			// 2. Decode as full protocol frame
			r := utils.NewByteReader(javaBytes)
			msg, err := codec.ProtocolCodec.Decode(r)
			if err != nil {
				t.Fatalf("ProtocolCodec.Decode: %v", err)
			}
			pm, ok := msg.(*frame.ProtocolMessage)
			if !ok {
				t.Fatalf("decoded message is not *frame.ProtocolMessage")
			}

			// 3. Auto-decode payload
			if err := pm.DecodePayload(); err != nil {
				t.Fatalf("DecodePayload: %v", err)
			}

			// 4. Re-encode the frame
			goBytes, err := pm.Bytes()
			if err != nil {
				t.Fatalf("ProtocolMessage.Bytes(): %v", err)
			}

			// 5. Byte-for-byte comparison with Java original
			if len(goBytes) != len(javaBytes) {
				t.Errorf("length mismatch: Go=%d, Java=%d", len(goBytes), len(javaBytes))
				return
			}
			for i := range javaBytes {
				if goBytes[i] != javaBytes[i] {
					t.Errorf("byte[%d] mismatch: Go=0x%02X, Java=0x%02X", i, goBytes[i], javaBytes[i])
					return
				}
			}
			passed++
		})
	}

	// 6. Also decode payload-only golden files (without protocol frame wrapper)
	payloadFiles, _ := filepath.Glob(filepath.Join(goldenDir, "payload_*.hex"))
	if len(payloadFiles) > 0 {
		t.Run("PayloadOnly", func(t *testing.T) {
			for _, hexFile := range payloadFiles {
				name := strings.TrimPrefix(strings.TrimSuffix(filepath.Base(hexFile), ".hex"), "payload_")
				t.Run(name, func(t *testing.T) {
					javaHex, _ := os.ReadFile(hexFile)
					javaBytes, _ := utils.HexToBytes(strings.TrimSpace(string(javaHex)))

					// Determine version from filename convention
					version := api.V2016
					if strings.Contains(name, "v2025") || strings.Contains(name, "2025") {
						version = api.V2025
					}

					// Load the codec for the payload type
					msgType := messageTypeFromPayload(name)
					cdc := api.GetCodec(version, msgType)
					if cdc == nil && version == api.V2016 {
						cdc = api.GetCodec(api.V2025, msgType)
					}
					if cdc == nil {
						t.Skipf("no codec registered for payload: %s", name)
					}

					r := utils.NewByteReader(javaBytes)
					decoded, err := cdc.Decode(r)
					if err != nil {
						t.Fatalf("decode payload: %v", err)
					}

					w := utils.NewByteWriter()
					if err := cdc.Encode(w, decoded); err != nil {
						t.Fatalf("re-encode payload: %v", err)
					}
					goBytes := w.Bytes()

					if len(goBytes) != len(javaBytes) {
						t.Errorf("payload length mismatch: Go=%d, Java=%d", len(goBytes), len(javaBytes))
						return
					}
					for i := range javaBytes {
						if goBytes[i] != javaBytes[i] {
							t.Errorf("payload byte[%d] mismatch: Go=0x%02X, Java=0x%02X", i, goBytes[i], javaBytes[i])
							return
						}
					}
					passed++
				})
			}
		})
	}

	t.Logf("Layer C golden regression: %d passed, %d failed", passed, failed)
	if failed > 0 {
		t.Errorf("%d golden files failed byte-for-byte comparison", failed)
	}
}

// messageTypeFromPayload returns the reflect.Type for a golden file payload name.
// This is a helper used when golden files contain payload bytes without the
// protocol frame wrapper.
//
// TODO: expand this map to cover all payload types as golden files grow.
func messageTypeFromPayload(name string) reflect.Type {
	// Map golden file payload names to their reflect types.
	// Example entries:
	//   "vehicle_login_2016"        → reflect.TypeOf((*gbt2016.VehicleLogin)(nil)).Elem()
	//   "vehicle_login_2025"        → reflect.TypeOf((*gbt2025.VehicleLoginV2025)(nil)).Elem()
	//   "platform_login_2016"       → reflect.TypeOf((*gbt2016.PlatformLogin)(nil)).Elem()
	//   "platform_login_2025"       → reflect.TypeOf((*gbt2016.PlatformLogin)(nil)).Elem()
	//   "vehicle_logout_2016"       → reflect.TypeOf((*gbt2016.VehicleLogout)(nil)).Elem()
	//   "vehicle_activate_2025"     → reflect.TypeOf((*gbt2025.VehicleActivate)(nil)).Elem()
	//   "vehicle_activate_resp"     → reflect.TypeOf((*gbt2025.VehicleActivateResponse)(nil)).Elem()
	//   "key_exchange_2025"         → reflect.TypeOf((*gbt2025.KeyExchangeData)(nil)).Elem()
	//   "real_time_data_2016"       → reflect.TypeOf((*gbt2016.RealTimeData)(nil)).Elem()
	//   "real_time_data_2025"       → reflect.TypeOf((*gbt2025.RealTimeV2025Data)(nil)).Elem()
	_ = name // placeholder — not yet implemented
	return nil
}

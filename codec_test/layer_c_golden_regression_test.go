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

// TestLayerC_GoldenRegression 验证 Go 能解码 Java 编码的金样文件,
// 并重编码出逐字节一致的结果。
//
// 对 golden/layer_c/ 中的每个 .hex 文件,测试会:
//  1. 读取 Java 编码的十六进制
//  2. 解码完整协议帧
//  3. 通过 DecodePayload() 自动解码数据单元
//  4. 通过 frame.ProtocolMessage.Bytes() 重编码该帧
//  5. 比较:Go 重编码的字节必须与 Java 原始字节一致
//
// 金样文件缺失时,测试优雅跳过。
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
			// 1. 读取 Java 编码的十六进制
			javaHex, err := os.ReadFile(hexFile)
			if err != nil {
				t.Fatalf("read golden file: %v", err)
			}
			javaBytes, err := utils.HexToBytes(strings.TrimSpace(string(javaHex)))
			if err != nil {
				t.Fatalf("hex decode: %v", err)
			}

			// 2. 按完整协议帧解码
			r := utils.NewByteReader(javaBytes)
			msg, err := codec.ProtocolCodec.Decode(r)
			if err != nil {
				t.Fatalf("ProtocolCodec.Decode: %v", err)
			}
			pm, ok := msg.(*frame.ProtocolMessage)
			if !ok {
				t.Fatalf("decoded message is not *frame.ProtocolMessage")
			}

			// 3. 自动解码数据单元
			if err := pm.DecodePayload(); err != nil {
				t.Fatalf("DecodePayload: %v", err)
			}

			// 4. 重编码该帧
			goBytes, err := pm.Bytes()
			if err != nil {
				t.Fatalf("ProtocolMessage.Bytes(): %v", err)
			}

			// 5. 与 Java 原文件逐字节比较
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

	// 6. 同时解码仅含数据单元的金样文件(不带协议帧外壳)
	payloadFiles, _ := filepath.Glob(filepath.Join(goldenDir, "payload_*.hex"))
	if len(payloadFiles) > 0 {
		t.Run("PayloadOnly", func(t *testing.T) {
			for _, hexFile := range payloadFiles {
				name := strings.TrimPrefix(strings.TrimSuffix(filepath.Base(hexFile), ".hex"), "payload_")
				t.Run(name, func(t *testing.T) {
					javaHex, _ := os.ReadFile(hexFile)
					javaBytes, _ := utils.HexToBytes(strings.TrimSpace(string(javaHex)))

					// 依据文件命名约定确定版本
					version := api.V2016
					if strings.Contains(name, "v2025") || strings.Contains(name, "2025") {
						version = api.V2025
					}

					// 加载该数据单元类型对应的编解码器
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

// messageTypeFromPayload 返回金样文件数据单元名称对应的 reflect.Type。
// 该辅助函数用于金样文件只含数据单元字节、
// 不带协议帧外壳的场景。
//
// TODO: 随着金样文件增多,扩展该映射以覆盖所有数据单元类型。
func messageTypeFromPayload(name string) reflect.Type {
	// 将金样文件数据单元名称映射到其反射类型。
	// 示例条目:
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
	_ = name // 占位符:尚未实现
	return nil
}

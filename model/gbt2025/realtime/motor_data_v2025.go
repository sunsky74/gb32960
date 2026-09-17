// Package realtime 定义 GB/T 32960.3-2025 实时数据子结构体。
//
// 注意:本包中的结构体不内嵌 gbt2025.GBT2025Body,以避免
// 循环导入(gbt2025 会导入本包以引用子记录类型,因此本包
// 无法再导入回 gbt2025)。每个结构体直接定义 Version()。
package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// MotorDataV2025 是一条 V2025 驱动电机记录(TLV 0x02)。
// 与 V2016 的 MotorData 相比,移除了 controllerVoltage 和 controllerCurrent。
// MotorState 映射到 Java 的 MotorState:0x01=CONSUMING, 0x02=GENERATING,
// 0x03=OFF, 0x04=READY, 0xFE=EXCEPTION, 0xFF=INVALID。
type MotorDataV2025 struct {
	MotorSeq              int     // 驱动电机序号
	MotorState            byte    // MotorState 枚举
	ControllerTemperature float64 // °C, offset=40
	MotorSpeed            float64 // rpm, offset=32000
	MotorTorque           float64 // N·m, offset=20000, scale=0.1
	MotorTemperature      float64 // °C, offset=40
}

func (m *MotorDataV2025) Version() api.GBTVersion { return api.V2025 }
func (m *MotorDataV2025) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*MotorDataV2025)(nil)).Elem())
}

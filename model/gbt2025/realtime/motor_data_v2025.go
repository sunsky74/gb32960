// Package realtime defines GB/T 32960.3-2025 realtime data sub-structs.
//
// NOTE: Structs in this package do NOT embed gbt2025.GBT2025Body to avoid a
// circular import (gbt2025 imports this package for sub-record types, so this
// package cannot import gbt2025 back). Each struct defines Version() directly.
package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// MotorDataV2025 is a single V2025 drive motor record (TLV 0x02).
// Compared to V2016 MotorData, REMOVES controllerVoltage and controllerCurrent.
// MotorState maps to Java MotorState: 0x01=CONSUMING, 0x02=GENERATING,
// 0x03=OFF, 0x04=READY, 0xFE=EXCEPTION, 0xFF=INVALID.
type MotorDataV2025 struct {
	MotorSeq              int     // 驱动电机序号
	MotorState            byte    // MotorState enum
	ControllerTemperature float64 // °C, offset=40
	MotorSpeed            float64 // rpm, offset=32000
	MotorTorque           float64 // N·m, offset=20000, scale=0.1
	MotorTemperature      float64 // °C, offset=40
}

func (m *MotorDataV2025) Version() api.GBTVersion { return api.V2025 }
func (m *MotorDataV2025) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*MotorDataV2025)(nil)).Elem())
}

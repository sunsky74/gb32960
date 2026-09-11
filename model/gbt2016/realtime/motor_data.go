package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// MotorData is one V2016 驱动电机 entry (TLV type 0x02 element).
// Field names mirror Java MotorData.
// MotorState is kept as byte because types/ has no MotorState enum;
// the codec layer will translate it (audit note: types package is fixed scope
// for this task — adding MotorState/EngineState/HighVoltageDCState is out of scope).
type MotorData struct {
	MotorSeq              int     // 驱动电机序号, 1~253
	MotorState            byte    // 驱动电机状态 (Java enum MotorState; raw byte here)
	ControllerTemperature float64 // 驱动电机控制器温度, °C (scale=1, offset=40)
	MotorSpeed            float64 // 驱动电机转速, r/min (scale=1, offset=20000)
	MotorTorque           float64 // 驱动电机转矩, N·m (scale=10, offset=2000)
	MotorTemperature      float64 // 驱动电机温度, °C (scale=1, offset=40)
	ControllerVoltage     float64 // 电机控制器输入电压, V (scale=10)
	ControllerCurrent     float64 // 电机控制器直流母线电流, A (scale=10, offset=1000)
}

func (m *MotorData) Version() api.GBTVersion { return api.V2016 }

func (m *MotorData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*MotorData)(nil)).Elem())
}

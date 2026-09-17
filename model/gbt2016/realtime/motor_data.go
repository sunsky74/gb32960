package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// MotorData 是 V2016 驱动电机的一条记录(TLV 类型 0x02 元素)。
// 字段名与 Java MotorData 保持一致。
// MotorState 保留为 byte,因为 types/ 中没有 MotorState 枚举;
// 编解码器层将负责转换它(审计说明:types 包在本任务中属于固定范围,
// 新增 MotorState/EngineState/HighVoltageDCState 超出范围)。
type MotorData struct {
	MotorSeq              int     // 驱动电机序号, 1~253
	MotorState            byte    // 驱动电机状态 (Java 枚举 MotorState;此处为原始字节)
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

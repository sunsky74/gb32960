package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// EngineData 是 V2016 发动机数据子记录(TLV 类型 0x04)。
// 字段名与 Java EngineData 保持一致。
// EngineState 保留为 byte,因为 types/ 中没有对应枚举;编解码器
// 层负责转换该原始字节(审计说明:types 包属于固定范围)。
type EngineData struct {
	EngineState         byte    // 发动机状态 (Java 枚举 EngineState;此处为原始字节)
	CrankshaftSpeed     int     // 曲轴转速, rpm
	FuelConsumptionRate float64 // 燃料消耗率, L/100km (scale=100)
}

func (m *EngineData) Version() api.GBTVersion { return api.V2016 }

func (m *EngineData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*EngineData)(nil)).Elem())
}

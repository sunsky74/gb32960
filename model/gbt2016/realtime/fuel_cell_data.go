package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// FuelCellData 是 V2016 燃料电池数据子记录(TLV 类型 0x03)。
// 字段名与 Java FuelCellData 保持一致。
// HighVoltageDCState 保留为 byte,因为 types/ 中没有对应枚举;
// 编解码器层负责转换该原始字节(审计说明:types 包在本任务中
// 属于固定范围)。
type FuelCellData struct {
	FuelCellVoltage                      float64   // 燃料电池电压, V (scale=10)
	FuelCellCurrent                      float64   // 燃料电池电流, A (scale=10)
	FuelConsumptionRate                  float64   // 燃料消耗率, kg/100km (scale=100)
	TotalNumberOfFcTp                    int       // 燃料电池温度探针总数 (探针计数 N)
	ProbeTemperatureValues               []float64 // 探针温度值, °C (scale=1, offset=40), 长度为 N
	HighestTempOfHydrogenSystem          float64   // 氢系统中最高温度, °C (scale=10, offset=40)
	HighestTempProbeCodeOfHydrogenSystem int       // 氢系统中最高温度探针代号
	HighestConOfHydrogen                 int       // 氢气最高浓度, ppm
	HighestHyConSensorCode               int       // 氢气最高浓度传感器代号
	HydrogenMaxPressure                  float64   // 氢气最高压力, MPa (scale=10)
	HydrogenMaxPressureSensorCode        int       // 氢气最高压力传感器代号
	HighVoltageDCState                   byte      // 高压 DC/DC 状态 (Java 枚举 HighVoltageDCState;此处为原始字节)
}

func (m *FuelCellData) Version() api.GBTVersion { return api.V2016 }

func (m *FuelCellData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*FuelCellData)(nil)).Elem())
}

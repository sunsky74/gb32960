package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
	"github.com/sunsky74/gb32960/types"
)

// VehicleData 是 V2016 整车数据子记录(TLV 类型 0x01)。
// 字段名与参考 Java 实现(VehicleData)一致。
type VehicleData struct {
	OperatingState      types.OperatingState // 车辆状态
	ChargingState       types.ChargingState  // 充电状态
	OperationMode       types.OperationMode  // 运行模式
	Speed               float64              // 车速, km/h (scale=10)
	Mileage             float64              // 累计里程, km (scale=10)
	Voltage             float64              // 总电压, V (scale=10)
	Current             float64              // 总电流, A (scale=10, offset=1000)
	SOC                 int                  // 充电 SOC 状态, %
	DC                  types.DCState        // 直流逆变器 DC/DC
	GearPosition        GearPosition         // 挡位 (结构:原始字节 + 派生标志 + gp 枚举)
	Insulance           int                  // 绝缘电阻, kΩ
	AccelerationValue   int                  // 加速踏板行程值, 0~100
	BrakePedalCondition int                  // 制动踏板状态, 0~100
}

func (m *VehicleData) Version() api.GBTVersion { return api.V2016 }

func (m *VehicleData) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*VehicleData)(nil)).Elem())
}

package gbt2016

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// VehicleLogin 是 V2016 车辆登入请求(命令 0x01)。
// 线格式布局:BeanTime(6B) + SerialNum(2B) + ICCID(20B) + Count(1B) + Length(1B) + Count*Length 字节的编码数据。
// V2016 对每个子系统编码使用单一共享的 Length;V2025 使用每个编码各自的长度。
type VehicleLogin struct {
	_         GBT2016Body
	BeanTime  model.BeanTime
	SerialNum int
	ICCID     string   // 20 字节定长字段
	Count     int      // 可充电储能子系统个数
	Length    int      // 每个条目共享的编码长度 (仅 V2016)
	Codes     []string // Count 个条目,每个 Length 字节
}

func (m *VehicleLogin) Version() api.GBTVersion { return api.V2016 }

func (m *VehicleLogin) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*VehicleLogin)(nil)).Elem())
}

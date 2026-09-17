package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// MotorDataList 包装 V2016 驱动电机列表(TLV 类型 0x02 消息体)。
// 与 Java MotorDataList 一致,后者继承 MessageList<MotorData>。
type MotorDataList struct {
	Count int         // 驱动电机个数 (成功解码后 == len(Items))
	Items []MotorData // 驱动电机条目序列
}

func (m *MotorDataList) Version() api.GBTVersion { return api.V2016 }

func (m *MotorDataList) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*MotorDataList)(nil)).Elem())
}

package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// RealTimeV2025Item 是一个容器,用于按原始 TLV 顺序保留实时信息体,
// 支持单帧内出现重复类型。
// 在 Java 中它不继承 GBT2025MessageBody,而是实现 Serializable。
// 这里它带有 Version()/Bytes() 以满足 model.MessageBody,从而可以
// 与其他 V2025 类型一起存储。
type RealTimeV2025Item struct {
	Type byte
	Body model.MessageBody
}

func (m *RealTimeV2025Item) Version() api.GBTVersion { return api.V2025 }
func (m *RealTimeV2025Item) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*RealTimeV2025Item)(nil)).Elem())
}

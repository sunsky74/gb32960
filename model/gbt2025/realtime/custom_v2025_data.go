package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// CustomV2025Data 是 V2025 自定义数据条目(TLV 0x80~0xFE)。
type CustomV2025Data struct {
	CustomKey byte // 自定义数据标识
	Length    int
	Data      []byte
}

func (m *CustomV2025Data) Version() api.GBTVersion { return api.V2025 }
func (m *CustomV2025Data) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*CustomV2025Data)(nil)).Elem())
}

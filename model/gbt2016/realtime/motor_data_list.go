package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// MotorDataList wraps the V2016 驱动电机 list (TLV type 0x02 body).
// Mirrors Java MotorDataList which extends MessageList<MotorData>.
type MotorDataList struct {
	Count int         // 驱动电机个数 (== len(Items) after a successful decode)
	Items []MotorData // sequence of motor entries
}

func (m *MotorDataList) Version() api.GBTVersion { return api.V2016 }

func (m *MotorDataList) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*MotorDataList)(nil)).Elem())
}

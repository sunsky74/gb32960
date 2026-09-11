package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	"github.com/sunsky74/gb32960/modelutil"
)

// RealTimeV2025Item is a holder for preserving realtime information bodies in
// their original TLV order, supporting repeated types within a single frame.
// In Java this does not extend GBT2025MessageBody; it implements Serializable.
// Here it carries Version()/Bytes() to satisfy model.MessageBody so it can be
// stored alongside other V2025 types.
type RealTimeV2025Item struct {
	Type byte
	Body model.MessageBody
}

func (m *RealTimeV2025Item) Version() api.GBTVersion { return api.V2025 }
func (m *RealTimeV2025Item) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*RealTimeV2025Item)(nil)).Elem())
}

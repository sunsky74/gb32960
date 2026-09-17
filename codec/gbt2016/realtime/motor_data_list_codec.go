package realtime

import (
	"fmt"
	"reflect"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
)

// MotorDataListCodec 编解码 V2016 驱动电机列表(TLV 类型 0x02 主体)。
// 线格式:Count(u8) + Count × MotorData。与 Java MotorDataListCodec 一致。
type MotorDataListCodec struct{}

func init() {
	api.Register[mdl.MotorDataList](api.V2016, &MotorDataListCodec{})
}

func (c *MotorDataListCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.MotorDataList{}
	m.Count = int(r.ReadUint8())
	if m.Count > 0 {
		elemCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.MotorData)(nil)).Elem())
		if elemCodec == nil {
			return nil, api.ErrCodecNotFound
		}
		m.Items = make([]mdl.MotorData, m.Count)
		for i := 0; i < m.Count; i++ {
			item, err := elemCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			m.Items[i] = *item.(*mdl.MotorData)
		}
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *MotorDataListCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.MotorDataList)
	// audit 2026-09-17 (M4):在写出任何内容之前拒绝计数/列表不匹配。
	// 此前 Count==0 而 Items 非空时会静默丢弃每个条目,
	// 产生畸形的帧。
	if m.Count != len(m.Items) {
		return fmt.Errorf("gb32960: motor count %d does not match items length %d", m.Count, len(m.Items))
	}
	w.WriteUint8(byte(m.Count))
	if len(m.Items) > 0 {
		elemCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.MotorData)(nil)).Elem())
		if elemCodec == nil {
			return api.ErrCodecNotFound
		}
		for i := range m.Items {
			if err := elemCodec.Encode(w, &m.Items[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

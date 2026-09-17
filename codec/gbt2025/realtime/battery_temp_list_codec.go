package realtime

import (
	"fmt"
	"reflect"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// BatteryTempListCodec 编解码 V2025 动力蓄电池温度列表
// (TLV 类型 0x08 主体)。线格式:
//
//	BatteryPackCount(u8) + BatteryPackCount × BatteryTemp
//
// 与 Java BatteryPackTemperatureListCodec 一致:BYTE1 哨兵计数
// (0xFE/0xFF) 表示其后没有条目,编码按原样写出哨兵字节
// (不再改写为 0xFF);普通计数必须与 Items 长度一致。
type BatteryTempListCodec struct{}

func init() {
	api.Register[mdl.BatteryTempList](api.V2025, &BatteryTempListCodec{})
}

func (c *BatteryTempListCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.BatteryTempList{}
	m.BatteryPackCount = int(r.ReadUint8())
	if !types.ErrByte1.IsInvalid(int64(m.BatteryPackCount)) {
		elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.BatteryTemp)(nil)).Elem())
		if elemCodec == nil {
			return nil, api.ErrCodecNotFound
		}
		m.Items = make([]mdl.BatteryTemp, m.BatteryPackCount)
		for i := 0; i < m.BatteryPackCount; i++ {
			item, err := elemCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			m.Items[i] = *item.(*mdl.BatteryTemp)
		}
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *BatteryTempListCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.BatteryTempList)
	// fix 2026-09-17: 表13 L220 —— 哨兵计数 0xFE(异常)/0xFF(无效)
	// 表示无条目,必须原样写出且列表必须为空;
	// 普通计数必须与 Items 长度一致,否则拒绝编码。
	if types.ErrByte1.IsInvalid(int64(m.BatteryPackCount)) {
		if len(m.Items) != 0 {
			return fmt.Errorf("gb32960: battery pack count %d is a sentinel but %d items present", m.BatteryPackCount, len(m.Items))
		}
		w.WriteUint8(byte(m.BatteryPackCount))
		return nil
	}
	if m.BatteryPackCount != len(m.Items) {
		return fmt.Errorf("gb32960: battery pack count %d does not match items length %d", m.BatteryPackCount, len(m.Items))
	}
	w.WriteUint8(byte(m.BatteryPackCount))
	elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.BatteryTemp)(nil)).Elem())
	if elemCodec == nil {
		return api.ErrCodecNotFound
	}
	for i := range m.Items {
		if err := elemCodec.Encode(w, &m.Items[i]); err != nil {
			return err
		}
	}
	return nil
}

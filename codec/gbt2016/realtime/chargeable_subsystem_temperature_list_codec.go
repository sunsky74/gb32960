package realtime

import (
	"fmt"
	"reflect"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// ChargeableSubsystemTemperatureListCodec 编解码 V2016 可充电储能装置温度
// 列表(TLV 类型 0x09 主体)。线格式:TemperatureCount(u8) + TemperatureCount ×
// ChargeableSubsystemTemperature 条目。
//
// 与 Java 的偏差(已报告):与电压列表相同的计数宽度不一致
// —— Java 读 1 字节,写 2 字节。Go 两侧都保持 u8,
// 因此往返是逐字节稳定的。与 Plan 2 列表模式一致。
type ChargeableSubsystemTemperatureListCodec struct{}

func init() {
	api.Register[mdl.ChargeableSubsystemTemperatureList](api.V2016, &ChargeableSubsystemTemperatureListCodec{})
}

func (c *ChargeableSubsystemTemperatureListCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.ChargeableSubsystemTemperatureList{}
	m.TemperatureCount = int(r.ReadUint8())
	// fix 2026-09-17: GB/T 32960.3-2016 表 B.7(L782)—— 可充电储能子系统个数为
	// 1 字节,有效值 1~250,“0xFE”表示异常 / “0xFF”表示无效。哨兵计数不携带
	// 任何子系统数据单元,按计数读取 N 个条目会下溢(与电压列表同款修复,
	// 计数宽度为 BYTE,故用 types.ErrByte1)。
	if m.TemperatureCount > 0 && !types.ErrByte1.IsInvalid(int64(m.TemperatureCount)) {
		elemCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.ChargeableSubsystemTemperature)(nil)).Elem())
		if elemCodec == nil {
			return nil, api.ErrCodecNotFound
		}
		m.Items = make([]mdl.ChargeableSubsystemTemperature, m.TemperatureCount)
		for i := 0; i < m.TemperatureCount; i++ {
			item, err := elemCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			m.Items[i] = *item.(*mdl.ChargeableSubsystemTemperature)
		}
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *ChargeableSubsystemTemperatureListCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.ChargeableSubsystemTemperatureList)
	// audit 2026-09-17 (M4):在写出任何内容之前拒绝计数/列表不匹配
	//(此前 Count==0 会静默丢弃每个条目 → 畸形帧)。
	// fix 2026-09-17: GB/T 32960.3-2016 表 B.7(L782)—— 哨兵计数(0xFE/0xFF)
	// 不携带任何条目,列表必须为空;计数按原样写出。普通计数仍必须与
	// 列表长度一致。
	if types.ErrByte1.IsInvalid(int64(m.TemperatureCount)) {
		if len(m.Items) != 0 {
			return fmt.Errorf("gb32960: temperature count %d is a sentinel but %d items present", m.TemperatureCount, len(m.Items))
		}
	} else if m.TemperatureCount != len(m.Items) {
		return fmt.Errorf("gb32960: temperature count %d does not match items length %d", m.TemperatureCount, len(m.Items))
	}
	w.WriteUint8(byte(m.TemperatureCount))
	if len(m.Items) > 0 {
		elemCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.ChargeableSubsystemTemperature)(nil)).Elem())
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

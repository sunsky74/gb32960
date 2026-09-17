package realtime

import (
	"fmt"
	"reflect"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// ChargeableSubsystemElectricListCodec 编解码 V2016 可充电储能装置电压
// 列表(TLV 类型 0x08 主体)。线格式:ElectricCount(u8) + ElectricCount ×
// ChargeableSubsystemElectric 条目。
//
// 与 Java 的偏差(已报告):Java 以 readByteAsInt 解码计数
// (1 字节),却以 writeShort 编码(2 字节)—— 这是 Java 已知的一处不一致,
// 破坏了 Java 自身 encode(decode(x)) 的逐字节稳定性。Go 移植版两侧
// 都保持 u8,因此往返是逐字节稳定的。这与 Plan 2 列表
// 模式一致(见 Plan 2 Part C4 MotorDataListCodec 样例),该模式把计数
// 按单个字节读写。
type ChargeableSubsystemElectricListCodec struct{}

func init() {
	api.Register[mdl.ChargeableSubsystemElectricList](api.V2016, &ChargeableSubsystemElectricListCodec{})
}

func (c *ChargeableSubsystemElectricListCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.ChargeableSubsystemElectricList{}
	m.ElectricCount = int(r.ReadUint8())
	// fix 2026-09-17: GB/T 32960.3-2016 表 B.5(L759)—— 可充电储能子系统个数为
	// 1 字节,有效值 1~250,“0xFE”表示异常 / “0xFF”表示无效。哨兵计数不携带
	// 任何子系统数据单元,按计数读取 N 个条目会下溢(镜像 M2 燃料电池模式,
	// 此处计数宽度为 BYTE,故用 types.ErrByte1)。
	if m.ElectricCount > 0 && !types.ErrByte1.IsInvalid(int64(m.ElectricCount)) {
		elemCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.ChargeableSubsystemElectric)(nil)).Elem())
		if elemCodec == nil {
			return nil, api.ErrCodecNotFound
		}
		m.Items = make([]mdl.ChargeableSubsystemElectric, m.ElectricCount)
		for i := 0; i < m.ElectricCount; i++ {
			item, err := elemCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			m.Items[i] = *item.(*mdl.ChargeableSubsystemElectric)
		}
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *ChargeableSubsystemElectricListCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.ChargeableSubsystemElectricList)
	// audit 2026-09-17 (M4):在写出任何内容之前拒绝计数/列表不匹配
	//(此前 Count==0 会静默丢弃每个条目 → 畸形帧)。
	// fix 2026-09-17: GB/T 32960.3-2016 表 B.5(L759)—— 哨兵计数(0xFE/0xFF)
	// 不携带任何条目,列表必须为空;计数按原样写出。普通计数仍必须与
	// 列表长度一致。
	if types.ErrByte1.IsInvalid(int64(m.ElectricCount)) {
		if len(m.Items) != 0 {
			return fmt.Errorf("gb32960: electric count %d is a sentinel but %d items present", m.ElectricCount, len(m.Items))
		}
	} else if m.ElectricCount != len(m.Items) {
		return fmt.Errorf("gb32960: electric count %d does not match items length %d", m.ElectricCount, len(m.Items))
	}
	w.WriteUint8(byte(m.ElectricCount))
	if len(m.Items) > 0 {
		elemCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdl.ChargeableSubsystemElectric)(nil)).Elem())
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

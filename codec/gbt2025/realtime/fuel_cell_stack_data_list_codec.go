package realtime

import (
	"fmt"
	"reflect"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// FuelCellStackDataListCodec 编解码 V2025 燃料电池电堆列表
// (TLV 类型 0x30 主体)。线格式:
//
//	StackCount(u8) + StackCount × FuelCellStackData
//
// 与 Java FuelCellStackDataListCodec 一致:当 StackCount 是 BYTE1
// 哨兵值 (0xFE/0xFF) 时其后没有条目;解码在计数之后停止,
// 编码原样写入计数字节而不写条目。
type FuelCellStackDataListCodec struct{}

func init() {
	api.Register[mdl.FuelCellStackDataList](api.V2025, &FuelCellStackDataListCodec{})
}

func (c *FuelCellStackDataListCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.FuelCellStackDataList{}
	m.StackCount = int(r.ReadUint8())
	if !types.ErrByte1.IsInvalid(int64(m.StackCount)) {
		elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.FuelCellStackData)(nil)).Elem())
		if elemCodec == nil {
			return nil, api.ErrCodecNotFound
		}
		m.Items = make([]mdl.FuelCellStackData, m.StackCount)
		for i := 0; i < m.StackCount; i++ {
			item, err := elemCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			m.Items[i] = *item.(*mdl.FuelCellStackData)
		}
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *FuelCellStackDataListCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.FuelCellStackDataList)
	// fix 2026-09-17: GB/T 32960.3-2025 表18(L281) —— 电堆个数 BYTE1 有效值
	// 1~253,0xFE 异常 / 0xFF 无效;哨兵计数后不携带条目,普通计数必须与列表
	// 长度一致。旧代码在哨兵计数后仍写出全部条目(不可解析)、普通计数不匹配
	// 时静默截断/超写。
	if types.ErrByte1.IsInvalid(int64(m.StackCount)) {
		if len(m.Items) != 0 {
			return fmt.Errorf("gb32960: fuel cell stack count %d is a sentinel but %d items present", m.StackCount, len(m.Items))
		}
	} else if m.StackCount != len(m.Items) {
		return fmt.Errorf("gb32960: fuel cell stack count %d does not match items length %d", m.StackCount, len(m.Items))
	}
	w.WriteUint8(byte(m.StackCount))
	if len(m.Items) > 0 {
		elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.FuelCellStackData)(nil)).Elem())
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

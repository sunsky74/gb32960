package realtime

import (
	"fmt"
	"reflect"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
)

// MotorDataV2025ListCodec 编解码 V2025 驱动电机列表
// (TLV 类型 0x02 主体)。线格式:MotorCount(u8) + MotorCount × MotorDataV2025。
// 与 Java MotorDataV2025ListCodec 一致。
type MotorDataV2025ListCodec struct{}

func init() {
	api.Register[mdl.MotorDataV2025List](api.V2025, &MotorDataV2025ListCodec{})
}

func (c *MotorDataV2025ListCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.MotorDataV2025List{}
	m.MotorCount = int(r.ReadUint8())
	// fix 2026-09-17: GB/T 32960.3-2025 表15(L238) —— 驱动电机个数 BYTE1 有效值
	// 1~253,0xFE 表示异常 / 0xFF 表示无效;哨兵计数后不携带条目,按计数读取
	// 254/255 个条目会下溢(旧代码缺陷)。
	if m.MotorCount > 0 && !types.ErrByte1.IsInvalid(int64(m.MotorCount)) {
		elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.MotorDataV2025)(nil)).Elem())
		if elemCodec == nil {
			return nil, api.ErrCodecNotFound
		}
		m.Items = make([]mdl.MotorDataV2025, m.MotorCount)
		for i := 0; i < m.MotorCount; i++ {
			item, err := elemCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			m.Items[i] = *item.(*mdl.MotorDataV2025)
		}
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *MotorDataV2025ListCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.MotorDataV2025List)
	// fix 2026-09-17: 表15(L238) —— 哨兵计数(0xFE/0xFF)不携带条目,普通计数必须
	// 与列表长度一致;校验先于写出,避免产生畸形帧(镜像 2016 MotorDataListCodec
	// 的 M4 校验)。
	if types.ErrByte1.IsInvalid(int64(m.MotorCount)) {
		if len(m.Items) != 0 {
			return fmt.Errorf("gb32960: motor count %d is a sentinel but %d items present", m.MotorCount, len(m.Items))
		}
	} else if m.MotorCount != len(m.Items) {
		return fmt.Errorf("gb32960: motor count %d does not match items length %d", m.MotorCount, len(m.Items))
	}
	w.WriteUint8(byte(m.MotorCount))
	if len(m.Items) > 0 {
		elemCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdl.MotorDataV2025)(nil)).Elem())
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

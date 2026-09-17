package realtime

import (
	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// EngineV2025Codec 编解码 V2025 发动机数据子记录
// (TLV 类型 0x04)。与 V2016 EngineData 相比,V2025 只有 CrankshaftSpeed
// (移除了 EngineState 与 FuelConsumptionRate)。字段顺序与 Java
// EngineDataV2025Codec 完全一致。
type EngineV2025Codec struct{}

func init() {
	api.Register[mdl.EngineV2025Data](api.V2025, &EngineV2025Codec{})
}

func (c *EngineV2025Codec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.EngineV2025Data{}
	m.CrankshaftSpeed = int(r.ReadUint16())

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *EngineV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.EngineV2025Data)
	w.WriteUint16(uint16(m.CrankshaftSpeed))
	return nil
}

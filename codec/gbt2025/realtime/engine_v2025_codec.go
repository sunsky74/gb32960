package realtime

import (
	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// EngineV2025Codec encodes/decodes the V2025 发动机数据 sub-record
// (TLV type 0x04). Compared to V2016 EngineData, V2025 has ONLY CrankshaftSpeed
// (removes EngineState and FuelConsumptionRate). Field order mirrors Java
// EngineDataV2025Codec exactly.
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

package realtime

import (
	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
)

// EngineDataCodec 编解码 V2016 发动机数据子记录(TLV 类型 0x04)。
// 字段顺序与转换器与 Java EngineDataCodec 一致:
//
//	EngineState(u8 原始字节)
//	CrankshaftSpeed(u16)
//	FuelConsumptionRate(u16 FuelConsumptionRateConverter)
type EngineDataCodec struct{}

func init() {
	api.Register[mdl.EngineData](api.V2016, &EngineDataCodec{})
}

func (c *EngineDataCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.EngineData{}
	m.EngineState = r.ReadUint8()
	m.CrankshaftSpeed = int(r.ReadUint16())
	m.FuelConsumptionRate = codec.FuelConsumptionRateConverter.Decode(int64(r.ReadUint16()))

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *EngineDataCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.EngineData)
	w.WriteUint8(m.EngineState)
	w.WriteUint16(uint16(m.CrankshaftSpeed))
	w.WriteUint16(uint16(codec.FuelConsumptionRateConverter.Encode(m.FuelConsumptionRate)))
	return nil
}

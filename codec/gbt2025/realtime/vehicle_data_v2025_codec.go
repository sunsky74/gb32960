// Package realtime 包含 GB/T 32960.3-2025 实时子记录编解码器。
//
// 每个编解码器在 init() 中自注册到 api.V2025。消费方必须
// 空白导入本包(或未来的 codec/all)才能使这些注册
// 对 api.GetCodec 以及 V2025 TLV 分发表可见,
// 分发表位于 codec/gbt2025/realtime_data_v2025_codec.go。
package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdlrt16 "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// VehicleDataV2025Codec 编解码 V2025 整车数据子记录
// (TLV 类型 0x01)。复用 V2016 VehicleData 结构体,但:
//
//   - 使用 CurrentConverter2025(offset=3000)而非 CurrentConverter2016(offset=1000)。
//   - 省略了 AccelerationValue 与 BrakePedalCondition(Java 中已注释掉)。
//   - GearPosition 通过 V2025 已注册的 GearPositionV2025Codec 解码
//     (1 字节,IsV2025=true;Java 同样调用
//     CodecMap.getV2025Codec(GearPosition.class),见 Java
//     VehicleDataV2025Codec.java L51-52)。
//
// fix 2026-09-17:此处原先查找 V2016 挡位编解码器,导致 IsV2025 恒为 false、
// 2025 帧的 bit7 有效标识语义失效;已改为查找 V2025 编解码器。
//
// 字段顺序与转换器与 Java VehicleDataV2025Codec 完全一致。
type VehicleDataV2025Codec struct{}

func init() {
	api.Register[mdlrt16.VehicleData](api.V2025, &VehicleDataV2025Codec{})
}

// Decode 与 Java VehicleDataV2025Codec.decodeBuffer 一致。
func (c *VehicleDataV2025Codec) Decode(r api.Reader) (api.Message, error) {
	m := &mdlrt16.VehicleData{}
	m.OperatingState = types.OperatingState(r.ReadUint8())
	m.ChargingState = types.ChargingState(r.ReadUint8())
	m.OperationMode = types.OperationMode(r.ReadUint8())
	m.Speed = codec.SpeedConverter.Decode(int64(r.ReadUint16()))
	m.Mileage = codec.MileageConverter.Decode(int64(r.ReadUint32()))
	m.Voltage = codec.VoltageConverter.Decode(int64(r.ReadUint16()))
	m.Current = codec.CurrentConverter2025.Decode(int64(r.ReadUint16()))
	m.SOC = int(r.ReadUint8())
	m.DC = types.DCState(r.ReadUint8())

	gpCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdlrt16.GearPosition)(nil)).Elem())
	if gpCodec == nil {
		return nil, api.ErrCodecNotFound
	}
	gp, err := gpCodec.Decode(r)
	if err != nil {
		return nil, err
	}
	m.GearPosition = *gp.(*mdlrt16.GearPosition)

	m.Insulance = int(r.ReadUint16())

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode 与 Java VehicleDataV2025Codec.encodeBuffer 一致。注意 V2025 通过
// GearPositionV2025Codec 为 GearPosition 写 1 字节(仅原点),并跳过
// AccelerationValue/BrakePedalCondition。
func (c *VehicleDataV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdlrt16.VehicleData)
	w.WriteUint8(byte(m.OperatingState))
	w.WriteUint8(byte(m.ChargingState))
	w.WriteUint8(byte(m.OperationMode))
	w.WriteUint16(uint16(codec.SpeedConverter.Encode(m.Speed)))
	w.WriteUint32(uint32(codec.MileageConverter.Encode(m.Mileage)))
	w.WriteUint16(uint16(codec.VoltageConverter.Encode(m.Voltage)))
	w.WriteUint16(uint16(codec.CurrentConverter2025.Encode(m.Current)))
	w.WriteUint8(byte(m.SOC))
	w.WriteUint8(byte(m.DC))

	gpCodec := api.GetCodec(api.V2025, reflect.TypeOf((*mdlrt16.GearPosition)(nil)).Elem())
	if gpCodec == nil {
		return api.ErrCodecNotFound
	}
	if err := gpCodec.Encode(w, &m.GearPosition); err != nil {
		return err
	}

	w.WriteUint16(uint16(m.Insulance))
	return nil
}

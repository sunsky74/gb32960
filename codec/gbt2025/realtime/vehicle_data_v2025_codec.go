// Package realtime contains GB/T 32960.3-2025 realtime sub-record codecs.
//
// Each codec self-registers under api.V2025 in init(). Consumers must
// blank-import this package (or codec/all once it exists) to make the
// registrations visible to api.GetCodec and to the V2025 TLV dispatch map
// in codec/gbt2025/realtime_data_v2025_codec.go.
package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/codec"
	mdlrt16 "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// VehicleDataV2025Codec encodes/decodes the V2025 整车数据 sub-record
// (TLV type 0x01). Reuses the V2016 VehicleData struct, but:
//
//   - Uses CurrentConverter2025 (offset=3000) instead of CurrentConverter2016 (offset=1000).
//   - OMITS AccelerationValue and BrakePedalCondition (Java commented out).
//   - GearPosition is decoded via the V2016 registered codec (1 byte; Java
//     does CodecMap.getV2016Codec(GearPosition.class) here too).
//
// Field order and converters mirror Java VehicleDataV2025Codec exactly.
type VehicleDataV2025Codec struct{}

func init() {
	api.Register[mdlrt16.VehicleData](api.V2025, &VehicleDataV2025Codec{})
}

// Decode mirrors Java VehicleDataV2025Codec.decodeBuffer.
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

	gpCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdlrt16.GearPosition)(nil)).Elem())
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

// Encode mirrors Java VehicleDataV2025Codec.encodeBuffer. Note V2025 writes
// 1 byte for GearPosition (origin only) via the V2016 codec, and skips
// AccelerationValue/BrakePedalCondition.
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

	gpCodec := api.GetCodec(api.V2016, reflect.TypeOf((*mdlrt16.GearPosition)(nil)).Elem())
	if gpCodec == nil {
		return api.ErrCodecNotFound
	}
	if err := gpCodec.Encode(w, &m.GearPosition); err != nil {
		return err
	}

	w.WriteUint16(uint16(m.Insulance))
	return nil
}

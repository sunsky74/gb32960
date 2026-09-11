package gbt2025

import (
	"fmt"
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdlrt16 "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	mdl "github.com/sunsky74/gb32960/model/gbt2025"
	v2025rt "github.com/sunsky74/gb32960/model/gbt2025/realtime"
	"github.com/sunsky74/gb32960/types"
	"github.com/sunsky74/gb32960/utils"
)

// realtimeCodecV2025For returns the V2025 codec registered for the given TLV
// type, looking it up fresh on each call.
//
// IMPORTANT (audit 2026-07-31): an earlier draft built a single dispatch
// map at init() time, mirroring the V2016 codec/gbt2016/realtime_data_codec.go
// pattern. That is fragile: Go does not guarantee init() order between
// non-dependent packages, so if codec/gbt2025 inits before
// codec/gbt2025/realtime (which is legal — neither imports the other),
// the map ends up full of nil entries and Decode fails with
// api.ErrUnknownTLVType on the first TLV byte. Looking up on each call
// sidesteps the init-order hazard entirely and still uses api.GetCodec
// (so the file stays decoupled from codec/gbt2025/realtime — the consumer's
// blank import is what makes registrations visible by the time Decode runs).
func realtimeCodecV2025For(rt types.RealTimeV2025Type) api.Codecer {
	switch rt {
	case types.RealTimeV2025Vehicle:
		return codecV2025For(mdlrt16.VehicleData{})
	case types.RealTimeV2025Motor:
		return codecV2025For(v2025rt.MotorDataV2025List{})
	case types.RealTimeV2025FuelCellEngine:
		return codecV2025For(v2025rt.FuelCellEngineV2025Data{})
	case types.RealTimeV2025Engine:
		return codecV2025For(v2025rt.EngineV2025Data{})
	case types.RealTimeV2025Location:
		return codecV2025For(v2025rt.LocationV2025Data{})
	case types.RealTimeV2025Alarm:
		return codecV2025For(v2025rt.AlarmV2025Data{})
	case types.RealTimeV2025MinParallelVoltage:
		return codecV2025For(v2025rt.MinParallelCellVoltageList{})
	case types.RealTimeV2025BatteryTemp:
		return codecV2025For(v2025rt.BatteryTempList{})
	case types.RealTimeV2025FuelCellStack:
		return codecV2025For(v2025rt.FuelCellStackDataList{})
	case types.RealTimeV2025SuperCapacitor:
		return codecV2025For(v2025rt.SuperCapacitorData{})
	case types.RealTimeV2025SuperCapExtremum:
		return codecV2025For(v2025rt.SuperCapacitorExtremumData{})
	case types.RealTimeV2025Signature:
		return codecV2025For(v2025rt.VehicleSignature{})
	case types.RealTimeV2025Custom:
		return codecV2025For(v2025rt.CustomV2025Data{})
	}
	return nil
}

// RealTimeDataV2025Codec encodes/decodes the V2025 realtime data report
// (command 0x02). Mirrors Java RealTimeDataV2025Codec: BeanTime via registered
// codec, then a TLV loop reading (flag, sub-record) pairs until the buffer
// is exhausted. Unknown flags surface as api.ErrUnknownTLVType (NOT a silent
// break — Java swallows them, but Plan 2 mandates the error).
type RealTimeDataV2025Codec struct{}

func init() {
	api.Register[mdl.RealTimeV2025Data](api.V2025, &RealTimeDataV2025Codec{})
}

// codecV2025For returns the V2025 codec registered for the type of v, or nil.
// Mirrors the V2016 codecFor indirection in codec/gbt2016/realtime_data_codec.go.
func codecV2025For(v any) api.Codecer {
	t := reflect.TypeOf(v)
	return api.GetCodec(api.V2025, t)
}

// Decode mirrors Java RealTimeDataV2025Codec.decodeBuffer / decodePayload /
// decodeByType: BeanTime via registered codec, then a TLV loop reading
// (flag, sub-record) pairs until the buffer is exhausted.
//
// The custom range 0x80~0xFE is dispatched via types.RealTimeV2025TypeByCode
// and decoded by the registered CustomV2025Data codec; the original flag byte
// is preserved on the resulting CustomV2025Data.CustomKey (matches Java
// decodeByType CUSTOM_DATA_FLAG branch).
func (c *RealTimeDataV2025Codec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.RealTimeV2025Data{}

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return nil, api.ErrCodecNotFound
	}
	bt, err := btCodec.Decode(r)
	if err != nil {
		return nil, err
	}
	m.BeanTime = *bt.(*model.BeanTime)

	for r.Remaining() > 0 {
		flag := r.ReadUint8()
		// Snapshot right after the flag read: for the signature TLV, SignData
		// covers everything before this flag byte (mirrors Java's
		// dumpBytes(start, readerIndex()-1-start) taken at the same point).
		consumedAtFlag := r.Consumed()
		rt, ok := types.RealTimeV2025TypeByCode(flag)
		if !ok {
			return nil, fmt.Errorf("%w: 0x%02X", api.ErrUnknownTLVType, flag)
		}

		// Custom range dispatches to CustomV2025Data codec; preserves raw flag.
		if rt == types.RealTimeV2025Custom {
			customCodec := realtimeCodecV2025For(types.RealTimeV2025Custom)
			if customCodec == nil {
				return nil, api.ErrCodecNotFound
			}
			item, err := customCodec.Decode(r)
			if err != nil {
				return nil, err
			}
			if err := r.Err(); err != nil {
				return nil, api.ErrBufferUnderflow
			}
			custom := item.(*v2025rt.CustomV2025Data)
			custom.CustomKey = flag
			m.CustomData = append(m.CustomData, *custom)
			m.Items = append(m.Items, v2025rt.RealTimeV2025Item{Type: flag, Body: custom})
			continue
		}

		codec := realtimeCodecV2025For(rt)
		if codec == nil {
			return nil, fmt.Errorf("%w: 0x%02X", api.ErrUnknownTLVType, flag)
		}

		item, err := codec.Decode(r)
		if err != nil {
			return nil, err
		}
		if err := r.Err(); err != nil {
			return nil, api.ErrBufferUnderflow
		}

		var body model.MessageBody
		switch rt {
		case types.RealTimeV2025Vehicle:
			m.VehicleData = item.(*mdlrt16.VehicleData)
			body = m.VehicleData
		case types.RealTimeV2025Motor:
			m.MotorDataList = item.(*v2025rt.MotorDataV2025List)
			body = m.MotorDataList
		case types.RealTimeV2025FuelCellEngine:
			m.FuelCellData = item.(*v2025rt.FuelCellEngineV2025Data)
			body = m.FuelCellData
		case types.RealTimeV2025Engine:
			m.EngineData = item.(*v2025rt.EngineV2025Data)
			body = m.EngineData
		case types.RealTimeV2025Location:
			m.LocationData = item.(*v2025rt.LocationV2025Data)
			body = m.LocationData
		case types.RealTimeV2025Alarm:
			m.AlarmData = item.(*v2025rt.AlarmV2025Data)
			body = m.AlarmData
		case types.RealTimeV2025MinParallelVoltage:
			m.MinParallelCellVoltages = item.(*v2025rt.MinParallelCellVoltageList)
			body = m.MinParallelCellVoltages
		case types.RealTimeV2025BatteryTemp:
			m.BatteryPackTemperatures = item.(*v2025rt.BatteryTempList)
			body = m.BatteryPackTemperatures
		case types.RealTimeV2025FuelCellStack:
			m.FuelCellStackDataList = item.(*v2025rt.FuelCellStackDataList)
			body = m.FuelCellStackDataList
		case types.RealTimeV2025SuperCapacitor:
			m.SuperCapacitorData = item.(*v2025rt.SuperCapacitorData)
			body = m.SuperCapacitorData
		case types.RealTimeV2025SuperCapExtremum:
			m.SuperCapacitorExtremumData = item.(*v2025rt.SuperCapacitorExtremumData)
			body = m.SuperCapacitorExtremumData
		case types.RealTimeV2025Signature:
			m.VehicleSignature = item.(*v2025rt.VehicleSignature)
			m.VehicleSignature.SignData = append([]byte(nil), consumedAtFlag[:len(consumedAtFlag)-1]...)
			body = m.VehicleSignature
		}
		m.Items = append(m.Items, v2025rt.RealTimeV2025Item{Type: flag, Body: body})
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode mirrors Java RealTimeDataV2025Codec.encodeBuffer / encodePayload:
// BeanTime via registered codec, then either the items list (when non-empty)
// in its recorded order, or the typed fields in the canonical Java order
// (Vehicle, Motor, FuelCellEngine, Engine, Location, Alarm, MinParallelVoltage,
// BatteryTemp, FuelCellStack, SuperCapacitor, SuperCapExtremum, Custom...,
// Signature).
//
// Items path: unlike Java — whose decode never adds custom data or the
// signature to items and therefore DROPS both when re-encoding a decoded
// frame — Go's Decode records every TLV in Items, so the items path re-emits
// the original wire byte-for-byte: order, repeats, custom data, and signature
// included. A VehicleSignature body gets its SignData refreshed from the
// bytes written so far (Java does the same via buffer.dumpBytes); a
// CustomV2025Data body writes its own CustomKey (no outer flag byte).
//
// api.Writer cannot report already-written bytes (needed for the SignData
// refresh), so everything is encoded into a local *utils.ByteWriter that can,
// then flushed once — the same pattern as protocolMessageCodec.Encode.
func (c *RealTimeDataV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.RealTimeV2025Data)

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return api.ErrCodecNotFound
	}

	local := utils.NewByteWriter()
	if err := btCodec.Encode(local, &m.BeanTime); err != nil {
		return err
	}

	if len(m.Items) > 0 {
		for i := range m.Items {
			item := &m.Items[i]
			if sig, ok := item.Body.(*v2025rt.VehicleSignature); ok {
				// Copy: ByteWriter.Bytes() aliases the live buffer
				sig.SignData = append([]byte(nil), local.Bytes()...)
			}
			if _, isCustom := item.Body.(*v2025rt.CustomV2025Data); !isCustom {
				local.WriteUint8(item.Type)
			}
			bodyType := reflect.TypeOf(item.Body)
			if bodyType.Kind() == reflect.Ptr {
				bodyType = bodyType.Elem()
			}
			codec := api.GetCodec(api.V2025, bodyType)
			if codec == nil {
				return fmt.Errorf("%w: 0x%02X", api.ErrUnknownTLVType, item.Type)
			}
			if err := codec.Encode(local, item.Body); err != nil {
				return err
			}
		}
		w.WriteBytes(local.Bytes())
		return nil
	}

	encodeTLV := func(flag types.RealTimeV2025Type, subMsg api.Message) error {
		codec := realtimeCodecV2025For(flag)
		if codec == nil {
			return fmt.Errorf("%w: 0x%02X", api.ErrUnknownTLVType, byte(flag))
		}
		local.WriteUint8(byte(flag))
		return codec.Encode(local, subMsg)
	}

	if m.VehicleData != nil {
		if err := encodeTLV(types.RealTimeV2025Vehicle, m.VehicleData); err != nil {
			return err
		}
	}
	if m.MotorDataList != nil {
		if err := encodeTLV(types.RealTimeV2025Motor, m.MotorDataList); err != nil {
			return err
		}
	}
	if m.FuelCellData != nil {
		if err := encodeTLV(types.RealTimeV2025FuelCellEngine, m.FuelCellData); err != nil {
			return err
		}
	}
	if m.EngineData != nil {
		if err := encodeTLV(types.RealTimeV2025Engine, m.EngineData); err != nil {
			return err
		}
	}
	if m.LocationData != nil {
		if err := encodeTLV(types.RealTimeV2025Location, m.LocationData); err != nil {
			return err
		}
	}
	if m.AlarmData != nil {
		if err := encodeTLV(types.RealTimeV2025Alarm, m.AlarmData); err != nil {
			return err
		}
	}
	if m.MinParallelCellVoltages != nil {
		if err := encodeTLV(types.RealTimeV2025MinParallelVoltage, m.MinParallelCellVoltages); err != nil {
			return err
		}
	}
	if m.BatteryPackTemperatures != nil {
		if err := encodeTLV(types.RealTimeV2025BatteryTemp, m.BatteryPackTemperatures); err != nil {
			return err
		}
	}
	if m.FuelCellStackDataList != nil {
		if err := encodeTLV(types.RealTimeV2025FuelCellStack, m.FuelCellStackDataList); err != nil {
			return err
		}
	}
	if m.SuperCapacitorData != nil {
		if err := encodeTLV(types.RealTimeV2025SuperCapacitor, m.SuperCapacitorData); err != nil {
			return err
		}
	}
	if m.SuperCapacitorExtremumData != nil {
		if err := encodeTLV(types.RealTimeV2025SuperCapExtremum, m.SuperCapacitorExtremumData); err != nil {
			return err
		}
	}

	if len(m.CustomData) > 0 {
		customCodec := realtimeCodecV2025For(types.RealTimeV2025Custom)
		if customCodec == nil {
			return api.ErrCodecNotFound
		}
		for i := range m.CustomData {
			if err := customCodec.Encode(local, &m.CustomData[i]); err != nil {
				return err
			}
		}
	}

	if m.VehicleSignature != nil {
		m.VehicleSignature.SignData = append([]byte(nil), local.Bytes()...)
		if err := encodeTLV(types.RealTimeV2025Signature, m.VehicleSignature); err != nil {
			return err
		}
	}

	w.WriteBytes(local.Bytes())
	return nil
}

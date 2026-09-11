package gbt2016

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/model"
	mdl "github.com/sunsky74/gb32960/model/gbt2016"
	mdlrt "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// realtimeCodecV2016For dispatches a TLV flag byte to the codec that decodes
// the corresponding realtime sub-record. The lookup runs fresh on each call.
//
// INIT-ORDER HAZARD FIX (audit 2026-07-31, mirrors V2025 sibling fix in
// codec/gbt2025/realtime_data_v2025_codec.go): an earlier draft built a
// single dispatch map at init() time. That is fragile: Go does not guarantee
// init() order between non-dependent packages, so if codec/gbt2016 inits
// before codec/gbt2016/realtime (which is legal — neither imports the other),
// the map ends up full of nil entries and Decode fails with
// api.ErrUnknownTLVType on the first TLV byte. Looking up on each call
// sidesteps the init-order hazard entirely and still uses api.GetCodec
// (so the file stays decoupled from codec/gbt2016/realtime — the consumer's
// blank import is what makes registrations visible by the time Decode runs).
//
// DEVIATION FROM PLAN 2 SAMPLE (reported): Plan 2 Part C3 references the
// realtime CODEC types via the alias `realtime` (e.g. `&realtime.VehicleDataCodec{}`),
// but that alias is also used for the realtime MODEL package — they would
// collide. Plan 2 also wrote the dispatched fields as `m.VoltageList` /
// `m.TempList`; the actual RealTimeData model exposes them as
// `ChargeableSubsystemElectricList` and `ChargeableSubsystemTemperatureList`
// (see model/gbt2016/realtime_data.go). We use the actual field names.
func realtimeCodecV2016For(rt types.RealTimeType) api.Codecer {
	switch rt {
	case types.RealTimeVehicle:
		return codecFor(mdlrt.VehicleData{})
	case types.RealTimeMotor:
		return codecFor(mdlrt.MotorDataList{})
	case types.RealTimeFuelCell:
		return codecFor(mdlrt.FuelCellData{})
	case types.RealTimeEngine:
		return codecFor(mdlrt.EngineData{})
	case types.RealTimeLocation:
		return codecFor(mdlrt.LocationData{})
	case types.RealTimeExtremum:
		return codecFor(mdlrt.ExtremumData{})
	case types.RealTimeAlarm:
		return codecFor(mdlrt.AlarmData{})
	case types.RealTimeVoltage:
		return codecFor(mdlrt.ChargeableSubsystemElectricList{})
	case types.RealTimeTemperature:
		return codecFor(mdlrt.ChargeableSubsystemTemperatureList{})
	}
	return nil
}

type RealTimeDataCodec struct{}

func init() {
	api.Register[mdl.RealTimeData](api.V2016, &RealTimeDataCodec{})
}

// codecFor returns the V2016 codec registered for the type of v, or nil.
// Using api.GetCodec keeps this file decoupled from codec/gbt2016/realtime
// (no direct import) — the consumer's blank import of that package is what
// makes the registrations visible. If the consumer forgot the blank import,
// the lookup returns nil and decode surfaces api.ErrCodecNotFound.
func codecFor(v any) api.Codecer {
	t := reflect.TypeOf(v)
	return api.GetCodec(api.V2016, t)
}

// Decode mirrors Java RealTimeDataCodec.decodeBuffer: BeanTime via registered
// codec, then a TLV loop reading (flag, sub-record) pairs until the buffer
// is exhausted. Unknown flags surface as api.ErrUnknownTLVType (NOT a silent
// break — Java swallows them, but Plan 2 mandates the error).
func (c *RealTimeDataCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.RealTimeData{}

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

		// 用户自定义数据(国标 2016 表 8: 0x80~0xFE): WORD 长度 + BYTE[N] 数据,
		// 按 flag → 数据体 原样存入 map(对应 Java customDataMap)
		if flag >= 0x80 && flag <= 0xFE {
			length := int(r.ReadUint16())
			data := r.ReadBytes(length)
			if err := r.Err(); err != nil {
				return nil, api.ErrBufferUnderflow
			}
			if m.CustomData == nil {
				m.CustomData = make(map[byte][]byte)
			}
			m.CustomData[flag] = data
			continue
		}

		rt := types.RealTimeType(flag)
		codec := realtimeCodecV2016For(rt)
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

		switch rt {
		case types.RealTimeVehicle:
			m.VehicleData = item.(*mdlrt.VehicleData)
		case types.RealTimeMotor:
			m.MotorDataList = item.(*mdlrt.MotorDataList)
		case types.RealTimeFuelCell:
			m.FuelCellData = item.(*mdlrt.FuelCellData)
		case types.RealTimeEngine:
			m.EngineData = item.(*mdlrt.EngineData)
		case types.RealTimeLocation:
			m.LocationData = item.(*mdlrt.LocationData)
		case types.RealTimeExtremum:
			m.ExtremumData = item.(*mdlrt.ExtremumData)
		case types.RealTimeAlarm:
			m.AlarmData = item.(*mdlrt.AlarmData)
		case types.RealTimeVoltage:
			m.ChargeableSubsystemElectricList = item.(*mdlrt.ChargeableSubsystemElectricList)
		case types.RealTimeTemperature:
			m.ChargeableSubsystemTemperatureList = item.(*mdlrt.ChargeableSubsystemTemperatureList)
		}
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode mirrors Java RealTimeDataCodec.encodeBuffer: BeanTime via registered
// codec, then each non-nil sub-record prefixed with its TLV flag byte.
// Sub-record emit order matches Java (Vehicle, Motor, FuelCell, Engine,
// Location, Extremum, Alarm, Voltage, Temperature).
func (c *RealTimeDataCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.RealTimeData)

	btCodec := api.GetCodec(api.V2016, reflect.TypeOf((*model.BeanTime)(nil)).Elem())
	if btCodec == nil {
		return api.ErrCodecNotFound
	}
	if err := btCodec.Encode(w, &m.BeanTime); err != nil {
		return err
	}

	encodeTLV := func(flag types.RealTimeType, subMsg api.Message) error {
		codec := realtimeCodecV2016For(flag)
		if codec == nil {
			return fmt.Errorf("%w: 0x%02X", api.ErrUnknownTLVType, byte(flag))
		}
		w.WriteUint8(byte(flag))
		return codec.Encode(w, subMsg)
	}

	if m.VehicleData != nil {
		if err := encodeTLV(types.RealTimeVehicle, m.VehicleData); err != nil {
			return err
		}
	}
	if m.MotorDataList != nil {
		if err := encodeTLV(types.RealTimeMotor, m.MotorDataList); err != nil {
			return err
		}
	}
	if m.FuelCellData != nil {
		if err := encodeTLV(types.RealTimeFuelCell, m.FuelCellData); err != nil {
			return err
		}
	}
	if m.EngineData != nil {
		if err := encodeTLV(types.RealTimeEngine, m.EngineData); err != nil {
			return err
		}
	}
	if m.LocationData != nil {
		if err := encodeTLV(types.RealTimeLocation, m.LocationData); err != nil {
			return err
		}
	}
	if m.ExtremumData != nil {
		if err := encodeTLV(types.RealTimeExtremum, m.ExtremumData); err != nil {
			return err
		}
	}
	if m.AlarmData != nil {
		if err := encodeTLV(types.RealTimeAlarm, m.AlarmData); err != nil {
			return err
		}
	}
	if m.ChargeableSubsystemElectricList != nil {
		if err := encodeTLV(types.RealTimeVoltage, m.ChargeableSubsystemElectricList); err != nil {
			return err
		}
	}
	if m.ChargeableSubsystemTemperatureList != nil {
		if err := encodeTLV(types.RealTimeTemperature, m.ChargeableSubsystemTemperatureList); err != nil {
			return err
		}
	}

	// 自定义数据(0x80~0xFE): key 即信息类型标志,按 key 升序输出保证编码确定性
	if len(m.CustomData) > 0 {
		keys := make([]int, 0, len(m.CustomData))
		for k := range m.CustomData {
			keys = append(keys, int(k))
		}
		sort.Ints(keys)
		for _, k := range keys {
			data := m.CustomData[byte(k)]
			w.WriteUint8(byte(k))
			w.WriteUint16(uint16(len(data)))
			w.WriteBytes(data)
		}
	}
	return nil
}

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

// realtimeCodecV2016For 把 TLV 标志字节分发到解码对应实时子记录的
// 编解码器。每次调用时重新查找。
//
// 初始化顺序风险修复(audit 2026-07-31,镜像 codec/gbt2025/realtime_data_v2025_codec.go
// 中的 V2025 同族修复):早期草稿在 init() 时构建单个分发 map。
// 这很脆弱:Go 不保证互不依赖的包之间的 init() 顺序,所以如果
// codec/gbt2016 先于 codec/gbt2016/realtime 初始化(这是合法的,两者互不导入),
// map 最终会充满 nil 条目,Decode 在第一个 TLV 字节上就失败并返回
// api.ErrUnknownTLVType。每次调用时查找可完全绕开初始化顺序风险,
// 并且仍然使用 api.GetCodec
// (这样本文件与 codec/gbt2016/realtime 保持解耦,消费方的
// 空白导入才是在 Decode 运行时让注册可见的关键)。
//
// 与 Plan 2 示例的偏差(已上报):Plan 2 Part C3 通过别名 `realtime`
// 引用实时 CODEC 类型(例如 `&realtime.VehicleDataCodec{}`),
// 但该别名也被实时 MODEL 包使用,二者会冲突。Plan 2 还把被分发的字段
// 写成 `m.VoltageList` / `m.TempList`;实际的 RealTimeData 模型把它们
// 暴露为 `ChargeableSubsystemElectricList` 和 `ChargeableSubsystemTemperatureList`
// (见 model/gbt2016/realtime_data.go)。我们使用实际的字段名。
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

// codecFor 返回为 v 的类型注册的 V2016 编解码器,不存在则为 nil。
// 使用 api.GetCodec 使本文件与 codec/gbt2016/realtime 解耦
// (不直接导入),消费方对该包的空白导入
// 才是让注册可见的关键。如果消费方忘记空白导入,
// 查找返回 nil,解码会上报 api.ErrCodecNotFound。
func codecFor(v any) api.Codecer {
	t := reflect.TypeOf(v)
	return api.GetCodec(api.V2016, t)
}

// Decode 镜像 Java RealTimeDataCodec.decodeBuffer:先经由已注册的
// 编解码器解码 BeanTime,然后进入 TLV 循环,持续读取(flag, 子记录)对,
// 直到缓冲区耗尽。未知 flag 上报为 api.ErrUnknownTLVType(不是静默
// 中断,Java 会吞掉它们,但 Plan 2 强制要求返回该错误)。
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

// Encode 镜像 Java RealTimeDataCodec.encodeBuffer:先经由已注册的
// 编解码器编码 BeanTime,然后输出每个非 nil 的子记录,各自前置其
// TLV 标志字节。子记录输出顺序与 Java 一致(Vehicle, Motor, FuelCell, Engine,
// Location, Extremum, Alarm, Voltage, Temperature)。
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
			// audit 2026-09-17 (L1):表 8 的自定义 flag 为 0x80~0xFE;拒绝任何
			// 其他 key(例如 0xFF),而不是发出一个会被解码器
			// 作为未知 TLV 类型拒绝的帧。
			if k < 0x80 || k > 0xFE {
				return fmt.Errorf("gb32960: custom data flag 0x%02X out of range 0x80~0xFE", k)
			}
			data := m.CustomData[byte(k)]
			w.WriteUint8(byte(k))
			w.WriteUint16(uint16(len(data)))
			w.WriteBytes(data)
		}
	}
	return nil
}

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

// realtimeCodecV2025For 返回为给定 TLV 类型注册的 V2025 编解码器,
// 每次调用时都重新查找。
//
// 重要(audit 2026-07-31):早期草稿在 init() 时构建单个分发表,参照了
// V2016 codec/gbt2016/realtime_data_codec.go 的模式。这很脆弱:Go 不
// 保证非依赖包之间的 init() 顺序,因此如果 codec/gbt2025 先于
// codec/gbt2025/realtime 初始化(这是合法的,两者互不 import),
// 该表最终会充满 nil 条目,Decode 在第一个 TLV 字节上就会以
// api.ErrUnknownTLVType 失败。每次调用时查找完全绕开了 init 顺序风险,
// 并且仍然使用 api.GetCodec(因此该文件与 codec/gbt2025/realtime 保持
// 解耦:消费方的空白导入才使注册在 Decode 运行时可见)。
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

// RealTimeDataV2025Codec 编码/解码 V2025 实时数据上报(命令 0x02)。
// 与 Java RealTimeDataV2025Codec 一致:BeanTime 通过注册的编解码器处理,
// 然后是 TLV 循环,逐一读取 (flag, sub-record) 对,直到缓冲区耗尽。
// 未知 flag 会以 api.ErrUnknownTLVType 暴露(不是静默 break:Java 会吞掉
// 它们,但 Plan 2 强制要求报错)。
type RealTimeDataV2025Codec struct{}

func init() {
	api.Register[mdl.RealTimeV2025Data](api.V2025, &RealTimeDataV2025Codec{})
}

// codecV2025For 返回为 v 的类型注册的 V2025 编解码器,未注册则为 nil。
// 与 codec/gbt2016/realtime_data_codec.go 中 V2016 的 codecFor 间接层一致。
func codecV2025For(v any) api.Codecer {
	t := reflect.TypeOf(v)
	return api.GetCodec(api.V2025, t)
}

// Decode 与 Java RealTimeDataV2025Codec.decodeBuffer / decodePayload /
// decodeByType 一致:BeanTime 通过注册的编解码器处理,然后是 TLV 循环,
// 逐一读取 (flag, sub-record) 对,直到缓冲区耗尽。
//
// 自定义范围 0x80~0xFE 通过 types.RealTimeV2025TypeByCode 分发,并由注册的
// CustomV2025Data 编解码器解码;原始 flag 字节保留在生成的
// CustomV2025Data.CustomKey 上(与 Java decodeByType 的 CUSTOM_DATA_FLAG
// 分支相符)。
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
		// 在读取 flag 之后立即快照:对于签名 TLV,SignData 覆盖此 flag 字节
		// 之前的全部内容(与 Java 在同一位置取得的
		// dumpBytes(start, readerIndex()-1-start) 一致)。
		consumedAtFlag := r.Consumed()
		rt, ok := types.RealTimeV2025TypeByCode(flag)
		if !ok {
			return nil, fmt.Errorf("%w: 0x%02X", api.ErrUnknownTLVType, flag)
		}

		// 自定义范围分发到 CustomV2025Data 编解码器;保留原始 flag。
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

// Encode 与 Java RealTimeDataV2025Codec.encodeBuffer / encodePayload 一致:
// BeanTime 通过注册的编解码器处理,然后要么按记录顺序输出 items 列表(非空时),
// 要么按 Java 的规范顺序输出各类型化字段
// (Vehicle, Motor, FuelCellEngine, Engine, Location, Alarm, MinParallelVoltage,
// BatteryTemp, FuelCellStack, SuperCapacitor, SuperCapExtremum, Custom...,
// Signature)。
//
// items 路径:与 Java 不同(Java 的解码从不把自定义数据或签名加入 items,
// 因此重编码已解码帧时会丢弃两者),Go 的 Decode 把每个 TLV 都记录进 Items,
// 所以 items 路径逐字节重发原始线上内容:顺序、重复、自定义数据和签名
// 都包含在内。VehicleSignature 主体会用已写入的字节刷新其 SignData(Java 通过
// buffer.dumpBytes 做同样的事);CustomV2025Data 主体写入自己的 CustomKey
// (无外层 flag 字节)。
//
// api.Writer 无法报告已写入的字节(SignData 刷新需要它们),因此所有内容都
// 编码进一个能做到这点的本地 *utils.ByteWriter,然后一次性刷出,这与
// protocolMessageCodec.Encode 的模式相同。
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
				// 复制:ByteWriter.Bytes() 返回的是活动缓冲区的别名
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

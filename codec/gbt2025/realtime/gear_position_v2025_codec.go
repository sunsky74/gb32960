package realtime

import (
	"github.com/sunsky74/gb32960/api"
	mdlrt16 "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// GearPositionV2025Codec 编解码打包在 V2025 VehicleData 内的挡位子记录。
// 线格式与 V2016 相同(1 字节);构造方式与 Java GearPositionV2025Codec 一致
// (后者构造 `new GearPosition(readByte(), true)`,即 IsV2025=true)。
// 注意:Java 的 effective 字段取值(bit7==1)与 2025.md L505 相反,Go 侧
// 按规范修正为 Effective=true ⇔ bit7=0(见下方 fix 注释),这是对 Java 的
// 有意偏离。模型结构体复用 model/gbt2016/realtime.GearPosition(与 Java 相同)。
type GearPositionV2025Codec struct{}

func init() {
	api.Register[mdlrt16.GearPosition](api.V2025, &GearPositionV2025Codec{})
}

// Decode 的位提取表达式与 V2016 GearPositionCodec 完全一致,
// 唯一区别是构造模型时 IsV2025=true。
// fix 2026-09-17:2025.md L505 —— bit7=1 表示挡位无效,Effective=true ⇔ bit7=0。
// Java 侧对应 VehicleDataV2025Codec 调用
// CodecMap.getV2025Codec(GearPosition.class)(Java VehicleDataV2025Codec.java L51-52),
// 若误用 V2016 编解码器则 isV2025=false,有效标识/驱动力/制动力位解析失效。
func (c *GearPositionV2025Codec) Decode(r api.Reader) (api.Message, error) {
	origin := r.ReadUint8()
	m := &mdlrt16.GearPosition{
		Origin:               origin,
		IsV2025:              true,
		Effective:            (origin >> 7 & 0x01) == 0,
		DrivingForceActive:   (origin >> 5 & 0x01) == 1,
		BrakingTorqueApplied: (origin >> 4 & 0x01) == 1,
		GP:                   types.GearPositionEnum(origin & 0x0F),
	}
	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

// Encode 写出原始 origin 字节。Java GearPositionV2025Codec.encodeBuffer
// 直接写出 msg.getOrigin(),因此派生字段在编码时不会被重新计算。
func (c *GearPositionV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdlrt16.GearPosition)
	w.WriteUint8(m.Origin)
	return nil
}

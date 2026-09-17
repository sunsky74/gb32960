// Package realtime 包含 V2016 实时子记录编解码器。
package realtime

import (
	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// GearPositionCodec 编解码打包在 VehicleData 内的 V2016 挡位子记录。
// 一个线上字节承载 bit 7 挡位有效标识(仅 V2025,1=挡位无效)、bit 5 驱动力、
// bit 4 制动力,以及 bits 3..0 的挡位枚举。构造方式与 Java GearPositionCodec
// 一致(后者构造 `new GearPosition(readByte(), false)`);但 Effective 推导
// 按规范修正,与 Java 不同(Java 的 effective 字段与规范相反,见下方 fix 注释)。
type GearPositionCodec struct{}

func init() {
	api.Register[mdl.GearPosition](api.V2016, &GearPositionCodec{})
}

// Decode 读取 1 字节,并按 Java GearPosition(byte, boolean) 构造函数的做法
// 精确推导 GearPosition 各字段:
//
//	effective            = ((origin >>> 7) & 0x01) == 0  // bit7=1 表示挡位无效
//	drivingForceActive   = ((origin >>> 5) & 0x01) == 1
//	brakingTorqueApplied = ((origin >>> 4) & 0x01) == 1
//	gp                   = types.GearPositionEnum(origin & 0x0F)
//
// fix 2026-09-17:2025.md L505(附录A.1 表A.1)—— bit7 = "1:挡位无效
// 0:挡位有效",故 Effective=true ⇔ bit7=0(旧实现 (origin>>7)==1 写反了)。
// 低四位映射(附录A.1):0x0 空挡,0x1..0xC 1挡..12挡,0xD 倒挡,
// 0xE 自动D挡,0xF 停车P挡 —— 见 types.GearPositionEnum。
// audit 2026-09-17:H2 —— 该文档此前针对旧的(错误的)枚举值引用了 Java
// GearPositionEnum.valueOf;解码表达式本身过去是、现在仍是 `origin & 0x0F`。
//
// 对 V2016 编解码器而言 IsV2025 恒为 false;bit7 在 2016 帧中为预留位
// (规范要求恒 0),因此本编解码器解出的 Effective 恒为 true,与字段
// 「仅 V2025 有效」的约定一致。
func (c *GearPositionCodec) Decode(r api.Reader) (api.Message, error) {
	origin := r.ReadUint8()
	m := &mdl.GearPosition{
		Origin:               origin,
		IsV2025:              false,
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

// Encode 写出原始 origin 字节。Java GearPositionCodec 直接写出
// msg.getOrigin(),因此派生字段在编码时不会被重新计算。
func (c *GearPositionCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.GearPosition)
	w.WriteUint8(m.Origin)
	return nil
}

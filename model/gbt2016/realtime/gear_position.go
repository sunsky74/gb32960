package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
	"github.com/sunsky74/gb32960/types"
)

// GearPosition 是打包在 VehicleData 内的 V2016 挡位子记录。
// 一个线上字节编码:位 7 挡位有效标识 (仅 V2025,1=挡位无效),位 5 驱动力,
// 位 4 制动力,位 3..0 档位 (P/R/N/D/...)。
// 线格式与参考 Java 实现(GearPosition)一致;但 Effective 语义按 2025.md
// 附录A.1 取值(Java 的 effective/isEffective() 与规范相反,Go 是有意修正,
// 见 gear_position_codec.go 的 fix 2026-09-17 注释)。
// 派生字段 Effective/DrivingForceActive/BrakingTorqueApplied/GP 由
// VehicleData 编解码器填充,因为该编解码器可以访问原始字节。
type GearPosition struct {
	Origin  byte // 原始线上字节
	IsV2025 bool // 是否是新国标 (V2016 为 false)
	// Effective 挡位数据有效标识:bit7=1 表示挡位无效 → Effective=false;
	// bit7=0 表示挡位有效 → Effective=true (2025.md L505 附录A.1 表A.1)。
	// 仅 V2025 有此语义;2016 该位为预留(恒 0),解码后 Effective 恒为 true。
	Effective            bool
	DrivingForceActive   bool                   // 有无驱动力标识
	BrakingTorqueApplied bool                   // 有无制动力标识
	GP                   types.GearPositionEnum // 档位 (解码后的低四位)
}

func (m *GearPosition) Version() api.GBTVersion { return api.V2016 }

func (m *GearPosition) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2016, m, reflect.TypeOf((*GearPosition)(nil)).Elem())
}

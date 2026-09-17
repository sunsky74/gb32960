package api

import "reflect"

var (
	v2016Codecs = map[reflect.Type]Codecer{}
	v2025Codecs = map[reflect.Type]Codecer{}
)

// Register 为特定的消息类型和协议版本注册编解码器。
// T 必须是消息结构体类型(例如 model.VehicleLogin)。
// 所有注册都发生在 init() 函数中,因此并发读取是安全的。
func Register[T Message](v GBTVersion, c Codecer) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	switch v {
	case V2016:
		v2016Codecs[t] = c
	case V2025:
		v2025Codecs[t] = c
	}
}

// GetCodec 获取某个消息类型和协议版本的编解码器。
// 若未注册任何编解码器则返回 nil。
func GetCodec(v GBTVersion, t reflect.Type) Codecer {
	switch v {
	case V2016:
		return v2016Codecs[t]
	case V2025:
		return v2025Codecs[t]
	}
	return nil
}

// RegisteredCount 返回给定版本已注册的编解码器数量。
// 可用于启动时的健全性检查(例如断言 codec/all 聚合包
// 已被导入)。由 audit 2026-07-31 添加(Plan 2 Task E0 Step 3)。
func RegisteredCount(v GBTVersion) int {
	switch v {
	case V2016:
		return len(v2016Codecs)
	case V2025:
		return len(v2025Codecs)
	}
	return 0
}

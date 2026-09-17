// Package all 空白导入每个 codec 子包,使导入
// "github.com/sunsky74/gb32960/codec/all" 时触发所有 codec init()
// 注册。消费方应导入本包,
// 而不是逐个列出各 codec 子包。
//
// 用法:
//
//	import _ "github.com/sunsky74/gb32960/codec/all"
package all

import (
	_ "github.com/sunsky74/gb32960/codec/gbt2016"
	_ "github.com/sunsky74/gb32960/codec/gbt2016/realtime"
	_ "github.com/sunsky74/gb32960/codec/gbt2025"
	_ "github.com/sunsky74/gb32960/codec/gbt2025/realtime"
)

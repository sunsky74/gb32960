// Package gbt2025 定义 GB/T 32960.3-2025 协议消息体结构体。
package gbt2025

import "github.com/sunsky74/gb32960/api"

// GBT2025Body 嵌入 V2025 消息结构体中,用于标识其协议版本。
type GBT2025Body struct{}

// Version 返回 V2025。
func (b GBT2025Body) Version() api.GBTVersion { return api.V2025 }

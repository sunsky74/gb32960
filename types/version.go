package types

import "github.com/sunsky74/gb32960/api"

// 协议常量
const (
	HeaderV2016 = "##"
	HeaderV2025 = "$$"
)

// GBTVersionByHeader 把 2 字节头部解析为协议版本。
func GBTVersionByHeader(header uint16) api.GBTVersion {
	switch header {
	case 8995: // "##"
		return api.V2016
	case 9252: // "$$"
		return api.V2025
	default:
		return api.V2016 // 默认兜底
	}
}

// Header 返回指定协议版本的线格式头部字节。
func Header(v api.GBTVersion) string {
	switch v {
	case api.V2016:
		return HeaderV2016
	case api.V2025:
		return HeaderV2025
	default:
		return HeaderV2016
	}
}

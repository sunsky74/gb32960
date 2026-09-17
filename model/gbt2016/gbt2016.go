package gbt2016

import "github.com/sunsky74/gb32960/api"

// GBT2016Body 内嵌于 V2016 消息结构体,用于标识其协议版本。
// 它自身不携带任何字段;内嵌它只是提升 Version() 方法,
// 使每个 V2016 消息体满足 model.MessageBody,
// 而无需重复定义版本常量。
type GBT2016Body struct{}

func (b GBT2016Body) Version() api.GBTVersion { return api.V2016 }

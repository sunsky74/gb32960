package model

import "github.com/sunsky74/gb32960/api"

// MessageBody 是所有消息体类型的接口。
// 每个消息体都知道自己的协议版本,并能把自己编码为字节。
type MessageBody interface {
	api.Message
	Version() api.GBTVersion
	Bytes() ([]byte, error)
}

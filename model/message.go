package model

import "github.com/sunsky74/gb32960/api"

// MessageBody is the interface for all message body types.
// Each message body knows its protocol version and can encode itself to bytes.
type MessageBody interface {
	api.Message
	Version() api.GBTVersion
	Bytes() ([]byte, error)
}

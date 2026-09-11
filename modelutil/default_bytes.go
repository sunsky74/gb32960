// Package modelutil provides the DefaultBytes helper used by all message body
// structs to encode themselves via the registered codec.
//
// This lives in a separate package (not model/) to break the circular import:
// model/gbt2016 and model/gbt2025 need DefaultBytes, but model/ itself must
// not be imported by them when model/ references gbt2016/gbt2025 types
// (via frame.ProtocolMessage).
package modelutil

import (
	"fmt"
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/utils"
)

// DefaultBytes is the default Bytes() implementation for message bodies.
// It looks up the registered codec and encodes the message.
// Callers must blank-import codec packages (or import codec/all) to trigger
// init() registration.
func DefaultBytes(v api.GBTVersion, msg api.Message, t reflect.Type) ([]byte, error) {
	codec := api.GetCodec(v, t)
	if codec == nil {
		return nil, fmt.Errorf("%w for %v", api.ErrCodecNotFound, t)
	}
	w := utils.NewByteWriter()
	if err := codec.Encode(w, msg); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

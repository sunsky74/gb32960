// Package modelutil 提供 DefaultBytes 辅助函数,供所有消息体结构体通过
// 注册的编解码器对自身编码。
//
// 它独立成包(而不是放在 model/)以打破循环导入:
// model/gbt2016 与 model/gbt2025 需要 DefaultBytes,但当 model/ 引用了
// gbt2016/gbt2025 的类型(经由 frame.ProtocolMessage)时,model/ 本身
// 不能被它们导入。
package modelutil

import (
	"fmt"
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/utils"
)

// DefaultBytes 是消息体的默认 Bytes() 实现。
// 它查找已注册的编解码器并对消息编码。
// 调用方必须空白导入编解码器包(或导入 codec/all)以触发
// init() 注册。
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

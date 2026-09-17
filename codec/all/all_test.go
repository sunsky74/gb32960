package all

import (
	"testing"

	"github.com/sunsky74/gb32960/api"
)

// TestAll_CodecsRegistered 是 Plan 2 Task E0 要求的防静默 nil 安全网:
// 导入 codec/all 必须触发两种协议版本的 codec 注册。
// 计数为零意味着 all.go 的空白导入列表中漏掉了某个 codec 子包,
// 没有这条断言,该故障只会表现为解码时 api.GetCodec
// 静默返回 nil codec。
func TestAll_CodecsRegistered(t *testing.T) {
	if got := api.RegisteredCount(api.V2016); got == 0 {
		t.Fatal("no V2016 codecs registered — did you forget to add a blank import to codec/all/all.go?")
	}
	if got := api.RegisteredCount(api.V2025); got == 0 {
		t.Fatal("no V2025 codecs registered — did you forget to add a blank import to codec/all/all.go?")
	}
}

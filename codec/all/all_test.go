package all

import (
	"testing"

	"github.com/sunsky74/gb32960/api"
)

// TestAll_CodecsRegistered is the anti-silent-nil safety net required by
// Plan 2 Task E0: importing codec/all must trigger registration of codecs
// for BOTH protocol versions. A zero count means a codec subpackage is
// missing from the blank-import list in all.go — without this assertion
// the failure would surface only as a silent nil-codec return from
// api.GetCodec at decode time.
func TestAll_CodecsRegistered(t *testing.T) {
	if got := api.RegisteredCount(api.V2016); got == 0 {
		t.Fatal("no V2016 codecs registered — did you forget to add a blank import to codec/all/all.go?")
	}
	if got := api.RegisteredCount(api.V2025); got == 0 {
		t.Fatal("no V2025 codecs registered — did you forget to add a blank import to codec/all/all.go?")
	}
}

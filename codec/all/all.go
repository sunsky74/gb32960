// Package all blank-imports every codec subpackage so that importing
// "github.com/sunsky74/gb32960/codec/all" triggers all codec init()
// registrations. Consumers should import this package instead of
// listing each codec subpackage separately.
//
// Usage:
//
//	import _ "github.com/sunsky74/gb32960/codec/all"
package all

import (
	_ "github.com/sunsky74/gb32960/codec/gbt2016"
	_ "github.com/sunsky74/gb32960/codec/gbt2016/realtime"
	_ "github.com/sunsky74/gb32960/codec/gbt2025"
	_ "github.com/sunsky74/gb32960/codec/gbt2025/realtime"
)

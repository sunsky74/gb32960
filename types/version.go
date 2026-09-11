package types

import "github.com/sunsky74/gb32960/api"

// Protocol constants
const (
	HeaderV2016 = "##"
	HeaderV2025 = "$$"
)

// GBTVersionByHeader parses the 2-byte header into a version.
func GBTVersionByHeader(header uint16) api.GBTVersion {
	switch header {
	case 8995: // "##"
		return api.V2016
	case 9252: // "$$"
		return api.V2025
	default:
		return api.V2016 // default fallback
	}
}

// Header returns the wire format header bytes for a version.
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

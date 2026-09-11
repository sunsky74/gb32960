package api

import "reflect"

var (
	v2016Codecs = map[reflect.Type]Codecer{}
	v2025Codecs = map[reflect.Type]Codecer{}
)

// Register registers a codec for a specific message type and protocol version.
// T must be the message struct type (e.g., model.VehicleLogin).
// All registration happens in init() functions — safe for concurrent reads.
func Register[T Message](v GBTVersion, c Codecer) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	switch v {
	case V2016:
		v2016Codecs[t] = c
	case V2025:
		v2025Codecs[t] = c
	}
}

// GetCodec retrieves the codec for a message type and protocol version.
// Returns nil if no codec is registered.
func GetCodec(v GBTVersion, t reflect.Type) Codecer {
	switch v {
	case V2016:
		return v2016Codecs[t]
	case V2025:
		return v2025Codecs[t]
	}
	return nil
}

// RegisteredCount returns the number of codecs registered for the given version.
// Useful for sanity checks at startup (e.g., asserting the codec/all aggregator
// was imported). Added by audit 2026-07-31 (Plan 2 Task E0 Step 3).
func RegisteredCount(v GBTVersion) int {
	switch v {
	case V2016:
		return len(v2016Codecs)
	case V2025:
		return len(v2025Codecs)
	}
	return 0
}

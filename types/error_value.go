package types

// DataErrorValue represents a pair of sentinel values (error + invalid).
// Used by ValueConverter to detect abnormal sensor readings.
type DataErrorValue struct {
	Error   int64 // Exception value (e.g., 65534 = 0xFFFE)
	Invalid int64 // Invalid value (e.g., 65535 = 0xFFFF)
}

// Predefined error value pairs matching Java DataErrorValue.
var (
	ErrByte1 = DataErrorValue{254, 255}                     // 0xFE, 0xFF
	ErrByte2 = DataErrorValue{65534, 65535}                 // 0xFFFE, 0xFFFF
	ErrByte4 = DataErrorValue{4_294_967_294, 4_294_967_295} // 0xFFFFFFFE, 0xFFFFFFFF
)

// IsInvalid returns true if v is either the error or invalid sentinel value.
func (e DataErrorValue) IsInvalid(v int64) bool {
	return v == e.Error || v == e.Invalid
}

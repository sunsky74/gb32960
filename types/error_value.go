package types

// DataErrorValue 表示一对哨兵值(错误值 + 无效值)。
// 由 ValueConverter 用于检测异常的传感器读数。
type DataErrorValue struct {
	Error   int64 // 异常值(如 65534 = 0xFFFE)
	Invalid int64 // 无效值(如 65535 = 0xFFFF)
}

// 预定义错误值对,与 Java DataErrorValue 一致。
var (
	ErrByte1 = DataErrorValue{254, 255}                     // 0xFE, 0xFF
	ErrByte2 = DataErrorValue{65534, 65535}                 // 0xFFFE, 0xFFFF
	ErrByte4 = DataErrorValue{4_294_967_294, 4_294_967_295} // 0xFFFFFFFE, 0xFFFFFFFF
)

// IsInvalid 在 v 等于错误值或无效哨兵值时返回 true。
func (e DataErrorValue) IsInvalid(v int64) bool {
	return v == e.Error || v == e.Invalid
}

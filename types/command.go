package types

// CommandV2016 holds a V2016 command definition, potentially covering a range of codes.
// Values aligned with the reference Java implementation (CommandV2016Type).
type CommandV2016 struct {
	Code byte
	Name string
	Min  byte
	Max  byte
}

// CommandV2025 holds a V2025 command definition.
// Values aligned with the reference Java implementation (CommandV2025Type).
type CommandV2025 struct {
	Code byte
	Name string
	Min  byte
	Max  byte
}

// NOTE: Command structs deliberately do NOT carry a MessageType().
// The reference Java implementation maps Command→message-body-type via an
// external switch (getV2016Body / getV2025Body). The Go mirror of that switch lives in
// model.payloadType() — see model/protocol.go. Do NOT add a MessageType()
// method here; Go has no override, so a stub returning nil would silently
// break DecodePayload (see audit 2026-07-31).

// V2016 command constants — aligned with Java CommandV2016Type.java.
// Key differences from earlier draft:
//   - 0x09~0x7F: RESERVED_FOR_UPLINK_DATA_SYSTEM (上行数据系统预留)
//   - 0x80/0x81/0x82: CONFIG_QUERY / CONFIG_SETUP / CONTROL
//   - 0x83~0xBF: RESERVED_FOR_DOWNLINK_DATA_SYSTEM (下行数据系统预留)
//   - 0xC0~0xFE: PLATFORM_EXCHANGE_CUSTOM_DATA (平台交换自定义数据)
var commandV2016Range = []CommandV2016{
	{0x01, "VEHICLE_LOGIN", 0x01, 0x01},
	{0x02, "REAL_TIME", 0x02, 0x02},
	{0x03, "REISSUE", 0x03, 0x03},
	{0x04, "VEHICLE_LOGOUT", 0x04, 0x04},
	{0x05, "PLATFORM_LOGIN", 0x05, 0x05},
	{0x06, "PLATFORM_LOGOUT", 0x06, 0x06},
	{0x07, "HEARTBEAT", 0x07, 0x07},
	{0x08, "CLOCK_CORRECT", 0x08, 0x08},
	{0x09, "RESERVED_FOR_UPLINK_DATA_SYSTEM", 0x09, 0x7F},
	{0x80, "CONFIG_QUERY", 0x80, 0x80},
	{0x81, "CONFIG_SETUP", 0x81, 0x81},
	{0x82, "CONTROL", 0x82, 0x82},
	{0x83, "RESERVED_FOR_DOWNLINK_DATA_SYSTEM", 0x83, 0xBF},
	{0xC0, "PLATFORM_EXCHANGE_CUSTOM_DATA", 0xC0, 0xFE},
}

// CommandV2016ByCode finds a command by its code byte.
func CommandV2016ByCode(code byte) *CommandV2016 {
	for i := range commandV2016Range {
		c := &commandV2016Range[i]
		if code >= c.Min && code <= c.Max {
			return c
		}
	}
	return nil
}

// V2025 command constants — aligned with Java CommandV2025Type.java.
// Key differences from V2016: adds ACTIVATE (0x09), ACTIVATE_RESULT (0x0A),
// DATA_UNIT_ENCRYPTION_KEY_EXCHANGE (0x0B); reserved uplink range starts at 0x0C.
var commandV2025Range = []CommandV2025{
	{0x01, "VEHICLE_LOGIN", 0x01, 0x01},
	{0x02, "REAL_TIME", 0x02, 0x02},
	{0x03, "REISSUE", 0x03, 0x03},
	{0x04, "VEHICLE_LOGOUT", 0x04, 0x04},
	{0x05, "PLATFORM_LOGIN", 0x05, 0x05},
	{0x06, "PLATFORM_LOGOUT", 0x06, 0x06},
	{0x07, "HEARTBEAT", 0x07, 0x07},
	{0x08, "CLOCK_CORRECT", 0x08, 0x08},
	{0x09, "ACTIVATE", 0x09, 0x09},
	{0x0A, "ACTIVATE_RESULT", 0x0A, 0x0A},
	{0x0B, "DATA_UNIT_ENCRYPTION_KEY_EXCHANGE", 0x0B, 0x0B},
	{0x0C, "RESERVED_FOR_UPLINK_DATA_SYSTEM", 0x0C, 0x7F},
	{0x80, "CONFIG_QUERY", 0x80, 0x80},
	{0x81, "CONFIG_SETUP", 0x81, 0x81},
	{0x82, "CONTROL", 0x82, 0x82},
	{0x83, "RESERVED_FOR_DOWNLINK_DATA_SYSTEM", 0x83, 0xBF},
	{0xC0, "PLATFORM_EXCHANGE_CUSTOM_DATA", 0xC0, 0xFE},
}

// CommandV2025ByCode finds a V2025 command by its code byte.
func CommandV2025ByCode(code byte) *CommandV2025 {
	for i := range commandV2025Range {
		c := &commandV2025Range[i]
		if code >= c.Min && code <= c.Max {
			return c
		}
	}
	return nil
}

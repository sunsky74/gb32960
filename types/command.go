package types

// CommandV2016 保存一条 V2016 命令定义,可覆盖一段编码范围。
// 取值与参考 Java 实现(CommandV2016Type)对齐。
type CommandV2016 struct {
	Code byte
	Name string
	Min  byte
	Max  byte
}

// CommandV2025 保存一条 V2025 命令定义。
// 取值与参考 Java 实现(CommandV2025Type)对齐。
type CommandV2025 struct {
	Code byte
	Name string
	Min  byte
	Max  byte
}

// 注意:Command 结构体刻意不携带 MessageType()。
// 参考 Java 实现通过外部 switch(getV2016Body / getV2025Body)把
// Command→消息体类型作映射。该 switch 在 Go 中的对应实现位于
// model.payloadType(),见 model/protocol.go。不要在此处添加 MessageType()
// 方法;Go 没有覆写机制,返回 nil 的桩实现会静默破坏
// DecodePayload(见 audit 2026-07-31)。

// V2016 命令常量,与 Java CommandV2016Type.java 对齐。
// 与早期草稿的主要差异:
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

// CommandV2016ByCode 按编码字节查找命令。
func CommandV2016ByCode(code byte) *CommandV2016 {
	for i := range commandV2016Range {
		c := commandV2016Range[i]
		if code >= c.Min && code <= c.Max {
			// fix 2026-09-17: 2016.md 表3 L55-65 —— 范围项(如
			// 0x09~0x7F 上行预留、0xC0~0xFE 平台交换自定义数据)的
			// Code 字段是范围基值,直接返回表项指针会让
			// frame.CommandCode 把实际线码(如 0x50)改写成基值,
			// 重编码即字节级损坏。返回表项副本并写入实际线码,
			// 静态表保持不变。
			c.Code = code
			return &c
		}
	}
	return nil
}

// V2025 命令常量,与 Java CommandV2025Type.java 对齐。
// 与 V2016 的主要差异:新增 ACTIVATE(0x09)、ACTIVATE_RESULT(0x0A)、
// DATA_UNIT_ENCRYPTION_KEY_EXCHANGE(0x0B);上行预留范围从 0x0C 开始。
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

// CommandV2025ByCode 按编码字节查找 V2025 命令。
func CommandV2025ByCode(code byte) *CommandV2025 {
	for i := range commandV2025Range {
		c := commandV2025Range[i]
		if code >= c.Min && code <= c.Max {
			// fix 2026-09-17: 2025.md 表3 L49-62 —— 与
			// CommandV2016ByCode 同理:范围项返回携带实际线码的副本,
			// 禁止改写静态表,保证重编码命令字节保真。
			c.Code = code
			return &c
		}
	}
	return nil
}

package realtime

import (
	"fmt"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// AlarmV2025Codec 编解码 V2025 报警数据子记录(TLV 类型 0x06)。
// V2025 报警有 28 个独立报警位打包在 u32 掩码中,外加 4 个 u32 码故障
// 列表(电池/电机/发动机/其他),外加一个由 (seq u8, level u8) 对组成的
// CommonAlertData 列表。(V2016 TLV 0x06 是极值;V2025 0x06 是报警。)
//
// 线格式与 Java AlarmV2025Codec 完全一致:
//
//	MaxAlarmLevel(u8)
//	AlarmBitIdentify(u32 → int64)
//	  → 28 个布尔报警位,按 LSB 优先解析(位 0..27)
//	BatteryFaultNum(u8) + BatteryFaultNum × u32   (仅在 0 < n <= 253 时)
//	MotorFaultNum(u8)   + MotorFaultNum   × u32
//	EngineFaultNum(u8)  + EngineFaultNum  × u32
//	OtherFaultNum(u8)   + OtherFaultNum   × u32
//	CommonAlertNum(u8)  + CommonAlertNum  × (u8 seq + u8 level)
//
// 位索引与 Java AlarmBits 一致(audit 2026-07-31):
//
//	0 TemperatureDifferential     10 BatteryConsistencyPoor      20 DriveMotorOverCurrent
//	1 BatteryHighTemperature      11 Insulation                  21 SuperCapacitorOverTemp
//	2 DeviceTypeOverVoltage       12 DCTemperature               22 SuperCapacitorOverVoltage
//	3 DeviceTypeUnderVoltage      13 BrakingSystem               23 DeviceThermalEvent
//	4 SOCLow                      14 DCStatus                    24 HydrogenLeakage
//	5 MonomerBatteryOverVoltage   15 DriveMotorControllerTemp    25 HydrogenPressureAbnormal
//	6 MonomerBatteryUnderVoltage  16 HighPressureInterlock       26 HydrogenTemperatureAbnormal
//	7 SOCHigh                     17 DriveMotorTemperature       27 FuelCellStackOverTemperature
//	8 SOCJump                     18 DeviceTypeOverFilling
//	9 DeviceTypeDontMatch         19 DriveMotorOverSpeed
type AlarmV2025Codec struct{}

// 28 位索引(在 32 位线格式掩码内按 LSB 优先)。位 0..18 与 V2016
// AlarmDataCodec 共享;位 19..27 是 V2025 新增。
const (
	v2025BitTemperatureDifferential         = 0
	v2025BitBatteryHighTemperature          = 1
	v2025BitDeviceTypeOverVoltage           = 2
	v2025BitDeviceTypeUnderVoltage          = 3
	v2025BitSOCLow                          = 4
	v2025BitMonomerBatteryOverVoltage       = 5
	v2025BitMonomerBatteryUnderVoltage      = 6
	v2025BitSOCHigh                         = 7
	v2025BitSOCJump                         = 8
	v2025BitDeviceTypeDontMatch             = 9
	v2025BitBatteryConsistencyPoor          = 10
	v2025BitInsulation                      = 11
	v2025BitDCTemperature                   = 12
	v2025BitBrakingSystem                   = 13
	v2025BitDCStatus                        = 14
	v2025BitDriveMotorControllerTemperature = 15
	v2025BitHighPressureInterlock           = 16
	v2025BitDriveMotorTemperature           = 17
	v2025BitDeviceTypeOverFilling           = 18
	v2025BitDriveMotorOverSpeed             = 19
	v2025BitDriveMotorOverCurrent           = 20
	v2025BitSuperCapacitorOverTemp          = 21
	v2025BitSuperCapacitorOverVoltage       = 22
	v2025BitDeviceThermalEvent              = 23
	v2025BitHydrogenLeakage                 = 24
	v2025BitHydrogenPressureAbnormal        = 25
	v2025BitHydrogenTemperatureAbnormal     = 26
	v2025BitFuelCellStackOverTemperature    = 27
)

func init() {
	api.Register[mdl.AlarmV2025Data](api.V2025, &AlarmV2025Codec{})
}

func (c *AlarmV2025Codec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.AlarmV2025Data{}
	m.MaxAlarmLevel = int(r.ReadUint8())
	m.AlarmBitIdentify = int64(r.ReadUint32())

	mask := m.AlarmBitIdentify
	m.TemperatureDifferential = isBitSetV2025(mask, v2025BitTemperatureDifferential)
	m.BatteryHighTemperature = isBitSetV2025(mask, v2025BitBatteryHighTemperature)
	m.DeviceTypeOverVoltage = isBitSetV2025(mask, v2025BitDeviceTypeOverVoltage)
	m.DeviceTypeUnderVoltage = isBitSetV2025(mask, v2025BitDeviceTypeUnderVoltage)
	m.SOCLow = isBitSetV2025(mask, v2025BitSOCLow)
	m.MonomerBatteryOverVoltage = isBitSetV2025(mask, v2025BitMonomerBatteryOverVoltage)
	m.MonomerBatteryUnderVoltage = isBitSetV2025(mask, v2025BitMonomerBatteryUnderVoltage)
	m.SOCHigh = isBitSetV2025(mask, v2025BitSOCHigh)
	m.SOCJump = isBitSetV2025(mask, v2025BitSOCJump)
	m.DeviceTypeDontMatch = isBitSetV2025(mask, v2025BitDeviceTypeDontMatch)
	m.BatteryConsistencyPoor = isBitSetV2025(mask, v2025BitBatteryConsistencyPoor)
	m.Insulation = isBitSetV2025(mask, v2025BitInsulation)
	m.DCTemperature = isBitSetV2025(mask, v2025BitDCTemperature)
	m.BrakingSystem = isBitSetV2025(mask, v2025BitBrakingSystem)
	m.DCStatus = isBitSetV2025(mask, v2025BitDCStatus)
	m.DriveMotorControllerTemperature = isBitSetV2025(mask, v2025BitDriveMotorControllerTemperature)
	m.HighPressureInterlock = isBitSetV2025(mask, v2025BitHighPressureInterlock)
	m.DriveMotorTemperature = isBitSetV2025(mask, v2025BitDriveMotorTemperature)
	m.DeviceTypeOverFilling = isBitSetV2025(mask, v2025BitDeviceTypeOverFilling)
	m.DriveMotorOverSpeed = isBitSetV2025(mask, v2025BitDriveMotorOverSpeed)
	m.DriveMotorOverCurrent = isBitSetV2025(mask, v2025BitDriveMotorOverCurrent)
	m.SuperCapacitorOverTemp = isBitSetV2025(mask, v2025BitSuperCapacitorOverTemp)
	m.SuperCapacitorOverVoltage = isBitSetV2025(mask, v2025BitSuperCapacitorOverVoltage)
	m.DeviceThermalEvent = isBitSetV2025(mask, v2025BitDeviceThermalEvent)
	m.HydrogenLeakage = isBitSetV2025(mask, v2025BitHydrogenLeakage)
	m.HydrogenPressureAbnormal = isBitSetV2025(mask, v2025BitHydrogenPressureAbnormal)
	m.HydrogenTemperatureAbnormal = isBitSetV2025(mask, v2025BitHydrogenTemperatureAbnormal)
	m.FuelCellStackOverTemperature = isBitSetV2025(mask, v2025BitFuelCellStackOverTemperature)

	m.BatteryFaultNum = int(r.ReadUint8())
	if isValidV2025Count(m.BatteryFaultNum) {
		m.BatteryFaultDatas = make([]int64, m.BatteryFaultNum)
		for i := 0; i < m.BatteryFaultNum; i++ {
			m.BatteryFaultDatas[i] = int64(r.ReadUint32())
		}
	}

	m.MotorFaultNum = int(r.ReadUint8())
	if isValidV2025Count(m.MotorFaultNum) {
		m.MotorFaultDatas = make([]int64, m.MotorFaultNum)
		for i := 0; i < m.MotorFaultNum; i++ {
			m.MotorFaultDatas[i] = int64(r.ReadUint32())
		}
	}

	m.EngineFaultNum = int(r.ReadUint8())
	if isValidV2025Count(m.EngineFaultNum) {
		m.EngineFaultDatas = make([]int64, m.EngineFaultNum)
		for i := 0; i < m.EngineFaultNum; i++ {
			m.EngineFaultDatas[i] = int64(r.ReadUint32())
		}
	}

	m.OtherFaultNum = int(r.ReadUint8())
	if isValidV2025Count(m.OtherFaultNum) {
		m.OtherFaultDatas = make([]int64, m.OtherFaultNum)
		for i := 0; i < m.OtherFaultNum; i++ {
			m.OtherFaultDatas[i] = int64(r.ReadUint32())
		}
	}

	m.CommonAlertNum = int(r.ReadUint8())
	if isValidV2025Count(m.CommonAlertNum) {
		m.CommonAlertDatas = make([]mdl.CommonAlertData, m.CommonAlertNum)
		for i := 0; i < m.CommonAlertNum; i++ {
			m.CommonAlertDatas[i] = mdl.CommonAlertData{
				Seq:   int(r.ReadUint8()),
				Level: int(r.ReadUint8()),
			}
		}
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *AlarmV2025Codec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.AlarmV2025Data)
	w.WriteUint8(byte(m.MaxAlarmLevel))

	// Java: getAlarmBitIdentify() == null ? buildAlarmBitIdentify(msg) : 存储的
	// 掩码;已解码的掩码原样写回,使预留位
	// 28..31 能在往返后存活。Go 没有 null:零值承担该角色
	// (解码出的 0 掩码意味着所有布尔均为 false,因此重建得到
	// 相同的 0,两条路径在线格式字节上从不分歧)。
	mask := m.AlarmBitIdentify
	if mask == 0 {
		mask = buildV2025AlarmBitIdentify(m)
	}
	w.WriteUint32(uint32(mask))

	// fix 2026-09-17: GB/T 32960.3-2025 表23(L342~L349) —— 四个故障总数 N1~N4
	// 与各自代码列表长度必须一致(普通计数 1~253)或列表为空(0/0xFE/0xFF)。
	if err := writeV2025Faults(w, m.BatteryFaultNum, m.BatteryFaultDatas); err != nil {
		return err
	}
	if err := writeV2025Faults(w, m.MotorFaultNum, m.MotorFaultDatas); err != nil {
		return err
	}
	if err := writeV2025Faults(w, m.EngineFaultNum, m.EngineFaultDatas); err != nil {
		return err
	}
	if err := writeV2025Faults(w, m.OtherFaultNum, m.OtherFaultDatas); err != nil {
		return err
	}

	commonCount := m.CommonAlertNum
	// fix 2026-09-17: 表23(L355~L356) —— 通用报警故障总数 N5 与等级列表长度必须
	// 一致(普通计数)或列表为空(0/哨兵);旧代码在计数与列表不一致时静默写出
	// 全部条目,产生畸形帧。
	if isValidV2025Count(commonCount) {
		if len(m.CommonAlertDatas) != commonCount {
			return fmt.Errorf("gb32960: common alert count %d does not match list length %d", commonCount, len(m.CommonAlertDatas))
		}
	} else if len(m.CommonAlertDatas) != 0 {
		return fmt.Errorf("gb32960: common alert count %d (0 or sentinel) cannot carry %d entries", commonCount, len(m.CommonAlertDatas))
	}
	w.WriteUint8(byte(commonCount))
	for i := range m.CommonAlertDatas {
		w.WriteUint8(byte(m.CommonAlertDatas[i].Seq))
		w.WriteUint8(byte(m.CommonAlertDatas[i].Level))
	}
	return nil
}

// isValidV2025Count 与 Java AlarmV2025Codec.isValidCount 一致:
// 计数必须处于 (0, 253] 才会读写列表主体。
func isValidV2025Count(count int) bool { return count > 0 && count <= 253 }

// writeV2025Faults 与 Java AlarmV2025Codec.writeFaults 一致:先写计数
// 字节,再把每个故障写为 u32。
//
// fix 2026-09-17: 表23(L342~L349) —— 计数与代码列表长度的编码校验:普通
// 计数(1~253)必须与列表长度完全一致;0 或哨兵(0xFE/0xFF)计数不携带
// 任何代码。旧代码在计数与列表不一致时静默写出全部条目,产生畸形帧。
func writeV2025Faults(w api.Writer, count int, faults []int64) error {
	if isValidV2025Count(count) {
		if len(faults) != count {
			return fmt.Errorf("gb32960: alarm fault count %d does not match fault list length %d", count, len(faults))
		}
	} else if len(faults) != 0 {
		return fmt.Errorf("gb32960: alarm fault count %d (0 or sentinel) cannot carry %d fault codes", count, len(faults))
	}
	w.WriteUint8(byte(count))
	for _, code := range faults {
		w.WriteUint32(uint32(code))
	}
	return nil
}

func isBitSetV2025(mask int64, bit int) bool { return mask&(1<<bit) != 0 }

// buildV2025AlarmBitIdentify 把 28 个布尔字段打包回 32 位
// 线格式掩码,与 Java AlarmV2025Codec.buildAlarmBitIdentify 一致。当
// AlarmBitIdentify 未设置(Java: null)时使用;已设置的掩码由
// Encode 直通,从而保留预留位 28..31。
func buildV2025AlarmBitIdentify(m *mdl.AlarmV2025Data) int64 {
	var mask int64
	set := func(bit int, b bool) {
		if b {
			mask |= 1 << bit
		}
	}
	set(v2025BitTemperatureDifferential, m.TemperatureDifferential)
	set(v2025BitBatteryHighTemperature, m.BatteryHighTemperature)
	set(v2025BitDeviceTypeOverVoltage, m.DeviceTypeOverVoltage)
	set(v2025BitDeviceTypeUnderVoltage, m.DeviceTypeUnderVoltage)
	set(v2025BitSOCLow, m.SOCLow)
	set(v2025BitMonomerBatteryOverVoltage, m.MonomerBatteryOverVoltage)
	set(v2025BitMonomerBatteryUnderVoltage, m.MonomerBatteryUnderVoltage)
	set(v2025BitSOCHigh, m.SOCHigh)
	set(v2025BitSOCJump, m.SOCJump)
	set(v2025BitDeviceTypeDontMatch, m.DeviceTypeDontMatch)
	set(v2025BitBatteryConsistencyPoor, m.BatteryConsistencyPoor)
	set(v2025BitInsulation, m.Insulation)
	set(v2025BitDCTemperature, m.DCTemperature)
	set(v2025BitBrakingSystem, m.BrakingSystem)
	set(v2025BitDCStatus, m.DCStatus)
	set(v2025BitDriveMotorControllerTemperature, m.DriveMotorControllerTemperature)
	set(v2025BitHighPressureInterlock, m.HighPressureInterlock)
	set(v2025BitDriveMotorTemperature, m.DriveMotorTemperature)
	set(v2025BitDeviceTypeOverFilling, m.DeviceTypeOverFilling)
	set(v2025BitDriveMotorOverSpeed, m.DriveMotorOverSpeed)
	set(v2025BitDriveMotorOverCurrent, m.DriveMotorOverCurrent)
	set(v2025BitSuperCapacitorOverTemp, m.SuperCapacitorOverTemp)
	set(v2025BitSuperCapacitorOverVoltage, m.SuperCapacitorOverVoltage)
	set(v2025BitDeviceThermalEvent, m.DeviceThermalEvent)
	set(v2025BitHydrogenLeakage, m.HydrogenLeakage)
	set(v2025BitHydrogenPressureAbnormal, m.HydrogenPressureAbnormal)
	set(v2025BitHydrogenTemperatureAbnormal, m.HydrogenTemperatureAbnormal)
	set(v2025BitFuelCellStackOverTemperature, m.FuelCellStackOverTemperature)
	return mask
}

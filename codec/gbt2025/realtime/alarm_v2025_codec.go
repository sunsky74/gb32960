package realtime

import (
	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2025/realtime"
)

// AlarmV2025Codec encodes/decodes the V2025 报警数据 sub-record (TLV type 0x06).
// V2025 alarm has 28 individual alarm bits packed in a u32 mask, plus 4 fault
// lists (battery/motor/engine/other) of u32 codes, plus a CommonAlertData list
// of (seq u8, level u8) pairs. (V2016 TLV 0x06 was Extremum; V2025 0x06 is Alarm.)
//
// Wire layout mirrors Java AlarmV2025Codec exactly:
//
//	MaxAlarmLevel(u8)
//	AlarmBitIdentify(u32 → int64)
//	  → 28 boolean alarm bits parsed LSB-first (bits 0..27)
//	BatteryFaultNum(u8) + BatteryFaultNum × u32   (only when 0 < n <= 253)
//	MotorFaultNum(u8)   + MotorFaultNum   × u32
//	EngineFaultNum(u8)  + EngineFaultNum  × u32
//	OtherFaultNum(u8)   + OtherFaultNum   × u32
//	CommonAlertNum(u8)  + CommonAlertNum  × (u8 seq + u8 level)
//
// Bit indices match Java AlarmBits (audit 2026-07-31):
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

// 28-bit indices (LSB-first within the 32-bit wire mask). Shared with V2016
// AlarmDataCodec for bits 0..18; bits 19..27 are V2025-only additions.
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

	// Java: getAlarmBitIdentify() == null ? buildAlarmBitIdentify(msg) : the
	// stored mask — a decoded mask is written back as-is so reserved bits
	// 28..31 survive the roundtrip. Go has no null: the zero value plays it
	// (a decoded 0 mask implies all-false booleans, so rebuilding yields the
	// same 0 and the two paths never disagree on wire bytes).
	mask := m.AlarmBitIdentify
	if mask == 0 {
		mask = buildV2025AlarmBitIdentify(m)
	}
	w.WriteUint32(uint32(mask))

	writeV2025Faults(w, m.BatteryFaultNum, m.BatteryFaultDatas)
	writeV2025Faults(w, m.MotorFaultNum, m.MotorFaultDatas)
	writeV2025Faults(w, m.EngineFaultNum, m.EngineFaultDatas)
	writeV2025Faults(w, m.OtherFaultNum, m.OtherFaultDatas)

	commonCount := m.CommonAlertNum
	w.WriteUint8(byte(commonCount))
	if isValidV2025Count(commonCount) && len(m.CommonAlertDatas) > 0 {
		for i := range m.CommonAlertDatas {
			w.WriteUint8(byte(m.CommonAlertDatas[i].Seq))
			w.WriteUint8(byte(m.CommonAlertDatas[i].Level))
		}
	}
	return nil
}

// isValidV2025Count mirrors Java AlarmV2025Codec.isValidCount:
// count must be in (0, 253] for the list body to be read/written.
func isValidV2025Count(count int) bool { return count > 0 && count <= 253 }

// writeV2025Faults mirrors Java AlarmV2025Codec.writeFaults: write the count
// byte then each fault as u32 (only when count is in valid range).
func writeV2025Faults(w api.Writer, count int, faults []int64) {
	w.WriteUint8(byte(count))
	if isValidV2025Count(count) && len(faults) > 0 {
		for _, code := range faults {
			w.WriteUint32(uint32(code))
		}
	}
}

func isBitSetV2025(mask int64, bit int) bool { return mask&(1<<bit) != 0 }

// buildV2025AlarmBitIdentify packs the 28 boolean fields back into the 32-bit
// wire mask, matching Java AlarmV2025Codec.buildAlarmBitIdentify. Used when
// AlarmBitIdentify is unset (Java: null); a set mask is passed through by
// Encode so reserved bits 28..31 are preserved.
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

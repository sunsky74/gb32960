package realtime

import (
	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
)

// AlarmDataCodec encodes/decodes the V2016 报警数据 sub-record (TLV type 0x07).
// Field order and bit layout mirror Java AlarmDataCodec exactly:
//
//	MaxAlarmLevel(u8)
//	AlarmBitIdentify(u32 → int64, sign-extended)
//	19 boolean alarm bits parsed from AlarmBitIdentify (bit 0..18, LSB first)
//	BatteryFaultNum(u8) + BatteryFaultNum × u32  (BatteryFaultDatas)
//	MotorFaultNum(u8)   + MotorFaultNum   × u32  (MotorFaultDatas)
//	EngineFaultNum(u8)  + EngineFaultNum  × u32  (EngineFaultDatas)
//	OtherFaultNum(u8)   + OtherFaultNum   × u32  (OtherFaultDatas)
//
// NOTE: the task brief's summary "faultCount(u8 → int) + faultCodeCount × u32
// fault list" was simplified to ONE list. The actual AlarmData model exposes
// FOUR independent fault lists (battery/motor/engine/other), matching Java.
// We follow the model + Java — see model/gbt2016/realtime/alarm_data.go.
type AlarmDataCodec struct{}

func init() {
	api.Register[mdl.AlarmData](api.V2016, &AlarmDataCodec{})
}

// alarm bit indices 0..18 (LSB-first within the 32-bit wire mask).
const (
	bitTemperatureDifferential         = 0
	bitBatteryHighTemperature          = 1
	bitDeviceTypeOverVoltage           = 2
	bitDeviceTypeUnderVoltage          = 3
	bitSocLow                          = 4
	bitMonomerBatteryOverVoltage       = 5
	bitMonomerBatteryUnderVoltage      = 6
	bitSocHigh                         = 7
	bitSocJump                         = 8
	bitDeviceTypeDontMatch             = 9
	bitBatteryConsistencyPoor          = 10
	bitInsulation                      = 11
	bitDcTemperature                   = 12
	bitBrakingSystem                   = 13
	bitDcStatus                        = 14
	bitDriveMotorControllerTemperature = 15
	bitHighPressureInterlock           = 16
	bitDriveMotorTemperature           = 17
	bitDeviceTypeOverFilling           = 18
)

func isBitSet(mask int64, bit int) bool { return mask&(1<<bit) != 0 }

func (c *AlarmDataCodec) Decode(r api.Reader) (api.Message, error) {
	m := &mdl.AlarmData{}
	m.MaxAlarmLevel = int(r.ReadUint8())
	m.AlarmBitIdentify = int64(r.ReadUint32())

	m.TemperatureDifferential = isBitSet(m.AlarmBitIdentify, bitTemperatureDifferential)
	m.BatteryHighTemperature = isBitSet(m.AlarmBitIdentify, bitBatteryHighTemperature)
	m.DeviceTypeOverVoltage = isBitSet(m.AlarmBitIdentify, bitDeviceTypeOverVoltage)
	m.DeviceTypeUnderVoltage = isBitSet(m.AlarmBitIdentify, bitDeviceTypeUnderVoltage)
	m.SocLow = isBitSet(m.AlarmBitIdentify, bitSocLow)
	m.MonomerBatteryOverVoltage = isBitSet(m.AlarmBitIdentify, bitMonomerBatteryOverVoltage)
	m.MonomerBatteryUnderVoltage = isBitSet(m.AlarmBitIdentify, bitMonomerBatteryUnderVoltage)
	m.SocHigh = isBitSet(m.AlarmBitIdentify, bitSocHigh)
	m.SocJump = isBitSet(m.AlarmBitIdentify, bitSocJump)
	m.DeviceTypeDontMatch = isBitSet(m.AlarmBitIdentify, bitDeviceTypeDontMatch)
	m.BatteryConsistencyPoor = isBitSet(m.AlarmBitIdentify, bitBatteryConsistencyPoor)
	m.Insulation = isBitSet(m.AlarmBitIdentify, bitInsulation)
	m.DcTemperature = isBitSet(m.AlarmBitIdentify, bitDcTemperature)
	m.BrakingSystem = isBitSet(m.AlarmBitIdentify, bitBrakingSystem)
	m.DcStatus = isBitSet(m.AlarmBitIdentify, bitDcStatus)
	m.DriveMotorControllerTemperature = isBitSet(m.AlarmBitIdentify, bitDriveMotorControllerTemperature)
	m.HighPressureInterlock = isBitSet(m.AlarmBitIdentify, bitHighPressureInterlock)
	m.DriveMotorTemperature = isBitSet(m.AlarmBitIdentify, bitDriveMotorTemperature)
	m.DeviceTypeOverFilling = isBitSet(m.AlarmBitIdentify, bitDeviceTypeOverFilling)

	m.BatteryFaultNum = int(r.ReadUint8())
	if m.BatteryFaultNum > 0 {
		m.BatteryFaultDatas = make([]int64, m.BatteryFaultNum)
		for i := 0; i < m.BatteryFaultNum; i++ {
			m.BatteryFaultDatas[i] = int64(r.ReadUint32())
		}
	}

	m.MotorFaultNum = int(r.ReadUint8())
	if m.MotorFaultNum > 0 {
		m.MotorFaultDatas = make([]int64, m.MotorFaultNum)
		for i := 0; i < m.MotorFaultNum; i++ {
			m.MotorFaultDatas[i] = int64(r.ReadUint32())
		}
	}

	m.EngineFaultNum = int(r.ReadUint8())
	if m.EngineFaultNum > 0 {
		m.EngineFaultDatas = make([]int64, m.EngineFaultNum)
		for i := 0; i < m.EngineFaultNum; i++ {
			m.EngineFaultDatas[i] = int64(r.ReadUint32())
		}
	}

	m.OtherFaultNum = int(r.ReadUint8())
	if m.OtherFaultNum > 0 {
		m.OtherFaultDatas = make([]int64, m.OtherFaultNum)
		for i := 0; i < m.OtherFaultNum; i++ {
			m.OtherFaultDatas[i] = int64(r.ReadUint32())
		}
	}

	if err := r.Err(); err != nil {
		return nil, api.ErrBufferUnderflow
	}
	return m, nil
}

func (c *AlarmDataCodec) Encode(w api.Writer, msg api.Message) error {
	m := msg.(*mdl.AlarmData)
	w.WriteUint8(byte(m.MaxAlarmLevel))

	mask := buildAlarmBitIdentify(m)
	w.WriteUint32(uint32(mask))

	w.WriteUint8(byte(m.BatteryFaultNum))
	for _, code := range m.BatteryFaultDatas {
		w.WriteUint32(uint32(code))
	}

	w.WriteUint8(byte(m.MotorFaultNum))
	for _, code := range m.MotorFaultDatas {
		w.WriteUint32(uint32(code))
	}

	w.WriteUint8(byte(m.EngineFaultNum))
	for _, code := range m.EngineFaultDatas {
		w.WriteUint32(uint32(code))
	}

	w.WriteUint8(byte(m.OtherFaultNum))
	for _, code := range m.OtherFaultDatas {
		w.WriteUint32(uint32(code))
	}
	return nil
}

// buildAlarmBitIdentify packs the 19 boolean fields back into the 32-bit wire
// mask, matching Java AlarmDataCodec.buildAlarmBitIdentify. If AlarmBitIdentify
// was set directly on the message we still recompute from the boolean fields —
// Java does the same and treats the booleans as source of truth on encode.
func buildAlarmBitIdentify(m *mdl.AlarmData) int64 {
	var mask int64
	set := func(bit int, b bool) {
		if b {
			mask |= 1 << bit
		}
	}
	set(bitTemperatureDifferential, m.TemperatureDifferential)
	set(bitBatteryHighTemperature, m.BatteryHighTemperature)
	set(bitDeviceTypeOverVoltage, m.DeviceTypeOverVoltage)
	set(bitDeviceTypeUnderVoltage, m.DeviceTypeUnderVoltage)
	set(bitSocLow, m.SocLow)
	set(bitMonomerBatteryOverVoltage, m.MonomerBatteryOverVoltage)
	set(bitMonomerBatteryUnderVoltage, m.MonomerBatteryUnderVoltage)
	set(bitSocHigh, m.SocHigh)
	set(bitSocJump, m.SocJump)
	set(bitDeviceTypeDontMatch, m.DeviceTypeDontMatch)
	set(bitBatteryConsistencyPoor, m.BatteryConsistencyPoor)
	set(bitInsulation, m.Insulation)
	set(bitDcTemperature, m.DcTemperature)
	set(bitBrakingSystem, m.BrakingSystem)
	set(bitDcStatus, m.DcStatus)
	set(bitDriveMotorControllerTemperature, m.DriveMotorControllerTemperature)
	set(bitHighPressureInterlock, m.HighPressureInterlock)
	set(bitDriveMotorTemperature, m.DriveMotorTemperature)
	set(bitDeviceTypeOverFilling, m.DeviceTypeOverFilling)
	return mask
}

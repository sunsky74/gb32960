package realtime

import (
	"fmt"

	"github.com/sunsky74/gb32960/api"
	mdl "github.com/sunsky74/gb32960/model/gbt2016/realtime"
	"github.com/sunsky74/gb32960/types"
)

// AlarmDataCodec 编解码 V2016 报警数据子记录(TLV 类型 0x07)。
// 字段顺序与位布局与 Java AlarmDataCodec 完全一致:
//
//	MaxAlarmLevel(u8)
//	AlarmBitIdentify(u32 → int64,符号扩展)
//	从 AlarmBitIdentify 解析出的 19 个布尔报警位(bit 0..18,低位在前)
//	BatteryFaultNum(u8) + BatteryFaultNum × u32  (BatteryFaultDatas)
//	MotorFaultNum(u8)   + MotorFaultNum   × u32  (MotorFaultDatas)
//	EngineFaultNum(u8)  + EngineFaultNum  × u32  (EngineFaultDatas)
//	OtherFaultNum(u8)   + OtherFaultNum   × u32  (OtherFaultDatas)
//
// 注意:任务简报的摘要 "faultCount(u8 → int) + faultCodeCount × u32
// fault list" 被简化成了单个列表。实际的 AlarmData 模型暴露
// 四个独立的故障列表(电池/电机/发动机/其他),与 Java 一致。
// 我们以模型 + Java 为准 —— 见 model/gbt2016/realtime/alarm_data.go。
type AlarmDataCodec struct{}

func init() {
	api.Register[mdl.AlarmData](api.V2016, &AlarmDataCodec{})
}

// 报警位下标 0..18(在 32 位线格式掩码内低位在前)。
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

// alarmDefinedBits 覆盖通用报警标志(表18)已定义的低位 0..18。
// Bits 19..31 为预留,编码时从 AlarmBitIdentify 原样保留
// (audit 2026-09-17:L4)。
const alarmDefinedBits = int64(1)<<19 - 1

// isBitSet 报告给定 bit 在原始线格式掩码中是否置位。
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

	// audit 2026-09-17:M2 —— N1..N4 使用表17 的字节哨兵值(0xFE 异常,0xFF 无效);
	// 故障条目只跟随正常计数,绝不跟随哨兵值。
	m.BatteryFaultNum = int(r.ReadUint8())
	if m.BatteryFaultNum > 0 && !types.ErrByte1.IsInvalid(int64(m.BatteryFaultNum)) {
		m.BatteryFaultDatas = make([]int64, m.BatteryFaultNum)
		for i := 0; i < m.BatteryFaultNum; i++ {
			m.BatteryFaultDatas[i] = int64(r.ReadUint32())
		}
	}

	m.MotorFaultNum = int(r.ReadUint8())
	if m.MotorFaultNum > 0 && !types.ErrByte1.IsInvalid(int64(m.MotorFaultNum)) {
		m.MotorFaultDatas = make([]int64, m.MotorFaultNum)
		for i := 0; i < m.MotorFaultNum; i++ {
			m.MotorFaultDatas[i] = int64(r.ReadUint32())
		}
	}

	m.EngineFaultNum = int(r.ReadUint8())
	if m.EngineFaultNum > 0 && !types.ErrByte1.IsInvalid(int64(m.EngineFaultNum)) {
		m.EngineFaultDatas = make([]int64, m.EngineFaultNum)
		for i := 0; i < m.EngineFaultNum; i++ {
			m.EngineFaultDatas[i] = int64(r.ReadUint32())
		}
	}

	m.OtherFaultNum = int(r.ReadUint8())
	if m.OtherFaultNum > 0 && !types.ErrByte1.IsInvalid(int64(m.OtherFaultNum)) {
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

	// audit 2026-09-17:L4 —— 这些布尔值只重建已定义的 bits 0..18;
	// AlarmBitIdentify 中预留的 bits 19..31 会被 OR 回填,使解码后的
	// 掩码在解码 → 编码后逐字节稳定地保留下来。
	mask := (m.AlarmBitIdentify &^ alarmDefinedBits) | buildAlarmBitIdentify(m)
	w.WriteUint32(uint32(mask))

	// audit 2026-09-17:M2/M4 —— 计数字节总是写出;仅当计数是正常数值且
	// 与列表长度匹配时,才跟随写出列表条目。
	if err := writeFaultList(w, "battery", m.BatteryFaultNum, m.BatteryFaultDatas); err != nil {
		return err
	}
	if err := writeFaultList(w, "motor", m.MotorFaultNum, m.MotorFaultDatas); err != nil {
		return err
	}
	if err := writeFaultList(w, "engine", m.EngineFaultNum, m.EngineFaultDatas); err != nil {
		return err
	}
	return writeFaultList(w, "other", m.OtherFaultNum, m.OtherFaultDatas)
}

// writeFaultList 写出一个表17 故障块:计数字节(总是写出),然后是
// u32 故障码。哨兵计数(0xFE 异常 / 0xFF 无效)必须携带空
// 列表;正常计数必须与列表长度匹配(audit 2026-09-17:M2/M4)。
func writeFaultList(w api.Writer, name string, count int, codes []int64) error {
	if types.ErrByte1.IsInvalid(int64(count)) {
		if len(codes) != 0 {
			return fmt.Errorf("gb32960: alarm %s fault count %d is a sentinel but list length is %d", name, count, len(codes))
		}
	} else if count != len(codes) {
		return fmt.Errorf("gb32960: alarm %s fault count %d does not match list length %d", name, count, len(codes))
	}
	w.WriteUint8(byte(count))
	for _, code := range codes {
		w.WriteUint32(uint32(code))
	}
	return nil
}

// buildAlarmBitIdentify 把 19 个布尔字段打包进 32 位线格式掩码已定义的低位
// (0..18),与 Java AlarmDataCodec.buildAlarmBitIdentify 一致。
// 对已定义的位而言,这些布尔值就是真理源;预留的 bits 19..31
// 由 Encode 从 m.AlarmBitIdentify 保留(audit 2026-09-17:L4)。
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

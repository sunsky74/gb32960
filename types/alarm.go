package types

// Alarm bit positions for the alarm bitmask.
const (
	AlarmBitTempDiff            = 0 // Temperature difference alarm
	AlarmBitBatteryHighTemp     = 1 // Battery high temperature alarm
	AlarmBitVehicleHighVoltage  = 2 // Vehicle high voltage alarm
	AlarmBitVehicleLowVoltage   = 3 // Vehicle low voltage alarm
	AlarmBitSOCLow              = 4 // SOC low alarm
	AlarmBitSingleCellOverVolt  = 5 // Single cell over-voltage alarm
	AlarmBitSingleCellUnderVolt = 6 // Single cell under-voltage alarm
	AlarmBitSOCJump             = 7 // SOC jump alarm
	AlarmBitChargeOverCurrent   = 8 // Charge over-current alarm
	// ... additional bits as defined in GB/T 32960
)

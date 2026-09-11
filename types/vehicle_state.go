package types

// OperatingState represents vehicle operating state.
type OperatingState byte

const (
	OpStateOn        OperatingState = 0x01
	OpStateOff       OperatingState = 0x02
	OpStateOther     OperatingState = 0x03
	OpStateException OperatingState = 0xFE
	OpStateInvalid   OperatingState = 0xFF
)

// ChargingState represents vehicle charging state.
type ChargingState byte

const (
	ChargeStateNotCharging ChargingState = 0x01
	ChargeStateCharging    ChargingState = 0x02
	ChargeStateComplete    ChargingState = 0x03
	ChargeStateException   ChargingState = 0xFE
	ChargeStateInvalid     ChargingState = 0xFF
)

// OperationMode represents vehicle operation mode.
type OperationMode byte

const (
	OpModeElectric  OperationMode = 0x01
	OpModeHybrid    OperationMode = 0x02
	OpModeFuel      OperationMode = 0x03
	OpModeException OperationMode = 0xFE
	OpModeInvalid   OperationMode = 0xFF
)

// DCState represents DC-DC converter state.
type DCState byte

const (
	DCStateOn        DCState = 0x01
	DCStateOff       DCState = 0x02
	DCStateException DCState = 0xFE
	DCStateInvalid   DCState = 0xFF
)

// GearPositionEnum represents gear position.
type GearPositionEnum byte

const (
	GearP     GearPositionEnum = 0x01
	GearR     GearPositionEnum = 0x02
	GearN     GearPositionEnum = 0x03
	GearD     GearPositionEnum = 0x04
	GearOther GearPositionEnum = 0x05
)

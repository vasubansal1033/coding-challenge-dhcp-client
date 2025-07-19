package main

// Field name constants for error context
const (
	FieldOpCode                = "OpCode"
	FieldHardwareType          = "HardwareType"
	FieldHardwareAddressLength = "HardwareAddressLength"
	FieldHops                  = "Hops"
	FieldTransactionID         = "TransactionID"
	FieldSeconds               = "Seconds"
	FieldFlags                 = "Flags"
	FieldClientIP              = "ClientIP"
	FieldYourIP                = "YourIP"
	FieldNextServerIP          = "NextServerIP"
	FieldRelayAgentIP          = "RelayAgentIP"
	FieldClientHardwareAddress = "ClientHardwareAddress"
	FieldServerHostName        = "ServerHostName"
	FieldBootFileName          = "BootFileName"
	FieldMagicCookie           = "MagicCookie"
)

// Size constants for DHCP fields
const (
	SizeClientHardwareAddress    = 16
	SizeServerHostName           = 64
	SizeBootFileName             = 128
	SizeMinimumDHCPMessageLength = 240
)

// DHCP option constants
const (
	OptionDHCPMessageType      = 53
	OptionClientIdentifier     = 61
	OptionParameterRequestList = 55
	OptionEnd                  = 255
	OptionPad                  = 0
)

// DHCP message type constants
const (
	DHCPDiscover = 1
	DHCPOffer    = 2
	DHCPRequest  = 3
	DHCPAck      = 5
)

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

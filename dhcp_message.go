package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

type DHCPMessage struct {
	OpCode                uint8           `json:"op"`      // 1 byte
	HardwareType          uint8           `json:"htype"`   // 1 byte
	HardwareAddressLength uint8           `json:"hlen"`    // 1 byte
	Hops                  uint8           `json:"hops"`    // 1 byte
	TransactionID         uint32          `json:"xid"`     // 4 bytes
	Seconds               uint16          `json:"secs"`    // 2 bytes
	Flags                 uint16          `json:"flags"`   // 2 bytes
	ClientIP              uint32          `json:"ciaddr"`  // 4 bytes
	YourIP                uint32          `json:"yiaddr"`  // 4 bytes
	NextServerIP          uint32          `json:"siaddr"`  // 4 bytes
	RelayAgentIP          uint32          `json:"giaddr"`  // 4 bytes
	ClientHardwareAddress []byte          `json:"chaddr"`  // 16 bytes
	ServerHostName        []byte          `json:"sname"`   // 64 bytes
	BootFileName          []byte          `json:"file"`    // 128 bytes
	MagicCookie           uint32          `json:"magic"`   // 4 bytes
	Options               map[byte][]byte `json:"options"` // DHCP options
}

// Serialize serializes the DHCPMessage into a byte slice with error handling.
func (m *DHCPMessage) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 300))
	write := func(data interface{}, field string) error {
		if err := binary.Write(buf, binary.BigEndian, data); err != nil {
			return fmt.Errorf("failed to write %s: %w", field, err)
		}
		return nil
	}

	if err := write(m.OpCode, FieldOpCode); err != nil {
		return nil, err
	}

	if err := write(m.HardwareType, FieldHardwareType); err != nil {
		return nil, err
	}

	if err := write(m.HardwareAddressLength, FieldHardwareAddressLength); err != nil {
		return nil, err
	}

	if err := write(m.Hops, FieldHops); err != nil {
		return nil, err
	}

	if err := write(m.TransactionID, FieldTransactionID); err != nil {
		return nil, err
	}

	if err := write(m.Seconds, FieldSeconds); err != nil {
		return nil, err
	}

	if err := write(m.Flags, FieldFlags); err != nil {
		return nil, err
	}

	if err := write(m.ClientIP, FieldClientIP); err != nil {
		return nil, err
	}

	if err := write(m.YourIP, FieldYourIP); err != nil {
		return nil, err
	}

	if err := write(m.NextServerIP, FieldNextServerIP); err != nil {
		return nil, err
	}

	if err := write(m.RelayAgentIP, FieldRelayAgentIP); err != nil {
		return nil, err
	}

	if len(m.ClientHardwareAddress) != SizeClientHardwareAddress {
		return nil, fmt.Errorf("%s must be %d bytes, got %d", FieldClientHardwareAddress, SizeClientHardwareAddress, len(m.ClientHardwareAddress))
	}

	if _, err := buf.Write(m.ClientHardwareAddress); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", FieldClientHardwareAddress, err)
	}

	if len(m.ServerHostName) != SizeServerHostName {
		return nil, fmt.Errorf("%s must be %d bytes, got %d", FieldServerHostName, SizeServerHostName, len(m.ServerHostName))
	}

	if _, err := buf.Write(m.ServerHostName); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", FieldServerHostName, err)
	}

	if len(m.BootFileName) != SizeBootFileName {
		return nil, fmt.Errorf("%s must be %d bytes, got %d", FieldBootFileName, SizeBootFileName, len(m.BootFileName))
	}

	if _, err := buf.Write(m.BootFileName); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", FieldBootFileName, err)
	}

	if err := write(m.MagicCookie, FieldMagicCookie); err != nil {
		return nil, err
	}

	// Add DHCP options
	if m.Options != nil {
		for code, value := range m.Options {
			if code == OptionPad || code == OptionEnd {
				continue // Skip pad and end options
			}

			if len(value) > 255 {
				return nil, fmt.Errorf("option %d too long: %d bytes", code, len(value))
			}

			buf.WriteByte(code)
			buf.WriteByte(byte(len(value)))
			buf.Write(value)
		}
	}

	// Add end option
	buf.WriteByte(OptionEnd)

	return buf.Bytes(), nil
}

// Deserialize parses a DHCPMessage from a byte slice with error handling.
func Deserialize(data []byte) (*DHCPMessage, error) {
	if len(data) < SizeMinimumDHCPMessageLength {
		return nil, fmt.Errorf("data too short for DHCP message: got %d bytes, want at least %d", len(data), SizeMinimumDHCPMessageLength)
	}

	buf := bytes.NewBuffer(data)
	m := &DHCPMessage{}
	read := func(data interface{}, field string) error {
		if err := binary.Read(buf, binary.BigEndian, data); err != nil {
			return fmt.Errorf("failed to read %s: %w", field, err)
		}
		return nil
	}

	if err := read(&m.OpCode, FieldOpCode); err != nil {
		return nil, err
	}

	if err := read(&m.HardwareType, FieldHardwareType); err != nil {
		return nil, err
	}

	if err := read(&m.HardwareAddressLength, FieldHardwareAddressLength); err != nil {
		return nil, err
	}

	if err := read(&m.Hops, FieldHops); err != nil {
		return nil, err
	}

	if err := read(&m.TransactionID, FieldTransactionID); err != nil {
		return nil, err
	}

	if err := read(&m.Seconds, FieldSeconds); err != nil {
		return nil, err
	}

	if err := read(&m.Flags, FieldFlags); err != nil {
		return nil, err
	}

	if err := read(&m.ClientIP, FieldClientIP); err != nil {
		return nil, err
	}

	if err := read(&m.YourIP, FieldYourIP); err != nil {
		return nil, err
	}

	if err := read(&m.NextServerIP, FieldNextServerIP); err != nil {
		return nil, err
	}

	if err := read(&m.RelayAgentIP, FieldRelayAgentIP); err != nil {
		return nil, err
	}

	m.ClientHardwareAddress = make([]byte, SizeClientHardwareAddress)
	if n, err := buf.Read(m.ClientHardwareAddress); err != nil || n != SizeClientHardwareAddress {
		return nil, fmt.Errorf("failed to read %s: %w (read %d bytes)", FieldClientHardwareAddress, err, n)
	}

	m.ServerHostName = make([]byte, SizeServerHostName)
	if n, err := buf.Read(m.ServerHostName); err != nil || n != SizeServerHostName {
		return nil, fmt.Errorf("failed to read %s: %w (read %d bytes)", FieldServerHostName, err, n)
	}

	m.BootFileName = make([]byte, SizeBootFileName)
	if n, err := buf.Read(m.BootFileName); err != nil || n != SizeBootFileName {
		return nil, fmt.Errorf("failed to read %s: %w (read %d bytes)", FieldBootFileName, err, n)
	}

	if err := read(&m.MagicCookie, FieldMagicCookie); err != nil {
		return nil, err
	}

	return m, nil
}

// String returns a human-readable representation of the DHCP message
func (m *DHCPMessage) String() string {
	var result strings.Builder

	result.WriteString("DHCP Message:\n")
	result.WriteString(fmt.Sprintf("  Op Code: %d (%s)\n", m.OpCode, m.opCodeString()))
	result.WriteString(fmt.Sprintf("  Hardware Type: %d (%s)\n", m.HardwareType, m.hardwareTypeString()))
	result.WriteString(fmt.Sprintf("  Hardware Address Length: %d\n", m.HardwareAddressLength))
	result.WriteString(fmt.Sprintf("  Hops: %d\n", m.Hops))
	result.WriteString(fmt.Sprintf("  Transaction ID: 0x%08x\n", m.TransactionID))
	result.WriteString(fmt.Sprintf("  Seconds: %d\n", m.Seconds))
	result.WriteString(fmt.Sprintf("  Flags: 0x%04x\n", m.Flags))
	result.WriteString(fmt.Sprintf("  Client IP: %s\n", m.ipToString(m.ClientIP)))
	result.WriteString(fmt.Sprintf("  Your IP: %s\n", m.ipToString(m.YourIP)))
	result.WriteString(fmt.Sprintf("  Next Server IP: %s\n", m.ipToString(m.NextServerIP)))
	result.WriteString(fmt.Sprintf("  Relay Agent IP: %s\n", m.ipToString(m.RelayAgentIP)))
	result.WriteString(fmt.Sprintf("  Client Hardware Address: %s\n", m.macToString(m.ClientHardwareAddress)))
	result.WriteString(fmt.Sprintf("  Server Host Name: '%s'\n", m.bytesToString(m.ServerHostName)))
	result.WriteString(fmt.Sprintf("  Boot File Name: '%s'\n", m.bytesToString(m.BootFileName)))
	result.WriteString(fmt.Sprintf("  Magic Cookie: 0x%08x\n", m.MagicCookie))

	if len(m.Options) > 0 {
		result.WriteString("  DHCP Options:\n")
		for code, value := range m.Options {
			result.WriteString(fmt.Sprintf("    %s: %s\n", m.optionCodeString(code), m.optionValueString(code, value)))
		}
	}

	return result.String()
}

// Helper methods for formatting
func (m *DHCPMessage) opCodeString() string {
	switch m.OpCode {
	case 1:
		return "Boot Request"
	case 2:
		return "Boot Reply"
	default:
		return "Unknown"
	}
}

func (m *DHCPMessage) hardwareTypeString() string {
	switch m.HardwareType {
	case 1:
		return "Ethernet"
	case 6:
		return "IEEE 802"
	case 15:
		return "Frame Relay"
	case 16:
		return "Asynchronous Transfer Mode (ATM)"
	default:
		return "Unknown"
	}
}

func (m *DHCPMessage) ipToString(ip uint32) string {
	if ip == 0 {
		return "0.0.0.0"
	}
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24),
		byte(ip>>16),
		byte(ip>>8),
		byte(ip))
}

func (m *DHCPMessage) macToString(mac []byte) string {
	if len(mac) < 6 {
		return "Invalid MAC"
	}
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

func (m *DHCPMessage) bytesToString(data []byte) string {
	// Find null terminator
	nullIndex := bytes.IndexByte(data, 0)
	if nullIndex == -1 {
		nullIndex = len(data)
	}
	return string(data[:nullIndex])
}

func (m *DHCPMessage) optionCodeString(code byte) string {
	switch code {
	case 1:
		return "Subnet Mask"
	case 3:
		return "Router"
	case 6:
		return "DNS Server"
	case 15:
		return "Domain Name"
	case 31:
		return "Perform Router Discovery"
	case 33:
		return "Static Route"
	case 43:
		return "Vendor-Specific Information"
	case 44:
		return "NetBIOS over TCP/IP Name Server"
	case 46:
		return "NetBIOS over TCP/IP Node Type"
	case 47:
		return "NetBIOS over TCP/IP Scope"
	case 50:
		return "Requested IP Address"
	case 51:
		return "IP Address Lease Time"
	case 53:
		return "DHCP Message Type"
	case 54:
		return "Server Identifier"
	case 55:
		return "Parameter Request List"
	case 57:
		return "Maximum DHCP Message Size"
	case 58:
		return "Renewal (T1) Time Value"
	case 59:
		return "Rebinding (T2) Time Value"
	case 60:
		return "Vendor Class Identifier"
	case 61:
		return "Client-identifier"
	case 66:
		return "TFTP Server Name"
	case 67:
		return "Bootfile Name"
	case 119:
		return "Domain Search"
	case 121:
		return "Classless Static Route"
	case 249:
		return "Private/Classless Static Route (Microsoft)"
	case 252:
		return "Private/Proxy autodiscovery"
	default:
		return fmt.Sprintf("Option %d", code)
	}
}

func (m *DHCPMessage) optionValueString(code byte, value []byte) string {
	switch code {
	case 53: // DHCP Message Type
		if len(value) > 0 {
			switch value[0] {
			case 1:
				return "DHCPDISCOVER"
			case 2:
				return "DHCPOFFER"
			case 3:
				return "DHCPREQUEST"
			case 4:
				return "DHCPDECLINE"
			case 5:
				return "DHCPACK"
			case 6:
				return "DHCPNAK"
			case 7:
				return "DHCPRELEASE"
			case 8:
				return "DHCPINFORM"
			default:
				return fmt.Sprintf("Unknown (%d)", value[0])
			}
		}
		return "Empty"
	case 1, 3, 6, 54: // IP addresses
		if len(value) == 4 {
			ip := binary.BigEndian.Uint32(value)
			return m.ipToString(ip)
		}
		return fmt.Sprintf("%v", value)
	case 51, 58, 59: // Time values
		if len(value) == 4 {
			time := binary.BigEndian.Uint32(value)
			return fmt.Sprintf("%d seconds", time)
		}
		return fmt.Sprintf("%v", value)
	case 61: // Client-identifier
		if len(value) > 1 {
			hwType := value[0]
			hwAddr := value[1:]
			return fmt.Sprintf("Type %d: %s", hwType, m.macToString(hwAddr))
		}
		return fmt.Sprintf("%v", value)
	case 55: // Parameter Request List
		var params []string
		for _, param := range value {
			params = append(params, m.optionCodeString(param))
		}
		return strings.Join(params, ", ")
	default:
		if len(value) == 0 {
			return "Empty"
		}
		// Try to convert to string if it looks like text
		if isPrintable(value) {
			return fmt.Sprintf("'%s'", string(value))
		}
		return fmt.Sprintf("%v", value)
	}
}

func isPrintable(data []byte) bool {
	for _, b := range data {
		if b < 32 || b > 126 {
			return false
		}
	}
	return true
}

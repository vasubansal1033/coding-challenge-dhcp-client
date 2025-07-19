package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type DHCPMessage struct {
	OpCode                uint8  `json:"op"`     // 1 byte
	HardwareType          uint8  `json:"htype"`  // 1 byte
	HardwareAddressLength uint8  `json:"hlen"`   // 1 byte
	Hops                  uint8  `json:"hops"`   // 1 byte
	TransactionID         uint32 `json:"xid"`    // 4 bytes
	Seconds               uint16 `json:"secs"`   // 2 bytes
	Flags                 uint16 `json:"flags"`  // 2 bytes
	ClientIP              uint32 `json:"ciaddr"` // 4 bytes
	YourIP                uint32 `json:"yiaddr"` // 4 bytes
	NextServerIP          uint32 `json:"siaddr"` // 4 bytes
	RelayAgentIP          uint32 `json:"giaddr"` // 4 bytes
	ClientHardwareAddress []byte `json:"chaddr"` // 16 bytes
	ServerHostName        []byte `json:"sname"`  // 64 bytes
	BootFileName          []byte `json:"file"`   // 128 bytes
	MagicCookie           uint32 `json:"magic"`  // 4 bytes
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

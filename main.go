package main

import (
	"fmt"
)

func main() {
	fmt.Println("DHCP client starting...")

	// Example usage for demonstration
	msg := &DHCPMessage{
		OpCode:                1,
		HardwareType:          1,
		HardwareAddressLength: 6,
		Hops:                  0,
		TransactionID:         0x12345678,
		Seconds:               0,
		Flags:                 0,
		ClientIP:              0,
		YourIP:                0,
		NextServerIP:          0,
		RelayAgentIP:          0,
		ClientHardwareAddress: make([]byte, SizeClientHardwareAddress),
		ServerHostName:        make([]byte, SizeServerHostName),
		BootFileName:          make([]byte, SizeBootFileName),
		MagicCookie:           0x63825363,
	}

	serialized, err := msg.Serialize()
	if err != nil {
		fmt.Printf("Error serializing DHCPMessage: %v\n", err)
		return
	}

	fmt.Printf("Serialized DHCPMessage (%d bytes)\n", len(serialized))
	deserialized, err := Deserialize(serialized)
	if err != nil {
		fmt.Printf("Error deserializing DHCPMessage: %v\n", err)
		return
	}

	fmt.Printf("Deserialized DHCPMessage: %+v\n", deserialized)
}

package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("DHCP client starting...")

	// Create a MAC address for testing
	macAddr := []byte{0x62, 0xf9, 0xb8, 0xfc, 0x9d, 0xff}

	// Create and start the DHCP client
	client := NewDHCPClient(macAddr)

	if err := client.Start(); err != nil {
		log.Fatalf("DHCP process failed: %v", err)
	}

	fmt.Println("DHCP process completed successfully!")
}

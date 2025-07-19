package main

import (
	"fmt"
	"net"
	"time"
)

func createUDPSendSocket() (*net.UDPConn, error) {
	// laddr, err := net.ResolveUDPAddr("udp", ":68")
	raddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:67")
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func createUDPReceiveSocket() (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", ":68")
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func main() {
	fmt.Println("DHCP client starting...")

	// Get a MAC address
	macAddr := []byte{0x62, 0xf9, 0xb8, 0xfc, 0x9d, 0xff}

	// construct DHCPDiscover message
	dhcpDiscover := &DHCPMessage{
		OpCode:                1, // Boot request
		HardwareType:          1, // Ethernet
		HardwareAddressLength: 6, // MAC address length
		Hops:                  0,
		TransactionID:         0x12345678,
		Seconds:               0,
		Flags:                 0,
		ClientHardwareAddress: make([]byte, 16), // Will be filled with MAC
		ServerHostName:        make([]byte, 64),
		BootFileName:          make([]byte, 128),
		MagicCookie:           0x63825363,
		Options:               make(map[byte][]byte),
	}

	// Copy MAC address to ClientHardwareAddress (first 6 bytes)
	copy(dhcpDiscover.ClientHardwareAddress, macAddr)

	// Add required DHCP options
	dhcpDiscover.Options[OptionDHCPMessageType] = []byte{DHCPDiscover}
	dhcpDiscover.Options[OptionClientIdentifier] = append([]byte{1}, macAddr...)                                       // Type 1 (Ethernet) + MAC
	dhcpDiscover.Options[OptionParameterRequestList] = []byte{1, 3, 6, 15, 31, 33, 43, 44, 46, 47, 119, 121, 249, 252} // Common DHCP options

	receiveSocket, err := createUDPReceiveSocket()
	if err != nil {
		fmt.Printf("Error creating receive socket: %v\n", err)
		return
	}

	defer receiveSocket.Close()

	receiveSocket.SetReadDeadline(time.Now().Add(10 * time.Second))

	sendSocket, err := createUDPSendSocket()
	if err != nil {
		fmt.Printf("Error creating send socket: %v\n", err)
		return
	}

	defer sendSocket.Close()

	msg, err := dhcpDiscover.Serialize()
	if err != nil {
		fmt.Printf("Error serializing DHCPDiscover message: %v\n", err)
		return
	}

	fmt.Printf("Sending DHCPDISCOVER (%d bytes)...\n", len(msg))
	_, err = sendSocket.Write(msg)
	if err != nil {
		fmt.Printf("Error sending DHCPDiscover message: %v\n", err)
		return
	}

	fmt.Println("DHCPDiscover message sent, waiting for response...")

	buf := make([]byte, 1024)
	n, addr, err := receiveSocket.ReadFromUDP(buf)
	if err != nil {
		fmt.Printf("Error reading from receive socket: %v\n", err)
		return
	}

	fmt.Printf("Received %d bytes from %s\n", n, addr.String())

	dhcpMessage, err := Deserialize(buf[:n])
	if err != nil {
		fmt.Printf("Error deserializing DHCP message: %v\n", err)
		return
	}

	fmt.Printf("Received DHCP message: %+v\n", dhcpMessage)

	// Check if it's a DHCPOFFER
	if msgType, exists := dhcpMessage.Options[OptionDHCPMessageType]; exists && len(msgType) > 0 {
		switch msgType[0] {
		case DHCPOffer:
			fmt.Println("Received DHCPOFFER!")
		case DHCPAck:
			fmt.Println("Received DHCPACK!")
		default:
			fmt.Printf("Received DHCP message type: %d\n", msgType[0])
		}
	}
}

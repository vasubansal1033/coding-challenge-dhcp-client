package main

import (
	"fmt"
	"net"
	"time"
)

// DHCPClient represents a DHCP client
type DHCPClient struct {
	macAddr       []byte
	transactionID uint32
	sendSocket    *net.UDPConn
	receiveSocket *net.UDPConn
}

// NewDHCPClient creates a new DHCP client
func NewDHCPClient(macAddr []byte) *DHCPClient {
	return &DHCPClient{
		macAddr:       macAddr,
		transactionID: 0x12345678, // You might want to generate this randomly
	}
}

// Start initiates the DHCP process
func (c *DHCPClient) Start() error {
	// Create sockets
	if err := c.createSockets(); err != nil {
		return fmt.Errorf("failed to create sockets: %w", err)
	}
	defer c.cleanup()

	fmt.Println("Starting DHCP process...")

	// Step 1: Send DHCPDISCOVER
	discoverMsg := c.createDHCPDiscover()
	fmt.Println("Sending DHCPDISCOVER...")
	if err := c.sendMessage(discoverMsg); err != nil {
		return fmt.Errorf("failed to send DHCPDISCOVER: %w", err)
	}

	// Step 2: Wait for DHCPOFFER
	fmt.Println("Waiting for DHCPOFFER...")
	offerMsg, err := c.waitForMessage(DHCPOffer, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to receive DHCPOFFER: %w", err)
	}

	fmt.Printf("Received DHCPOFFER:\n%s", offerMsg.String())

	// Step 3: Send DHCPREQUEST
	requestMsg := c.createDHCPRequest(offerMsg)
	fmt.Println("Sending DHCPREQUEST...")
	if err := c.sendMessage(requestMsg); err != nil {
		return fmt.Errorf("failed to send DHCPREQUEST: %w", err)
	}

	// Step 4: Wait for DHCPACK or DHCPNAK
	fmt.Println("Waiting for DHCPACK/DHCPNAK...")
	responseMsg, err := c.waitForMessage(DHCPAck, 10*time.Second)
	if err != nil {
		// Try waiting for DHCPNAK
		responseMsg, err = c.waitForMessage(DHCPNak, 5*time.Second)
		if err != nil {
			return fmt.Errorf("failed to receive DHCPACK/DHCPNAK: %w", err)
		}
	}

	fmt.Printf("Received response:\n%s", responseMsg.String())

	// Check if we got ACK or NAK
	if msgType, exists := responseMsg.Options[OptionDHCPMessageType]; exists && len(msgType) > 0 {
		switch msgType[0] {
		case DHCPAck:
			fmt.Println("DHCPACK received! IP address successfully assigned.")
			fmt.Printf("Assigned IP: %s\n", responseMsg.ipToString(responseMsg.YourIP))
			return nil
		case DHCPNak:
			fmt.Println("DHCPNAK received! IP address assignment failed.")
			return fmt.Errorf("DHCP server rejected the request")
		default:
			return fmt.Errorf("unexpected message type: %d", msgType[0])
		}
	}

	return fmt.Errorf("no message type found in response")
}

// createSockets creates the UDP sockets for sending and receiving
func (c *DHCPClient) createSockets() error {
	var err error

	// Create send socket
	c.sendSocket, err = createUDPSendSocket()
	if err != nil {
		return fmt.Errorf("failed to create send socket: %w", err)
	}

	// Create receive socket
	c.receiveSocket, err = createUDPReceiveSocket()
	if err != nil {
		c.sendSocket.Close()
		return fmt.Errorf("failed to create receive socket: %w", err)
	}

	// Set read timeout
	c.receiveSocket.SetReadDeadline(time.Now().Add(10 * time.Second))

	return nil
}

// cleanup closes the sockets
func (c *DHCPClient) cleanup() {
	if c.sendSocket != nil {
		c.sendSocket.Close()
	}
	if c.receiveSocket != nil {
		c.receiveSocket.Close()
	}
}

// createDHCPDiscover creates a DHCPDISCOVER message
func (c *DHCPClient) createDHCPDiscover() *DHCPMessage {
	msg := &DHCPMessage{
		OpCode:                1, // Boot request
		HardwareType:          1, // Ethernet
		HardwareAddressLength: 6, // MAC address length
		Hops:                  0,
		TransactionID:         c.transactionID,
		Seconds:               0,
		Flags:                 0,
		ClientHardwareAddress: make([]byte, 16),
		ServerHostName:        make([]byte, 64),
		BootFileName:          make([]byte, 128),
		MagicCookie:           0x63825363,
		Options:               make(map[byte][]byte),
	}

	// Copy MAC address to ClientHardwareAddress (first 6 bytes)
	copy(msg.ClientHardwareAddress, c.macAddr)

	// Add required DHCP options
	msg.Options[OptionDHCPMessageType] = []byte{DHCPDiscover}
	msg.Options[OptionClientIdentifier] = append([]byte{1}, c.macAddr...) // Type 1 (Ethernet) + MAC
	msg.Options[OptionParameterRequestList] = []byte{1, 3, 6, 15, 31, 33, 43, 44, 46, 47, 119, 121, 249, 252}

	return msg
}

// createDHCPRequest creates a DHCPREQUEST message based on the received offer
func (c *DHCPClient) createDHCPRequest(offerMsg *DHCPMessage) *DHCPMessage {
	msg := &DHCPMessage{
		OpCode:                1, // Boot request
		HardwareType:          1, // Ethernet
		HardwareAddressLength: 6, // MAC address length
		Hops:                  0,
		TransactionID:         c.transactionID, // Same transaction ID
		Seconds:               0,
		Flags:                 0,
		ClientHardwareAddress: make([]byte, 16),
		ServerHostName:        make([]byte, 64),
		BootFileName:          make([]byte, 128),
		MagicCookie:           0x63825363,
		Options:               make(map[byte][]byte),
	}

	// Copy MAC address to ClientHardwareAddress (first 6 bytes)
	copy(msg.ClientHardwareAddress, c.macAddr)

	// Add required DHCP options
	msg.Options[OptionDHCPMessageType] = []byte{DHCPRequest}
	msg.Options[OptionClientIdentifier] = append([]byte{1}, c.macAddr...) // Type 1 (Ethernet) + MAC

	// Request the offered IP address
	if offeredIP, exists := offerMsg.Options[OptionRequestedIPAddress]; exists {
		msg.Options[OptionRequestedIPAddress] = offeredIP
	} else {
		// If no requested IP in offer, use the YourIP field
		msg.Options[OptionRequestedIPAddress] = []byte{
			byte(offerMsg.YourIP >> 24),
			byte(offerMsg.YourIP >> 16),
			byte(offerMsg.YourIP >> 8),
			byte(offerMsg.YourIP),
		}
	}

	// Identify the server that made the offer
	if serverID, exists := offerMsg.Options[OptionServerIdentifier]; exists {
		msg.Options[OptionServerIdentifier] = serverID
	} else {
		// If no server identifier in offer, use the NextServerIP field
		msg.Options[OptionServerIdentifier] = []byte{
			byte(offerMsg.NextServerIP >> 24),
			byte(offerMsg.NextServerIP >> 16),
			byte(offerMsg.NextServerIP >> 8),
			byte(offerMsg.NextServerIP),
		}
	}

	// Request the same parameters as in DISCOVER
	msg.Options[OptionParameterRequestList] = []byte{1, 3, 6, 15, 31, 33, 43, 44, 46, 47, 119, 121, 249, 252}

	return msg
}

// sendMessage sends a DHCP message
func (c *DHCPClient) sendMessage(msg *DHCPMessage) error {
	data, err := msg.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	_, err = c.sendSocket.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// waitForMessage waits for a specific DHCP message type
func (c *DHCPClient) waitForMessage(expectedType byte, timeout time.Duration) (*DHCPMessage, error) {
	c.receiveSocket.SetReadDeadline(time.Now().Add(timeout))

	buf := make([]byte, 1024)
	for {
		n, addr, err := c.receiveSocket.ReadFromUDP(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read from socket: %w", err)
		}

		fmt.Printf("Received %d bytes from %s\n", n, addr.String())

		msg, err := Deserialize(buf[:n])
		if err != nil {
			fmt.Printf("Failed to deserialize message: %v\n", err)
			continue
		}

		fmt.Printf("Received message:\n%s", msg.String())

		// Check if this is the expected message type
		if msgType, exists := msg.Options[OptionDHCPMessageType]; exists && len(msgType) > 0 {
			if msgType[0] == expectedType {
				return msg, nil
			}

			// If not the expected type, continue waiting
			fmt.Printf("Received message type %d, waiting for %d\n",
				msg.Options[OptionDHCPMessageType][0], expectedType)
		}

	}
}

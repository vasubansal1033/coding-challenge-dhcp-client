# DHCP Client

A Go implementation of a DHCP (Dynamic Host Configuration Protocol) client that can request and obtain IP addresses from DHCP servers.

## Overview

This project implements a complete DHCP client following RFC 2131. It performs the full DHCP exchange:
1. **DHCPDISCOVER** - Client broadcasts to find available DHCP servers
2. **DHCPOFFER** - Server responds with an IP address offer
3. **DHCPREQUEST** - Client requests the offered IP address
4. **DHCPACK/NAK** - Server confirms or rejects the request

## Features

- ✅ Complete DHCP message serialization/deserialization
- ✅ Full DHCP exchange (DISCOVER → OFFER → REQUEST → ACK/NAK)
- ✅ Human-readable message formatting
- ✅ Proper error handling and logging
- ✅ Support for DHCP options
- ✅ MAC address and IP address formatting
- ✅ UDP socket management

## Project Structure

```
├── main.go              # Main application entry point
├── dhcp_client.go       # DHCP client logic and exchange handling
├── dhcp_message.go      # DHCP message struct and serialization
├── dhcp_sockets.go      # UDP socket creation and management
├── constants.go         # DHCP constants and option codes
└── README.md           # This file
```

## Usage

### Prerequisites

- Go 1.16 or later
- Network access (for DHCP communication)

### Building

```bash
go build -o dhcpclient
```

### Running

```bash
# Run the DHCP client
./dhcpclient
```

**Note:** The client uses port 68 for receiving, which may conflict with your system's DHCP client. For testing, consider:
- Using a virtual machine
- Using a different port (modify `dhcp_sockets.go`)
- Temporarily stopping your system's DHCP client

### Example Output

```
DHCP client starting...
Starting DHCP process...
Sending DHCPDISCOVER...
Waiting for DHCPOFFER...
Received 300 bytes from 192.168.1.1:67
Received DHCPOFFER:
DHCP Message:
  Op Code: 2 (Boot Reply)
  Hardware Type: 1 (Ethernet)
  Hardware Address Length: 6
  Hops: 0
  Transaction ID: 0x12345678
  Seconds: 0
  Flags: 0x0000
  Client IP: 0.0.0.0
  Your IP: 192.168.1.100
  Next Server IP: 0.0.0.0
  Relay Agent IP: 0.0.0.0
  Client Hardware Address: 62:f9:b8:fc:9d:ff
  Server Host Name: ''
  Boot File Name: ''
  Magic Cookie: 0x63825363
  DHCP Options:
    DHCP Message Type: DHCPOFFER
    Server Identifier: 192.168.1.1
    IP Address Lease Time: 86400 seconds
    Subnet Mask: 255.255.255.0
    Router: 192.168.1.1
    DNS Server: 8.8.8.8
Sending DHCPREQUEST...
Waiting for DHCPACK/DHCPNAK...
Received 300 bytes from 192.168.1.1:67
Received response:
DHCP Message:
  Op Code: 2 (Boot Reply)
  Hardware Type: 1 (Ethernet)
  Hardware Address Length: 6
  Hops: 0
  Transaction ID: 0x12345678
  Seconds: 0
  Flags: 0x0000
  Client IP: 0.0.0.0
  Your IP: 192.168.1.100
  Next Server IP: 0.0.0.0
  Relay Agent IP: 0.0.0.0
  Client Hardware Address: 62:f9:b8:fc:9d:ff
  Server Host Name: ''
  Boot File Name: ''
  Magic Cookie: 0x63825363
  DHCP Options:
    DHCP Message Type: DHCPACK
    Server Identifier: 192.168.1.1
    IP Address Lease Time: 86400 seconds
    Subnet Mask: 255.255.255.0
    Router: 192.168.1.1
    DNS Server: 8.8.8.8
✅ DHCPACK received! IP address successfully assigned.
Assigned IP: 192.168.1.100
DHCP process completed successfully!
```

## Technical Details

### DHCP Message Format

The implementation follows RFC 2131 for DHCP message structure:

- **Fixed Header** (236 bytes): Op code, hardware info, transaction ID, IP addresses, etc.
- **DHCP Options** (variable): Message type, client identifier, requested parameters, etc.

### Key Components

1. **DHCPMessage**: Core struct representing a DHCP packet
2. **DHCPClient**: Manages the complete DHCP exchange process
3. **Socket Management**: Handles UDP communication on ports 67/68
4. **Message Serialization**: Converts structs to/from byte arrays
5. **Human-Readable Output**: Formats messages for debugging

### Supported DHCP Options

- Message Type (53)
- Client Identifier (61)
- Parameter Request List (55)
- Requested IP Address (50)
- Server Identifier (54)
- And many more...

## Development

### Adding New Features

1. **New DHCP Options**: Add constants to `constants.go` and update `optionCodeString()` in `dhcp_message.go`
2. **Message Types**: Extend the DHCP exchange in `dhcp_client.go`
3. **Error Handling**: Add proper error handling and logging

### Testing

```bash
# Build and run
go build && ./dhcpclient

# Run with verbose output (modify main.go for more logging)
go run .
```

## Limitations

- Currently uses a hardcoded MAC address for testing
- No IP address assignment to network interface (requires root privileges)
- Limited to basic DHCP options
- No DHCP renewal/release functionality

## Future Enhancements

- [ ] Random transaction ID generation
- [ ] DHCP renewal and release
- [ ] Network interface configuration
- [ ] Support for more DHCP options
- [ ] DHCP server implementation
- [ ] Configuration file support
- [ ] Multiple network interface support

## References

- [RFC 2131 - DHCP](https://datatracker.ietf.org/doc/html/rfc2131)
- [RFC 2132 - DHCP Options](https://datatracker.ietf.org/doc/html/rfc2132)
- [Coding Challenge #94 - DHCP Client](https://codingchallenges.substack.com/p/coding-challenge-94-dhcp-client)

## License

This project is for educational purposes. Feel free to use and modify as needed. 
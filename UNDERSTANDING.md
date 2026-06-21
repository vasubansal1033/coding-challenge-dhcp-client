# DHCP Client - My Notes

I built this DHCP client in Go as part of the [Coding Challenge #94](https://codingchallenges.substack.com/p/coding-challenge-94-dhcp-client). This is my own step by step writeup of what I learned while implementing it, not a polished docs page.

## What DHCP Actually Does

Before writing any code, I had to understand the protocol. DHCP is how a machine gets an IP address without manual configuration. The client does not know its IP yet, so it broadcasts on the network and a DHCP server replies with an offer.

The full exchange looks like this:

1. **DHCPDISCOVER** - client says "I need an IP"
2. **DHCPOFFER** - server says "here is one you can have"
3. **DHCPREQUEST** - client says "I want that one"
4. **DHCPACK** or **DHCPNAK** - server confirms or rejects

There is also **DHCPRELEASE** for giving the IP back when you are done. That is the last piece I still need to wire up properly.

## How I Built It (Step by Step)

### Step 1: Understand the Packet Format

DHCP messages are not JSON. They are a fixed binary layout defined in RFC 2131. I started with a `DHCPMessage` struct in `dhcp_message.go` that mirrors the packet:

- A fixed header (op code, hardware type, transaction ID, IP fields, MAC address, etc.)
- A magic cookie (`0x63825363`) that marks where DHCP options begin
- A variable list of options (message type, subnet mask, DNS, lease time, and so on)

The first real work was `Serialize()` and `Deserialize()`. Getting the byte order right mattered a lot. Everything is big endian. I used Go's `encoding/binary` package for that.

I also added size checks so bad data fails early instead of producing a corrupt packet. For example, `ClientHardwareAddress` must always be 16 bytes even though only the first 6 hold the MAC.

### Step 2: Open the Right Sockets

DHCP uses UDP. The client sends to port **67** (server) and listens on port **68** (client).

I put socket setup in `dhcp_sockets.go`:

- **Send socket**: dials `255.255.255.255:67` so the DISCOVER goes out as a broadcast
- **Receive socket**: listens on `:68` for replies

One thing I learned the hard way: port 68 is often already taken by macOS's own DHCP client (`ipconfig` or similar). If bind fails, or if you never get a reply, check `lsof -i :68` first.

### Step 3: Send a DHCPDISCOVER

With serialization and sockets in place, I built my first DISCOVER message in `main.go`:

- Op code `1` (boot request)
- Hardware type `1` (Ethernet), length `6`
- A fixed transaction ID (`0x12345678` for now, should randomize later)
- My test MAC address in `chaddr`
- Options: message type DISCOVER, client identifier (type 1 + MAC), and a parameter request list asking for subnet mask, router, DNS, etc.

Then serialize, send on the UDP socket, and wait for something to come back on port 68.

### Step 4: Read Replies in a Human Friendly Way

Raw bytes are useless for debugging. I added a `String()` method on `DHCPMessage` that prints:

- Header fields with readable labels
- IP addresses formatted as `x.x.x.x`
- MAC as `aa:bb:cc:dd:ee:ff`
- DHCP options decoded by name (e.g. "DHCP Message Type: DHCPOFFER" instead of `53: [2]`)

This made a huge difference when I was staring at server responses trying to figure out what went wrong.

### Step 5: Refactor and Complete the Exchange

Once DISCOVER/OFFER worked, I moved the logic into a proper `DHCPClient` struct in `dhcp_client.go` and pulled socket code into its own file.

The `Start()` method runs the full state machine:

1. Create sockets (with deferred cleanup)
2. Send **DHCPDISCOVER**
3. Wait for **DHCPOFFER** (with timeout, skip unrelated message types)
4. Build **DHCPREQUEST** from the offer:
   - Same transaction ID
   - Option 50 (requested IP) from the offer's `yiaddr` or option 50
   - Option 54 (server identifier) so the server knows which offer we accepted
5. Wait for **DHCPACK** or **DHCPNAK**
6. Print the assigned IP on success

`waitForMessage()` loops on the receive socket until it sees the expected message type or times out. DHCP can be noisy, so ignoring wrong message types is important.

`main.go` is now thin: create a MAC, create the client, call `Start()`.

## Project Layout

```
main.go           entry point
dhcp_client.go    DHCP exchange logic (DISCOVER → OFFER → REQUEST → ACK)
dhcp_message.go   packet struct, serialize/deserialize, pretty printing
dhcp_sockets.go   UDP sockets on ports 67 and 68
constants.go      field names, sizes, option codes, message types
```

## Running It

```bash
go build -o dhcpclient
sudo ./dhcpclient    # may need root for port 68
```

On my machine I sometimes hit a timeout waiting for DHCPOFFER. That was not always a code bug. Common causes:

- Port 68 already in use by the OS DHCP client
- No DHCP server on the network (e.g. testing on a network without one)
- Need broadcast flag (`0x8000`) in some environments (still on my list to verify)

## What I Still Need to Do

The coding challenge has one more step I have not finished yet:

1. **Apply the IP to the NIC** when DHCPACK arrives (set address, subnet mask, maybe gateway from options)
2. **DHCPRELEASE** to clear the local IP and tell the server we are done with the lease
3. **Test the full cycle**: release, then request a fresh IP

Right now the client prints the assigned IP but does not actually configure the network interface. Constants for `DHCPRelease` (type 7) are already in `constants.go`, so the groundwork is there.

## Things I Would Improve Later

- Random transaction ID per run
- Broadcast flag in DISCOVER/REQUEST for clients without an IP yet
- `SO_REUSEADDR` / `SO_BROADCAST` on sockets
- Pick network interface by name instead of hardcoded MAC
- Renewal (T1/T2 timers from lease options)

## References

- [RFC 2131 - DHCP](https://datatracker.ietf.org/doc/html/rfc2131)
- [RFC 2132 - DHCP Options](https://datatracker.ietf.org/doc/html/rfc2132)
- [Coding Challenge #94](https://codingchallenges.substack.com/p/coding-challenge-94-dhcp-client)

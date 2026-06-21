package main

import (
	"encoding/binary"
	"net"
	"testing"
	"time"
)

func mockDHCPServer(t *testing.T, serverPort int, replyPort int, done <-chan struct{}) {
	t.Helper()

	addr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort("127.0.0.1", itoa(serverPort)))
	if err != nil {
		t.Fatalf("resolve server addr: %v", err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		t.Fatalf("listen mock server: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1500)
	for {
		select {
		case <-done:
			return
		default:
		}

		_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return
		}
		if n < 240 {
			continue
		}

		xid := binary.BigEndian.Uint32(buf[4:8])
		chaddr := make([]byte, 16)
		copy(chaddr, buf[28:44])

		offer := buildMockOffer(xid, chaddr)
		replyAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: replyPort}
		if _, err := conn.WriteToUDP(offer, replyAddr); err != nil {
			t.Logf("mock server write: %v", err)
		}
	}
}

func buildMockOffer(xid uint32, chaddr []byte) []byte {
	pkt := make([]byte, 300)
	pkt[0] = 2 // boot reply
	pkt[1] = 1
	pkt[2] = 6
	pkt[3] = 0
	binary.BigEndian.PutUint32(pkt[4:8], xid)
	binary.BigEndian.PutUint16(pkt[10:12], 0)
	copy(pkt[28:44], chaddr)
	binary.BigEndian.PutUint32(pkt[236:240], 0x63825363)

	opts := []byte{
		53, 1, DHCPOffer,
		54, 4, 127, 0, 0, 1,
		255,
	}
	copy(pkt[240:], opts)
	return pkt[:240+len(opts)]
}

func testClientWithMAC(t *testing.T, mac []byte, serverPort int) error {
	t.Helper()

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		return err
	}
	defer conn.Close()

	done := make(chan struct{})
	defer close(done)
	go mockDHCPServer(t, serverPort, conn.LocalAddr().(*net.UDPAddr).Port, done)
	time.Sleep(50 * time.Millisecond)

	serverAddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort("127.0.0.1", itoa(serverPort)))
	if err != nil {
		return err
	}

	client := &DHCPClient{
		macAddr:       mac,
		transactionID: 0x12345678,
		sendSocket:    conn,
		receiveSocket: conn,
	}

	discover, err := client.createDHCPDiscover().Serialize()
	if err != nil {
		return err
	}
	if _, err := conn.WriteToUDP(discover, serverAddr); err != nil {
		return err
	}

	_, err = client.waitForMessage(DHCPOffer, 3*time.Second)
	return err
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func TestHardcodedMACIsValidFormat(t *testing.T) {
	mac := []byte{0x62, 0xf9, 0xb8, 0xfc, 0x9d, 0xff}
	if len(mac) != 6 {
		t.Fatalf("expected 6 bytes, got %d", len(mac))
	}
	if mac[0]&1 != 0 {
		t.Fatalf("multicast MAC not valid for DHCP client chaddr: %x", mac[0])
	}
}

func TestDHCPOFFERWithHardcodedMACAgainstMockServer(t *testing.T) {
	mac := []byte{0x62, 0xf9, 0xb8, 0xfc, 0x9d, 0xff}
	if err := testClientWithMAC(t, mac, 17670); err != nil {
		t.Fatalf("hardcoded MAC failed against mock server: %v", err)
	}
}

func TestDHCPOFFERWithRealInterfaceMACAgainstMockServer(t *testing.T) {
	mac := []byte{0x80, 0xa9, 0x97, 0x25, 0xdc, 0x88} // en0 Wi-Fi
	if err := testClientWithMAC(t, mac, 17671); err != nil {
		t.Fatalf("real en0 MAC failed against mock server: %v", err)
	}
}
func TestProductionSplitSocketsReceiveOnPort68(t *testing.T) {
	mac := []byte{0x62, 0xf9, 0xb8, 0xfc, 0x9d, 0xff}
	serverPort := 17680

	recv, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 68})
	if err != nil {
		t.Skipf("cannot bind udp4 :68 (need root?): %v", err)
	}
	defer recv.Close()

	sendAddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort("127.0.0.1", itoa(serverPort)))
	if err != nil {
		t.Fatalf("resolve send addr: %v", err)
	}
	send, err := net.DialUDP("udp4", nil, sendAddr)
	if err != nil {
		t.Fatalf("dial send socket: %v", err)
	}
	defer send.Close()

	done := make(chan struct{})
	defer close(done)
	go mockDHCPServer(t, serverPort, 68, done)
	time.Sleep(50 * time.Millisecond)

	client := &DHCPClient{
		macAddr:       mac,
		transactionID: 0x12345678,
		sendSocket:    send,
		receiveSocket: recv,
	}

	if err := client.sendMessage(client.createDHCPDiscover()); err != nil {
		t.Fatalf("send discover: %v", err)
	}
	if _, err := client.waitForMessage(DHCPOffer, 2*time.Second); err != nil {
		t.Fatalf("split sockets on port 68 failed with hardcoded MAC: %v", err)
	}
}

func TestDHCPOFFERWithRandomMACAgainstMockServer(t *testing.T) {
	mac := []byte{0x02, 0x11, 0x22, 0x33, 0x44, 0x55}
	if err := testClientWithMAC(t, mac, 17672); err != nil {
		t.Fatalf("random locally-admin MAC failed against mock server: %v", err)
	}
}

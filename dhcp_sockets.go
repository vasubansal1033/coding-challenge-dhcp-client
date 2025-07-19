package main

import "net"

func createUDPSendSocket() (*net.UDPConn, error) {
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

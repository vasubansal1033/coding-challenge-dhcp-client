package main

import "net"

func createUDPSendSocket() (*net.UDPConn, error) {
	raddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:67")
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp4", nil, raddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func createUDPReceiveSocket() (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp4", ":68")
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

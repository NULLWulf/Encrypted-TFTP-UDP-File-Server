package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"log"
	"math/rand"
	"net"
)

func NewTFTPServer() (*TFTPProtocol, error) {
	addr := &net.UDPAddr{IP: net.IPv4zero, Port: Port}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("Error starting server:", err)
		return nil, err
	}
	return &TFTPProtocol{conn: conn, raddr: addr}, nil
}
func NewTFTPServer2() (*TFTPProtocol, error, int) {
	// Random port
	port := rand.Intn(65535)
	addr := &net.UDPAddr{IP: net.IPv4zero, Port: port}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("Error starting server:", err)
		return nil, err, 0
	}
	return &TFTPProtocol{conn: conn, raddr: addr}, nil, port
}

func RunServerMode() {
	udpServer, err := NewTFTPServer()
	if err != nil {
		log.Println("Error creating server:", err)
		return
	}
	defer udpServer.Close()
	udpServer.handleConnectionsUDP2() // Launch in separate goroutine
	select {}
}

func (c *TFTPProtocol) handleConnectionsUDP2() {
	buf := make([]byte, 516)
	for {
		// Read message
		n, raddr, err := c.conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("Error reading message:", err)
			continue
		}
		// decode message
		msg := buf[:n]
		c.handleRequestWithRecovery(raddr, msg)
	}
}

func (c *TFTPProtocol) handleRequestWithRecovery(raddr *net.UDPAddr, msg []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic:", r)
		}
	}()
	c.handleRequest(raddr, msg)
}

func (c *TFTPProtocol) handleRequest(addr *net.UDPAddr, buf []byte) {
	code := binary.BigEndian.Uint16(buf[:2])
	switch tftp.TFTPOpcode(code) {
	case tftp.TFTPOpcodeRRQ:
		c.handleRRQ(addr, buf)
	case tftp.TFTPOpcodeWRQ:
		// send error packet
		c.sendError(11, "Write requests are not supported at this time")
	case tftp.TFTPOpcodeERROR:
		log.Println("Received ERROR packet, Terminating Connection...")
		return
	case tftp.TFTPOpcodeTERM:
		log.Println("Received TERM, Terminating Connection...")
		return
	default:
		log.Println("Packet context invalid, sending error packet...")
		c.sendError(4, "Illegal TFTP operation")
		return
	}
}

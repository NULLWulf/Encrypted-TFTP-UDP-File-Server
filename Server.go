package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"log"
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

func RunServerMode() {
	udpServer, err := NewTFTPServer()
	if err != nil {
		log.Println("Error creating server:", err)
		return
	}
	defer udpServer.Close()
	udpServer.handleConnectionsUDP() // launch in separate goroutine
	select {}
}

// handleConnectipnUDP handles a single udp "connection"
func (c *TFTPProtocol) handleConnectionsUDP() {
	buf := make([]byte, 516)
	go func() {
		for {
			// read message
			n, raddr, err := c.conn.ReadFromUDP(buf)
			if err != nil {
				log.Println("Error reading message:", err)
				continue
			}
			// decode message
			msg := buf[:n]
			c.handleRequest(raddr, msg)
			// close connection
			if err != nil {
				log.Printf("Error closing connection: %s\n", err)
			}
		}
	}()
}

func (c *TFTPProtocol) handleRequest(addr *net.UDPAddr, buf []byte) {
	code := binary.BigEndian.Uint16(buf[:2])
	switch tftp.TFTPOpcode(code) {
	case tftp.TFTPOpcodeRRQ:
		c.handleRRQ(addr, buf)
		break
	case tftp.TFTPOpcodeWRQ:
		// send error packet
		c.sendError(addr, 11, "Write requests are not supported at this time")
	case tftp.TFTPOpcodeERROR:
		log.Println("Received ERROR packet, Terminating Connection...")
		return
	case tftp.TFTPOpcodeTERM:
		log.Println("Received TERM, Terminating Connection...")
		return
	default:
		log.Println("Packet context invalid, sending error packet...")
		c.sendError(addr, 4, "Illegal TFTP operation")
		return
	}
}

func (c *TFTPProtocol) SetTransferSize(size uint32) {
	c.xferSize = size
}

package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"hash/crc32"
	"log"
	"math/big"
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
	udpServer.handleConnectionsUDP2() // launch in separate goroutine
	select {}
}

// handleConnectipnUDP handles a single udp "connection"
func (c *TFTPProtocol) handleConnectionsUDP() {
	buf := make([]byte, 516)
	//go func() {
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
}

func (c *TFTPProtocol) handleConnectionsUDP2() {
	buf := make([]byte, 516)
	for {
		// read message
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
	return
	//return nil
}

func (c *TFTPProtocol) receiveKeyPair(buf []byte, addr *net.UDPAddr) {
	c.dhke = new(DHKESession) // Create a new DHKE session
	c.dhke.GenerateKeyPair()  // Generate a new key pair for server
	// get bytes conver to big int
	px, py := new(big.Int), new(big.Int) // Create new big ints to hold clients public keys
	var err error
	px.SetBytes(buf[:32]) // Set the big ints to the clients public keys
	py.SetBytes(buf[32:]) // Set the big ints to the clients public keys
	c.dhke.sharedKey, err = c.dhke.generateSharedKey(px, py)
	if err != nil {
		log.Printf("Error generating shared key: %v\n", err.Error())
		c.sendError(11, "Error generating shared key")
		return
	}
	if err != nil {
		log.Printf("Error generating shared key: %v\n", err.Error())
		c.sendError(11, "Error generating shared key")
		return
	}
	key := make([]byte, 64)
	copy(c.dhke.pubKeyX.Bytes(), key[:32])
	copy(c.dhke.pubKeyY.Bytes(), key[32:])
	_, err = c.conn.WriteToUDP(buf, addr)
	if err != nil {
		log.Printf("Error sending key pair: %v\n", err.Error())
		return
	}
	log.Printf("Shared Key Chechksum %d\n", crc32.ChecksumIEEE(c.dhke.sharedKey))
	log.Printf("Sent key pair to %s\n", addr.String())

	return
}

func (c *TFTPProtocol) handleRequest(addr *net.UDPAddr, buf []byte) {
	// TODO Insert pre-request handler here
	c.ADti(len(buf))
	if len(buf) == 64 {
		c.receiveKeyPair(buf, addr)
		return
	}

	code := binary.BigEndian.Uint16(buf[:2])
	switch tftp.TFTPOpcode(code) {
	case tftp.TFTPOpcodeRRQ:
		c.handleRRQ(addr, buf)
		break
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

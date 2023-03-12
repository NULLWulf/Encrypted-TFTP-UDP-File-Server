package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"log"
	"net"
)

// handleConnectipnUDP handles a single udp "connection"
func (c *TFTPProtocol) handleConnectionUDP() {
	buf := make([]byte, 1024)
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
		}
	}()

}

func RunServerMode() {
	udpServer, err := NewTFTPServer()
	if err != nil {
		log.Println("Error creating server:", err)
		return
	}
	defer udpServer.Close()
	go udpServer.handleConnectionUDP() // launch in separate goroutine
	select {}
}

func (c *TFTPProtocol) handleRequest(addr *net.UDPAddr, buf []byte) {

	code := binary.BigEndian.Uint16(buf[:2])
	switch tftp.TFTPOpcode(code) {
	case tftp.TFTPOpcodeRRQ:
		log.Println("Received RRQ")
		var req tftp.Request
		err := req.Parse(buf)
		if err != nil {
			log.Println("Error parsing request:", err)
			return
		}
		log.Printf("Received %d bytes from %s for file %s \n", string(buf), addr, string(req.Filename))
		err, img := IQ.AddNewAndReturnImg(string(req.Filename))
		if err != nil {
			log.Println("Error adding new image:", err)
			return
		}
		log.Printf("Sending %d bytes to %s for file %s \n", len(img), addr, string(req.Filename))
		// Encode the Person struct as JSON.
		//dataPack, _ := tftp.NewData(0, img)
		//packet := dataPack.ToBytes()
		_, err = c.conn.WriteToUDP(img, addr)
		if err != nil {
			log.Println("Error sending data packet:", err)
			return
		}
		break
	case tftp.TFTPOpcodeWRQ:
		log.Println("Received WRQ")
		break
	case tftp.TFTPOpcodeDATA:
		log.Println("Received DATA")
		break
	case tftp.TFTPOpcodeACK:
		log.Println("Received ACK")
		break
	case tftp.TFTPOpcodeERROR:
		log.Println("Received ERROR")
		break
	case tftp.TFTPOpcodeTERM:
		log.Println("Received TERM")
		break
	default:

	}
}

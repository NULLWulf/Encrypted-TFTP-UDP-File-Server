package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"fmt"
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
	udpServer.handleConnectionUDP() // launch in separate goroutine
	select {}
}

func (c *TFTPProtocol) handleRequest(addr *net.UDPAddr, buf []byte) {
	code := binary.BigEndian.Uint16(buf[:2])
	switch tftp.TFTPOpcode(code) {
	case tftp.TFTPOpcodeRRQ:
		c.handleRRQ(addr, buf)
		break
	case tftp.TFTPOpcodeWRQ:
		log.Println("Received WRQ")
		break
	case tftp.TFTPOpcodeACK:
		log.Println("Received ACK")
		break
	case tftp.TFTPOpcodeTERM:
		log.Println("Received TERM, Terminating Transfer...")
		return
	default:

	}
}

func (c *TFTPProtocol) SetTransferSize(size uint32) {
	c.xferSize = size
}

func (c *TFTPProtocol) handleRRQ(addr *net.UDPAddr, buf []byte) {
	log.Println("Received RRQ")
	var req tftp.Request
	err := req.Parse(buf)
	if err != nil {
		c.sendError(addr, 4, "Illegal TFTP operation")
		return
	}
	log.Printf("Received %d bytes from %s for file %s \n", len(buf), addr, string(req.Filename))
	err, img := IQ.AddNewAndReturnImg(string(req.Filename))
	if err != nil {
		c.sendError(addr, 10, "File not found")
		return
	}
	c.SetProtocolOptions(req.Options, len(img))
	opAck := tftp.NewOpt1(c.blockSize, c.xferSize, c.blockSize, []byte("netascii"))

	_, err = c.conn.WriteToUDP(opAck.ToBytes(), addr)
	if err != nil {
		log.Println("Error sending data packet:", err)
		return
	}

	c.dataBlocks, err = tftp.PrepareData(img, int(c.blockSize))
	if err != nil {
		return
	}
	fmt.Sprintf("Sending %d blocks", len(c.dataBlocks))

	c.StartTftpSenderLoop()
}

func (c *TFTPProtocol) StartTftpSenderLoop() error {
	c.base = 1
	c.nextExpectedBlock = 1
	c.ackBlocks = make(map[uint16]bool)
	c.bufferedBlocks = make(map[uint16]*tftp.Data)

	for {
		for i := 0; i < int(c.windowSize) && int(c.nextExpectedBlock)+i < len(c.dataBlocks); i++ {
			dataPacket := c.dataBlocks[int(c.nextExpectedBlock)+i-1]
			_, err := c.conn.WriteToUDP(dataPacket.ToBytes(), c.raddr)
			if err != nil {
				return fmt.Errorf("Failed to send data packet: %v", err)
			}
		}

		ackBuf := make([]byte, 4)
		n, _, err := c.conn.ReadFromUDP(ackBuf)
		if err != nil {
			return fmt.Errorf("Failed to read from UDP connection: %v", err)
		}
		ackBuf = ackBuf[:n]

		opcode := tftp.TFTPOpcode(binary.BigEndian.Uint16(ackBuf[:2]))

		switch opcode {
		case tftp.TFTPOpcodeACK:
			ackPacket := &tftp.Ack{}
			if err := ackPacket.Parse(ackBuf); err == nil {
				if ackPacket.BlockNumber >= c.base {
					c.base = ackPacket.BlockNumber + 1
					c.nextExpectedBlock = c.base
				}
			} else {
				return fmt.Errorf("Failed to parse ACK packet: %v", err)
			}
		case tftp.TFTPOpcodeERROR:
			var errPacket tftp.Error
			if err := errPacket.Parse(ackBuf); err == nil {
				return fmt.Errorf("Error packet received: %v", errPacket)
			} else {
				return fmt.Errorf("Failed to parse error packet: %v", err)
			}
		default:
			// Ignore any other opcodes
		}

		if c.base > uint16(len(c.dataBlocks)) {
			break
		}
	}
	return nil
}

func (c *TFTPProtocol) sendError(addr *net.UDPAddr, errCode uint16, errMsg string) {
	errPack := tftp.NewErr(errCode, []byte(errMsg))
	_, err := c.conn.WriteToUDP(errPack.ToBytes(), addr)
	if err != nil {
		log.Println("Error sending error packet:", err)
		return
	}
}

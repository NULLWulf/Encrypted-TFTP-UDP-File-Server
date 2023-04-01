package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
)

func (c *TFTPProtocol) handleRRQ(addr *net.UDPAddr, buf []byte) {
	var req tftp.Request
	err := req.Parse(buf)
	if err != nil {
		parseErr := fmt.Errorf("error parsing request: %s", err)
		log.Printf("Error parsing request: %s\n", parseErr)
		c.sendError(22, parseErr.Error())
		return
	}
	log.Printf("Received %d bytes from %s for file %s \n", len(buf), addr, string(req.Filename))
	err, img := IQ.AddNewAndReturnImg(string(req.Filename))
	if err != nil {
		c.sendError(20, "File not found")
		return
	}
	c.SetProtocolOptions(nil, 0)
	opAck := tftp.NewOpt1(c.blockSize, c.xferSize, c.blockSize, []byte("octet"))
	c.dataBlocks, err = PrepareData(img, int(c.blockSize))
	if err != nil {
		c.sendError(4, "Illegal TFTP operation")
		return
	}
	n, err := c.conn.WriteToUDP(opAck.ToBytes(), addr)
	c.ADto(n)
	if err != nil {
		c.sendError(10, "Illegal TFTP operation")
		return
	}
	err = c.sender(addr)
	if err != nil {
		c.sendError(21, "Illegal TFTP operation")
		return
	}
}

func (c *TFTPProtocol) sender(addr *net.UDPAddr) error {
	// Initialize variables
	var ack tftp.Ack
	log.Println("Starting sender transfer TFTP loop")
	packet := make([]byte, 516)
	windowSize := 8
	base := 1
	nextSeqNum := 1
	n, _ := c.conn.Read(packet)
	c.ADti(n)
	packet = packet[:n]
	err := ack.Parse(packet)
	log.Printf("Initial ACK received: %v\n", ack)
	if err != nil {
		return errors.New("error parsing ack packet: " + err.Error())
	}
	if ack.BlockNumber != 0 {
		return errors.New("error parsing ack packet: block number should be 0")
	}

	// Loop until all data blocks have been sent and acknowledged
	for base <= len(c.dataBlocks) {

		// Send packets within the window size
		for nextSeqNum < base+windowSize && nextSeqNum <= len(c.dataBlocks) {
			if rand.Float64() < .20 && DropPax {
				log.Printf("Dropped packet %d\n", nextSeqNum)
				nextSeqNum--
				continue
			}

			packet = c.dataBlocks[nextSeqNum-1].ToBytes()
			n, err = c.conn.WriteToUDP(packet, addr)
			c.ADto(n)
			log.Printf("Sending packet %d, %v", nextSeqNum, len(packet))
			if err != nil {
				c.sendAbort()
				return fmt.Errorf("error sending data packet: %s", err)
			}
			nextSeqNum++

			n, _ = c.conn.Read(packet)
			c.ADti(n)
			packet = packet[:n]
			opcode := binary.BigEndian.Uint16(packet[:2])
			switch tftp.TFTPOpcode(opcode) {
			case tftp.TFTPOpcodeACK:
				// may need to check if the ack is for the correct packet
				err = ack.Parse(packet)
				if err != nil {
					log.Printf("Error parsing ACK packet: %s\n", err)
				}
				log.Printf("Received ACK for packet %d\n", ack.BlockNumber)
				if ack.BlockNumber >= uint16(base) {
					base = int(ack.BlockNumber + 1)
				}
			default:
				log.Printf("Received unexpected packet: %v\n", packet)
			}

		}

		// Check if all packets have been sent and acknowledged
		if base > len(c.dataBlocks) {
			log.Printf("All packets sent and acknowledged\n")
			break
		}
	}

	return nil
}

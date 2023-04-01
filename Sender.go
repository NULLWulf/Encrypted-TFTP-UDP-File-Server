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

// sender is the main loop for the sender side of the TFTP protocol
// It sends data blocks and waits for ACKs.  If an ACK is not received
// within the timeout period, the data block is resent.  If an error
// occurs, the error is logged and the loop is exited.
func (c *TFTPProtocol) sender(addr *net.UDPAddr) error {
	// Initialize variables
	var ack tftp.Ack //
	log.Println("Starting sender transfer TFTP loop")
	packet := make([]byte, 520)                            //Byte buffer
	windowSize, base, nextSeqNum, dropProb := 8, 1, 1, .20 //Window size, base, next sequence number, and drop probability
	n, _ := c.conn.Read(packet)                            //Read the initial ACK
	c.ADti(n)                                              // Add to the total bytes received to running outgoing total
	packet = packet[:n]                                    //Trim the packet to the size of the data received
	err := ack.Parse(packet)                               //
	log.Printf("Initial ACK received: %v\n", ack)
	if err != nil {
		return errors.New("error parsing ack packet: " + err.Error())
	}
	if ack.BlockNumber != 0 {
		c.sendError(3, "Expected initial block number to be 0")
		return errors.New("error parsing ack packet: block number should be 0, expecting initial block")
	}

	// Loop until all data blocks have been sent and acknowledged
	for base <= len(c.dataBlocks) {

		// Send packets within the window size
		for nextSeqNum < base+windowSize && nextSeqNum <= len(c.dataBlocks) {
			if rand.Float64() < dropProb && DropPax { //DropPax is a global variable that is set to true if the user wants to simulate packet loss
				// uses probability of 0.2 to drop a packet
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

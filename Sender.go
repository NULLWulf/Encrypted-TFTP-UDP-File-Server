package main

import (
	"CSC445_Assignment2/tftp"
	"errors"
	"fmt"
	"log"
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
	c.dataBlocks, err = tftp.PrepareData(img, int(c.blockSize))
	if err != nil {
		c.sendError(4, "Illegal TFTP operation")
		return
	}

	//start := time.Now().UnixNano()
	_, err = c.conn.WriteToUDP(opAck.ToBytes(), addr)
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

//func (c *TFTPProtocol) startTftpSenderLoop(start int64) error {
//	log.Printf("Starting Sender TFTP Loop\n")
//	packet := make([]byte, 516)
//	var ack tftp.Ack
//	err := error(nil) // placeholder to avoid shadowing
//	c.nextSeqNum = 1  // settings to 0 for first data packet
//	c.base = 1
//	c.retryCount = 0
//	c.maxRetries = 5
//	_, _ = c.conn.Read(packet)
//	err = ack.Parse(packet)
//	if err != nil {
//		return errors.New("error parsing ack packet: " + err.Error())
//	}
//	if ack.BlockNumber != 0 {
//		return errors.New("error parsing ack packet: block number should be 0")
//	}
//
//	for {
//		if c.nextSeqNum < c.base+c.windowSize && c.nextSeqNum <= uint16(len(c.dataBlocks)) {
//			// send data packet
//			// Send data packet, c.nextSeqNum is the next sequence number to send
//			// -1 because nextSeqNum is 1 indexed
//			_, err = c.conn.Write(c.dataBlocks[c.nextSeqNum-1].ToBytes())
//			c.nextSeqNum++
//			if err != nil {
//				c.sendAbort()
//				return errors.New("error sending data packet: " + err.Error())
//			}
//
//			_, err := c.conn.Read(packet)
//			ack.Parse(packet)
//			if ack.BlockNumber == c.nextSeqNum {
//				c.base = c.nextSeqNum
//				c.retries = 0
//			}
//			}
//		}
//	}
//}

func (c *TFTPProtocol) sender(addr *net.UDPAddr) error {
	// Initialize variables
	var ack tftp.Ack
	log.Println("Starting sender transfer TFTP loop")
	packet := make([]byte, 516)
	windowSize := 8
	base := 1
	nextSeqNum := 1
	n, _ := c.conn.Read(packet)
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
			packet = c.dataBlocks[nextSeqNum-1].ToBytes()
			//_, err = c.conn.Write(packet)
			_, err = c.conn.WriteToUDP(packet, addr)

			log.Printf("Sending packet %d\n", nextSeqNum)
			if err != nil {
				c.sendAbort()
				return fmt.Errorf("error sending data packet: %s", err)
			}

			nextSeqNum++
		}

		// Check if all packets have been sent and acknowledged
		if base > len(c.dataBlocks) {
			break
		}
	}

	return nil
}

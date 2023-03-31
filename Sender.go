package main

import (
	"CSC445_Assignment2/tftp"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

func (c *TFTPProtocol) handleRRQ(addr *net.UDPAddr, buf []byte) {
	var req tftp.Request
	err := req.Parse(buf)
	if err != nil {
		parseErr := fmt.Errorf("error parsing request: %s", err)
		log.Printf("Error parsing request: %s\n", parseErr)
		c.sendError(addr, 22, parseErr.Error())
		return
	}
	log.Printf("Received %d bytes from %s for file %s \n", len(buf), addr, string(req.Filename))
	err, img := IQ.AddNewAndReturnImg(string(req.Filename))
	if err != nil {
		c.sendError(addr, 20, "File not found")
		return
	}
	c.SetProtocolOptions(nil, 0)
	opAck := tftp.NewOpt1(c.blockSize, c.xferSize, c.blockSize, []byte("octet"))
	c.dataBlocks, err = tftp.PrepareData(img, int(c.blockSize))
	if err != nil {
		log.Println("Error preparing data blocks: ", err)
		c.sendError(addr, 4, "Illegal TFTP operation")
		return
	}

	start := time.Now().UnixNano()
	_, err = c.conn.WriteToUDP(opAck.ToBytes(), addr)
	if err != nil {
		log.Println("Error sending opack packet: ", err)
		c.sendError(addr, 10, "Illegal TFTP operation")
		return
	}

	err = c.startTftpSenderLoop(start)
	if err != nil {
		log.Println("Fatal error in TFTP Sender Loop: ", err)
		c.sendError(addr, 21, "Illegal TFTP operation")
		return
	}
}

func (c *TFTPProtocol) startTftpSenderLoop(start int64) error {
	log.Printf("Starting Sender TFTP Loop\n")
	dataPacket := make([]byte, 516)
	n := 0            // number of bytes read
	err := error(nil) // placeholder to avoid shadowing
	c.nextSeqNum = 1  // settings to 0 for first data packet
	c.base = 1
	c.retryCount = 0
	c.maxRetries = 5
	c.timeout := time.Duration(c.backoff) * time.Millisecond
	n, _ = c.conn.Read(dataPacket)

	for {
		if c.nextSeqNum < c.base+c.windowSize && c.nextSeqNum <= uint16(len(c.dataBlocks)) {
			// send data packet
			// Send data packet, c.nextSeqNum is the next sequence number to send
			// -1 because nextSeqNum is 1 indexed
			_, err = c.conn.Write(c.dataBlocks[c.nextSeqNum-1].ToBytes())
			if err != nil {
				c.sendAbort()
				return errors.New("error sending data packet: " + err.Error())
			}

			timer := time.NewTimer(time.Duration(c.timeout))
			defer time.Stop()

		} else {

		}

	}
}

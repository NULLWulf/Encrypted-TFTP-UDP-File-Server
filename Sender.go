package main

import (
	"CSC445_Assignment2/tftp"
	"fmt"
	"log"
	"net"
	"time"
)

func (c *TFTPProtocol) handleRRQ(addr *net.UDPAddr, buf []byte) {
	log.Println("Received RRQ")
	var req tftp.Request
	err := req.Parse(buf)
	if err != nil {
		parseErr := fmt.Errorf("error parsing request: %s", err)
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
		log.Println("Error sending data packet: ", err)
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
	log.Printf("Starting TFTP Sender Transfer Loop")
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

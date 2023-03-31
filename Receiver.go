package main

import (
	"CSC445_Assignment2/tftp"
	"log"
	"net"
)

func (c *TFTPProtocol) TftpClientTransferLoop(addr net.Addr) error {
	log.Printf("Starting Receiver TFTP Transfer Loop\n")
	//nextSeqNum := uint16(0)

	return nil
}

func (c *TFTPProtocol) receiveDataPacket(dataPack *tftp.Data) {
	blockNumber := dataPack.BlockNumber
	if blockNumber == c.nextExpectedBlock {
		c.bufferedBlocks[blockNumber] = dataPack
		c.ackBlocks[blockNumber] = true

		for c.bufferedBlocks[c.base] != nil {
			c.dataBlocks = append(c.dataBlocks, c.bufferedBlocks[c.base])
			delete(c.bufferedBlocks, c.base)
			c.base++
			c.nextExpectedBlock++
		}
	} else if blockNumber > c.nextExpectedBlock && blockNumber < c.base+c.windowSize {
		// Buffer out-of-order packet
		c.bufferedBlocks[blockNumber] = dataPack
		c.ackBlocks[blockNumber] = true
	}
}

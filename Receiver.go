package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"errors"
	"log"
	"net"
)

// TftpClientTransferLoop is the main loop for the client side of the transfer
func (c *TFTPProtocol) TftpClientTransferLoop(cn *net.UDPConn) (err error, finish bool) {
	conn := *cn
	log.Printf("Starting Receiver TFTP Transfer Loop\n")
	c.receivedPackets = make(map[uint16]*tftp.Data)
	dataPacket := make([]byte, 1024)
	//n := 0           // number of bytes read
	err = error(nil) // placeholder to avoid shadowing
	lb := false      // last data block received
	c.nextSeqNum = 0 // settings to 0 for first data packet
	ack := tftp.NewAck(c.nextSeqNum)
	log.Printf("Sending initial ACK packet: %v\n", ack)
	c.nextSeqNum++ // increment for first data packet
	dataPacket = ack.ToBytes()
	_, err = conn.Write(dataPacket)
	if err != nil {
		c.sendAbort()
		return errors.New("error sending initial ACK packet: " + err.Error()), false
	}
	for { // loop until packet received
		dataPacket = make([]byte, 1024) // reset data packet
		n, err := conn.Read(dataPacket)
		dataPacket, _ = decrypt(dataPacket[:n], c.dhke.aes512Key)

		opcode := binary.BigEndian.Uint16(dataPacket[:2])

		switch tftp.TFTPOpcode(opcode) {
		case tftp.TFTPOpcodeERROR:
			c.handleErrPacket(dataPacket)
		case tftp.TFTPOpcodeTERM:
			return errors.New("termination packet received"), false
		case tftp.TFTPOpcodeDATA:
			lb = c.receiveDataPacket(dataPacket) // handle data packet
		default:
			if err != nil {
				return errors.New("error reading packet: " + err.Error()), false
			}
		}
		// if last data block received, end transfer
		if lb {
			log.Printf("Last data block received, ending transfer\n")
			return nil, true
		}
	}
}

// receiveDataPacket handles a data packet and returns true if the last data
// note that this function needs a key to decrypt the data
func (c *TFTPProtocol) receiveDataPacket(dataPacket []byte) bool {
	var dataPack tftp.Data
	err := dataPack.Parse(dataPacket, nil)
	if err != nil || dataPack.BlockNumber != c.nextSeqNum {
		// Duplicate packet or out of order packet
		c.sendAck(c.nextSeqNum - 1) // Send ACK for previous packet
		return false
	}
	if dataPack.Checksm != tftp.Checksum(dataPack.Data) {
		c.sendAck(c.nextSeqNum - 1) // Send ACK for previous packet
		return false
	}
	// Append data to file
	if !c.appendFileDate(&dataPack) { // Append data to file, if duplicate packet, return false
		return false
	}
	// Send ACK for this packet on routine
	if len(dataPack.Data) < 512 {
		// Last data block received, end of file
		log.Printf("Last data block received, end of file\n")
		c.sendAck(dataPack.BlockNumber)
		return true
	}
	c.sendAck(c.nextSeqNum) // Send ACK for this packet
	c.nextSeqNum++          // Increment for next packet
	return false            // Not last data block
}

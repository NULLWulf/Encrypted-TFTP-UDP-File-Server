package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"errors"
	"log"
	"net"
)

func (c *TFTPProtocol) TftpClientTransferLoop(addr net.Addr) (err error, finish bool) {
	log.Printf("Starting Receiver TFTP Transfer Loop\n")
	dataPacket := make([]byte, 516)
	n := 0           // number of bytes read
	err = error(nil) // placeholder to avoid shadowing
	lb := false      // last data block received
	c.nextSeqNum = 0 // settings to 0 for first data packet
	ack := tftp.NewAck(c.nextSeqNum)
	c.nextSeqNum++ // increment for first data packet
	_, err = c.conn.Write(ack.ToBytes())
	if err != nil {
		c.sendAbort()
		return errors.New("error sending initial ACK packet: " + err.Error()), false
	}
	for {
		// receive data packet from server
		n, err = c.conn.Read(dataPacket)
		dataPacket = dataPacket[:n] // trim packet to size of data
		// get opcode
		opcode := binary.BigEndian.Uint16(dataPacket[:2])
		switch tftp.TFTPOpcode(opcode) {
		case tftp.TFTPOpcodeERROR:
			// error packet received, handle it, won't necessarily end transfer
			// transfer end other than termination determine by excess timeouts
			c.handleErrPacket(dataPacket)
			break
		case tftp.TFTPOpcodeTERM:
			return errors.New("termination packet received"), false
		case tftp.TFTPOpcodeDATA:
			// data packet received, handle ituint16
			lb = c.receiveDataPacket(dataPacket)
			break
		default:
			if err != nil {
				return errors.New("error reading packet: " + err.Error()), false
			}
		}

		if lb {
			break
		}
	}

	return nil, true
}

// receiveDataPacket receives a data packet and handles it
// if the data packet is the next expected packet, it is added to the dataBlocks
// and an ACK is sent for it
// if the data packet is a duplicate of a previously received packet, an ACK is
// sent for the previous packet
// if the data packet is not the next expected packet, an ACK is sent for the
// previous packet
func (c *TFTPProtocol) receiveDataPacket(dataPacket []byte) (endOfFile bool) {
	var dataPack tftp.Data
	err := dataPack.Parse(dataPacket)
	if err == nil && dataPack.BlockNumber == c.nextSeqNum {
		c.dataBlocks = append(c.dataBlocks, &dataPack)
		// send ACK for this packet on go routine
		go c.sendAck(c.nextSeqNum)
		c.nextSeqNum++
		if len(dataPacket) < 516 {
			// last data block received, end of file
			return true
		}
	} else { // duplicate packet or out of order packet
		// send ACK for previous packet on go routine
		go c.sendAck(c.nextSeqNum - 1)
	}
	return false
}

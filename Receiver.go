package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"errors"
	"log"
	"net"
)

func (c *TFTPProtocol) TftpClientTransferLoop(cn *net.UDPConn) (err error, finish bool) {
	conn := *cn
	log.Printf("Starting Receiver TFTP Transfer Loop\n")
	c.receivedPackets = make(map[uint16]*tftp.Data)
	dataPacket := make([]byte, 520)
	//n := 0           // number of bytes read
	err = error(nil) // placeholder to avoid shadowing
	lb := false      // last data block received
	c.nextSeqNum = 0 // settings to 0 for first data packet
	c.base = 1
	ack := tftp.NewAck(c.nextSeqNum)
	log.Printf("Sending initial ACK packet: %v\n", ack)
	c.nextSeqNum++ // increment for first data packet
	_, err = conn.Write(ack.ToBytes())
	if err != nil {
		c.sendAbort()
		return errors.New("error sending initial ACK packet: " + err.Error()), false
	}
	for {
		n, err := conn.Read(dataPacket)
		c.ADti(n)
		dataPacket = dataPacket[:n] // trim packet to size of data
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
			log.Printf("Last data block received, ending transfer\n")
			return nil, true
		}
	}

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
		if dataPack.Checksm == tftp.Checksum(dataPack.Data) {
			log.Printf("\n-----------------\nReceived data packet block number: %d\nFirst 10 Bytes: %v\nLength %d\n-----------------\n", dataPack.BlockNumber, dataPack.Data[0:10], len(dataPack.Data))
			//c.dataBlocks = append(c.dataBlocks, &dataPack)
			c.appendFileDate(&dataPack)
			// send ACK for this packet on  routinel
			if len(dataPack.Data) < 512 {
				// last data block received, end of file
				log.Printf("Last data block received, end of file\n")
				c.sendAck(dataPack.BlockNumber)
				return true
			}
			c.sendAck(c.nextSeqNum)

			c.nextSeqNum++
		} else {
			// detail about failure
			//log.Printf("\n-----------------\nReceived data packet block number: %d\nFirst 10 Bytes: %v\nLength %d\nChecksum: %v\n-----------------\n", dataPack.BlockNumber, dataPack.Data[0:10], len(dataPack.Data), dataPack.Checksm)
			log.Printf("Calc Checksum: %v\n", tftp.Checksum(dataPack.Data))
			log.Printf("Received Checksum: %v\n", dataPack.Checksm)
			//log.Printf("Checksum failed, sending ACK for previous packet\n")
			c.sendAck(c.nextSeqNum - 1)
		}
	} else { // duplicate packet or out of order packet
		// send ACK for previous packet on  routine
		c.sendAck(c.nextSeqNum - 1)
	}
	return false
}

package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"net"
)

type TFTPProtocol struct {
	conn              *net.UDPConn
	raddr             *net.UDPAddr
	fileData          *[]byte
	xferSize          uint32
	blockSize         uint16
	windowSize        uint16
	key               []byte
	dataBlocks        []*tftp.Data
	base              uint16                // Base of the window
	nextExpectedBlock uint16                // Next expected block number
	ackBlocks         map[uint16]bool       // Map to keep track of acknowledged blocks
	bufferedBlocks    map[uint16]*tftp.Data // Map to buffer out-of-order blocks
}

func (c *TFTPProtocol) Close() error {
	return c.conn.Close()
}

// SetProtocolOptions sets the protocol options for the TFTP protocol
// using static values for the time being
func (c *TFTPProtocol) SetProtocolOptions(options map[string][]byte, l int) {
	if l != 0 {
		c.SetTransferSize(uint32(l))
	}
	if options["tsize"] != nil && c.xferSize == 0 {
		c.SetTransferSize(binary.BigEndian.Uint32(options["tsize"]))
	}
	if options["blksize"] != nil {
		c.blockSize = binary.BigEndian.Uint16(options["blksize"])
	}
	if options["windowsize"] != nil {
		c.windowSize = binary.BigEndian.Uint16(options["windowsize"])
	}
	if options["key"] != nil {
		c.key = options["key"]
	}

	c.key = []byte("1234567890123456")
	c.blockSize = 512
	c.windowSize = 4
}

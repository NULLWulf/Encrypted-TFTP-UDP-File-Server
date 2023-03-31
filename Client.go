// StartImgReqTFTP Starts image request over UDP taking in parameters needing to create
// initial Read Request TFTP packet.
// 1) Attempt to resolve host address
// 2) Dial host address tion
// 3) Create request packet
// 4) Send Request Packet
// 5) Wait for OACK Packet
// if error packet, returns error code and message from packet
// if not expected packet, Returns received opcode
// and closes the
// 6) Begin Sliding Window Protocol of Data
package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

func NewTFTPClient() (*TFTPProtocol, error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:7500")
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return nil, err
	}

	return &TFTPProtocol{conn: conn, raddr: remoteAddr, xferSize: 0}, nil
}

func (c *TFTPProtocol) RequestFile(url string) (tData []byte, err error) {
	reqPack, _ := tftp.NewReq([]byte(url), []byte("octet"), 0, nil)
	packet, _ := reqPack.ToBytes()
	c.SetProtocolOptions(nil, 0)
	_, err = c.conn.Write(packet)
	if err != nil {
		log.Printf("Error sending request packet: %s\n", err)
		return
	}
	n, _, err := c.conn.ReadFromUDP(packet)
	packet = packet[:n]
	switch opcode := binary.BigEndian.Uint16(packet[:2]); opcode {
	case uint16(tftp.TFTPOpcodeERROR):
		var errPack tftp.Error
		err := errPack.Parse(packet)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error code %d: %s", errPack.ErrorCode, errPack.ErrorMessage)
	case uint16(tftp.TFTPOpcodeTERM):
		return nil, fmt.Errorf("server terminated connection")
	case uint16(tftp.TFTPOpcodeOACK):
		var oackPack tftp.OptionAcknowledgement
		err := oackPack.Parse(packet)
		if err != nil {
			return nil, fmt.Errorf("error parsing OACK packet: %s", err)
		}
		err = c.TftpClientTransferLoop()
		if err != nil {
			return nil, fmt.Errorf("error in transfer loop: %s", err)
		}
	}

	return *c.fileData, nil
}

func (c *TFTPProtocol) TftpClientTransferLoop() error {
	log.Printf("Starting Client TFTP Transfer Loop")
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

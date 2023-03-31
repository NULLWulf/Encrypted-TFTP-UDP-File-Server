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
	d, err := c.preDataTransfer()
	if err != nil {
		log.Printf("Error in preDataTransfer: %s\n", err)
		return nil, err
	}
	return d, nil
}

func (c *TFTPProtocol) preDataTransfer() ([]byte, error) {
	packet := make([]byte, 516)
	cont := true
	err := error(nil)
	finish := false

	for {
		n, addr, tErr := c.conn.ReadFromUDP(packet)
		if tErr != nil {
			return nil, fmt.Errorf("error reading packet: %s", err)
		}
		packet = packet[:n]
		code := binary.BigEndian.Uint16(packet[:2])
		switch tftp.TFTPOpcode(code) {
		case tftp.TFTPOpcodeERROR:
			log.Printf("Error packet received: %s\n", packet)
			var errPack tftp.Error
			err = errPack.Parse(packet)
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("error code %d: %s", errPack.ErrorCode, errPack.ErrorMessage)
		case tftp.TFTPOpcodeTERM:
			log.Printf("Server terminated connection: %s\n", packet)
			return nil, fmt.Errorf("server terminated connection")
		case tftp.TFTPOpcodeOACK:
			log.Printf("Received OACK packet: %s\n", packet)
			var oackPack tftp.OptionAcknowledgement
			err = oackPack.Parse(packet)
			if err != nil {
				return nil, fmt.Errorf("error parsing OACK packet: %s", err)
			}
			err, finish = c.TftpClientTransferLoop(addr)
			if err != nil {
				return nil, fmt.Errorf("error in transfer loop: %s", err)
			}
			cont = false
		default:
			log.Printf("Received unexpected packet: %s\n", packet)
		}
		if !cont {
			break
		}
	}

	if finish {
		d := tftp.RebuildData(c.dataBlocks)
		return d, nil
	}
	// WIP
	return nil, nil
}

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
	"errors"
	"fmt"
	"log"
	"net"
)

type TFTPProtocol struct {
	conn     *net.UDPConn
	raddr    *net.UDPAddr
	fileData *[]byte
	xferSize uint64
}

func NewTFTPClient() (*TFTPProtocol, error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:7501")
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return nil, err
	}

	return &TFTPProtocol{conn: conn, raddr: remoteAddr, xferSize: 0}, nil
}

func (c *TFTPProtocol) Close() error {
	return c.conn.Close()
}

func (c *TFTPProtocol) RequestFile(url string) (tData []byte, err error) {
	packet := make([]byte, 512)
	// Make a new request packet
	//BlockSize := 512
	//WindowSize := 1
	//PrivateKey := "1234567890123456"
	reqPack, _ := tftp.NewReq([]byte(url), []byte("octet"), 0, nil)
	packet, _ = reqPack.ToBytes()

	_, err = c.conn.WriteToUDP(packet, c.raddr)
	if err != nil {
		log.Printf("Error sending request packet: %s\n", err)
		return
	}

	n, _, err := c.conn.ReadFromUDP(packet)
	packet = packet[:n]
	opcode := tftp.TFTPOpcode(binary.BigEndian.Uint16(packet[:2]))
	if err != nil {
		log.Printf("Error reading reply from UDP server: %s\n", err)
		return
	}

	if opcode == tftp.TFTPOpcodeERROR {
		//process error packet
		var errPack tftp.Error
		err := errPack.Parse(packet)
		if err != nil {
			errSt := fmt.Sprintf("error packet received... code: %d message: %s\n", errPack.ErrorCode, errPack.ErrorMessage)
			log.Println(errSt)
			return nil, errors.New(errSt)
		}
	}

	if opcode != tftp.TFTPOpcodeOACK {
		errSt := fmt.Sprintf("returned packet opcode is neither OACK or ACK.. opcode: %d packet_t: %s\n", opcode, opcode.String())
		log.Println(errSt)
		return nil, errors.New(errSt)
	}

	// Process OACK packet
	var oackPack tftp.OptionAcknowledgement
	err = oackPack.Parse(packet)
	err = c.StartDataClientXfer(512)
	if err != nil {
		return nil, err
	}

	return *c.fileData, nil
}

func (c *TFTPProtocol) StartDataClientXfer(blocksize uint32) (err error) {
	var dataPack tftp.Data
	var errPack tftp.Error

	pckBfr := make([]byte, blocksize+4)
	dataBuffer := *c.fileData
	dataBuffer = make([]byte, 0)
	n := 0
	closeXfer := false

	for {
		n, c.raddr, err = c.conn.ReadFromUDP(pckBfr)
		pckBfr = pckBfr[:n]
		opcode := tftp.TFTPOpcode(binary.BigEndian.Uint16(pckBfr[:2]))
		switch opcode {
		case tftp.TFTPOpcodeDATA:
			err = dataPack.Parse(pckBfr)
			copy(dataBuffer, dataPack.Data)
			ackPack := tftp.NewAck(dataPack.BlockNumber)
			pckBfr = ackPack.ToBytes()
			_, err = c.conn.WriteToUDP(pckBfr, c.raddr)
		case tftp.TFTPOpcodeERROR:
			_ = errPack.Parse(pckBfr)
			err = fmt.Errorf("error packet received... code: %d message: %s\n", errPack.ErrorCode, errPack.ErrorMessage)
			closeXfer = true
		default:
			closeXfer = true
		}

		if closeXfer {
			break
		}
	}
	if err != nil {
		return
	}

	return
}

package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
)

type TFTPClient struct {
	conn *net.UDPConn
}

func NewTFTPClient() (*TFTPClient, error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:7501")
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return nil, err
	}

	return &TFTPClient{conn}, nil
}

func (c *TFTPClient) Close() error {
	return c.conn.Close()
}

// StartImgReqTFTP Starts image request over UDP taking in parameters needing to create
// initial Read Request TFTP packet.
// 1) Attempt to resolve host address
// 2) Dial host address
// TODO Send Private Key for encryption
// TODO Send Window Size
// TODO Send Acceptable Block Size
// 3) Create request packet
// 4) Send Request Packet
// 5) Wait for OACK Packet
// if error packet, returns error code and message from packet
// if not expected packet, Returns received opcode
// and closes the
// 6) Begin Sliding Window Protocol of Data
func (c *TFTPClient) RequestFile(url string) (tData []byte, err error) {

	client, err := NewTFTPClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	packet := make([]byte, 512)
	// Make a new request packet
	BlockSize := 512
	WindowSize := 1
	PrivateKey := "1234567890123456"
	reqPack, _ := tftp.NewTFTPRequest([]byte(url), []byte("octet"), 0, map[string]string{"blksize": string(rune(BlockSize)), "windowsize": string(rune(WindowSize)), "key": PrivateKey})
	packet, _ = reqPack.ToBytes()
	//remoteAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:7501")
	//
	//if err != nil {
	//	// TODO Make error more elegant
	//	log.Printf("Error attempting to resolve host address: %s\n", err)
	//	return
	//}
	//
	//conn, err := net.DialUDP("udp", nil, remoteAddr)
	//if err != nil {
	//	// TODO Make error more elegant
	//	log.Printf("Error connecting to host: %s\n", err)
	//	return
	//}
	//defer func(conn *net.UDPConn) {
	//	err := conn.Close()
	//	if err != nil {
	//		log.Printf("Error closing connection: %s\n", err)
	//		return
	//	}
	//}(conn)

	_, err = conn.WriteToUDP(packet, remoteAddr)
	if err != nil {
		log.Printf("Error sending request packet: %s\n", err)
		return
	}

	n, _, err := conn.ReadFromUDP(packet)
	packet = packet[:n]
	opcode := tftp.TFTPOpcode(binary.BigEndian.Uint16(packet[:2]))
	if err != nil {
		log.Printf("Error reading reply from UDP server: %s\n", err)
		return
	}

	if opcode == tftp.TFTPOpcodeERROR {
		//process error packet
		var errPack tftp.TFTPError
		err := errPack.ReadFromBytes(packet)
		if err != nil {
			errSt := fmt.Sprintf("error packet received... code: %d message: %s\n", errPack.ErrorCode, errPack.ErrorMessage)
			log.Println(errSt)
			return nil, errors.New(errSt)
		}
	}

	if opcode != tftp.TFTPOpcodeOACK {
		errSt := fmt.Sprintf("returned packet opcode is neither OACK or ACK.. opcode: %d packet_t: %s\n", opcode, tftp.TFTPOpcode(opcode))
		log.Println(errSt)
		return nil, errors.New(errSt)
	}

	// If checks clear, begin sliding window protocol
	StartDataClientXfer(conn)
	if err != nil {
		return nil, err
	}
	return
}

func StartDataClientXfer(conn *net.UDPConn) {
	pckBfr := make([]byte, 1024)
	dataBuffer := make([]byte, 0)
	n := 0
	var raddr *net.UDPAddr
	var err error
	closeXfer := false

	for {
		n, raddr, err = conn.ReadFromUDP(pckBfr)
		pckBfr = pckBfr[:n]
		opcode := tftp.TFTPOpcode(binary.BigEndian.Uint16(pckBfr[:2]))
		switch opcode {
		case tftp.TFTPOpcodeDATA:
			var dataPack tftp.TFTPData
			err = dataPack.ReadFromBytes(pckBfr)
			dataBuffer = append(dataBuffer, dataPack.Data...)
			ackPack := tftp.NewTFTPAcknowledgement(dataPack.BlockNumber)
			pckBfr, _ = ackPack.ToBytes()
			_, err = conn.WriteToUDP(pckBfr, raddr)
		case tftp.TFTPOpcodeERROR:
			var errPack tftp.TFTPError
			_ = errPack.ReadFromBytes(pckBfr)
			er := fmt.Sprintf("error packet received... code: %d message: %s\n", errPack.ErrorCode, errPack.ErrorMessage)
			err = errors.New(er)
			closeXfer = true
		default:
			closeXfer = true
		}

		if closeXfer {
			break
		}
	}
	if err != nil {
		log.Printf("Error in data transfer: %s\n", err)
		return
	}

	return
}

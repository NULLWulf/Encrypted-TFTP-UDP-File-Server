package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
)

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
func StartImgReqTFTP(url string) (tData []byte, err error) {

	packet := make([]byte, 512)
	// Make a new request packet
	reqPack, _ := tftp.NewTFTPRequest([]byte(url), []byte("octet"), 0, nil)
	packet, _ = reqPack.ToBytes()
	remoteAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:7501")

	if err != nil {
		// TODO Make error more elegant
		log.Printf("Error attempting to resolve host address: %s\n", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		// TODO Make error more elegant
		log.Printf("Error connecting to host: %s\n", err)
		return
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %s\n", err)
			return
		}
	}(conn)

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
	tData, err = StartDataTransferTFTP(conn)
	if err != nil {
		return nil, err
	}
	return
}

func StartDataTransferTFTP(conn *net.UDPConn) (data []byte, err error) {
	return
}

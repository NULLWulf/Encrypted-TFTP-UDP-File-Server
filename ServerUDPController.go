package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

func main() {
	fmt.Println("Starting TFTP server on port", Port)

	addr, err := net.ResolveUDPAddr("udp", string(rune(Port)))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}

	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	for {
		buf := make([]byte, 516)
		_, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("error reading from UDP:", err)
			continue
		}

		go handleRequest(conn, addr, buf)
	}
}

func handleRequest(conn *net.UDPConn, addr *net.UDPAddr, buf []byte) {
	switch tftp.TFTPOpcode(binary.BigEndian.Uint16(buf[:2])) {
	case tftp.TFTPOpcodeRRQ:
		fmt.Println("Received RRQ")
		break
	case tftp.TFTPOpcodeWRQ:
		fmt.Println("Received WRQ")
		break
	case tftp.TFTPOpcodeDATA:
		fmt.Println("Received DATA")
		break
	case tftp.TFTPOpcodeACK:
		fmt.Println("Received ACK")
		break
	case tftp.TFTPOpcodeERROR:
		fmt.Println("Received ERROR")
		break
	case tftp.TFTPOpcodeTERM:
		fmt.Println("Received TERM")
		break
	default:

	}
}

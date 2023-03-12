package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

// handleConnectipnUDP handles a single udp "connection"
func handleConnectionUDP(conn *net.UDPConn) {
	buf := make([]byte, 1024)
	for {
		// read message
		n, raddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("Error reading message:", err)
			continue
		}
		// decode message
		msg := buf[:n]
		handleRequest(conn, raddr, msg)
	}
}

func RunServerMode() {
	addr := &net.UDPAddr{IP: net.IPv4zero, Port: Port}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("Error starting server:", err)
		return
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing the connection:", err)
		}
	}(conn)

	log.Printf("Server listening on: %s\n", addr)

	handleConnectionUDP(conn)
}

func handleRequest(conn *net.UDPConn, addr *net.UDPAddr, buf []byte) {
	switch tftp.TFTPOpcode(binary.BigEndian.Uint16(buf[:2])) {
	case tftp.TFTPOpcodeRRQ:
		log.Println("Received RRQ")
		var req tftp.Request
		err := req.Parse(buf)
		if err != nil {
			log.Println("Error parsing request:", err)
			return
		}
		// Encode the Person struct as JSON.
		log.Println(string(req.Filename))
		jsonBytes, err := json.Marshal(req)
		if err != nil {
			fmt.Println("Error encoding JSON:", err)
			return
		}
		// Print the JSON as a string.
		fmt.Println(string(jsonBytes))

		break
	case tftp.TFTPOpcodeWRQ:
		log.Println("Received WRQ")
		break
	case tftp.TFTPOpcodeDATA:
		log.Println("Received DATA")
		break
	case tftp.TFTPOpcodeACK:
		log.Println("Received ACK")
		break
	case tftp.TFTPOpcodeERROR:
		log.Println("Received ERROR")
		break
	case tftp.TFTPOpcodeTERM:
		log.Println("Received TERM")
		break
	default:

	}
}

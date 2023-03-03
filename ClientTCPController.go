package main

import (
	"CSC445_Assignment2/tftp"
	"fmt"
	"log"
	"math/rand"
	"net"
)

const BlockSize = 512

func tcpClientImageRequest(url string) (imgBuff []byte, err error) {

	conn, err := net.Dial("tcp", Address)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %s", err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing connection:", err)
			return
		}
	}(conn)

	buffer, err := initTFTPReqPacket(url)
	_, err = conn.Write(buffer)
	if err != nil {
		log.Println("Error sending message:", err)
		return
	}
	n, err := conn.Read(buffer)
	oack := &tftp.TFTPOptionAcknowledgement{}
	err = oack.ReadFromBytes(buffer[:n])
	if err != nil {
		log.Println("Error receiving reply:", err)
		return
	}
	
	return imgBuff, nil
}

func initTFTPReqPacket(url string) ([]byte, error) {
	reqOpt := make(map[string]string)
	reqOpt["blksize"] = string(rune(BlockSize))
	reqOpt["maxWindowSize"] = "1"
	reqOpt["randKey"] = string(rune(rand.Intn(1000)))
	req, err := tftp.NewTFTPRequest([]byte(url), []byte("octet"), 0, reqOpt)
	if err != nil {
		return nil, fmt.Errorf("error creating TFTP request: %s", err)
	}
	pax, err := req.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("error converting TFTP request to bytes: %s", err)
	}
	return pax, nil
}

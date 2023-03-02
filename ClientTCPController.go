package main

import (
	"CSC445_Assignment2/tftp"
	"log"
	"math/rand"
	"net"
)

const BlockSize = 512

func tcpClientImageRequest(url string) (err error) {

	reqOpt := make(map[string]string)
	reqOpt["blksize"] = string(rune(BlockSize))
	reqOpt["maxWindowSize"] = "1"
	reqOpt["randKey"] = string(rune(rand.Intn(1000)))
	req, _ := tftp.NewTFTPRequest([]byte(url), []byte("octet"), 0, reqOpt)
	reqB, _ := req.ToBytes()
	conn, err := net.Dial("tcp", Address)
	if err != nil {
		log.Printf("Error connecting to server: %s\n", err)
		return err
	}
	defer conn.Close()

	if err != nil {
		log.Println("Error connecting to server: ", err)
		return
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing connection:", err)
		}
	}(conn)

	//msg = XOREncode(msg)
	// send message and measure round-trip time
	_, err = conn.Write(reqB)
	if err != nil {
		log.Println("Error sending message:", err)
		return
	}
	reply := make([]byte, BlockSize+2)
	n, err := conn.Read(reply)
	oack := &tftp.TFTPOptionAcknowledgement{}
	err = oack.ReadFromBytes(reply[:n])
	if err != nil {
		log.Println("Error receiving reply:", err)
		return
	}

	reply = make([]byte, BlockSize+2)
	n, err = conn.Read(reply)
	ack := &tftp.TFTPAcknowledgement{}
	err = p.ReadFromBytes(reply[:n])
	if err != nil {
		log.Println("Error receiving reply:", err)
		return
	}

	return nil
}

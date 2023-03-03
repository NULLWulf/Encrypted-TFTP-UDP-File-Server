package main

import (
	"CSC445_Assignment2/tftp"
	"log"
	"net"
)

func listenAndServeTFTP() (err error) {
	port := "69"
	conn, err := net.ListenPacket("udp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatalf("Error listening on port %s: %v", port, err)
		return
	}
	defer func(conn net.PacketConn) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(conn)

	buf := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			return err
		}

		packet := buf[:n]
		tftp.HandlePacket(packet)

	}
}

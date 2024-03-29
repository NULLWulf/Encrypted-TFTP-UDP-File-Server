package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"log"
	"math/big"
	"net"
)

// NewTFTPClient method constructs a new TFTPProtocol struct
func NewTFTPClient() (*TFTPProtocol, error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", Address)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return nil, err
	}

	return &TFTPProtocol{conn: conn, raddr: remoteAddr, xferSize: 0}, nil
}

// RequestFile method sends a request packet to the server and begins the transfer process
func (c *TFTPProtocol) RequestFile(url string) (data []byte, transTime float64, err error) {
	defer func() { // Recover from panic in case of key generation failure
		if r := recover(); r != nil {
			log.Printf("Panic recovered in RequestFile: %v", r)
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	log.Printf("Starting RequestFile\n")

	c.dhke = new(DHKESession) // Make a new DHKE session
	c.dhke.GenerateKeyPair()  // Generate the key pair

	options := make(map[string][]byte)       // Create a map for the options
	options["keyx"] = c.dhke.pubKeyX.Bytes() // Set the x public key to the map
	options["keyy"] = c.dhke.pubKeyY.Bytes() // Set the y public key to the map

	reqPack, _ := tftp.NewReq([]byte(url), []byte("octet"), 0, options)
	packet, _ := reqPack.ToBytes()

	c.SetProtocolOptions(options, 0) // Sets the protocol options
	_, err = c.conn.Write(packet)    // Sends the request packet

	if err != nil {
		log.Printf("Error sending request packet: %s\n", err)
		return nil, 0, err
	}
	err = c.preDataTransfer() // Starts the transfer process
	c.EndTime()               // Ends the timer
	if err != nil {
		log.Printf("Error in preDataTransfer: %s\n", err)
		return nil, 0, err
	}
	data = c.rebuildData() // Rebuilds the data from the data packets received
	return data, 0, nil
}

// PreDataTransfer method handles the OACK packet and any error packets
func (c *TFTPProtocol) preDataTransfer() error {
	packet := make([]byte, 1024)
	err := error(nil)
	n, tErr := c.conn.Read(packet)
	if tErr != nil {
		return fmt.Errorf("error reading packet: %s", err)
	}
	packet = packet[:n]
	code := binary.BigEndian.Uint16(packet[:2])
	switch tftp.TFTPOpcode(code) {
	case tftp.TFTPOpcodeERROR:
		log.Printf("Error packet received: %s\n", packet)
		var errPack tftp.Error
		errPack.Parse(packet)
		// Sleep for 1 second to allow the server to close the connection
		panic("Received Error Packet When Expecting OACK, assumed key exchange failed")
	case tftp.TFTPOpcodeTERM:
		log.Printf("Received Termination packet from server: %s\n", c.conn.RemoteAddr().String())
		panic("Received Termination Packet When Expecting OACK, assumed key exchange failed")

	case tftp.TFTPOpcodeOACK:
		log.Printf("Received oack from server: %s\n", c.conn.RemoteAddr().String())
		oackPack := new(tftp.OptionAcknowledgement)
		err = oackPack.Parse(packet) // Parse the options packet which contains the server key pair
		if err != nil {
			c.sendError(0, "Error parsing OACK packet")
			panic("Error parsing OACK packet")
		}
		px, py := new(big.Int), new(big.Int) // create new big ints for the x and y values
		px.SetBytes(oackPack.KeyX)           // convert x,y values into Big Ints
		py.SetBytes(oackPack.KeyY)
		c.dhke.sharedKey, err = c.dhke.generateSharedKey(px, py) // generate the shared key
		if err != nil {                                          // if there is an error, send an error packet and panic
			c.sendError(0, "Error generating shared key")
			panic("Error generating shared key")
		}
		log.Printf("Shared Key: %d\n", crc32.ChecksumIEEE(c.dhke.sharedKey))

		err, _ = c.TftpClientTransferLoop(c.conn) // starts the transfer loop, returns error and bool
		// signifying if the transfer is complete or not, and error would terminate the transfer
		if err != nil {
			return fmt.Errorf("error in transfer loop: %s", err)
		}
	default:
		log.Printf("Received unexpected packet: %s\n", packet)
	}

	//
	return nil
}

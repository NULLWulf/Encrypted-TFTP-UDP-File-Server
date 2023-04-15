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

import "C"
import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"net"
)

// NewTFTPClient method constructs a new TFTPProtocol struct
// TODO - add in a parameter for the port number and address
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
// / and returns the data
func (c *TFTPProtocol) RequestFile(url string) (err error, data []byte, transTime float64) {
	// make a map with a key field
	options := make(map[string][]byte)
	c.dhke = new(DHKESession)
	c.dhke.GenerateKeyPair()
	options["keyx"] = c.dhke.pubKeyX.Bytes() // set the x public key to the map
	options["keyy"] = c.dhke.pubKeyY.Bytes() // set the y public key to the map
	// random 256 bit key
	// set the key to the map
	reqPack, _ := tftp.NewReq([]byte(url), []byte("octet"), 0, options)
	packet, _ := reqPack.ToBytes()
	c.SetProtocolOptions(options, 0) // sets the protocol options
	c.StartTime()                    // starts the timer
	n, err := c.conn.Write(packet)   // sends the request packet
	c.ADto(n)                        // adds the number of bytes sent to the total bytes sent running total
	if err != nil {
		log.Printf("Error sending request packet: %s\n", err)
		return err, nil, 0
	}
	err = c.preDataTransfer() // starts the transfer process
	c.EndTime()               // ends the timer
	if err != nil {
		log.Printf("Error in preDataTransfer: %s\n", err)
		return err, nil, 0
	}
	data = c.rebuildData()    // rebuilds the data from the data packets received
	c.DisplayStats(len(data)) // displays the stats regarding the transfer

	return nil, data, 0
}

// preDataTransfer method handles the OACK packet and any error packets
func (c *TFTPProtocol) preDataTransfer() error {
	packet := make([]byte, 516)
	cont := true
	err := error(nil)

	// loop until we receive an error, term or oack packet
	// any errors or terms at this stage simply terminate the connection
	for {
		n, tErr := c.conn.Read(packet)
		c.ADti(n)
		if tErr != nil {
			return fmt.Errorf("error reading packet: %s", err)
		}
		packet = packet[:n]
		code := binary.BigEndian.Uint16(packet[:2])
		switch tftp.TFTPOpcode(code) {
		case tftp.TFTPOpcodeERROR:
			log.Printf("Error packet received: %s\n", packet)
			var errPack tftp.Error
			err = errPack.Parse(packet)
			if err != nil {
				return err
			}
			return fmt.Errorf("error code %d: %s", errPack.ErrorCode, errPack.ErrorMessage)
		case tftp.TFTPOpcodeTERM:
			log.Printf("Server terminated connection: %s\n", packet)
			return fmt.Errorf("server terminated connection")
		case tftp.TFTPOpcodeOACK:
			log.Printf("Received OACK packet: %s\n", packet)
			var oackPack tftp.OptionAcknowledgement
			err = oackPack.Parse(packet)
			// get the shared secret
			px, py := new(big.Int), new(big.Int)
			px.SetBytes(oackPack.KeyX)
			py.SetBytes(oackPack.KeyY)
			c.dhke.sharedKey = c.dhke.generateSharedKey(px, py)
			log.Printf("Shared Key: %s\n", c.dhke.sharedKey)
			if err != nil {
				return fmt.Errorf("error parsing OACK packet: %s", err)
			}
			err, cont = c.TftpClientTransferLoop(c.conn) // starts the transfer loop, returns error and bool
			// signifying if the transfer is complete or not, and error would terminate the transfer
			if err != nil {
				return fmt.Errorf("error in transfer loop: %s", err)
			}
		default:
			log.Printf("Received unexpected packet: %s\n", packet)
		}
		if cont { // if the transfer is complete, return nil error
			log.Printf("Transfer complete\n")
			return nil
		}
	}
}

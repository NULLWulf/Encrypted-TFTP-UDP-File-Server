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
	"fmt"
	"hash/crc32"
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
	options := make(map[string][]byte)
	c.dhke = new(DHKESession)                // Make a new DHKE session
	c.dhke.GenerateKeyPair()                 // Generate the key pair
	options["keyx"] = c.dhke.pubKeyX.Bytes() // set the x public key to the map
	options["keyy"] = c.dhke.pubKeyY.Bytes() // set the y public key to the map
	reqPack, _ := tftp.NewReq([]byte(url), []byte("octet"), 0, options)
	packet, _ := reqPack.ToBytes()
	c.SetProtocolOptions(options, 0) // sets the protocol options
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
		err = errPack.Parse(packet)
		if err != nil {
			return err
		}
		return fmt.Errorf("error code %d: %s", errPack.ErrorCode, errPack.ErrorMessage)
	case tftp.TFTPOpcodeTERM:
		log.Printf("Received Termination packet from server: %s\n", c.conn.RemoteAddr().String())
		return fmt.Errorf("server terminated connection")
	case tftp.TFTPOpcodeOACK:
		log.Printf("Received oack from server: %s\n", c.conn.RemoteAddr().String())
		var oackPack tftp.OptionAcknowledgement
		err = oackPack.Parse(packet)
		px, py := new(big.Int), new(big.Int)
		px.SetBytes(oackPack.KeyX)
		py.SetBytes(oackPack.KeyY)
		c.dhke.sharedKey, err = c.dhke.generateSharedKey(px, py)
		if err != nil {
			log.Printf("Error generating shared key: %v. Retrying", err.Error())
			return err
		}
		log.Printf("Shared Key: %d\n", crc32.ChecksumIEEE(c.dhke.sharedKey))
		if err != nil {
			return fmt.Errorf("error parsing OACK packet: %s", err)
		}
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

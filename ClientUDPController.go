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
	"errors"
	"fmt"
	"log"
	"net"
)

type TFTPProtocol struct {
	conn              *net.UDPConn
	raddr             *net.UDPAddr
	fileData          *[]byte
	xferSize          uint32
	blockSize         uint16
	windowSize        uint16
	key               []byte
	dataBlocks        []*tftp.Data
	base              uint16                // Base of the window
	nextExpectedBlock uint16                // Next expected block number
	ackBlocks         map[uint16]bool       // Map to keep track of acknowledged blocks
	bufferedBlocks    map[uint16]*tftp.Data // Map to buffer out-of-order blocks
}

func NewTFTPClient() (*TFTPProtocol, error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:7500")
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return nil, err
	}

	return &TFTPProtocol{conn: conn, raddr: remoteAddr, xferSize: 0}, nil
}

func NewTFTPServer() (*TFTPProtocol, error) {
	addr := &net.UDPAddr{IP: net.IPv4zero, Port: Port}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("Error starting server:", err)
		return nil, err
	}
	return &TFTPProtocol{conn: conn, raddr: addr}, nil
}

func (c *TFTPProtocol) Close() error {
	return c.conn.Close()
}

func (c *TFTPProtocol) RequestFile(url string) (tData []byte, err error) {
	options := make(map[string][]byte)
	options["blksize"] = []byte("512")
	options["key"] = tftp.GetRandomKey()

	reqPack, _ := tftp.NewReq([]byte(url), []byte("octet"), 0, options)
	packet, _ := reqPack.ToBytes()
	c.SetProtocolOptions(options, 0)
	_, err = c.conn.Write(packet)
	if err != nil {
		log.Printf("Error sending request packet: %s\n", err)
		return
	}

	n, _, err := c.conn.ReadFromUDP(packet)
	packet = packet[:n]
	fmt.Println("Length of packet: ", len(packet))
	opcode := binary.BigEndian.Uint16(packet[:2])
	if err != nil {
		log.Printf("Error reading reply from UDP server: %s\n", err)
		return
	}

	if opcode == 5 {
		//process error packet
		var errPack tftp.Error
		_ = errPack.Parse(packet)

		errSt := fmt.Sprintf("error packet received... code: %d message: %s\n", errPack.ErrorCode, errPack.ErrorMessage)
		log.Println(errSt)
		return nil, errors.New(errSt)

	}
	var oackPack tftp.OptionAcknowledgement
	err = oackPack.Parse(packet)

	log.Printf(oackPack.String())
	err = c.TftpClientTransferLoop()
	if err != nil {
		return nil, err
	}

	return *c.fileData, nil
}

func (c *TFTPProtocol) TftpClientTransferLoop() error {
	pckBfr := make([]byte, c.blockSize+4)
	c.base = 1
	c.nextExpectedBlock = 1
	c.ackBlocks = make(map[uint16]bool)
	c.bufferedBlocks = make(map[uint16]*tftp.Data)

	for {
		v, _, err := c.conn.ReadFromUDP(pckBfr)
		if err != nil {
			return fmt.Errorf("Failed to read from UDP connection: %v", err)
		}
		pckBfr = pckBfr[:v]
		opcode := tftp.TFTPOpcode(binary.BigEndian.Uint16(pckBfr[:2]))

		switch opcode {
		case tftp.TFTPOpcodeDATA:
			var dataPack tftp.Data
			if err := dataPack.Parse(pckBfr); err == nil {
				c.receiveDataPacket(&dataPack)
			} else {
				return fmt.Errorf("failed to parse data packet: %v", err)
			}

			if len(dataPack.Data) < int(c.blockSize) {
				// End of transfer
				// verify all blocks received
				*c.fileData = tftp.RebuildData(c.dataBlocks)
				break
			}

			// Send ACK for the last contiguous block received
			ackPack := tftp.NewAck(c.base - 1)
			if _, err := c.conn.WriteToUDP(ackPack.ToBytes(), c.raddr); err != nil {
				return fmt.Errorf("failed to write ACK packet to UDP connection: %v", err)
			}

		case tftp.TFTPOpcodeERROR:
			var errPack tftp.Error
			if err := errPack.Parse(pckBfr); err == nil {
				return fmt.Errorf("error packet received: %v", errPack)
			} else {
				return fmt.Errorf("failed to parse error packet: %v", err)
			}

		default:
			// Ignore any other opcodes
		}
	}
}

func (c *TFTPProtocol) SetProtocolOptions(options map[string][]byte, l int) {
	if l != 0 {
		c.SetTransferSize(uint32(l))
	}
	if options["tsize"] != nil && c.xferSize == 0 {
		c.SetTransferSize(binary.BigEndian.Uint32(options["tsize"]))
	}
	if options["blksize"] != nil {
		c.blockSize = binary.BigEndian.Uint16(options["blksize"])
	}
	if options["windowsize"] != nil {
		c.windowSize = binary.BigEndian.Uint16(options["windowsize"])
	}
	if options["key"] != nil {
		c.key = options["key"]
	}
}

func (c *TFTPProtocol) receiveDataPacket(dataPack *tftp.Data) {
	blockNumber := dataPack.BlockNumber
	if blockNumber == c.nextExpectedBlock {
		c.bufferedBlocks[blockNumber] = dataPack
		c.ackBlocks[blockNumber] = true

		for c.bufferedBlocks[c.base] != nil {
			c.dataBlocks = append(c.dataBlocks, c.bufferedBlocks[c.base])
			delete(c.bufferedBlocks, c.base)
			c.base++
			c.nextExpectedBlock++
		}
	} else if blockNumber > c.nextExpectedBlock && blockNumber < c.base+c.windowSize {
		// Buffer out-of-order packet
		c.bufferedBlocks[blockNumber] = dataPack
		c.ackBlocks[blockNumber] = true
	}
}

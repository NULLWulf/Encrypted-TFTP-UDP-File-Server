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
	conn       *net.UDPConn
	raddr      *net.UDPAddr
	fileData   *[]byte
	xferSize   uint32
	blockSize  uint16
	windowSize uint16
	key        []byte
	dataBlocks []*tftp.Data
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
	packet := make([]byte, 516)
	options := make(map[string][]byte)
	options["blksize"] = []byte("512")
	options["key"] = tftp.GetRandomKey()

	reqPack, _ := tftp.NewReq([]byte(url), []byte("octet"), 0, options)
	packet, _ = reqPack.ToBytes()
	c.SetProtocolOptions(options, 0)
	_, err = c.conn.Write(packet)
	if err != nil {
		log.Printf("Error sending request packet: %s\n", err)
		return
	}

	n, _, err := c.conn.ReadFromUDP(packet)
	packet = packet[:n]
	fmt.Println("Length of packet: ", len(packet))
	opcode := tftp.TFTPOpcode(binary.BigEndian.Uint16(packet[:2]))
	if err != nil {
		log.Printf("Error reading reply from UDP server: %s\n", err)
		return
	}

	if opcode == tftp.TFTPOpcodeERROR {
		//process error packet
		var errPack tftp.Error
		err := errPack.Parse(packet)
		if err != nil {
			errSt := fmt.Sprintf("error packet received... code: %d message: %s\n", errPack.ErrorCode, errPack.ErrorMessage)
			log.Println(errSt)
			return nil, errors.New(errSt)
		}
	}

	if opcode != tftp.TFTPOpcodeOACK {
		errSt := fmt.Sprintf("returned packet opcode is neither OACK or ACK.. opcode: %d packet_t: %s\n", opcode, opcode.String())
		log.Println(errSt)
		return nil, errors.New(errSt)
	}

	// Process OACK packet
	var oackPack tftp.OptionAcknowledgement
	err = oackPack.Parse(packet)

	if tsizeBytes, ok := oackPack.Options["tsize"]; ok && c.xferSize == 0 {
		c.SetTransferSize(binary.BigEndian.Uint32(tsizeBytes))
	}

	log.Printf(oackPack.String())
	err = c.StartDataClientXfer()
	if err != nil {
		return nil, err
	}

	return *c.fileData, nil
}

func (c *TFTPProtocol) StartDataClientXfer() (err error) {
	var dataPack tftp.Data
	var errPack tftp.Error
	var v int
	pckBfr := make([]byte, c.blockSize+4)
	var n uint16
	n = 0
	closeXfer := false

	ackPack := tftp.NewAck(0)
	_, err = c.conn.WriteToUDP(ackPack.ToBytes(), c.raddr)
	for {
		v, c.raddr, err = c.conn.ReadFromUDP(pckBfr)
		pckBfr = pckBfr[:v]
		opcode := tftp.TFTPOpcode(binary.BigEndian.Uint16(pckBfr[:2]))
		switch opcode {
		case tftp.TFTPOpcodeDATA:
			err = dataPack.Parse(pckBfr)
			if dataPack.BlockNumber == n {
				c.dataBlocks = append(c.dataBlocks, &dataPack)
				log.Printf("Received data packet: %d\n", dataPack.BlockNumber)
				n++
			}
			ackPack = tftp.NewAck(n)
			_, err = c.conn.WriteToUDP(ackPack.ToBytes(), c.raddr)
		case tftp.TFTPOpcodeERROR:
			_ = errPack.Parse(pckBfr)
			err = fmt.Errorf("error packet received... code: %d message: %s\n", errPack.ErrorCode, errPack.ErrorMessage)
			closeXfer = true
		default:
			closeXfer = true
		}

		if closeXfer {
			break
		}
	}
	if err != nil {
		return
	}

	return
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

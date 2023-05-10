package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"log"
	"math/big"
	"net"
	"sort"
	"time"
)

type TFTPProtocol struct {
	conn            *net.UDPConn          // UDP connection
	raddr           *net.UDPAddr          // Remote address
	xferSize        uint32                // Size of the file to be transferred
	blockSize       uint16                // Block size of the data packets
	windowSize      uint16                //Sliding window size
	key             []byte                // Key
	dataBlocks      []*tftp.Data          //Data packets to be sent
	base            uint16                // Base of the window
	nextSeqNum      uint16                // Next expected block number
	totalFrames     int                   // Total number of frames
	dataThroughIn   int                   // Data throughput in
	dataThroughOut  int                   // Data throughput out
	requestStart    int64                 // Time when the request was sent
	requestEnd      int64                 // Time when the request was received
	receivedPackets map[uint16]*tftp.Data // Received packets
	dhke            *DHKESession          // Diffie Hellman Key Exchange
}

// SetProtocolOptions sets the protocol options for the TFTP protocol
// using static values for the time being
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
	if options["keyy"] != nil {
		v := new(big.Int)
		v.SetBytes(options["keyy"])
		c.key = options["keyy"]
	}
	if options["keyx"] != nil {
		// convert to big int
		v := new(big.Int)
		v.SetBytes(options["keyx"])
		c.key = options["keyx"]
	}

	c.blockSize = 512
	c.windowSize = uint16(WindowSize)
}

func (c *TFTPProtocol) sendError(errCode uint16, errMsg string) {
	log.Printf("Sending error packet: %d %s\n", errCode, errMsg)
	errPack := tftp.NewErr(errCode, []byte(errMsg))
	_, err := c.conn.Write(errPack.ToBytes())
	if err != nil {
		log.Println("Error sending error packet:", err)
		return
	}
}

func (c *TFTPProtocol) sendErrorClient(errCode uint16, errMsg string, raddr *net.UDPAddr) {
	log.Printf("Sending error packet: %d %s\n", errCode, errMsg)
	errPack := tftp.NewErr(errCode, []byte(errMsg))
	_, err := c.conn.WriteToUDP(errPack.ToBytes(), raddr)
	if err != nil {
		log.Println("Error sending error packet:", err)
		return
	}
}

func (c *TFTPProtocol) sendAbort() {
	c.sendError(9, "Aborting transfer")
}

func (c *TFTPProtocol) sendAck(nextSeqNum uint16) {
	ack := tftp.NewAck(nextSeqNum)
	ackPack, _ := encrypt(ack.ToBytes(), c.dhke.aes512Key)
	n, err := c.conn.Write(ackPack)
	c.ADto(n)
	if err != nil {
		log.Println("Error sending ACK packet:", err)
		return
	}
}

// handleErrPacket handles an error packet but currently just sends an error
// back so relying on timeout to close the connection.  Should probably
// implement a proper connection close.
func (c *TFTPProtocol) handleErrPacket(packet []byte) {
	var errPack tftp.Error
	err := errPack.Parse(packet)
	if err != nil {
		log.Println("Error parsing error packet:", err)
		c.sendError(22, "Error parsing error packet")
		return
	}
	c.sendError(errPack.ErrorCode, string(errPack.ErrorMessage))
	return
}

func (c *TFTPProtocol) SetTransferSize(size uint32) {
	c.xferSize = size
}

// appendFileDate appends the file date to the data packet
// and also keeps track of duplicate packets and discards \
// any already stored.  duplicate packets are checked via a
// struct in the TFTP protocol struct
func (c *TFTPProtocol) appendFileDate(data *tftp.Data) bool {
	// Check if the packet is already stored
	if _, exists := c.receivedPackets[data.BlockNumber]; exists {
		log.Println("Duplicate packet, discarding")
		return false
	} else {
		//log.Println("New packet, storing")
	}
	c.receivedPackets[data.BlockNumber] = data
	c.totalFrames++
	return true
}

// ADto is the cumulative data throughput out in bytes
func (c *TFTPProtocol) ADto(n int) {
	c.dataThroughOut += n
}

// ADti is the cumulative data throughput in in bytes
func (c *TFTPProtocol) ADti(n int) {
	c.dataThroughIn += n
}

func (c *TFTPProtocol) Close() error {
	return c.conn.Close()
}

func (c *TFTPProtocol) rebuildData() []byte {
	var data []byte
	keys := make([]int, 0, len(c.receivedPackets))

	// Get and sort the keys
	for k := range c.receivedPackets {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	// Rebuild the data using the sorted keys
	for _, key := range keys {
		data = append(data, c.receivedPackets[uint16(key)].Data...)
	}
	return data
}

func (c *TFTPProtocol) StartTime() {
	// Start the protocol
	c.requestStart = time.Now().UnixNano()
}

func (c *TFTPProtocol) EndTime() {
	// End the protocol
	c.requestEnd = time.Now().UnixNano()
}

// DisplayStats displays the stats for the protocol
// such as throughput, total frames, total bytes, Mbps
func (c *TFTPProtocol) DisplayStats(n int) {
	log.Println("Total frames received:", c.totalFrames)
	log.Println("Total bytes received:", c.dataThroughIn)
	log.Println("Total bytes sent:", c.dataThroughOut)
	nanos := time.Duration(c.requestEnd - c.requestStart)
	bytesToMegaBit := (float64(c.dataThroughIn+c.dataThroughOut) * 8) / 1000000
	through := bytesToMegaBit / nanos.Seconds()
	log.Println("Raw throughput i/o: ", through, "Mbps")
	log.Println("Perceived Throughput: ", float64(n*8/1000000)/nanos.Seconds(), "Mbps")
}

func PrepareData(data []byte, blockSize int, xorKey []byte) (dataQueue []*tftp.Data, err error) {
	// Create a slice of TFTPData packets
	blocks := len(data) / blockSize
	if len(data)%blockSize != 0 {
		blocks++
	}
	log.Printf("Length of data: %d, Block size: %d, Blocks: %d", len(data), blockSize, blocks)
	dataQueue = make([]*tftp.Data, blocks)

	// Populate the slice with TFTPData packets
	for i := 0; i < blocks; i++ {
		// Calculate the start and end indices of the data
		start := i * blockSize
		end := start + blockSize
		//log.Printf("[%d:%d]", start, end)
		if end > len(data) {
			end = len(data)
		}
		// Create the TFTPData packet
		// data que append
		dataQueue[i], err = tftp.NewData(uint16(i)+1, data[start:end], xorKey)
		//log.Printf("Data packet %d: %v", i, dataQueue[i].BlockNumber)
		//log.Printf("\n-----------------\nBuild data packet block number: %d\nFirst 10 Bytes: %v\nLength %d\n-----------------\n", dataQueue[i].BlockNumber, dataQueue[i].Data[0:10], len(dataQueue[i].Data))

		if err != nil {
			return
		}
	}
	log.Printf("Finished preparing data, %d blocks", len(dataQueue))
	return
}

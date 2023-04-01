package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"log"
	"net"
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
	retries         []int                 // Number of retries for each block
	retryCount      int                   // Number of retries for the current block
	maxRetries      int                   // Maximum number of retries
	backoff         int                   // Backoff time
	timeout         int                   // Timeout
	totalFrames     int                   // Total number of frames
	dataThroughIn   int                   // Data throughput in
	dataThroughOut  int                   // Data throughput out
	requestStart    int64                 // Time when the request was sent
	requestEnd      int64                 // Time when the request was received
	receivedPackets map[uint16]*tftp.Data // Received packets
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

	c.key = []byte("1234567890123456")
	c.blockSize = 512
	c.windowSize = 4
}

func (c *TFTPProtocol) sendError(errCode uint16, errMsg string) {
	log.Printf("Sending error packet: %d %s\n", errCode, errMsg)
	errPack := tftp.NewErr(errCode, []byte(errMsg))
	n, err := c.conn.Write(errPack.ToBytes())
	if err != nil {
		log.Println("Error sending error packet:", err)
		return
	}
	c.ADto(n)
}

func (c *TFTPProtocol) sendAbort() {
	c.sendError(9, "Aborting transfer")
}

func (c *TFTPProtocol) sendAck(nextSeqNum uint16) {
	ack := tftp.NewAck(nextSeqNum)
	n, err := c.conn.Write(ack.ToBytes())
	c.ADto(n)
	log.Printf("Sending ACK packet: %d\n", ack.BlockNumber)
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

func (c *TFTPProtocol) appendFileDate(data *tftp.Data) {
	// Check if the packet is already stored
	if _, exists := c.receivedPackets[data.BlockNumber]; exists {
		log.Println("Duplicate packet, discarding")
		return
	}
	c.receivedPackets[data.BlockNumber] = data
	c.totalFrames++
	return
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
	for i := 1; i <= c.totalFrames; i++ {
		data = append(data, c.receivedPackets[uint16(i)].Data...)
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

func (c *TFTPProtocol) DisplayStats() {
	log.Println("Total frames received:", c.totalFrames)
	log.Println("Total bytes received:", c.dataThroughIn)
	log.Println("Total bytes sent:", c.dataThroughOut)
	nanos := time.Duration(c.requestEnd - c.requestStart)
	bytesToMegaBit := float64(c.dataThroughIn+c.dataThroughOut) / 8
	through := bytesToMegaBit / nanos.Seconds()
	log.Println("Raw throughput: ", through, "Mbps")
}

func PrepareData(data []byte, blockSize int) (dataQueue []*tftp.Data, err error) {
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
		if end > len(data) {
			end = len(data)
		}
		// Create the TFTPData packet
		// data que append
		dataQueue[i], err = tftp.NewData(uint16(i)+1, data[start:end])
		// on a something percent chance duplicate the packet and add it to the queue
		// and decrement the drops counter and increase the probability of dropping
		// Randomly duplicate a data block before adding it to the queue

		if err != nil {
			return
		}
	}
	return
}

package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"log"
	"math/big"
	"net"
	"time"
)

// handleRRQ is the entry point for the sender side of the TFTP protocol
// when a RRQ is received.  It parses the request, sends an OACK, and
// enters the sender loop.
func (c *TFTPProtocol) handleRRQ(addr *net.UDPAddr, buf []byte) {
	var req tftp.Request
	err := req.Parse(buf) // Parse the request
	if err != nil {
		parseErr := fmt.Errorf("error parsing request: %s", err)
		log.Printf("Error parsing request: %s\n", parseErr)
		c.sendError(5, parseErr.Error())
		return
	}
	log.Printf("Received %d bytes from %s for file %s \n", len(buf), addr.String(), string(req.Filename))
	//err, img := IQ.AddNewAndReturnImg(string(req.Filename)) //Add the image to the image queue
	file, err := ProxyRequest(string(req.Filename))

	if err != nil {
		c.sendError(5, "File not found")
		return
	}
	c.dhke = new(DHKESession) // Create a new DHKE session
	c.dhke.GenerateKeyPair()  // Generate a new key pair for server
	// get bytes conver to big int
	px, py := new(big.Int), new(big.Int)                     // Create new big ints to hold clients public keys
	px.SetBytes(req.Options["keyx"])                         // Set the big ints to the clients public keys
	py.SetBytes(req.Options["keyy"])                         // Set the big ints to the clients public keys
	c.dhke.sharedKey, err = c.dhke.generateSharedKey(px, py) // Generate the shared key
	if err != nil {
		log.Printf("Error generating shared key: %v\n", err.Error())
		c.sendErrorClient(11, "Error generating shared key", addr)
		return
	}
	c.SetProtocolOptions(req.Options, 0) //Set the protocol options
	log.Printf("Shared Key Chechksum %d\n", crc32.ChecksumIEEE(c.dhke.sharedKey))
	// Lazy interface to new option packets
	opAck2 := tftp.OptionAcknowledgement{
		Opcode: tftp.TFTPOpcodeOACK,
		KeyX:   c.dhke.pubKeyX.Bytes(),
		KeyY:   c.dhke.pubKeyY.Bytes(),
	}

	//c.dataBlocks, err = PrepareData(file, int(c.blockSize), c.dhke.sharedKey) //Prepare the data blocks
	c.dataBlocks, err = PrepareData(file, int(c.blockSize), c.dhke.aes512Key) //Prepare the data blocks
	if err != nil {
		c.sendErrorClient(5, "Error preparing data blocks", addr)
		return
	}
	_, err = c.conn.WriteToUDP(opAck2.ToBytes(), addr) //Send the OACK
	if err != nil {
		c.sendErrorClient(6, "Error writing to UDP", addr)
		return
	}
	err = c.sender(addr)
	if err != nil {
		log.Printf("Error sending file: %v\n", err.Error())
		c.sendErrorClient(5, "Error sending file", addr)
		return
	}
}

// sender is the main loop for the sender side of the TFTP protocol
// It sends data blocks and waits for ACKs.  If an ACK is not received
// within the timeout period, the data block is resent.  If an error
// occurs, the error is logged and the loop is exited.
func (c *TFTPProtocol) sender(addr *net.UDPAddr) error {
	var ack tftp.Ack
	log.Println("Starting sender transfer TFTP loop")
	packet := make([]byte, 1024)                                     //Byte slice "buffer"
	base, nextSeqNum := 1, 1                                         //Initialize the base, next sequence number, and drop probability
	tOuts, mDelay, iDelay := 0, 30*time.Second, 500*time.Millisecond //Initialize the timeout counter, max delay, and initial delay
	delay := iDelay                                                  // set initial to delay to current delay value
	n, _ := c.conn.Read(packet)                                      //Read the initial ACK
	log.Printf("Initial ACK received: %v\n", n)
	packet = packet[:n] //Trim the packet to the size of the data received
	//packet = tftp.Xor(packet, c.dhke.aes512Key)                      //XOR the packet with the shared key
	err := ack.Parse(packet) // Parse the ACK
	//packet, _ = decrypt(packet, c.dhke.aes512Key)
	log.Printf("Initial ACK received: %v\n", ack)
	if err != nil {
		return errors.New("error parsing ack packet: " + err.Error())
	}
	if ack.BlockNumber != 0 { //Check if the block number is 0 got initial ACK
		c.sendError(3, "Expected initial block number to be 0")
		return errors.New("error parsing ack packet: block number should be 0, expecting initial block")
	}

	// Loop until all data blocks have been sent and acknowledged
	// Send packets within the window size
	for nextSeqNum < base+WindowSize && nextSeqNum <= len(c.dataBlocks) {
		packet = c.dataBlocks[nextSeqNum-1].ToBytes() //Get the data block and convert it to a byte sliceft
		packet, _ = encrypt(packet, c.dhke.aes512Key)
		n, err = c.conn.WriteToUDP(packet, addr) //Send the data block
		c.ADto(n)                                // Add to the total bytes sent to running outgoing total
		nextSeqNum++                             //Increment the next sequence number
		n, _ = c.conn.Read(packet)               //Read the ACK
		packet = packet[:n]                      //Trim the packet to the size of the data received
		packet, _ = decrypt(packet, c.dhke.aes512Key)
		if nErr, ok := err.(net.Error); ok && nErr.Timeout() { //Check if the error is a timeout error
			log.Printf("Timeout, resending unacknowledged packets\n") //If it is a timeout error, log it and increment the timeout counter
			tOuts++                                                   //Increment the timeout counter
			delay = iDelay * (1 << tOuts)                             // 2^tOuts
			if delay > mDelay {                                       // If the delay is greater than the max delay, set the delay to the max delay
				delay = mDelay
			}
			c.conn.SetReadDeadline(time.Now().Add(delay)) //Set the read deadline to the current time plus the delay
			if tOuts >= 5 {
				log.Println("Closing connection due to 5 consecutive unacknowledged packets")
				return fmt.Errorf("connection closed after 5 consecutive unacknowledged packets")
			}
			continue // Continue to the next iteration of the loop while not incrementing the next sequence number
		}

		tOuts = 0 // Reset consecutive timeouts counter when an ACK is received

		opcode := binary.BigEndian.Uint16(packet[:2]) //Get the opcode from the packet
		switch tftp.TFTPOpcode(opcode) {
		case tftp.TFTPOpcodeACK: //If the opcode is an ACK
			err = ack.Parse(packet) //Parse the ACK
			if err != nil {
				log.Printf("Error parsing ACK packet: %s\n", err)
			}
			//log.Printf("Received ACK for packet %d\n", ack.BlockNumber)
			if ack.BlockNumber >= uint16(base) { //If the block number is greater than or equal to the base number
				base = int(ack.BlockNumber + 1) //Set the base to the block number plus 1
			}
		default: // Default case for unexpected packets
			log.Printf("Received unexpected packet: %v\n", packet)
			log.Printf("Window size: %d, base: %d, nextSeqNum: %d\n", WindowSize, base, nextSeqNum)
		}

	}

	// Check if all packets have been sent and acknowledged
	if base > len(c.dataBlocks) {
		log.Printf("All packets sent and acknowledged\n")
		return nil
	}

	return nil
}

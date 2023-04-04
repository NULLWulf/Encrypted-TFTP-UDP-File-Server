package main

import (
	"CSC445_Assignment2/tftp"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/rand"
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
	log.Printf("Received %d bytes from %s for file %s \n", len(buf), addr, string(req.Filename))
	err, img := IQ.AddNewAndReturnImg(string(req.Filename)) //Add the image to the image queue
	if err != nil {
		c.sendError(5, "File not found")
		return
	}
	c.SetProtocolOptions(req.Options, 0)                                         //Set the protocol options
	opAck := tftp.NewOpt1(c.blockSize, c.xferSize, c.blockSize, []byte("octet")) //Create the OACK packet
	c.dataBlocks, err = PrepareData(img, int(c.blockSize), c.key)                //Prepare the data blocks
	if err != nil {
		c.sendError(5, "Error preparing data blocks")
		return
	}
	n, err := c.conn.WriteToUDP(opAck.ToBytes(), addr) //Send the OACK
	c.ADto(n)
	if err != nil {
		c.sendError(5, "Error sending OACK")
		return
	}
	err = c.sender(addr)

	if err != nil {
		c.sendError(5, "Illegal TFTP operation")
		return
	}
}

// sender is the main loop for the sender side of the TFTP protocol
// It sends data blocks and waits for ACKs.  If an ACK is not received
// within the timeout period, the data block is resent.  If an error
// occurs, the error is logged and the loop is exited.
func (c *TFTPProtocol) sender(addr *net.UDPAddr) error {
	defer c.conn.SetReadDeadline(time.Time{})
	var ack tftp.Ack
	log.Println("Starting sender transfer TFTP loop")
	packet := make([]byte, 520)                                      //Byte slice "buffer"
	base, nextSeqNum, dropProb := 1, 1, .20                          //Initialize the base, next sequence number, and drop probability
	tOuts, mDelay, iDelay := 0, 30*time.Second, 500*time.Millisecond //Initialize the timeout counter, max delay, and initial delay
	delay := iDelay                                                  // set initial to delay to current delay value
	n, _ := c.conn.Read(packet)                                      //Read the initial ACK
	c.ADti(n)                                                        // Add to the total bytes received to running outgoing total
	packet = packet[:n]                                              //Trim the packet to the size of the data received
	err := ack.Parse(packet)                                         // Parse the ACK
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
		if rand.Float64() < dropProb && DropPax && nextSeqNum > WindowSize { //DropPax is a global variable that is set to true if the user wants to simulate packet loss
			// uses probability of 0.2 to drop a packet
			log.Printf("Dropped packet %d\n", nextSeqNum)
			nextSeqNum-- //Decrement the next sequence number so that the next packet sent will be the same
			continue     //Continue to the next iteration of the loop
		}

		packet = c.dataBlocks[nextSeqNum-1].ToBytes() //Get the data block and convert it to a byte slice
		n, err = c.conn.WriteToUDP(packet, addr)      //Send the data block
		c.ADto(n)                                     // Add to the total bytes sent to running outgoing total
		log.Printf("Sending packet %d, %v", nextSeqNum, len(packet))
		nextSeqNum++ //Increment the next sequence number

		n, _ = c.conn.Read(packet)                             //Read the ACK
		c.ADti(n)                                              // Add to the total bytes received to running outgoing total
		if nErr, ok := err.(net.Error); ok && nErr.Timeout() { //Check if the error is a timeout error
			log.Printf("Timeout, resending unacknowledged packets\n") //If it is a timeout error, log it and increment the timeout counter
			tOuts++                                                   //Increment the timeout counter
			// Calculate the delay for the next retry using exponential backoff
			delay = iDelay * (1 << tOuts) // 2^tOuts
			if delay > mDelay {           // If the delay is greater than the max delay, set the delay to the max delay
				delay = mDelay
			}
			c.conn.SetReadDeadline(time.Now().Add(delay)) //Set the read deadline to the current time plus the delay
			if tOuts >= 5 {
				log.Println("Closing connection due to 5 consecutive unacknowledged packets")
				return fmt.Errorf("connection closed after 5 consecutive unacknowledged packets")
			}
			continue // Continue to the next iteration of the loop
		}

		tOuts = 0 // Reset consecutiveTimeouts counter when an ACK is received

		packet = packet[:n]                           //Trim the packet to the size of the data received
		opcode := binary.BigEndian.Uint16(packet[:2]) //Get the opcode from the packet
		switch tftp.TFTPOpcode(opcode) {
		case tftp.TFTPOpcodeACK: //If the opcode is an ACK
			err = ack.Parse(packet) //Parse the ACK
			if err != nil {
				log.Printf("Error parsing ACK packet: %s\n", err)
			}
			log.Printf("Received ACK for packet %d\n", ack.BlockNumber)
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

/*
sender2 is a modified version of sliding window that attempts to use concurrency to listen for ack packets while sending data packets.
It is not currently used in the program, but it is left here for future reference.
*/
func (c *TFTPProtocol) sender2(addr *net.UDPAddr) error {
	// ... (unchanged code)\/
	// Initialize variables
	var ack tftp.Ack //
	log.Println("Starting sender transfer TFTP loop")
	packet := make([]byte, 520)                            //Byte buffer
	WindowSize, base, nextSeqNum, dropProb := 8, 1, 1, .20 //Window size, base, next sequence number, and drop probability
	n, _ := c.conn.Read(packet)                            //Read the initial ACK
	c.ADti(n)                                              // Add to the total bytes received to running outgoing total
	packet = packet[:n]                                    //Trim the packet to the size of the data received
	err := ack.Parse(packet)                               //
	// set timeout
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	log.Printf("Initial ACK received: %v\n", ack)
	if err != nil {
		return errors.New("error parsing ack packet: " + err.Error())
	}
	if ack.BlockNumber != 0 {
		c.sendError(3, "Expected initial block number to be 0")
		return errors.New("error parsing ack packet: block number should be 0, expecting initial block")
	}

	// Create channels for ACKs and errors
	ackCh := make(chan tftp.Ack, WindowSize)
	errCh := make(chan error)

	// Goroutine to receive ACKs
	go func() {
		for {
			n, _ := c.conn.Read(packet)
			c.ADti(n)
			packet = packet[:n]
			opcode := binary.BigEndian.Uint16(packet[:2])
			switch tftp.TFTPOpcode(opcode) {
			case tftp.TFTPOpcodeACK:
				err := ack.Parse(packet)
				if err != nil {
					errCh <- fmt.Errorf("error parsing ACK packet: %s", err)
					return
				}
				ackCh <- ack
			default:
				errCh <- fmt.Errorf("received unexpected packet: %v", packet)
				return
			}
		}
	}()

	// Send packets within the window size
	for nextSeqNum < base+WindowSize && nextSeqNum <= len(c.dataBlocks) {
		if rand.Float64() < dropProb && DropPax {
			log.Printf("Dropped packet %d\n", nextSeqNum)
			continue // do not increment nextSeqNum when a packet is dropped
		}
		packet = c.dataBlocks[nextSeqNum-1].ToBytes()
		n, err = c.conn.WriteToUDP(packet, addr)
		c.ADto(n)
		log.Printf("Sending packet %d, %v", nextSeqNum, len(packet))
		if err != nil {
			c.sendAbort()
			return fmt.Errorf("error sending data packet: %s", err)
		}
		nextSeqNum++
	}

	// Check for ACKs with a timeout
	select {
	case ack := <-ackCh:
		log.Printf("Received ACK for packet %d\n", ack.BlockNumber)
		if ack.BlockNumber >= uint16(base) {
			base = int(ack.BlockNumber + 1)
		}
	case err := <-errCh:
		log.Printf("Error: %s\n", err)
		return err
	case <-time.After(time.Duration(500) * time.Millisecond):
		// Timeout: resend unacknowledged packets within the window
		log.Printf("Timeout, resending unacknowledged packets\n")
		nextSeqNum = base
	}

	// Check if all packets have been sent and acknowledged
	if base > len(c.dataBlocks) {
		log.Printf("All packets sent and acknowledged\n")
		return nil
	}
	return nil
}

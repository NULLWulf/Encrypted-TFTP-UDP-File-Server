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
	c.SetProtocolOptions(nil, 0)                                                 //Set the protocol options
	opAck := tftp.NewOpt1(c.blockSize, c.xferSize, c.blockSize, []byte("octet")) //Create the OACK packet
	c.dataBlocks, err = PrepareData(img, int(c.blockSize))                       //Prepare the data blocks
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
	//err = c.sender2(addr)

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
	var ack tftp.Ack
	log.Println("Starting sender transfer TFTP loop")
	packet := make([]byte, 520)                                     //Byte slice "buffer"
	base, nextSeqNum, dropProb, consecutiveTimeouts := 1, 1, .20, 0 //Window size, base, next sequence number, and drop probability
	n, _ := c.conn.Read(packet)                                     //Read the initial ACK
	c.ADti(n)                                                       // Add to the total bytes received to running outgoing total
	packet = packet[:n]                                             //Trim the packet to the size of the data received
	err := ack.Parse(packet)                                        // Parse the ACK
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
		if rand.Float64() < dropProb && DropPax { //DropPax is a global variable that is set to true if the user wants to simulate packet loss
			// uses probability of 0.2 to drop a packet
			log.Printf("Dropped packet %d\n", nextSeqNum)
			nextSeqNum--
			continue
		}

		packet = c.dataBlocks[nextSeqNum-1].ToBytes()
		n, err = c.conn.WriteToUDP(packet, addr)
		c.ADto(n)
		log.Printf("Sending packet %d, %v", nextSeqNum, len(packet))
		if err != nil {
			c.sendAbort()
			c.conn.SetReadDeadline(time.Time{})
			return fmt.Errorf("error sending data packet: %s", err)
		}
		nextSeqNum++

		c.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _ = c.conn.Read(packet)
		c.ADti(n)
		if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
			log.Printf("Timeout, resending unacknowledged packets\n")
			consecutiveTimeouts++
			if consecutiveTimeouts >= 5 {
				log.Println("Closing connection due to 5 consecutive unacknowledged packets")
				c.conn.SetReadDeadline(time.Time{})
				return fmt.Errorf("connection closed after 5 consecutive unacknowledged packets")
			}

			continue
		}

		// Reset consecutiveTimeouts counter when an ACK is received
		consecutiveTimeouts = 0
		packet = packet[:n]
		opcode := binary.BigEndian.Uint16(packet[:2])
		switch tftp.TFTPOpcode(opcode) {
		case tftp.TFTPOpcodeACK:
			// may need to check if the ack is for the correct packet
			err = ack.Parse(packet)
			if err != nil {
				log.Printf("Error parsing ACK packet: %s\n", err)
			}
			log.Printf("Received ACK for packet %d\n", ack.BlockNumber)
			if ack.BlockNumber >= uint16(base) {
				base = int(ack.BlockNumber + 1)
			}
		default:
			log.Printf("Received unexpected packet: %v\n", packet)
		}

	}

	// Check if all packets have been sent and acknowledged
	if base > len(c.dataBlocks) {
		log.Printf("All packets sent and acknowledged\n")
		// set read deadline to infinite
		c.conn.SetReadDeadline(time.Time{})
		return nil
	}

	c.conn.SetReadDeadline(time.Time{}) // set read deadline to infinite
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

package tftp

import (
	"encoding/binary"

	"errors"
)

type Ack struct {
	Opcode      TFTPOpcode
	BlockNumber uint16
}

func NewAck(blockNumber uint16) *Ack {
	return &Ack{
		Opcode:      TFTPOpcodeACK,
		BlockNumber: blockNumber,
	}
}

func (ack *Ack) Parse(packet []byte) error {
	// Check that the packet is at least 4 bytes long
	if len(packet) < 4 {
		return errors.New("packet too short")
	}

	// Check that the opcode is valid
	opcode := binary.BigEndian.Uint16(packet[:2])
	if opcode != uint16(TFTPOpcodeACK) {
		return errors.New("invalid opcode")
	}

	// Parse the block number
	blockNumber := binary.BigEndian.Uint16(packet[2:4])

	// Set the fields in the Ack packet
	ack.Opcode = TFTPOpcodeACK
	ack.BlockNumber = blockNumber

	return nil
}

func (ack *Ack) ToBytes() []byte {
	// Allocate a byte slice to hold the packet.
	packet := make([]byte, 4)

	// Set the opcode and block number in the packet.
	binary.BigEndian.PutUint16(packet[:2], uint16(ack.Opcode))
	binary.BigEndian.PutUint16(packet[2:4], ack.BlockNumber)

	return packet
}

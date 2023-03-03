package tftp

import (
	"encoding/binary"
	"errors"
)

// TFTPAcknowledgement represents a TFTP acknowledgement packet.
type TFTPAcknowledgement struct {
	Opcode      TFTPOpcode
	BlockNumber uint16
}

// NewTFTPAcknowledgement creates a new TFTPAcknowledgement object with the given block number.
func NewTFTPAcknowledgement(blockNumber uint16) *TFTPAcknowledgement {
	return &TFTPAcknowledgement{
		Opcode:      TFTPOpcodeACK,
		BlockNumber: blockNumber,
	}
}

// ReadFromBytes reads a TFTPAcknowledgement packet from a byte slice.
func (ack *TFTPAcknowledgement) ReadFromBytes(packet []byte) error {
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

	// Set the fields in the acknowledgement packet
	ack.Opcode = TFTPOpcodeACK
	ack.BlockNumber = blockNumber

	return nil
}

// ToBytes converts a TFTPAcknowledgement packet to a byte slice.
func (ack *TFTPAcknowledgement) ToBytes() ([]byte, error) {
	// Allocate a byte slice to hold the packet.
	packet := make([]byte, 4)

	// Set the opcode and block number in the packet.
	binary.BigEndian.PutUint16(packet[:2], uint16(ack.Opcode))
	binary.BigEndian.PutUint16(packet[2:4], ack.BlockNumber)

	return packet, nil
}

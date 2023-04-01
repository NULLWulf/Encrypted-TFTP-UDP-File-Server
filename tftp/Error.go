package tftp

import (
	"encoding/binary"
	"errors"
)

// Error TFTPError represents a TFTP error packet.
type Error struct {
	Opcode       TFTPOpcode
	ErrorCode    uint16
	ErrorMessage []byte
}

// NewErr creates a new TFTP error packet.
func NewErr(errorCode uint16, errorMessage []byte) *Error {
	return &Error{
		Opcode:       TFTPOpcodeERROR,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	}
}

// ToBytes converts the error packet to a byte slice.
func (err *Error) ToBytes() []byte {
	// Allocate a byte slice to hold the packet.
	packet := make([]byte, 4+len(err.ErrorMessage))
	// Set the opcode and error code in the packet.
	binary.BigEndian.PutUint16(packet[:2], uint16(err.Opcode))
	binary.BigEndian.PutUint16(packet[2:4], err.ErrorCode)
	// Copy the error message into the packet.
	copy(packet[4:], err.ErrorMessage)
	return packet
}

// Parse parses a TFTP error packet.
func (err *Error) Parse(packet []byte) error {
	// Check that the packet is at least 5 bytes long
	if len(packet) < 5 {
		return errors.New("packet too short")
	}

	// Check that the opcode is valid
	opcode := binary.BigEndian.Uint16(packet[:2])
	if opcode != uint16(TFTPOpcodeERROR) {
		return errors.New("invalid opcode")
	}
	// Parse the error code
	errorCode := binary.BigEndian.Uint16(packet[2:4])

	// Parse the error message
	errorMessage := packet[4:]

	// Set the fields in the error packet
	err.Opcode = TFTPOpcodeERROR
	err.ErrorCode = errorCode
	err.ErrorMessage = errorMessage

	return nil
}

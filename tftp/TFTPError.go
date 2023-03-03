package tftp

import (
	"encoding/binary"
	"errors"
)

// TFTPError represents a TFTP error packet.
type TFTPError struct {
	Opcode       TFTPOpcode
	ErrorCode    uint16
	ErrorMessage string
}

// NewTFTPError creates a new TFTPError object with the given error code and message.
func NewTFTPError(errorCode uint16, errorMessage string) *TFTPError {
	return &TFTPError{
		Opcode:       TFTPOpcodeERROR,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	}
}

// ToBytes converts a TFTPError packet to a byte slice.
func (err *TFTPError) ToBytes() ([]byte, error) {
	// Allocate a byte slice to hold the packet.
	packet := make([]byte, 4+len(err.ErrorMessage))

	// Set the opcode and error code in the packet.
	binary.BigEndian.PutUint16(packet[:2], uint16(err.Opcode))
	binary.BigEndian.PutUint16(packet[2:4], err.ErrorCode)

	// Copy the error message into the packet.
	copy(packet[4:], []byte(err.ErrorMessage))

	return packet, nil
}

// ReadFromBytes reads a TFTPError packet from a byte slice.
func (err *TFTPError) ReadFromBytes(packet []byte) error {
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
	errorMessage := string(packet[4:])

	// Set the fields in the error packet
	err.Opcode = TFTPOpcodeERROR
	err.ErrorCode = errorCode
	err.ErrorMessage = errorMessage

	return nil
}

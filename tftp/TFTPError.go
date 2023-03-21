package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// Error TFTPError represents a TFTP error packet.
type Error struct {
	Opcode       TFTPOpcode
	ErrorCode    uint16
	ErrorMessage []byte
}

func NewErr(errorCode uint16, errorMessage []byte) *Error {
	return &Error{
		Opcode:       TFTPOpcodeERROR,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	}
}

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

type OptionAcknowledgement struct {
	Opcode     TFTPOpcode
	Options    map[string][]byte
	Windowsize uint16
	XferSize   uint32
	BlkSize    uint16
	Key        []byte
}

func NewOpt1(windowsize uint16, xfer uint32, blksize uint16, key []byte) *OptionAcknowledgement {
	options := make(map[string][]byte, 4)

	if xfer != 0 {
		options["tsize"] = make([]byte, 4)
		binary.BigEndian.PutUint32(options["tsize"], uint32(xfer))
	}
	if windowsize != 0 {
		options["windowsize"] = make([]byte, 2)
		binary.BigEndian.PutUint16(options["windowsize"], windowsize)
	}
	if blksize != 0 {
		options["blksize"] = make([]byte, 2)
		binary.BigEndian.PutUint16(options["blksize"], blksize)
	}
	if key != nil {
		options["key"] = key
	}
	return &OptionAcknowledgement{
		Opcode:     TFTPOpcodeOACK,
		Options:    options,
		Windowsize: windowsize,
		XferSize:   xfer,
		BlkSize:    blksize,
		Key:        key,
	}
}

func (oack *OptionAcknowledgement) ToBytes() []byte {
	// Create a byte buffer to store the options.
	var buf bytes.Buffer
	for k, v := range oack.Options {
		buf.WriteString(k)
		buf.WriteByte(0)
		buf.Write(v)
		buf.WriteByte(0)
	}

	// Calculate the total length of the packet.
	packetLength := 2 + buf.Len()

	// Allocate a byte slice to hold the packet.
	packet := make([]byte, packetLength)

	// Set the opcode in the first two bytes of the packet.
	binary.BigEndian.PutUint16(packet[:2], uint16(oack.Opcode))

	// Copy the options into the remaining bytes of the packet.
	copy(packet[2:], buf.Bytes())

	return packet
}

func (oack *OptionAcknowledgement) Parse(packet []byte) error {
	if len(packet) < 2 {
		return fmt.Errorf("packet too short")
	}

	// Extract the opcode from the first two bytes of the packet.
	oack.Opcode = TFTPOpcode(binary.BigEndian.Uint16(packet[:2]))

	// Parse the options from the remaining bytes of the packet.
	oack.Options = make(map[string][]byte)
	optionBytes := packet[2:]
	for len(optionBytes) > 0 {
		// Find the next null byte to split the option name and value.
		nullIndex := bytes.IndexByte(optionBytes, 0)
		if nullIndex < 0 {
			return fmt.Errorf("invalid option: %v", optionBytes)
		}
		name := string(optionBytes[:nullIndex])

		// Move past the null byte to the start of the value.
		optionBytes = optionBytes[nullIndex+1:]

		if len(optionBytes) == 0 {
			return fmt.Errorf("invalid option: %v", packet)
		}

		// Find the length of the option value.
		valueLen := bytes.IndexByte(optionBytes, 0)
		if valueLen < 0 {
			valueLen = len(optionBytes)
		}

		value := make([]byte, valueLen)
		copy(value, optionBytes[:valueLen])

		// Move past the null byte and the option value to the start of the next option name.
		optionBytes = optionBytes[valueLen+1:]

		oack.Options[name] = value
	}

	// Extract optional parameters from the parsed options
	if windowsizeBytes, ok := oack.Options["windowsize"]; ok {
		oack.Windowsize = binary.BigEndian.Uint16(windowsizeBytes)
	}
	if tsizeBytes, ok := oack.Options["tsize"]; ok {
		oack.XferSize = binary.BigEndian.Uint32(tsizeBytes)
	}
	if blksizeBytes, ok := oack.Options["blksize"]; ok {
		oack.BlkSize = binary.BigEndian.Uint16(blksizeBytes)
	}
	if keyBytes, ok := oack.Options["key"]; ok {
		oack.Key = keyBytes
	}

	return nil
}

func (oa *OptionAcknowledgement) String() string {
	var optionsStr string
	for k, v := range oa.Options {
		optionsStr += k + "=" + string(v) + ","
	}
	return fmt.Sprintf("OptionAcknowledgement{ Opcode: %v, Options: {%s} }", oa.Opcode, optionsStr)
}

package tftp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type OptionAcknowledgement struct {
	Opcode  TFTPOpcode
	Options map[string][]byte
}

func NewOpt(options map[string][]byte) *OptionAcknowledgement {
	return &OptionAcknowledgement{
		Opcode:  TFTPOpcodeOACK,
		Options: options,
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

		// Find the next null byte to split the option value and the next option name.
		nullIndex = bytes.IndexByte(optionBytes, 0)
		if nullIndex < 0 {
			return fmt.Errorf("invalid option: %v", optionBytes)
		}
		value := string(optionBytes[:nullIndex])

		// Move past the null byte to the start of the next option name.
		optionBytes = optionBytes[nullIndex+1:]

		oack.Options[name] = []byte(value)
	}

	return nil
}

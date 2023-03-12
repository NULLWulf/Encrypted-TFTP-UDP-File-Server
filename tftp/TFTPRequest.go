package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type Request struct {
	Opcode       TFTPOpcode
	Filename     []byte
	Mode         []byte
	Options      map[string][]byte
	TransferSize uint32
	WindowSize   uint16
}

func (r *Request) Parse(packet []byte) (*Request, error) {
	// Check that the packet is at least 2 bytes long
	if len(packet) < 2 {
		return nil, errors.New("packet too short")
	}

	// Check that the opcode is valid
	opcode := binary.BigEndian.Uint16(packet[:2])
	if opcode != uint16(TFTPOpcodeRRQ) {
		return nil, errors.New("invalid opcode")
	}

	// Parse the filename and mode
	nullIndex := bytes.IndexByte(packet[2:], 0)
	if nullIndex < 0 {
		return nil, errors.New("filename not null-terminated")
	}
	filename := packet[2 : 2+nullIndex]
	modeStartIndex := 2 + nullIndex + 1
	nullIndex = bytes.IndexByte(packet[modeStartIndex:], 0)
	if nullIndex < 0 {
		return nil, errors.New("mode not null-terminated")
	}
	mode := packet[modeStartIndex : modeStartIndex+nullIndex]

	// Parse the options
	options := make(map[string][]byte)
	if nullIndex+modeStartIndex+1 < len(packet) {
		optionsBytes := packet[nullIndex+modeStartIndex+1:]
		for len(optionsBytes) > 0 {
			// Find the next null byte to split the option name and value.
			nullIndex := bytes.IndexByte(optionsBytes, 0)
			if nullIndex < 0 {
				return nil, fmt.Errorf("invalid option: %v", optionsBytes)
			}
			name := optionsBytes[:nullIndex]

			// Move past the null byte to the start of the value.
			optionsBytes = optionsBytes[nullIndex+1:]

			// Find the next null byte to split the option value and the next option name.
			nullIndex = bytes.IndexByte(optionsBytes, 0)
			if nullIndex < 0 {
				return nil, fmt.Errorf("invalid option: %v", optionsBytes)
			}
			value := optionsBytes[:nullIndex]

			// Move past the null byte to the start of the next option name.
			optionsBytes = optionsBytes[nullIndex+1:]

			options[string(name)] = value
		}
	}

	// Construct and return the read request packet
	request := &Request{
		Opcode:   TFTPOpcodeRRQ,
		Filename: filename,
		Mode:     mode,
		Options:  options,
	}
	return request, nil
}

func (r *Request) ToBytes() ([]byte, error) {
	// Check that the filename is not empty
	if r.Filename == nil || string(r.Filename) == "" {
		return nil, errors.New("filename is empty")
	}

	// Construct the request packet
	var optionsString []byte
	for key, value := range r.Options {
		optionsString = append(optionsString, []byte(key)...)
		optionsString = append(optionsString, 0)
		optionsString = append(optionsString, value...)
		optionsString = append(optionsString, 0)
	}
	packet := make([]byte, len(r.Filename)+1+len(r.Mode)+1+len(optionsString)+2)
	copy(packet, r.Filename)
	packet[len(r.Filename)] = 0
	copy(packet[len(r.Filename)+1:], r.Mode)
	packet[len(r.Filename)+1+len(r.Mode)] = 0
	copy(packet[len(r.Filename)+1+len(r.Mode)+1:], optionsString)
	packet[len(packet)-2] = 0

	// Set the opcode based on whether transfer size is specified
	if r.TransferSize > 0 {
		binary.BigEndian.PutUint16(packet[:2], uint16(TFTPOpcodeWRQ))
		binary.BigEndian.PutUint32(packet[len(packet)-2-4:], r.TransferSize)
	} else {
		binary.BigEndian.PutUint16(packet[:2], uint16(TFTPOpcodeRRQ))
	}

	// Return the request packet
	return packet, nil
}

func NewReq(filename []byte, mode []byte, transferSize uint32, options map[string][]byte) (*Request, error) {
	// Check that the filename is not empty
	if filename == nil || string(filename) == "" {
		return nil, errors.New("filename is empty")
	}

	if !bytes.Equal(mode, []byte("netascii")) && !bytes.Equal(mode, []byte("octet")) && !bytes.Equal(mode, []byte("mail")) {
		return nil, errors.New("invalid mode")
	}
	// Construct the request packet
	var opcode uint16
	if transferSize > 0 {
		opcode = uint16(TFTPOpcodeWRQ)
		if options == nil {
			options = make(map[string][]byte)
		}
		options["tsize"] = []byte(fmt.Sprintf("%d", transferSize))
	} else {
		opcode = uint16(TFTPOpcodeRRQ)
	}

	// Construct and return the TFTPRequest struct
	request := &Request{
		Opcode:   TFTPOpcode(opcode),
		Filename: filename,
		Mode:     mode,
		Options:  options,
	}
	return request, nil
}

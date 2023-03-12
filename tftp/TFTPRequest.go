package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

type Request struct {
	Opcode       TFTPOpcode
	Filename     []byte
	Mode         []byte
	Options      map[string][]byte
	TransferSize uint16
	WindowSize   uint16
}

func (r *Request) Parse(p []byte) error {
	bs := bytes.Split(p[2:], []byte{0})
	if len(bs) < 2 {
		return fmt.Errorf("missing filename or mode")
	}
	r.Filename = bs[0]
	r.Mode = bs[1]
	if len(bs) < 4 {
		return nil
	}
	r.Options = make(map[string][]byte)
	for i := 2; i+1 < len(bs); i += 2 {
		r.Options[string(bs[i])] = bs[i+1]
	}
	return nil
}
func (r *Request) ToBytes() ([]byte, error) {
	// Check that the filename is not empty
	if len(r.Filename) == 0 {
		return nil, errors.New("empty filename")
	}

	// Construct the request packet
	var packet []byte
	packet = append(packet, byte(r.Opcode>>8), byte(r.Opcode))
	packet = append(packet, r.Filename...)
	packet = append(packet, 0)
	packet = append(packet, r.Mode...)
	packet = append(packet, 0)
	for key, value := range r.Options {
		packet = append(packet, []byte(key)...)
		packet = append(packet, 0)
		packet = append(packet, value...)
		packet = append(packet, 0)
	}
	if r.TransferSize > 0 {
		packet = append(packet, []byte("tsize")...)
		packet = append(packet, 0)
		sizeBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(sizeBytes, r.TransferSize)
		packet = append(packet, sizeBytes...)
		packet = append(packet, 0)
	}
	if r.WindowSize > 0 {
		packet = append(packet, []byte("windowsize")...)
		packet = append(packet, 0)
		sizeBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(sizeBytes, r.WindowSize)
		packet = append(packet, sizeBytes...)
		packet = append(packet, 0)
	}
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

func (r *Request) String() {
	var optionsString string
	for key, value := range r.Options {
		optionsString += fmt.Sprintf("%s=%s,", key, value)
	}
	if len(optionsString) > 0 {
		optionsString = optionsString[:len(optionsString)-1]
	}

	log.Printf("Opcode=%d, Filename=%s, Mode=%s, Options={%s}, TransferSize=%d, WindowSize=%d\n",
		r.Opcode, r.Filename, r.Mode, optionsString, r.TransferSize, r.WindowSize)
}

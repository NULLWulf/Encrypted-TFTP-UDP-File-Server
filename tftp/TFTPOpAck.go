package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type OptionAcknowledgement struct {
	Opcode  TFTPOpcode
	Options map[string][]byte
}

type OptionAcknowledgementv2 struct {
	Opcode     TFTPOpcode
	Windowsize uint16
	XferSize   uint32
	BlkSize    uint16
	Key        []byte
}

func NewOptv2(windowsize uint16, xfer uint32, blksize uint16, key []byte) *OptionAcknowledgementv2 {
	return &OptionAcknowledgementv2{
		Opcode:     TFTPOpcodeOACK,
		Windowsize: windowsize,
		XferSize:   xfer,
		BlkSize:    blksize,
		Key:        key,
	}
}

func NewOpt(options map[string][]byte, xfer uint32) *OptionAcknowledgement {
	if xfer != 0 {
		options["tsize"] = make([]byte, 4)
		binary.BigEndian.PutUint32(options["tsize"], uint32(xfer))
	}
	return &OptionAcknowledgement{
		Opcode:  TFTPOpcodeOACK,
		Options: options,
	}
}

func (oack *OptionAcknowledgementv2) ToBytesV2() []byte {
	// Create a byte buffer to store the options.
	var buf bytes.Buffer

	if oack.Windowsize != 0 {
		buf.WriteString("windowsize")
		buf.WriteByte(0)
		buf.Write(make([]byte, 2))
		binary.BigEndian.PutUint16(buf.Bytes()[buf.Len()-2:], oack.Windowsize)
		buf.WriteByte(0)
	}

	if oack.XferSize != 0 {
		buf.WriteString("tsize")
		buf.WriteByte(0)
		buf.Write(make([]byte, 4))
		binary.BigEndian.PutUint32(buf.Bytes()[buf.Len()-4:], oack.XferSize)
		buf.WriteByte(0)
	}

	if oack.BlkSize != 0 {
		buf.Write([]byte("blksize"))
		buf.WriteByte(0)
		buf.Write(make([]byte, 2))
		binary.BigEndian.PutUint16(buf.Bytes()[buf.Len()-2:], oack.BlkSize)
		buf.WriteByte(0)
	}

	if oack.Key != nil {
		buf.Write([]byte("key"))
		buf.WriteByte(0)
		buf.Write(oack.Key)
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

	return nil
}

func (oa *OptionAcknowledgement) String() string {
	var optionsStr string
	for k, v := range oa.Options {
		optionsStr += k + "=" + string(v) + ","
	}
	return fmt.Sprintf("OptionAcknowledgement{ Opcode: %v, Options: {%s} }", oa.Opcode, optionsStr)
}

//func (o *OptionAcknowledgementv2) ParseOptions(data []byte) error {
//	// Create a new map to hold the options.
//	options := make(map[string][]byte)
//
//	// Iterate over the data and parse the options.
//	var optionName string
//	for i := 0; i < len(data); i++ {
//		if data[i] == 0 {
//			if optionName == "" {
//				// This is the first null byte, so the next bytes are the option name.
//				optionName = string(data[:i])
//				if len(data[i+1:]) == 0 {
//					// If there are no more bytes, this is an empty option, so set the value to an empty byte slice.
//					options[optionName] = []byte{}
//					break
//				}
//			} else {
//				// This is the second null byte, so the next bytes are the option value.
//				optionValue := data[len(optionName)+1 : i]
//				options[optionName] = optionValue
//				optionName = ""
//				if len(data[i+1:]) == 0 {
//					// If there are no more bytes, we're done parsing options.
//					break
//				}
//			}
//		}
//	}
//
//	// Set the fields in the OACKPacket based on the options.
//	if windowsizeBytes, ok := options["windowsize"]; ok {
//		o.Windowsize = binary.BigEndian.Uint16(windowsizeBytes)
//	}
//	if tsizeBytes, ok := options["tsize"]; ok {
//		o.XferSize = int64(binary.BigEndian.Uint32(tsizeBytes)
//	}
//	if blksizeBytes, ok := options["blksize"]; ok {
//		o.BlkSize = int(binary.BigEndian.Uint16(blksizeBytes))
//	}
//	if keyBytes, ok := options["key"]; ok {
//		o.Key = keyBytes
//	}
//
//	// Check if any unknown options were included in the packet.
//	for option := range options {
//		if option != "windowsize" && option != "tsize" && option != "blksize" && option != "key" {
//			return fmt.Errorf("unknown option: %s", option)
//		}
//	}
//
//	return nil
//}

func (optAck *OptionAcknowledgementv2) ParseV2(data []byte) error {
	if len(data) < 10 { // Minimum length required: 2 + 2 + 4 + 2
		return errors.New("insufficient data length")
	}

	optAck.Opcode = TFTPOpcode(binary.BigEndian.Uint16(data[0:2]))
	optAck.Windowsize = binary.BigEndian.Uint16(data[2:4])
	optAck.XferSize = binary.BigEndian.Uint32(data[4:8])
	optAck.BlkSize = binary.BigEndian.Uint16(data[8:10])
	optAck.Key = data[10 : len(data)-1]

	return nil
}

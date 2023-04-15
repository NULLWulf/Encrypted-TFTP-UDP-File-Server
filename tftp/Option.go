package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
)

type OptionAcknowledgement struct {
	Opcode     TFTPOpcode
	Windowsize uint16
	XferSize   uint32
	BlkSize    uint16
	Timeout    uint16
	Key        []byte
	KeyX, KeyY []byte
}

func New(opcode TFTPOpcode) *OptionAcknowledgement {
	return &OptionAcknowledgement{
		Opcode: opcode,
	}
}

func (oa *OptionAcknowledgement) Parse(data []byte) error {
	if len(data) < 2 {
		return errors.New("invalid data length")
	}

	oa.Opcode = TFTPOpcode(binary.BigEndian.Uint16(data[:2]))
	options := strings.Split(string(data[2:]), "\x00")

	for i := 0; i < len(options)-1; i += 2 {
		switch options[i] {
		case "windowsize":
			val, _ := strconv.ParseUint(options[i+1], 10, 16)
			oa.Windowsize = uint16(val)
		case "tsize":
			val, _ := strconv.ParseUint(options[i+1], 10, 32)
			oa.XferSize = uint32(val)
		case "blksize":
			val, _ := strconv.ParseUint(options[i+1], 10, 16)
			oa.BlkSize = uint16(val)
		case "timeout":
			val, _ := strconv.ParseUint(options[i+1], 10, 16)
			oa.Timeout = uint16(val)
		case "key":
			oa.Key = []byte(options[i+1])
		case "keyx":
			oa.KeyX = []byte(options[i+1])
		case "keyy":
			oa.KeyY = []byte(options[i+1])
		}
	}

	return nil
}

func (oa *OptionAcknowledgement) ToBytes() []byte {
	buf := new(bytes.Buffer)
	// Write Opcode
	binary.Write(buf, binary.BigEndian, oa.Opcode)
	// Write WindowSize
	if oa.Windowsize > 0 {
		buf.WriteString("windowsize")
		buf.WriteByte(0)
		buf.WriteString(strconv.Itoa(int(oa.Windowsize)))
		buf.WriteByte(0)
	}

	// Write XferSize
	if oa.XferSize > 0 {
		buf.WriteString("tsize")
		buf.WriteByte(0)
		buf.WriteString(strconv.Itoa(int(oa.XferSize)))
		buf.WriteByte(0)
	}

	// Write BlkSize
	if oa.BlkSize > 0 {
		buf.WriteString("blksize")
		buf.WriteByte(0)
		buf.WriteString(strconv.Itoa(int(oa.BlkSize)))
		buf.WriteByte(0)
	}

	// Write Timeout
	if oa.Timeout > 0 {
		buf.WriteString("timeout")
		buf.WriteByte(0)
		buf.WriteString(strconv.Itoa(int(oa.Timeout)))
		buf.WriteByte(0)
	}

	// Write key
	if len(oa.Key) > 0 {
		buf.WriteString("key")
		buf.WriteByte(0)
		buf.Write(oa.Key)
		buf.WriteByte(0)
	}

	// Write keyx
	if len(oa.KeyX) > 0 {
		buf.WriteString("keyx")
		buf.WriteByte(0)
		buf.Write(oa.KeyX)
		buf.WriteByte(0)
	}

	// Write keyy
	if len(oa.KeyY) > 0 {
		buf.WriteString("keyy")
		buf.WriteByte(0)
		buf.Write(oa.KeyY)
		buf.WriteByte(0)
	}
	return buf.Bytes()
}

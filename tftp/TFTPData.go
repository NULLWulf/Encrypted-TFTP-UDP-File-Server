package tftp

import (
	"encoding/binary"
	"errors"
)

// TFTPData represents a TFTP data packet.
type Data struct {
	Opcode      TFTPOpcode
	BlockNumber uint16
	Data        []byte
}

func (d *Data) ToBytes() []byte {
	// Construct the data packet
	packet := make([]byte, 2+2+len(d.Data))
	binary.BigEndian.PutUint16(packet[:2], uint16(TFTPOpcodeDATA))
	binary.BigEndian.PutUint16(packet[2:4], d.BlockNumber)
	copy(packet[4:], d.Data)

	// Return the data packet
	return packet
}

func (d *Data) Parse(packet []byte) error {
	// Check that the packet is at least 4 bytes long
	if len(packet) < 4 {
		return errors.New("packet too short")
	}

	// Check that the opcode is valid
	opcode := binary.BigEndian.Uint16(packet[:2])
	if opcode != uint16(TFTPOpcodeDATA) {
		return errors.New("invalid opcode")
	}

	// Parse the block number and data
	blockNumber := binary.BigEndian.Uint16(packet[2:4])
	data := packet[4:]

	d.Opcode = TFTPOpcodeDATA
	d.BlockNumber = blockNumber
	d.Data = data

	return nil
}

func NewData(blockNumber uint16, data []byte) (*Data, error) {
	// Check that the data is not empty
	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	// Construct and return the TFTPData struct
	dataPacket := &Data{
		Opcode:      TFTPOpcodeDATA,
		BlockNumber: blockNumber,
		Data:        data,
	}
	return dataPacket, nil
}

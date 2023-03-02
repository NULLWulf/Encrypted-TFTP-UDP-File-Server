package tftp

import (
	"encoding/binary"
	"errors"
)

// TFTPData represents a TFTP data packet.
type TFTPData struct {
	Opcode      TFTPOpcode
	BlockNumber uint16
	Data        []byte
}

func (d *TFTPData) ToBytes() ([]byte, error) {
	// Check that the data is not empty
	if len(d.Data) == 0 {
		return nil, errors.New("data is empty")
	}

	// Construct the data packet
	packet := make([]byte, 2+2+len(d.Data))
	binary.BigEndian.PutUint16(packet[:2], uint16(TFTPOpcodeDATA))
	binary.BigEndian.PutUint16(packet[2:4], d.BlockNumber)
	copy(packet[4:], d.Data)

	// Return the data packet
	return packet, nil
}

func (d *TFTPData) ReadFromBytes(packet []byte) error {
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

func NewTFTPData(blockNumber uint16, data []byte) (*TFTPData, error) {
	// Check that the data is not empty
	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	// Construct the data packet
	packet := make([]byte, 2+2+len(data))
	binary.BigEndian.PutUint16(packet[:2], uint16(TFTPOpcodeDATA))
	binary.BigEndian.PutUint16(packet[2:4], blockNumber)
	copy(packet[4:], data)

	// Construct and return the TFTPData struct
	dataPacket := &TFTPData{
		Opcode:      TFTPOpcodeDATA,
		BlockNumber: blockNumber,
		Data:        data,
	}
	return dataPacket, nil
}

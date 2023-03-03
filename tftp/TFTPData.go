package tftp

import (
	"encoding/binary"
	"errors"
	"log"
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

	// Construct and return the TFTPData struct
	dataPacket := &TFTPData{
		Opcode:      TFTPOpcodeDATA,
		BlockNumber: blockNumber,
		Data:        data,
	}
	return dataPacket, nil
}

func PrepareTFTPDataPackets(data []byte, blockSize int) (dataQueue []*TFTPData, err error) {
	// Create a slice of TFTPData packets
	blocks := len(data) / blockSize
	log.Printf("Length of data: %d, Block size: %d, Blocks: %d", len(data), blockSize, blocks)
	if len(data)%blockSize != 0 {
		blocks++
	}
	dataQueue = make([]*TFTPData, blocks)

	// Populate the slice with TFTPData packets
	for i := 0; i < blocks; i++ {
		// Calculate the start and end indices of the data
		start := i * blockSize
		end := start + blockSize
		if end > len(data) {
			end = len(data)
		}

		// Create the TFTPData packet
		dataQueue[i], err = NewTFTPData(uint16(i+1), data[start:end])
		if err != nil {
			return
		}
	}
	return
}

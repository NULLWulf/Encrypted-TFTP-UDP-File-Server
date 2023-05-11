package tftp

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

// Data struct represents a TFTP data packet
type Data struct {
	Opcode      TFTPOpcode
	BlockNumber uint16
	Checksum    uint32
	Data        []byte
}

// ToBytes method converts the TFTPData struct to a byte array packet
func (d *Data) ToBytes() []byte {
	// Construct the data packet
	packet := make([]byte, 2+2+4+len(d.Data))
	binary.BigEndian.PutUint16(packet[:2], uint16(TFTPOpcodeDATA))
	binary.BigEndian.PutUint16(packet[2:4], d.BlockNumber)
	binary.BigEndian.PutUint32(packet[4:8], d.Checksum)
	copy(packet[8:], d.Data)

	// Return the data packet
	return packet
} //data = Xor(data, xorKey)
//data = data
// Calculate the checksum

// Parse method parses a byte array into a TFTPData struct
func (d *Data) Parse(packet []byte, xorKey []byte) error {
	// Check that the packet is at least 8 bytes long
	if len(packet) < 8 {
		return errors.New("packet too short")
	}

	// Parse the block number, checksum, and data
	blockNumber := binary.BigEndian.Uint16(packet[2:4])
	checksum := binary.BigEndian.Uint32(packet[4:8])
	data := packet[8:]

	d.Opcode = TFTPOpcodeDATA
	d.BlockNumber = blockNumber
	d.Checksum = checksum
	d.Data = data

	return nil
}

// NewData method constructs a new TFTPData struct
func NewData(blockNumber uint16, data []byte, xorKey []byte) (*Data, error) {
	// Check that the data is not empty
	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	checksum := crc32.ChecksumIEEE(data)

	// Construct and return the TFTPData struct
	dataPacket := &Data{
		Opcode:      TFTPOpcodeDATA,
		BlockNumber: blockNumber,
		Checksum:    checksum,
		Data:        data,
	}
	return dataPacket, nil
}

// Checksum method calculates and returns the CRC32 checksum of the data
func Checksum(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

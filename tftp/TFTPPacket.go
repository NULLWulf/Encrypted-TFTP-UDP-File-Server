package tftp

import "encoding/binary"

type TFTPOpcode uint16

const (
	TFTPOpcodeRRQ   TFTPOpcode = 1
	TFTPOpcodeWRQ   TFTPOpcode = 2
	TFTPOpcodeDATA  TFTPOpcode = 3
	TFTPOpcodeACK   TFTPOpcode = 4
	TFTPOpcodeERROR TFTPOpcode = 5
	TFTPOpcodeOACK  TFTPOpcode = 6
	__tftUnused     TFTPOpcode = 7
	TFTPOpcodeTERM  TFTPOpcode = 8
)

func (o TFTPOpcode) String() string {
	switch o {
	case TFTPOpcodeRRQ:
		return "RRQ"
	case TFTPOpcodeWRQ:
		return "WRQ"
	case TFTPOpcodeDATA:
		return "DATA"
	case TFTPOpcodeACK:
		return "ACK"
	case TFTPOpcodeERROR:
		return "ERROR"
	case TFTPOpcodeOACK:
		return "OACK"
	default:
		return "INVALID"
	}
}

func HandlePacket(packet []byte) {
	opcode := TFTPOpcode(binary.BigEndian.Uint16(packet[:2]))
	switch opcode {
	case TFTPOpcodeRRQ:
		// Handle read request
	case TFTPOpcodeWRQ:
		// Handle write request
	case TFTPOpcodeDATA:
		// Handle data packet
		// Handle acknowledgement packet
	case TFTPOpcodeACK:
		// Handle error packet
	case TFTPOpcodeERROR:
		// Handle error packet
	case TFTPOpcodeOACK:
		// Handle option acknowledgement packet
	case TFTPOpcodeTERM:
	default:
		// Handle invalid opcode
	}
}

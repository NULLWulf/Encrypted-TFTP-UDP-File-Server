package tftp

import "encoding/binary"

type TFTPOpcode uint16

const (
	TFTPOpcodeRRQ   TFTPOpcode = 01
	TFTPOpcodeWRQ   TFTPOpcode = 02
	TFTPOpcodeDATA  TFTPOpcode = 03
	TFTPOpcodeACK   TFTPOpcode = 04
	TFTPOpcodeERROR TFTPOpcode = 05
	TFTPOpcodeOACK  TFTPOpcode = 06
	__tftUnused     TFTPOpcode = 07
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
	case TFTPOpcodeACK:
		// Handle acknowledgement packet
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

package tftp

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

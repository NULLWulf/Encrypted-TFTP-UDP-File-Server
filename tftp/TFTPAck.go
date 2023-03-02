package tftp

// TFTPAcknowledgement represents a TFTP acknowledgement packet.
type TFTPAcknowledgement struct {
	Opcode      TFTPOpcode
	BlockNumber uint16
}

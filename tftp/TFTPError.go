package tftp

// TFTPError represents a TFTP error packet.
type TFTPError struct {
	Opcode       TFTPOpcode
	ErrorCode    uint16
	ErrorMessage string
}

package tftp

type TFTPOptions struct {
	blksize   uint16
	tsize     uint16
	timeout   uint16
	multicast uint8
	wsize     uint16
	key       uint16
}

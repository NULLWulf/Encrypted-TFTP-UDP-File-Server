package tftp

type Option string

const (
	BlockSize    Option = "blksize"
	WindowSize   Option = "windowsize"
	TransferSize Option = "tsize"
	Key          Option = "key"
)

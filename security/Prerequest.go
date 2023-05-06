package security

import (
	"bytes"
	"hash/crc32"
	"log"
	"math/big"
)

type Prerequest struct {
	data    bytes.Buffer
	PubKeyX []byte
	PubKeyY []byte
}

func NewPreRequest() *Prerequest {
	// 11 Opcode
	buf := bytes.Buffer{}
	buf.Write([]byte{0x0b}) // 11 Opcode
	buf.Write([]byte{0x00})

	return &Prerequest{
		data: buf,
	}
}

func ParsePreRequest(data []byte) *Prerequest {
	p := NewPreRequest()
	log.Printf("ParsePreRequest: %v", string(p.data.Bytes()))
	fields := bytes.Split(data, []byte{0x00})
	for i := 0; i < len(fields); i++ {
		switch i {
		case 1:
			p.PubKeyX = fields[i]
		case 2:
			p.PubKeyY = fields[i]
		}
	}
	return p
}

func (p *Prerequest) AddKeyX(int *big.Int) {
	p.data.Write(int.Bytes())
	p.data.Write([]byte{0x00})
	p.PubKeyX = int.Bytes()
}

func (p *Prerequest) AddKeyY(int *big.Int) {
	p.data.Write(int.Bytes())
	p.data.Write([]byte{0x00})
	p.PubKeyY = int.Bytes()
}

func (p *Prerequest) AddKeyYBytes(bytes []byte) {
	p.data.Write(bytes)
	p.data.Write([]byte{0x00})
	p.PubKeyY = bytes
}

func (p *Prerequest) AddKeyXBytes(bytes []byte) {
	p.data.Write(bytes)
	p.data.Write([]byte{0x00})
	p.PubKeyX = bytes
}

func (p *Prerequest) BytesWChecksum() []byte {
	checksum := crc32.Checksum(p.data.Bytes(), crc32.IEEETable)
	p.data.Write([]byte{byte(checksum >> 24), byte(checksum >> 16), byte(checksum >> 8), byte(checksum)})
	p.data.Write([]byte{0x00})
	return p.data.Bytes()
}

func (p *Prerequest) Bytes() []byte {
	return p.data.Bytes()
}

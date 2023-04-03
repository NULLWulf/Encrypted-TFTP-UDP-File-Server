package tftp

import (
	"crypto/rand"
	"log"
)

/*
Tests related to the TFTP packets, probably will remove before sending in
TODO - remove this file before sending in
*/

type Test struct {
}

// Request TestTFTPRequestAck Test Ack
func (t Test) Request() {
	// Create a new request packet
	// random byte key
	optMap := make(map[string][]byte)
	optMap["door"] = []byte("door")

	optMap["key"] = GetRandomKey()
	optMap["blksize"] = []byte("512")

	request, err := NewReq([]byte("test.txt"), []byte("octet"), 512, optMap)
	if err != nil {
		log.Fatal(err)
	}
	// Convert the request packet to a byte slice
	packet, err := request.ToBytes()
	if err != nil {
		log.Fatal(err)
	}

	err = request.Parse(packet)
	if err != nil {
		return
	}

	bsize, _ := request.ToBytes()
	log.Printf("Request Packet: %d", len(bsize))
	TestEncryptDecrypt(bsize)
}

//func (t Test) Data() {
//	by
//	//data, err := NewData(1, make([]byte, 512),
//	if err != nil {
//		log.Fatal(err)
//	}
//	packet := data.ToBytes()
//	var data2 Data
//	err = data2.Parse(packet)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	bsize := data2.ToBytes()
//	log.Printf("Data Packet: %d", len(bsize))
//	TestEncryptDecrypt(bsize)
//
//}

func (t Test) Error() {
	// make 2 byte error code
	errPack := NewErr(02, []byte("File not found"))
	packet := errPack.ToBytes()
	var err2 Error
	err := err2.Parse(packet)
	if err != nil {
		return
	}
	bsize := err2.ToBytes()
	log.Printf("Error Packet: %d", len(bsize))
	TestEncryptDecrypt(bsize)
}

func (t Test) Ack() {
	ack := NewAck(02)
	packet := ack.ToBytes()
	var ack2 Ack
	err := ack2.Parse(packet)
	if err != nil {
		return
	}
	bsize := ack2.ToBytes()
	log.Printf("Ack Packet: %d", len(bsize))
	TestEncryptDecrypt(bsize)

}

func (t Test) Test() {
	//t.Request()
	//t.Data()
	//t.Error()
	//t.Ack()
	//t.Oack()

}

func Xor(data []byte, key []byte) []byte {
	ciphertext := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		ciphertext[i] = data[i] ^ key[i%len(key)]
	}
	return ciphertext
}

func DecryptXOR(ciphertext []byte, key []byte) []byte {
	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i++ {
		plaintext[i] = ciphertext[i] ^ key[i%len(key)]
	}
	return plaintext
}

func TestEncryptDecrypt(data []byte) {
	key := make([]byte, 128)
	_, err := rand.Read(key)
	if err != nil {
		return
	}

	ciphertext := Xor(data, key)

	plaintext := DecryptXOR(ciphertext, key)

	if string(plaintext) != string(data) {
		log.Fatal("Error: plaintext != data")
	} else {
		log.Printf("Success: plaintext == data")
	}
}

func GetRandomKey() []byte {
	key := make([]byte, 128)
	_, err := rand.Read(key)
	if err != nil {
		return nil
	}
	return key
}

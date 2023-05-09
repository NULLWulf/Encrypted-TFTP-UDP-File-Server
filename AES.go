package main

import (
	"CSC445_Assignment2/tftp"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"hash/crc32"
	"io"
	"log"
	"os"
)

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the plaintext using aesGCM.Seal
	// Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Get the nonce size
	nonceSize := aesGCM.NonceSize()

	// Extract the nonce from the encrypted data
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func AESTester() {
	sharedKey := DHKETester()
	AES := deriveAESKey256(sharedKey)

	img, err := ProxyRequest("")
	blocks, err := PrepareData(img, 512, AES)
	if err != nil {
		log.Printf("Error preparing data: %s\n", err.Error())
	}
	encrypted := make([][]byte, len(blocks))
	for i, block := range blocks {
		encrypted[i], err = encrypt(block.ToBytes(), AES)
		if err != nil {
			log.Printf("Error encrypting block: %s\n", err.Error())
		}
	}
	decrypted := make([][]byte, len(blocks))
	for i, block := range encrypted {
		log.Printf("Decrypting block %d\n", i)
		if i == 189 {
			log.Println("Block 189")
		}
		decrypted[i], err = decrypt(block, AES)
		if err != nil {
			log.Printf("Error decrypting block: %s\n", err.Error())
		}
	}

	data := new(tftp.Data)
	dataPacks := make([]*tftp.Data, len(blocks))
	reformed := make([]byte, 0)
	for i, block := range decrypted {
		err := data.Parse(block, nil)
		if err != nil {
			return
		}
		dataPacks[i] = data
		reformed = append(reformed, data.Data...)
	}

	// Check to see if reformed and img are the same
	if crc32.ChecksumIEEE(reformed) == crc32.ChecksumIEEE(img) {
		log.Println("Success!")
	}
	os.WriteFile("test-reformed.jpg", reformed, 0644)
}

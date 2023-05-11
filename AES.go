package main

import (
	"CSC445_Assignment2/tftp"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"hash/crc32"
	"io"
	"log"
)

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new Galois Counter Mode with the block cipher
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce. Never use more than 2^32 random nonces with a given key
	// because of the risk of a repeat.
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the data using AES-GCM
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	// Return the ciphertext
	return ciphertext, nil
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new Galois Counter Mode with the block cipher
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Get the nonce size
	nonceSize := aesGCM.NonceSize()

	// Split the ciphertext into the nonce and the encrypted data
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	// Return the plaintext
	return plaintext, nil
}

func AESTester() {
	sharedKey := DHKETester()         // Generate a shared key
	AES := deriveAESKey256(sharedKey) // Derive a 256-bit AES key from the shared key

	img, _ := ProxyRequest("https://rare-gallery.com/uploads/posts/577429-star-wars-high.jpg") // Get the image via HTTP

	blocks, _ := PrepareData(img, 512, AES) // Prepare the data for encryption
	log.Printf("Number of blocks: %d\n", len(blocks))

	// Make a new slice of byte slices to hold the encrypted blocks
	encrypted := make([][]byte, len(blocks))
	for i, block := range blocks {
		// Encrypt each block passed in the AES key
		encrypted[i], _ = encrypt(block.ToBytes(), AES)
	}

	data := new(tftp.Data)      // Create a new data packet
	reformed := make([]byte, 0) // Create a new byte slice to hold the reformed image
	for _, block := range encrypted {
		pt, _ := decrypt(block, AES)              // Decrypt the block
		data.Parse(pt, nil)                       // Parse the decrypted block into a data packet
		reformed = append(reformed, data.Data...) // Append the data to the reformed image
	}

	// Check to see if reformed image and image are the same
	if crc32.ChecksumIEEE(reformed) != crc32.ChecksumIEEE(img) {
		log.Fatalf("Reformed and original are not the same!")
	}
	log.Printf("Reformed and original are the same! Checksum: %d\n", crc32.ChecksumIEEE(reformed))
}

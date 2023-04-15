package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
)

type DHKESession struct {
	privateKey []byte
	pubKeyX    *big.Int
	pubKeyY    *big.Int
	sharedKey  []byte
}

// Encrypt takes a plaintext and key as byte slices, and returns the encrypted
// data as a byte slice (IV + ciphertext). It uses AES in CFB mode for encryption.
func (d *DHKESession) Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	// Create a new AES cipher with the provided key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a byte slice to store the IV and ciphertext
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	// Generate a random IV and store it in the beginning of the ciphertext slice
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Create a CFB stream encrypter using the AES cipher, and encrypt the plaintext
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// Decrypt takes a ciphertext and key as byte slices, and returns the decrypted
// data as a byte slice. It uses AES in CFB mode for decryption.
func (d *DHKESession) Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// Create a new AES cipher with the provided key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Check if the ciphertext is too short (should include IV + encrypted data)
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract the IV from the beginning of the ciphertext slice
	iv := ciphertext[:aes.BlockSize]

	// Remove the IV from the ciphertext
	ciphertext = ciphertext[aes.BlockSize:]

	// Create a CFB stream decrypter using the AES cipher, and decrypt the ciphertext
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}

// GenerateKeyPair a public-private key pair for the DHKE using the elliptic curve P-256.
func (d *DHKESession) GenerateKeyPair() {
	curve := elliptic.P256()
	d.privateKey, d.pubKeyX, d.pubKeyY, _ = elliptic.GenerateKey(curve, rand.Reader)
}

// GenerateSharedKey generates the shared key using the private key and the public key of the other party.
func (d *DHKESession) generateSharedKey(pubKeyX, pubKeyY *big.Int) {
	curve := elliptic.P256()
	x, _ := curve.ScalarMult(pubKeyX, pubKeyY, d.privateKey)
	d.sharedKey = x.Bytes()
}

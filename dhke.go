package main

import (
	"crypto/aes"
	"crypto/cipher"
	_ "crypto/ecdh"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"math/big"
)

type DHKESession struct {
	privateKey []byte
	pubKeyX    *big.Int
	pubKeyY    *big.Int
	sharedKey  []byte
	aes512Key  []byte
}

func (d *DHKESession) EncryptAes(data []byte, key []byte) []byte {
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		panic(err)
	}

	block, _ := aes.NewCipher(key)

	// Encrypt the plaintext using CBC mode
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(data))
	mode.CryptBlocks(ciphertext, data)

	// Prepend the IV to the ciphertext
	encryptedPacket := append(iv, ciphertext...)

	return encryptedPacket
}

// decrypt decrypts an encrypted packet using AES-CBC mode
func Decrypt(encryptedPacket []byte, key []byte) ([]byte, error) {
	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Extract the IV and ciphertext from the encrypted packet
	iv := encryptedPacket[:aes.BlockSize]
	ciphertext := encryptedPacket[aes.BlockSize:]

	// Decrypt the ciphertext using CBC mode
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Unpad the plaintext to remove PKCS#7 padding
	plaintext = unpad(plaintext)

	return plaintext, nil
}

// unpad removes PKCS#7 padding from the input data
func unpad(data []byte) []byte {
	padding := int(data[len(data)-1])
	return data[:len(data)-padding]
}

// GenerateKeyPair a public-private key pair for the DHKE using the elliptic curve P-256.
func (d *DHKESession) GenerateKeyPair() {
	curve := elliptic.P256()
	d.privateKey, d.pubKeyX, d.pubKeyY, _ = elliptic.GenerateKey(curve, rand.Reader)
}

func (d *DHKESession) generateSharedKey(pubKeyX, pubKeyY *big.Int) ([]byte, error) {
	curve := elliptic.P256()

	if !curve.IsOnCurve(pubKeyX, pubKeyY) {
		return nil, errors.New("public key is not on the curve")
	}

	x, y := curve.ScalarMult(pubKeyX, pubKeyY, d.privateKey)
	if x == nil || y == nil {
		return nil, errors.New("invalid input to ScalarMult")
	}

	d.aes512Key = (deriveAESKey512(x.Bytes()))
	return x.Bytes(), nil
}

// Derive the AES key from the shared secret using SHA-256
func deriveAESKey256(sharedSecret []byte) []byte {
	hash := sha256.Sum256(sharedSecret)
	return hash[:]
}

func deriveAESKey512(sharedSecret []byte) []byte {
	hash := sha512.Sum512(sharedSecret)
	return hash[:]
}

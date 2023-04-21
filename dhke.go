package main

import (
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

// GenerateKeyPair a public-private key pair for the DHKE using the elliptic curve P-256.
func (d *DHKESession) GenerateKeyPair() {
	curve := elliptic.P256()
	d.privateKey, d.pubKeyX, d.pubKeyY, _ = elliptic.GenerateKey(curve, rand.Reader)
}

// generateSharedKey generates the shared key using the public key and the private key
func (d *DHKESession) generateSharedKey(pubKeyX, pubKeyY *big.Int) ([]byte, error) {
	curve := elliptic.P256()

	// Check if the public key is on the curve
	// If not, generate a new key pair
	if !curve.IsOnCurve(pubKeyX, pubKeyY) {
		d.GenerateKeyPair() // Generate a new key pair
		return nil, errors.New("public key is not on the curve")
	}

	var x, y *big.Int
	// Generate the shared key until it is not nil
	for generated := false; !generated; {
		x, y = curve.ScalarMult(pubKeyX, pubKeyY, d.privateKey)
		if x != nil && y != nil {
			generated = true
		}
	}

	d.aes512Key = deriveAESKey512(x.Bytes())
	return x.Bytes(), nil
}

// Derive the AES key from the shared secret using SHA-256
func deriveAESKey256(sharedSecret []byte) []byte {
	hash := sha256.Sum256(sharedSecret)
	return hash[:]
}

// Derive the AES key from the shared secret using SHA-512
func deriveAESKey512(sharedSecret []byte) []byte {
	hash := sha512.Sum512(sharedSecret)
	return hash[:]
}

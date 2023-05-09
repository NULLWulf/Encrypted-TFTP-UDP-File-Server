package main

import (
	_ "crypto/ecdh"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"log"
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

func (d *DHKESession) curveCheck(pubKeyX, pubKeyY *big.Int) bool {
	curve := elliptic.P256()
	return curve.IsOnCurve(pubKeyX, pubKeyY)
}

// generateSharedKey generates the shared key using the public key and the private key
func (d *DHKESession) generateSharedKey(pubKeyX, pubKeyY *big.Int) ([]byte, error) {
	curve := elliptic.P256()

	var x, y *big.Int
	// Generate the shared key until it is not nil
	for generated := false; !generated; {
		x, y = curve.ScalarMult(pubKeyX, pubKeyY, d.privateKey)
		if x != nil && y != nil {
			generated = true
		}
	}

	d.aes512Key = deriveAESKey256(x.Bytes())
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

func DHKETester() []byte {
	client := new(DHKESession)
	server := new(DHKESession)
	client.GenerateKeyPair()
	server.GenerateKeyPair()
	key, err := client.generateSharedKey(server.pubKeyX, server.pubKeyY)
	if err != nil {
		log.Printf("Error: %s", err)
		return nil
	}
	sharedKey, err := server.generateSharedKey(client.pubKeyX, client.pubKeyY)
	if err != nil {
		log.Printf("Error: %s", err)
		return nil
	}

	// Asset keys are the same
	if string(key) != string(sharedKey) {
		log.Printf("Error: keys are not the same")
		return nil
	}
	log.Printf("Keys are the same: %s", key)

	return sharedKey

}

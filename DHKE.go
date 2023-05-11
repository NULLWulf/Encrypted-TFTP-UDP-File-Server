package main

import (
	"CSC445_Assignment2/tftp"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"log"
	"math/big"
)

// DHKESession used to store out keys ephemerally
type DHKESession struct {
	privateKey []byte
	pubKeyX    *big.Int
	pubKeyY    *big.Int
	sharedKey  []byte
	aes512Key  []byte
}

// GenerateKeyPair a public-private key pair for the DHKE using the elliptic curve P-256.
func (d *DHKESession) GenerateKeyPair() {
	// Set curve to NIST P-256, in reality this is a pretty weak curve
	curve := elliptic.P256()

	// Generate the private key, public key, and check if the public key is on the curve
	d.privateKey, d.pubKeyX, d.pubKeyY, _ = elliptic.GenerateKey(curve, rand.Reader)

	// If the public key is not on the curve, generate a new key pair
	for !d.curveCheck(d.pubKeyX, d.pubKeyY) {
		// until a key pair is generated that is on the curve
		log.Printf("Keys are not on the curve, generating new keys")
		d.privateKey, d.pubKeyX, d.pubKeyY, _ = elliptic.GenerateKey(curve, rand.Reader)
	}
}

// Helper function to check if the public key is on the curve
func (d *DHKESession) curveCheck(pubKeyX, pubKeyY *big.Int) bool {
	curve := elliptic.P256()
	return curve.IsOnCurve(pubKeyX, pubKeyY)
}

// Generate the shared key using our private key and the public key of the other party
func (d *DHKESession) generateSharedKey(pubKeyX, pubKeyY *big.Int) ([]byte, error) {
	curve := elliptic.P256()
	var x, y *big.Int // Instantiate x and y Big Ints (256 bits)

	// If by some chance the public key is not on the curve, return an error
	// This can help to prevent attacks such as MITM or key corruption
	if !d.curveCheck(pubKeyX, pubKeyY) {
		return nil, errors.New("public key is not on the curve")
	}

	// Generate the shared key until it is not nil
	for generated := false; !generated; {
		// Low level API for scalar multiplication, can be very testy and have
		// to handle errors manually
		x, y = curve.ScalarMult(pubKeyX, pubKeyY, d.privateKey)
		if x != nil && y != nil {
			generated = true
		}
	}

	// Derive the AES key from the shared secret using SHA-256
	d.aes512Key = deriveAESKey256(x.Bytes())
	// Return the shared secret as bytes in case we want to use it for something else
	return x.Bytes(), nil
}

// Derive the AES key from the shared secret using SHA-256
func deriveAESKey256(sharedSecret []byte) []byte {
	hash := sha256.Sum256(sharedSecret)
	return hash[:]
}

func DHKETester() []byte {
	// Instantiate a new DHKE session for the client
	client := new(DHKESession)
	// Instantiate a new DHKE session for the server
	server := new(DHKESession)
	// Generate the key pair for the client
	client.GenerateKeyPair()
	// Generate the key pair for the server
	server.GenerateKeyPair()
	// Generate the shared key for the client
	key, _ := client.generateSharedKey(server.pubKeyX, server.pubKeyY)
	log.Printf("Client Key Checksum: %d", tftp.Checksum(key))
	// Generate the shared key for the server
	sharedKey, _ := server.generateSharedKey(client.pubKeyX, client.pubKeyY)
	log.Printf("Server Key Checksum: %d", tftp.Checksum(sharedKey))

	// Assert keys are the same
	if string(key) != string(sharedKey) {
		log.Printf("Error: keys are not the same")
		return nil
	}
	log.Printf("Keys are the same, Checksum: %d", tftp.Checksum(key))

	return sharedKey

}

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
)

// PBKDF2Key transforms a key into a 32 bit key.
func PBKDF2Key(key []byte) ([]byte, error) {
	salt := make([]byte, 32)

	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	return pbkdf2.Key(key, salt, 4096, 32, sha3.New384), nil
}

// EncryptGCM encrypts the plaintext with AES/GCM.
func EncryptGCM(plaintext []byte, key []byte) ([]byte, error) {
	// Create a new cipher.
	c, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	// Use GCM - Galois/Counter Mode.
	gcm, err := cipher.NewGCM(c)

	if err != nil {
		return nil, err
	}

	// GCM requires that we have a cryptographically secure nonce which will be passed
	// to our Seal function.
	nonce := make([]byte, gcm.NonceSize())

	// Get the nonce.
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Now seal the deal.
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptGCM decrypts an AES/GCM encrypted ciphertext.
func DecryptGCM(ciphertext []byte, key []byte) ([]byte, error) {
	// Create a new cipher.
	c, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	// Use GCM - Galois/Counter Mode.
	gcm, err := cipher.NewGCM(c)

	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()

	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("the size of the nonce is greater than the length of the ciphertext")
	}

	// Separate the nonce from the ciphertext.
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Now decrypt the ciphertext.
	return gcm.Open(nil, nonce, ciphertext, nil)
}

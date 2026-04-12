package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// Encrypt a plaintext message using AES-GCM.
func Encrypt(plaintext, key []byte) (ciphertext []byte, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	// Generate a random nonce. Ensure it is unique for each encryption with the same key.
	nonce, err = GenerateRandomBytes(aesGCM.NonceSize())
	if err != nil {
		return nil, nil, err
	}

	// Encrypt the plaintext using the nonce.
	ciphertext = aesGCM.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt a ciphertext message using AES-GCM.
func Decrypt(ciphertext, key, nonce []byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt the ciphertext using the nonce.
	plaintext, err = aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GenerateRandomBytes random bytes using crypto/rand, which is secure for cryptographic purposes.
func GenerateRandomBytes(size int) ([]byte, error) {
	bytes := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}

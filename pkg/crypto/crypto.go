// Package crypto provides cryptographic operations for encrypting and decrypting data.
//
// This package implements AES-256-GCM encryption for securing sensitive data.
// It uses authenticated encryption to ensure both confidentiality and integrity.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// KeySize defines the size of encryption keys in bytes (256 bits).
const KeySize = 32

// KeyGen generates a new random encryption key of KeySize bytes.
// It uses crypto/rand for cryptographically secure random generation.
//
// Returns the generated key or an error if random generation fails.
func KeyGen() ([]byte, error) {
	key := make([]byte, KeySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with the provided key.
// The nonce is prepended to the ciphertext in the output.
//
// Parameters:
//   - key: 32-byte AES encryption key
//   - plaintext: data to encrypt
//
// Returns the encrypted data (nonce + ciphertext) or an error.
func Encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the provided key.
// Expects the nonce to be prepended to the ciphertext (as produced by Encrypt).
//
// Parameters:
//   - key: 32-byte AES encryption key (same as used for encryption)
//   - ciphertext: encrypted data with prepended nonce
//
// Returns the decrypted plaintext or an error if decryption fails.
func Decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	nonce := ciphertext[:nonceSize]
	data := ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// Package crypto provides ChaCha20-Poly1305 encryption and decryption
// with Argon2id key derivation for the Gouija two-stage proxy.
package crypto

import (
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

// AuthChallengeSize is the size of the authentication challenge
const AuthChallengeSize = 32

// AuthResponseSize is the size of the HMAC-SHA256 response
const AuthResponseSize = sha256.Size // 32 bytes

const (
	// NonceSize is the size of the nonce for ChaCha20-Poly1305
	NonceSize = chacha20poly1305.NonceSize // 12 bytes

	// KeySize is the size of the encryption key
	KeySize = chacha20poly1305.KeySize // 32 bytes

	// Overhead is the authentication tag overhead
	Overhead = chacha20poly1305.Overhead // 16 bytes

	// DefaultSalt is the default salt for key derivation
	DefaultSalt = "gouija-salt-v1"

	// DefaultAuthSalt is the default salt for auth key derivation
	DefaultAuthSalt = "gouija-auth-v1"

	// Argon2id parameters
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
)

var (
	// ErrDecryptionFailed indicates that decryption failed (authentication failure or corruption)
	ErrDecryptionFailed = errors.New("crypto: decryption failed")

	// ErrCiphertextTooShort indicates that the ciphertext is too short to contain a nonce
	ErrCiphertextTooShort = errors.New("crypto: ciphertext too short")
)

// Cipher provides encryption and decryption using ChaCha20-Poly1305
type Cipher struct {
	aead cipher.AEAD
}

// NewCipher creates a new Cipher from a pre-shared key.
// The key is processed through Argon2id to derive a 256-bit encryption key.
// If salt is empty, DefaultSalt is used.
func NewCipher(presharedKey, salt string) (*Cipher, error) {
	if presharedKey == "" {
		return nil, errors.New("crypto: pre-shared key cannot be empty")
	}
	if salt == "" {
		salt = DefaultSalt
	}

	// Derive key using Argon2id
	derivedKey := argon2.IDKey(
		[]byte(presharedKey),
		[]byte(salt),
		argonTime,
		argonMemory,
		argonThreads,
		KeySize,
	)

	// Create ChaCha20-Poly1305 AEAD cipher
	aead, err := chacha20poly1305.New(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("crypto: failed to create AEAD cipher: %w", err)
	}

	// Zero the derived key from stack (best effort)
	for i := range derivedKey {
		derivedKey[i] = 0
	}

	return &Cipher{aead: aead}, nil
}

// Encrypt encrypts plaintext using ChaCha20-Poly1305.
// Returns: [12-byte nonce][ciphertext + 16-byte auth tag]
func (c *Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	// Generate random nonce
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("crypto: failed to generate nonce: %w", err)
	}

	// Allocate output buffer: nonce + ciphertext + auth tag
	ciphertext := make([]byte, NonceSize+len(plaintext)+Overhead)

	// Copy nonce to beginning
	copy(ciphertext[:NonceSize], nonce)

	// Encrypt and append to nonce
	c.aead.Seal(ciphertext[NonceSize:NonceSize], nonce, plaintext, nil)

	return ciphertext, nil
}

// EncryptTo encrypts plaintext and appends the result to dst.
// This allows reusing buffers to reduce allocations.
// Returns the extended dst slice.
func (c *Cipher) EncryptTo(dst, plaintext []byte) ([]byte, error) {
	// Generate random nonce
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("crypto: failed to generate nonce: %w", err)
	}

	// Append nonce to dst
	dst = append(dst, nonce...)

	// Encrypt and append to dst
	dst = c.aead.Seal(dst, nonce, plaintext, nil)

	return dst, nil
}

// Decrypt decrypts ciphertext encrypted by Encrypt.
// Input format: [12-byte nonce][ciphertext + 16-byte auth tag]
func (c *Cipher) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < NonceSize+Overhead {
		return nil, ErrCiphertextTooShort
	}

	// Extract nonce
	nonce := ciphertext[:NonceSize]
	encryptedData := ciphertext[NonceSize:]

	// Decrypt
	plaintext, err := c.aead.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// DecryptTo decrypts ciphertext and appends the result to dst.
// This allows reusing buffers to reduce allocations.
// Returns the extended dst slice.
func (c *Cipher) DecryptTo(dst, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < NonceSize+Overhead {
		return nil, ErrCiphertextTooShort
	}

	// Extract nonce
	nonce := ciphertext[:NonceSize]
	encryptedData := ciphertext[NonceSize:]

	// Decrypt
	plaintext, err := c.aead.Open(dst, nonce, encryptedData, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// EncryptedSize returns the size of the ciphertext for a given plaintext size.
func EncryptedSize(plaintextSize int) int {
	return NonceSize + plaintextSize + Overhead
}

// PlaintextSize returns the size of the plaintext for a given ciphertext size.
// Returns -1 if the ciphertext size is invalid.
func PlaintextSize(ciphertextSize int) int {
	size := ciphertextSize - NonceSize - Overhead
	if size < 0 {
		return -1
	}
	return size
}

// GenerateChallenge generates a random challenge for authentication.
func GenerateChallenge() ([]byte, error) {
	challenge := make([]byte, AuthChallengeSize)
	if _, err := rand.Read(challenge); err != nil {
		return nil, fmt.Errorf("crypto: failed to generate challenge: %w", err)
	}
	return challenge, nil
}

// ComputeAuthResponse computes HMAC-SHA256 of the challenge using the given key.
func ComputeAuthResponse(key, challenge []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(challenge)
	return mac.Sum(nil)
}

// VerifyAuthResponse verifies an HMAC-SHA256 response against the expected response.
// Uses constant-time comparison to prevent timing attacks.
func VerifyAuthResponse(key, challenge, response []byte) bool {
	expected := ComputeAuthResponse(key, challenge)
	return subtle.ConstantTimeCompare(expected, response) == 1
}

// DeriveAuthKey derives a key for HMAC authentication from the pre-shared key.
// This uses a different salt than the encryption key derivation.
// If salt is empty, DefaultAuthSalt is used.
func DeriveAuthKey(presharedKey, salt string) []byte {
	if salt == "" {
		salt = DefaultAuthSalt
	}
	return argon2.IDKey(
		[]byte(presharedKey),
		[]byte(salt),
		argonTime,
		argonMemory,
		argonThreads,
		32, // 256-bit key for HMAC
	)
}

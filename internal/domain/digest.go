package domain

import (
	"crypto/sha256"
	"encoding/hex"
)

// Digest computes the content digest (SHA-256, hex) — the one law of
// every digest in the machine: raw bytes, no normalization, the
// digest is the data's exact identity.
func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

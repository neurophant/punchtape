package yamlio

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"

	"github.com/goccy/go-yaml"
)

// Digest computes the content digest (SHA-256, hex).
func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// NormalizeEOL normalizes line breaks to LF: cross-OS stability — a
// file carried between CRLF systems does not change the machine's
// computations.
func NormalizeEOL(data []byte) []byte {
	if !bytes.ContainsRune(data, '\r') {
		return data
	}
	return bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
}

// DigestNorm — the digest of normalized content: line breaks before
// the hash, the same semantics on any OS.
func DigestNorm(data []byte) string {
	return Digest(NormalizeEOL(data))
}

// DigestShort — the first 16 digest characters for human-readable
// places.
func DigestShort(data []byte) string {
	return Digest(data)[:16]
}

// Unmarshal decodes exactly one YAML document leniently — the read
// path of the machine's own caches (unknown-field strictness is an
// intake discipline, not a reread one).
func Unmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}

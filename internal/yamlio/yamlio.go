// Package yamlio — shared rules for working with canon YAML files:
// strict decoding (an unknown field is an error), atomic writes
// (temp file + rename), content digests.
package yamlio

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"

	"gopkg.in/yaml.v3"
)

// unknownFieldRe extracts the line and the extra field's name from
// a yaml error.
var unknownFieldRe = regexp.MustCompile(`line (\d+): field (\S+) not found`)

// DecodeStrict decodes data into a value strictly: any unknown
// field, wrong type or value outside the enumeration — a one-line
// error, without stack traces in the face.
func DecodeStrict(data []byte, v any) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	err := dec.Decode(v)
	if err == nil {
		return nil
	}
	var typeErr *yaml.TypeError
	if errors.As(err, &typeErr) && len(typeErr.Errors) > 0 {
		first := typeErr.Errors[0]
		if m := unknownFieldRe.FindStringSubmatch(first); m != nil {
			return fmt.Errorf("line %s: unknown field %s", m[1], m[2])
		}
		return fmt.Errorf("%s", oneLine(first))
	}
	return fmt.Errorf("%s", oneLine(err.Error()))
}

func oneLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// Marshal emits canonical YAML: struct fields in declaration order,
// map keys sorted — the same data gives a bit-for-bit identical
// output.
func Marshal(v any) ([]byte, error) {
	out, err := yaml.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	return out, nil
}

// WriteAtomic writes data to a file in one action: first a temp
// file next to the target, then a rename. A reader at any moment
// sees either the whole old content or the new one.
func WriteAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("dir %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // cleanup when the rename did not happen

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("temp file write: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("temp file sync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("temp file close: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename to %s: %w", path, err)
	}
	return nil
}

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

// DecodeLiteral — a decoder for a one-line byte literal: \n, \t,
// \r, \\ and \xNN expand into bytes, the rest passes as is. Seeds
// and byte expectations live in one YAML line — control bytes and
// arbitrary sequences are expressed with escapes.
func DecodeLiteral(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			sb.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'n':
			sb.WriteByte('\n')
		case 't':
			sb.WriteByte('\t')
		case 'r':
			sb.WriteByte('\r')
		case '\\':
			sb.WriteByte('\\')
		case 'x':
			if b, ok := hexByte(s, i+1); ok {
				sb.WriteByte(b)
				i += 2
				continue
			}
			sb.WriteByte('\\')
			sb.WriteByte('x')
		default:
			sb.WriteByte('\\')
			sb.WriteByte(s[i])
		}
	}
	return sb.String()
}

// DecodeHexOnly — a decoder for the strict expectation literal: it
// expands only \xNN. Any other escape in an expectation is an
// encoding error: control characters are written by YAML itself
// (double quotes, block scalar), a literal backslash by the \x5C
// escape; validation (ValidateExpectLiteral) catches the violation
// at intake with a recipe, not with an eternal red run.
func DecodeHexOnly(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) || s[i+1] != 'x' {
			sb.WriteByte(s[i])
			continue
		}
		if b, ok := hexByte(s, i+2); ok {
			sb.WriteByte(b)
			i += 3
			continue
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
}

// ValidateExpectLiteral — strict validation of the expectation
// literal: every backslash must start a \xNN (a literal slash is
// \x5C). Literal "\n"/"\t"/"\r" — an encoding error: a control
// character was intended, two bytes came out.
func ValidateExpectLiteral(s string) error {
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			continue
		}
		if s[i+1] == 'x' {
			if _, ok := hexByte(s, i+2); ok {
				i += 3
				continue
			}
		}
		return fmt.Errorf("%s", canondata.T("format.error.escape", canondata.M{
			"at": fmt.Sprintf("%d", i+1),
		}))
	}
	return nil
}

// hexByte — two hexadecimal characters at position pos — one byte.
func hexByte(s string, pos int) (byte, bool) {
	if pos < 0 || pos+1 >= len(s) {
		return 0, false
	}
	hi, ok1 := hexVal(s[pos])
	lo, ok2 := hexVal(s[pos+1])
	if !ok1 || !ok2 {
		return 0, false
	}
	return hi<<4 | lo, true
}

func hexVal(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// Package fprint — the fingerprint laws: what a proof depends on and
// how each dependency becomes bytes. A behavior's dependency is the
// content of its claims — the exact line ranges it stands on; a
// row's proof is the row's own bytes, its behavior's claims
// fingerprint and the environment hash. Pure file reads and
// digests: nothing watches anything, the same numbers come out on
// every operating system.
package fprint

import (
	"sort"
	"strings"

	"github.com/neurophant/punchtape/internal/domain"
)

// AbsentDigest — the spelling of a claimed range whose lines no
// longer exist: absence is a fact the proof must see.
const AbsentDigest = "absent"

// EnvIdentity — the environment a command runs in: the machine's own
// identity and the declared clean environment actually passed. The
// values enter the hash: a session that changes what the commands see
// changes the proofs — rows go stale, never falsely green.
type EnvIdentity struct {
	Version string
	Arch    string
	Kernel  string
	Env     []string // "NAME=value", sorted; the declared clean set
}

// EnvHash — the environment's digest.
func EnvHash(e EnvIdentity) string {
	entries := append([]string(nil), e.Env...)
	sort.Strings(entries)
	var b strings.Builder
	b.WriteString(e.Version)
	b.WriteByte(0)
	b.WriteString(e.Arch)
	b.WriteByte(0)
	b.WriteString(e.Kernel)
	for _, kv := range entries {
		b.WriteByte(0)
		b.WriteString(kv)
	}
	return domain.Digest([]byte(b.String()))
}

// Proof — the row's proof fingerprint: the row's canonical bytes, its
// behavior's claims fingerprint, the environment hash. While it
// stands, the row is never executed again; when it moves, the row is
// stale.
func Proof(rowBytes []byte, claimsDigest, envHash string) string {
	var b strings.Builder
	b.Write(rowBytes)
	b.WriteByte(0)
	b.WriteString(claimsDigest)
	b.WriteByte(0)
	b.WriteString(envHash)
	return domain.Digest([]byte(b.String()))
}

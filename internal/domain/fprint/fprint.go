// Package fprint — the fingerprint laws: what a proof depends on and
// how each dependency becomes bytes. A behavior's fingerprint is the
// content of its scope map — the author-declared set of files the
// behavior stands on; a row's proof is the row's own bytes, its
// behavior's scope fingerprint and the environment hash. Pure file
// reads and digests: nothing watches anything, the same numbers come
// out on every operating system.
package fprint

import (
	"path"
	"sort"
	"strings"

	"github.com/neurophant/punchtape/internal/domain"
)

// FileDigest — one scope file's measured digest: the sha256 of the
// file's exact bytes, no transformation — different bytes, a
// different digest. Absence is a fact the proof must see.
type FileDigest struct {
	Path   string
	Digest string // content digest, or AbsentDigest
}

// AbsentDigest — the spelling of a scope file that no longer exists.
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

// ScopeHash — the behavior's fingerprint: the digest of the sorted
// (path, digest-or-absent) list of the files its scope map resolves
// to. A scope file that does not exist spells absent — the map names
// what must stand, the hash sees the present.
func ScopeHash(files []FileDigest) string {
	set := append([]FileDigest(nil), files...)
	sort.Slice(set, func(i, j int) bool { return set[i].Path < set[j].Path })
	var b strings.Builder
	for _, f := range set {
		b.WriteString(f.Path)
		b.WriteByte(0)
		b.WriteString(f.Digest)
		b.WriteByte(0)
	}
	return domain.Digest([]byte(b.String()))
}

// Proof — the row's proof fingerprint: the row's canonical bytes, its
// behavior's scope fingerprint, the environment hash. While it
// stands, the row is never executed again; when it moves, the row is
// stale.
func Proof(rowBytes []byte, scopeHash, envHash string) string {
	var b strings.Builder
	b.Write(rowBytes)
	b.WriteByte(0)
	b.WriteString(scopeHash)
	b.WriteByte(0)
	b.WriteString(envHash)
	return domain.Digest([]byte(b.String()))
}

// MatchGlob — whether a root-relative file path belongs to a scope
// glob: '**' matches any depth of directories, '*' and '?' and a
// '[...]' class match inside one segment.
func MatchGlob(glob, path string) bool {
	return matchSegs(strings.Split(glob, "/"), strings.Split(path, "/"))
}

func matchSegs(pat, name []string) bool {
	if len(pat) == 0 {
		return len(name) == 0
	}
	if pat[0] == "**" {
		// '**' swallows zero or more segments; the rest must match what
		// remains.
		for skip := 0; skip <= len(name); skip++ {
			if matchSegs(pat[1:], name[skip:]) {
				return true
			}
		}
		return false
	}
	if len(name) == 0 {
		return false
	}
	if ok, err := path.Match(pat[0], name[0]); err != nil || !ok {
		return false
	}
	return matchSegs(pat[1:], name[1:])
}

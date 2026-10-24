package doc

// The row's verdict against its expectation: pure comparison, no
// interpretation. The match spellings are two — the exact bytes, or
// the substring behind the "contains: " prefix; every field absent
// from the expectation is not checked. The exit code is always
// measured: an expectation without rc wants zero.

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// ContainsPrefix — the substring-match spelling of out and files.
const ContainsPrefix = "contains: "

// Compare — the measured facts against the expectation; every
// returned string is one broken fact, in the expectation's own
// terms. ReadFile answers a root-relative file's current bytes; an
// absent file is an honest failure, not an error of the comparison.
func Compare(e Expect, exit int, stdout string, readFile func(string) ([]byte, error)) []string {
	var facts []string
	wantRC := 0
	if e.RC != nil {
		wantRC = *e.RC
	}
	if exit != wantRC {
		facts = append(facts, fmt.Sprintf("exit: got %d, want %d", exit, wantRC))
	}
	if e.Out != "" {
		if !match(e.Out, []byte(stdout)) {
			facts = append(facts, fmt.Sprintf("stdout: %s", spellMismatch(e.Out, stdout)))
		}
	}
	for _, path := range sortedKeys(e.Files) {
		want := e.Files[path]
		data, err := readFile(path)
		if err != nil {
			facts = append(facts, fmt.Sprintf("file %s: %v", path, err))
			continue
		}
		if !match(want, data) {
			facts = append(facts, fmt.Sprintf("file %s: %s", path, spellMismatch(want, string(data))))
		}
	}
	return facts
}

// match — one wanted spelling against the measured bytes.
func match(want string, got []byte) bool {
	if strings.HasPrefix(want, ContainsPrefix) {
		return bytes.Contains(got, []byte(strings.TrimPrefix(want, ContainsPrefix)))
	}
	return bytes.Equal(got, []byte(want))
}

// spellMismatch — the mismatch as the expectation's own terms: the
// wanted spelling and a bounded slice of what was measured.
func spellMismatch(want, got string) string {
	const bound = 80
	if len(got) > bound {
		got = got[:bound] + "…"
	}
	return fmt.Sprintf("want %q, got %q", want, got)
}

// sortedKeys — a map's keys in order, for a deterministic walk.
func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

package doc

// The line-range claims' laws: one range spelling everywhere —
// "path/file.go:12-48" — parsed through the fingerprint package's
// one parser, and the span helpers the gates and the coverage
// arithmetic share.

import (
	"strconv"

	"github.com/neurophant/punchtape/internal/domain/fprint"
)

// ParseLineRange — one claim's shape through the one parser: a
// root-relative path, a first line, a last line.
func ParseLineRange(s string) (rel string, first, last int, err error) {
	return fprint.ParseLineRange(s)
}

// DemandID — the D spelling of one line number.
func DemandID(n int) string { return "D" + strconv.Itoa(n) }

// ClaimCarries — whether one claim names real lines of the tree:
// the spelling parses and the tree's line count covers it. The
// count answers -1 when the resolver is absent or the file does not
// stand.
func ClaimCarries(claim string, fileLines func(rel string) (int, bool)) bool {
	rel, first, last, err := fprint.ParseLineRange(claim)
	if err != nil {
		return false
	}
	if fileLines == nil {
		return true
	}
	n, ok := fileLines(rel)
	return ok && last <= n && first >= 1
}

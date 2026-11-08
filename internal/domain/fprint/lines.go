package fprint

// The line-range fingerprint laws: the one range spelling —
// "path/file.go:12-48", a root-relative path, a first line, a last
// line, first <= last — the digest of a range's own bytes, and the
// coverage arithmetic over the tree's line counts. A range is the
// unit of coverage and the unit of staleness: a changed line moves
// its range's digest and nothing else; a line no range holds is a
// hole, aggregated to maximal ranges per file.

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/domain"
	"github.com/neurophant/punchtape/internal/domain/zone"
)

// ParseLineRange — one range spelling into its parts; any other
// shape answers its own error.
func ParseLineRange(s string) (rel string, first, last int, err error) {
	notOne := fmt.Errorf("%q is not a line range path:first-last", s)
	pathPart, rangePart, found := strings.Cut(s, ":")
	if !found || zone.ValidateRel(pathPart) != nil {
		return "", 0, 0, notOne
	}
	lo, hi, dotted := strings.Cut(rangePart, "-")
	if !dotted {
		return "", 0, 0, notOne
	}
	a, aerr := lineNo(lo)
	b, berr := lineNo(hi)
	if aerr != nil || berr != nil || a < 1 || b < a {
		return "", 0, 0, notOne
	}
	return pathPart, a, b, nil
}

// lineNo — a plain decimal number from 1 up: no sign, no leading
// zeros.
func lineNo(s string) (int, error) {
	if s == "" || (len(s) > 1 && s[0] == '0') {
		return 0, fmt.Errorf("%q is not a line number", s)
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, fmt.Errorf("%q is not a line number", s)
		}
	}
	return strconv.Atoi(s)
}

// Token12 — the first twelve hex of a digest: the short spelling
// the faces print and the proofs must carry. The caller passes the
// digest of the thing the token names — the claims list's own
// digest, the demand quote's digest — and the token is its head.
func Token12(d string) string {
	if len(d) > 12 {
		return d[:12]
	}
	return d
}

// FileDigest — one walked file's line in the tree's own
// fingerprint: its root-relative path and its bytes' digest;
// UnreadableDigest spells a file the walk could not read.
type FileDigest struct {
	Path   string
	Digest string
}

// UnreadableDigest — the line a file the walk could not read spells
// in the tree's fingerprint.
const UnreadableDigest = "unreadable"

// TreeDigest — the working tree's own fingerprint: the sha256 over
// the sorted "path\0digest" lines of every counted file — binaries
// included by their bytes.
func TreeDigest(files []FileDigest) string {
	sorted := append([]FileDigest(nil), files...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Path < sorted[j].Path })
	var b bytes.Buffer
	for _, f := range sorted {
		b.WriteString(f.Path)
		b.WriteByte(0)
		b.WriteString(f.Digest)
		b.WriteByte('\n')
	}
	return domain.Digest(b.Bytes())
}

// CountLines — the line count of a file's bytes: the number of
// lines, an empty file carrying none.
func CountLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	n := 1 + bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] == '\n' {
		n--
	}
	return n
}

// LineRangeDigest — the sha256 of a range's own bytes: the file's
// bytes from the first line's start through the last line's end,
// exactly as they stand. AbsentDigest when the lines do not exist
// in the bytes given.
func LineRangeDigest(data []byte, first, last int) string {
	lines := CountLines(data)
	if first < 1 || last > lines {
		return AbsentDigest
	}
	start := offsetOf(data, first)
	end := offsetOf(data, last+1) // one past the last line's newline
	return domain.Digest(data[start:end])
}

// offsetOf — the byte offset of a line's start; the offset past the
// end for the line count plus one. Lines are the spans between
// newlines, the final line allowed to run without one.
func offsetOf(data []byte, line int) int {
	if line <= 1 {
		return 0
	}
	seen := 1
	for i := 0; i < len(data); i++ {
		if data[i] != '\n' {
			continue
		}
		seen++
		if seen == line {
			return i + 1
		}
	}
	return len(data)
}

// FileLineCount — one tree file's own facts for the coverage
// universe: its path and its line count. A file the universe cannot
// own carries the reason instead of a count.
type FileLineCount struct {
	Path  string
	Lines int
	// Outside is "" when the file is coverage material; otherwise the
	// named reason it stands outside the universe wholesale.
	Outside string
}

// Range — one maximal span of lines, first to last, inside one file.
type Range struct {
	First int
	Last  int
}

// FileExclusion — one file outside the coverage universe, with the
// reason a reader can weigh.
type FileExclusion struct {
	Path   string
	Reason string
}

// Coverage — the tree against the claims and the loud exclusions:
// the universe's size, the held part, the open lines per file
// aggregated to maximal ranges, and the files outside the universe
// with their reasons.
type Coverage struct {
	Total    int
	Claimed  int
	Holes    map[string][]Range
	Excluded []FileExclusion
	// ExcludedLines — the lines only a loud exclusion holds, inside
	// the counted universe: the share the state prints is measured
	// over these, never over the spellings' own span.
	ExcludedLines int
}

// LineCoverage — the claims and the loud exclusions against the
// tree's line counts: pure set arithmetic over (file, line) pairs,
// one pass. A loud exclusion covers like a claim; a spelling that
// does not parse, a range past a file's own lines or a file outside
// the universe holds nothing — the lines it names stay honest
// holes.
func LineCoverage(claims []string, excl []string, files []FileLineCount) Coverage {
	linesOf := map[string]int{}
	for _, f := range files {
		if f.Outside == "" && f.Lines > 0 {
			linesOf[f.Path] = f.Lines
		}
	}
	claimHeld := map[string]map[int]bool{}
	exclHeld := map[string]map[int]bool{}
	hold := func(spelling string, held map[string]map[int]bool) {
		rel, first, last, err := ParseLineRange(spelling)
		if err != nil {
			return
		}
		limit, known := linesOf[rel]
		if !known {
			return
		}
		if last > limit {
			last = limit
		}
		if held[rel] == nil {
			held[rel] = map[int]bool{}
		}
		for n := first; n <= last; n++ {
			held[rel][n] = true
		}
	}
	for _, claim := range claims {
		hold(claim, claimHeld)
	}
	for _, ex := range excl {
		hold(ex, exclHeld)
	}
	cov := Coverage{Holes: map[string][]Range{}}
	for _, f := range files {
		if f.Outside != "" {
			cov.Excluded = append(cov.Excluded, FileExclusion{Path: f.Path, Reason: f.Outside})
			continue
		}
		cov.Total += f.Lines
		claimed, excluded := claimHeld[f.Path], exclHeld[f.Path]
		var open []int
		for n := 1; n <= f.Lines; n++ {
			if claimed[n] || excluded[n] {
				cov.Claimed++
				if excluded[n] {
					cov.ExcludedLines++
				}
				continue
			}
			open = append(open, n)
		}
		if len(open) > 0 {
			cov.Holes[f.Path] = maximalRanges(open)
		}
	}
	sort.Slice(cov.Excluded, func(i, j int) bool { return cov.Excluded[i].Path < cov.Excluded[j].Path })
	return cov
}

// maximalRanges — ascending line numbers to maximal ranges: adjacent
// numbers merge, every gap cuts.
func maximalRanges(nums []int) []Range {
	var out []Range
	for _, n := range nums {
		if len(out) > 0 && n == out[len(out)-1].Last+1 {
			out[len(out)-1].Last = n
			continue
		}
		out = append(out, Range{First: n, Last: n})
	}
	return out
}

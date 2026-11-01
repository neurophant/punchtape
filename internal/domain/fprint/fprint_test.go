// The fingerprint laws' battery: the range grammar's every shape,
// the range digest's exact slices, the line counts, and the proof's
// move on every dependency's move.
package fprint

import (
	"testing"

	"github.com/neurophant/punchtape/internal/domain"
)

func TestParseLineRangeEveryShape(t *testing.T) {
	for _, ok := range []string{"a.go:1-1", "src/a.go:12-48", "deep/dir/x.txt:3-3", "a b.go:1-2"} {
		rel, first, last, err := ParseLineRange(ok)
		if err != nil || rel == "" || first < 1 || last < first {
			t.Fatalf("%q must parse: %v", ok, err)
		}
	}
	for _, broken := range []string{
		"a.go", "a.go:1", "a.go:1-", "a.go:-2", "a.go:0-2", "a.go:5-3",
		"a.go:01-2", "a.go:+1-2", "a.go:1-2:3", "/abs.go:1-2", "../x:1-2",
		"a.go:1-2,x", "",
	} {
		if _, _, _, err := ParseLineRange(broken); err == nil {
			t.Fatalf("%q must be refused", broken)
		}
	}
}

func TestCountLines(t *testing.T) {
	cases := []struct {
		data []byte
		want int
	}{
		{nil, 0},
		{[]byte(""), 0},
		{[]byte("one"), 1},
		{[]byte("one\n"), 1},
		{[]byte("one\ntwo\nthree"), 3},
		{[]byte("one\ntwo\nthree\n"), 3},
		{[]byte("\n\n"), 2},
	}
	for _, c := range cases {
		if got := CountLines(c.data); got != c.want {
			t.Fatalf("CountLines(%q) = %d, want %d", c.data, got, c.want)
		}
	}
}

func TestRangeDigestSlices(t *testing.T) {
	data := []byte("alpha\nbeta\ngamma\ndelta\n")
	// The digest of lines [2..3] is the digest of the file's own
	// bytes from line two's start through line three's end.
	if got := LineRangeDigest(data, 2, 3); got != domain.Digest([]byte("beta\ngamma\n")) {
		t.Fatalf("the slice must be the file's own bytes: %s", got)
	}
	if got := LineRangeDigest(data, 4, 4); got != domain.Digest([]byte("delta\n")) {
		t.Fatalf("the last line carries its own bytes: %s", got)
	}
	// A changed line inside the range moves the digest; a change
	// outside it does not.
	moved := []byte("alpha\nBETA\ngamma\ndelta\n")
	if LineRangeDigest(moved, 2, 3) == LineRangeDigest(data, 2, 3) {
		t.Fatal("a changed line inside the range must move the digest")
	}
	if LineRangeDigest(moved, 4, 4) != LineRangeDigest(data, 4, 4) {
		t.Fatal("a change outside the range must not move its digest")
	}
	// Absent lines answer absence.
	for _, r := range [][2]int{{0, 1}, {1, 5}, {5, 6}} {
		if got := LineRangeDigest(data, r[0], r[1]); got != AbsentDigest {
			t.Fatalf("lines %v do not exist: %s", r, got)
		}
	}
	if got := LineRangeDigest(nil, 1, 1); got != AbsentDigest {
		t.Fatalf("no bytes carry no lines: %s", got)
	}
}

func TestProofMovesWithItsDependencies(t *testing.T) {
	p := Proof([]byte("run: x"), "claims", "env")
	if p != Proof([]byte("run: x"), "claims", "env") {
		t.Fatalf("proof must be stable")
	}
	if p == Proof([]byte("run: y"), "claims", "env") {
		t.Fatalf("row bytes must move the proof")
	}
	if p == Proof([]byte("run: x"), "claims2", "env") {
		t.Fatalf("the claims digest must move the proof")
	}
	if p == Proof([]byte("run: x"), "claims", "env2") {
		t.Fatalf("the environment must move the proof")
	}
}

// TestLineCoverageArithmetic — the coverage universe against the
// claims and the loud exclusions: pure set arithmetic over lines,
// holes aggregated to maximal ranges, outside files named with
// their reasons.
func TestLineCoverageArithmetic(t *testing.T) {
	files := []FileLineCount{
		{Path: "a.txt", Lines: 5},
		{Path: "b.txt", Lines: 3},
		{Path: "nul.bin", Lines: 7, Outside: "a NUL byte"},
		{Path: "empty.txt", Lines: 0, Outside: "no lines"},
	}
	// The whole universe claimed: no holes, the outside files never
	// enter the arithmetic.
	cov := LineCoverage([]string{"a.txt:1-5", "b.txt:1-3"}, nil, files)
	if cov.Total != 8 || cov.Claimed != 8 || len(cov.Holes) != 0 {
		t.Fatalf("the claimed universe must close: %+v", cov)
	}
	if len(cov.Excluded) != 2 || cov.Excluded[0].Path != "empty.txt" || cov.Excluded[1].Path != "nul.bin" {
		t.Fatalf("the outside files must stand named, sorted: %+v", cov.Excluded)
	}
	// Holes aggregate to maximal ranges: one open line inside the
	// claims, and the open edges of a partial claim.
	cov = LineCoverage([]string{"a.txt:2-4", "b.txt:2-3"}, nil, files)
	if cov.Total != 8 || cov.Claimed != 5 {
		t.Fatalf("the partial claims count their own lines: %+v", cov)
	}
	if len(cov.Holes) != 2 ||
		len(cov.Holes["a.txt"]) != 2 ||
		cov.Holes["a.txt"][0] != (Range{First: 1, Last: 1}) ||
		cov.Holes["a.txt"][1] != (Range{First: 5, Last: 5}) ||
		len(cov.Holes["b.txt"]) != 1 ||
		cov.Holes["b.txt"][0] != (Range{First: 1, Last: 1}) {
		t.Fatalf("holes must aggregate per file: %+v", cov.Holes)
	}
	// Adjacent holes merge into one maximal range.
	cov = LineCoverage([]string{"a.txt:1-1", "a.txt:5-5"}, nil, files)
	if len(cov.Holes["a.txt"]) != 1 || cov.Holes["a.txt"][0] != (Range{First: 2, Last: 4}) {
		t.Fatalf("adjacent holes must merge: %+v", cov.Holes)
	}
	// Loud exclusions cover like claims; claims past a file's own
	// lines, broken spellings and claims on outside or absent files
	// count nothing.
	cov = LineCoverage([]string{"a.txt:1-1"}, []string{"a.txt:2-5"}, files)
	if cov.Claimed != 5 || len(cov.Holes) != 1 || len(cov.Holes["b.txt"]) != 1 {
		t.Fatalf("exclusions must cover like claims: %+v", cov)
	}
	if cov = LineCoverage([]string{"b.txt:1-9", "nul.bin:1-1", "gone.txt:1-1", "junk"}, nil, files); cov.Claimed != 3 {
		t.Fatalf("unreal claims count nothing: %+v", cov)
	}
	// No claims at all: the whole universe is holes.
	cov = LineCoverage(nil, nil, files)
	if cov.Total != 8 || cov.Claimed != 0 || len(cov.Holes) != 2 {
		t.Fatalf("a bare tree is whole open: %+v", cov)
	}
}

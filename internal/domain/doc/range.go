package doc

// The demand stream's own vocabulary: the line universe's one law —
// a text's non-empty trimmed lines, in order — and the range grammar
// over it, both directions: a range list ("D3", "D5..D9",
// comma-joined) parses to its line numbers, and a number set spells
// back into the same grammar with runs collapsed. The demand
// stream's non-empty lines number from D1 in file order; a range
// names them alone — one spelling everywhere the machine reads or
// prints them.

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// NonEmptyLines — a text's non-empty lines, trimmed, in order: the
// one law of the demand stream's universe. Blank lines frame and
// carry nothing; every other line is a demand line.
func NonEmptyLines(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// ParseRanges — the range strings as one sorted set of line numbers.
// A broken spelling answers its own error; the caller names the
// field it rode in.
func ParseRanges(ranges []string) ([]int, error) {
	return ParseRangesWithin(ranges, 0)
}

// RangeError — a range naming a line past the stream's own count: the
// generation of a..b never runs past the stream, the error names the
// line and the count.
type RangeError struct {
	Line int
	Max  int
}

// Error — the overflow's own fact.
func (e *RangeError) Error() string {
	return fmt.Sprintf("the range names line D%d past the stream's %d lines", e.Line, e.Max)
}

// ParseRangesWithin — ParseRanges against a stream of max lines: a
// range whose last line runs past the stream answers a *RangeError
// before any number is generated. A max below one leaves every
// spelling to the plain grammar — no stream to bound by.
func ParseRangesWithin(ranges []string, max int) ([]int, error) {
	var out []int
	seen := map[int]bool{}
	for _, list := range ranges {
		for _, elem := range strings.Split(list, ",") {
			a, b, err := parseRange(strings.TrimSpace(elem))
			if err != nil {
				return nil, err
			}
			if max > 0 && b > max {
				return nil, &RangeError{Line: b, Max: max}
			}
			for n := a; n <= b; n++ {
				if !seen[n] {
					seen[n] = true
					out = append(out, n)
				}
			}
		}
	}
	sort.Ints(out)
	return out, nil
}

// parseRange — one element: D<n> or D<a>..D<b>, ascending.
func parseRange(elem string) (int, int, error) {
	lo, hi, dotted := strings.Cut(elem, "..")
	a, err := demandNumber(lo)
	if err != nil {
		return 0, 0, err
	}
	if !dotted {
		return a, a, nil
	}
	b, err := demandNumber(hi)
	if err != nil || b < a {
		return 0, 0, fmt.Errorf("the range %q must ascend", elem)
	}
	return a, b, nil
}

// demandNumber — "D" and a plain decimal number from 1 up: one
// spelling, no leading zeros, no sign.
func demandNumber(s string) (int, error) {
	notOne := fmt.Errorf("%q is not a demand line D<n>", s)
	if !strings.HasPrefix(s, "D") {
		return 0, notOne
	}
	digits := s[1:]
	if digits == "" || (len(digits) > 1 && digits[0] == '0') {
		return 0, notOne
	}
	for i := 0; i < len(digits); i++ {
		if digits[i] < '0' || digits[i] > '9' {
			return 0, notOne
		}
	}
	n, err := strconv.Atoi(digits)
	if err != nil || n < 1 {
		return 0, notOne
	}
	return n, nil
}

// SpellRanges — the numbers in the range grammar, consecutive runs
// collapsed: [19, 23, 24, 25] spells "D19, D23..D25". An empty set
// spells the empty string.
func SpellRanges(nums []int) string {
	if len(nums) == 0 {
		return ""
	}
	sorted := append([]int(nil), nums...)
	sort.Ints(sorted)
	var b strings.Builder
	start, prev := sorted[0], sorted[0]
	spell := func(end int) {
		switch {
		case end == start:
			fmt.Fprintf(&b, "D%d", start)
		case end == start+1:
			fmt.Fprintf(&b, "D%d, D%d", start, end)
		default:
			fmt.Fprintf(&b, "D%d..D%d", start, end)
		}
	}
	for _, n := range sorted[1:] {
		if n == prev {
			continue
		}
		if n == prev+1 {
			prev = n
			continue
		}
		spell(prev)
		b.WriteString(", ")
		start, prev = n, n
	}
	spell(prev)
	return b.String()
}

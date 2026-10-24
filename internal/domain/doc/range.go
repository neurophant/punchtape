package doc

// The wish-line ranges' one grammar, both directions: a range list
// ("W3", "W5..W9", comma-joined) parses to its line numbers, and a
// number set spells back into the same grammar with runs collapsed.
// The wish's lines number from W1 in file order; a range names them
// alone — one spelling everywhere the machine reads or prints them.

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ParseRanges — the range strings as one sorted set of line numbers.
// A broken spelling answers its own error; the caller names the
// field it rode in.
func ParseRanges(ranges []string) ([]int, error) {
	var out []int
	seen := map[int]bool{}
	for _, list := range ranges {
		for _, elem := range strings.Split(list, ",") {
			a, b, err := parseRange(strings.TrimSpace(elem))
			if err != nil {
				return nil, err
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

// parseRange — one element: W<n> or W<a>..W<b>, ascending.
func parseRange(elem string) (int, int, error) {
	lo, hi, dotted := strings.Cut(elem, "..")
	a, err := wishNumber(lo)
	if err != nil {
		return 0, 0, err
	}
	if !dotted {
		return a, a, nil
	}
	b, err := wishNumber(hi)
	if err != nil || b < a {
		return 0, 0, fmt.Errorf("the range %q must ascend", elem)
	}
	return a, b, nil
}

// wishNumber — "W" and a plain decimal number from 1 up: one
// spelling, no leading zeros, no sign.
func wishNumber(s string) (int, error) {
	notOne := fmt.Errorf("%q is not a wish line W<n>", s)
	if !strings.HasPrefix(s, "W") {
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
// collapsed: [19, 23, 24, 25] spells "W19, W23..W25". An empty set
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
			fmt.Fprintf(&b, "W%d", start)
		case end == start+1:
			fmt.Fprintf(&b, "W%d, W%d", start, end)
		default:
			fmt.Fprintf(&b, "W%d..W%d", start, end)
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

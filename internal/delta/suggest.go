package delta

import (
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
)

// With the rejection line the machine attaches a corrected hint —
// but only when the correction is uniquely determined. Semantics is
// never guessed: everything non-deterministic stays without a hint.
// The decode surface needs no form repairs: a failure is loud,
// addressed and final — the only surviving hint is the delta kind.

var kindKinds = []string{
	KindSpec, KindRun,
}

// editDistance — Levenshtein distance for short words.
func editDistance(a, b string) int {
	la, lb := len(a), len(b)
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = minInt(minInt(cur[j-1]+1, prev[j]+1), prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[lb]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SuggestKind — an unknown delta kind: the single known kind at an
// edit distance <=2 becomes a hint for the kind line.
func SuggestKind(kind string) (string, bool) {
	got := strings.ToLower(strings.TrimSpace(kind))
	if got == "" || knownKind(got) {
		return "", false
	}
	var best string
	bestD := 3
	ambig := false
	for _, k := range kindKinds {
		d := editDistance(got, k)
		if d > 2 {
			continue
		}
		if d < bestD {
			best, bestD, ambig = k, d, false
		} else if d == bestD {
			ambig = true
		}
	}
	if best != "" && !ambig {
		return canondata.T("format.suggest.kind", canondata.M{"kind": best}), true
	}
	return "", false
}

func knownKind(k string) bool {
	for _, known := range kindKinds {
		if k == known {
			return true
		}
	}
	return false
}

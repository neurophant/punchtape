package delta

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
)

// With the rejection line the machine attaches a corrected
// submission line — but only when the correction is uniquely determined.
// Semantics is never guessed: a hint is a mechanical fix of the form;
// everything non-deterministic stays without a hint.

var (
	kindKinds = []string{
		KindIntent, KindLint, KindClarify, KindSpec, KindChecks, KindAssert,
		KindSplit, KindCode, KindFix, KindProbe,
	}
	lineErrRe    = regexp.MustCompile(`line (\d+)`)
	indicatorRe  = regexp.MustCompile(`^[@` + "`" + `|%>&*!{}[\],#]`)
	mapValueLine = regexp.MustCompile(`^(\s*-?\s*[\w.-]+:\s)(.+)$`)
)

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

// SuggestLine returns the corrected line when the correction is
// uniquely determined; ok=false — no hint.
func SuggestLine(text string, decodeErr error, kind string) (string, bool) {
	// (1) Unknown delta kind: the single known kind at an edit
	// distance <=2 becomes a hint for the kind line.
	if hint, ok := SuggestKind(kind); ok {
		return hint, true
	}
	// (2)-(3) A decode error with a line number: an indicator-leading
	// scalar and ": " inside the value are fixed with a quote —
	// mechanically and uniquely (ambiguous constructs are not in the
	// whitelist).
	if decodeErr == nil {
		return "", false
	}
	m := lineErrRe.FindStringSubmatch(decodeErr.Error())
	if m == nil {
		return "", false
	}
	lineNo := 0
	fmt.Sscanf(m[1], "%d", &lineNo)
	lines := strings.Split(text, "\n")
	if lineNo < 1 || lineNo > len(lines) {
		return "", false
	}
	mm := mapValueLine.FindStringSubmatch(lines[lineNo-1])
	if mm == nil {
		return "", false
	}
	prefix, value := mm[1], strings.TrimRight(mm[2], " \t")
	if value == "" || strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") {
		return "", false
	}
	msg := decodeErr.Error()
	// An indicator-leading scalar (|, > etc.) gets quoted; a value
	// starting with { or [ is a structure, not a scalar: quoting its
	// fragment would be a garbage hint — a rejection is more honest.
	if strings.Contains(msg, canondata.T("format.match.cannot-start-token")) &&
		!strings.HasPrefix(value, "{") && !strings.HasPrefix(value, "[") &&
		indicatorRe.MatchString(value) {
		return prefix + doubleQuote(value), true
	}
	if strings.Contains(msg, canondata.T("format.match.mapping-values-short")) &&
		strings.Contains(value, ": ") {
		return prefix + doubleQuote(value), true
	}
	return "", false
}

// DiagnoseYAML — the root cause of a dialect decode error in one line.
// Only actual pitfall patterns are checked: an anchor (not supported),
// a comma in an unquoted flow-map value (splits it into entries), a
// block scalar inside a flow map. Everything else — ok=false, the
// rejection stays as is.
func DiagnoseYAML(text string, decodeErr error) (string, bool) {
	if decodeErr == nil {
		return "", false
	}
	msg := decodeErr.Error()
	if strings.Contains(msg, "unknown anchor") {
		return canondata.T("format.diag.anchor"), true
	}
	m := lineErrRe.FindStringSubmatch(msg)
	if m == nil {
		return "", false
	}
	lineNo := 0
	fmt.Sscanf(m[1], "%d", &lineNo)
	lines := strings.Split(text, "\n")
	if lineNo < 1 || lineNo > len(lines) {
		return "", false
	}
	line := lines[lineNo-1]
	// A block scalar inside a flow map — the parser says "cannot start
	// any token"; the cause is on the line with {.
	if strings.Contains(msg, canondata.T("format.match.cannot-start-token")) && strings.Contains(line, "{") {
		return canondata.T("format.diag.flow-block", canondata.M{"line": m[1]}), true
	}
	// A comma in an unquoted flow-map value: a segment after the comma
	// without "key:" — the value is torn apart; the characteristic
	// error classes are a foreign field ("field … not found in type",
	// "unknown field") and a type mismatch (pieces of the value text
	// became keys).
	if decodeConfusedClass(msg) && flowCommaSplit(line) {
		return canondata.T("format.diag.flow-comma", canondata.M{"line": m[1]}), true
	}
	if strings.Contains(msg, "did not find expected key") {
		return canondata.T("format.diag.long-scalar"), true
	}
	return "", false
}

// decodeConfusedClass — the characteristic decode-error classes of a
// torn flow value: a foreign field ("field … not found in type",
// "unknown field") and a type mismatch (pieces of the value text
// became keys).
func decodeConfusedClass(msg string) bool {
	return unknownFieldRepairRe.MatchString(msg) ||
		strings.Contains(msg, "field ") && strings.Contains(msg, " not found in type") ||
		strings.Contains(msg, "cannot unmarshal")
}

// flowCommaSplit — whether the line's flow map has a segment after a
// comma without "key:" (outside quotes): the comma tore the value apart.
func flowCommaSplit(line string) bool {
	open := strings.Index(line, "{")
	closeIdx := strings.LastIndex(line, "}")
	if open < 0 || closeIdx <= open {
		return false
	}
	for _, s := range splitFlowSegments(line[open+1 : closeIdx]) {
		if !strings.Contains(s, ":") {
			return true
		}
	}
	return false
}

// splitFlowSegments splits the inside of a flow map on commas outside
// quotes; nested brackets are NOT tracked — callers bail on them.
func splitFlowSegments(inner string) []string {
	var segs []string
	var cur strings.Builder
	var quote byte
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
			cur.WriteByte(c)
		case c == '\'' || c == '"':
			quote = c
			cur.WriteByte(c)
		case c == ',':
			segs = append(segs, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	segs = append(segs, cur.String())
	return segs
}

// doubleQuote wraps the value in double quotes, escaping the backslash
// and the quote itself.
func doubleQuote(v string) string {
	return "\"" + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(v) + "\""
}

func knownKind(k string) bool {
	for _, known := range kindKinds {
		if k == known {
			return true
		}
	}
	return false
}

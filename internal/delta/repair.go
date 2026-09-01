// A whitelist for submission format repair: deterministic and closed.
// Only purely formal defects are repaired — a duplicated "# TOKENS
// <n>" trailer, a typo in a known field name (the single candidate at
// an edit distance of at most 2), an unbalanced scalar quote, a tab
// in the indentation outside a literal block, and a flow value torn
// by an unquoted comma. Everything outside the
// list — a rejection as today; semantics is never repaired. Every
// repair is returned as a line for the submission reply.
package delta

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// knownFieldSet — the field names of all delta structs (yaml keys),
// collected by reflection once at package start.
var knownFieldSet = collectKnownFields()

func collectKnownFields() map[string]bool {
	names := map[string]bool{}
	visited := map[reflect.Type]bool{}
	var walk func(t reflect.Type)
	walk = func(t reflect.Type) {
		for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Map || t.Kind() == reflect.Array {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct || visited[t] {
			return
		}
		visited[t] = true
		for i := 0; i < t.NumField(); i++ {
			tag := t.Field(i).Tag.Get("yaml")
			if tag == "" || tag == "-" {
				continue
			}
			key := strings.Split(tag, ",")[0]
			if key != "" {
				names[key] = true
			}
			walk(t.Field(i).Type)
		}
	}
	for _, v := range []any{
		&Delta{}, &Intent{}, &Lint{}, &Clarify{}, &Spec{}, &SpecOperation{},
		&canon.Requirement{}, &canon.Scenario{}, &canon.When{}, &canon.Assertion{}, &Assert{},
		&Split{}, &FileChange{}, &Code{}, &Fix{}, &Probe{},
	} {
		walk(reflect.TypeOf(v))
	}
	return names
}

// editDistanceAtMost — Levenshtein distance with early exit: true
// when the distance does not exceed limit.
func editDistanceAtMost(a, b string, limit int) bool {
	if abs(len(a)-len(b)) > limit {
		return false
	}
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		rowMin := cur[0]
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = minInt(prev[j]+1, minInt(cur[j-1]+1, prev[j-1]+cost))
			if cur[j] < rowMin {
				rowMin = cur[j]
			}
		}
		if rowMin > limit {
			return false
		}
		prev, cur = cur, prev
	}
	return prev[len(b)] <= limit
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var (
	tokenTailRe          = regexp.MustCompile(`^# TOKENS \d+$`)
	mappingRe            = regexp.MustCompile(`^(\s*)([A-Za-z0-9_.-]+):(?:[ \t]+(.*))?$`)
	literalOpenRe        = regexp.MustCompile(`^(\s*)[^ #][^:]*:(\s*[|>][-+0-9]*)?\s*$`)
	unknownFieldRepairRe = regexp.MustCompile(`line (\d+): unknown field (\S+)`)
)

// RepairFormat repairs the submission format by whitelist and returns
// the normalized text and "fixed: …" lines for the reply. A submission
// without a "# TOKENS <n>" trailer is legal and is not repaired.
func RepairFormat(text string) (string, []string) {
	var notes []string

	// (1) Duplicated trailer: consecutive "# TOKENS <n>" lines at the
	// end — the last one stays, the counter is booked by it as usual.
	lines := strings.Split(text, "\n")
	end := len(lines)
	for end > 0 && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	tails := 0
	for end-tails-1 >= 0 && tokenTailRe.MatchString(strings.TrimSpace(lines[end-tails-1])) {
		tails++
	}
	if tails > 1 {
		cut := lines[:end-tails+1]
		rest := lines[end:]
		lines = append(cut, rest...)
		notes = append(notes, canondata.T("format.fix.tokens-trailer",
			canondata.M{"count": fmt.Sprintf("%d", tails)}))
	}
	text = strings.Join(lines, "\n")

	// (2) Quotes and indentation: a text pass outside literal blocks —
	// block contents (product code) are untouchable.
	lines = strings.Split(text, "\n")
	blockIndent := -1
	prevIndent := ""
	for i, line := range lines {
		trimmed := strings.TrimRight(line, " \t\r")
		if trimmed == "" {
			continue
		}
		indentLen := len(line) - len(strings.TrimLeft(line, " \t"))
		if blockIndent >= 0 {
			if indentLen > blockIndent {
				prevIndent = line[:indentLen]
				continue
			}
			blockIndent = -1
		}
		if m := literalOpenRe.FindStringSubmatch(line); m != nil && m[2] != "" {
			blockIndent = indentLen
			prevIndent = m[1]
			continue
		}
		// A tab in the indentation outside a block: replaced with the
		// previous meaningful line's indentation — YAML forbids tabs in
		// the indentation entirely.
		if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " \t") {
			lead := line[:indentLen]
			if strings.Contains(lead, "\t") {
				lines[i] = prevIndent + strings.TrimLeft(line, " \t")
				notes = append(notes, canondata.T("format.fix.tab-indent",
					canondata.M{"line": fmt.Sprintf("%d", i+1)}))
				line = lines[i]
			}
		}
		// An unclosed quote of a SHORT scalar: the value starts with a
		// quote that occurs an odd number of times in the line — the
		// quote is closed at the end of the line. Long scalars are not
		// repaired: a broken quote in a long line usually means a real
		// newline inside the value (block scalar territory); a
		// mechanical close would repair the wrong thing — a rejection
		// is more honest (the threshold comes from delivery data). Only
		// the block form `key: "`: a corrupted quote inside a flow
		// mapping is ambiguous (it is unclear where the string ends) —
		// the repair whitelist does not touch such lines; a rejection as
		// today.
		if m := mappingRe.FindStringSubmatch(line); m != nil && m[3] != "" {
			q := m[3][0]
			if q == '"' || q == '\'' {
				if len(m[3]) <= canondata.Limit("repair.quote-value-max") &&
					strings.Count(m[3], string(q))%2 == 1 {
					lines[i] = line + string(q)
					notes = append(notes, canondata.T("format.fix.closed-quote",
						canondata.M{"line": fmt.Sprintf("%d", i+1)}))
				}
			}
		}
		prevIndent = line[:len(line)-len(strings.TrimLeft(line, " \t"))]
	}
	text = strings.Join(lines, "\n")

	// (3)-(5) Decode-driven repairs: a known field typo, quoting a
	// scalar with ": ", a closed mapping with a single entry instead
	// of a list — each repairs exactly one line, then the decode
	// repeats. Everything outside the whitelist — a rejection as today.
	type fieldRename struct {
		idx      int
		from     string
		noteIdxs []int
	}
	renames := []fieldRename{}
	for attempt := 0; attempt < 8; attempt++ {
		var d Delta
		err := yamlio.DecodeStrict([]byte(text), &d)
		if err == nil {
			break
		}
		msg := err.Error()

		// A "no worse than before" guard: a text edit is applied only
		// when the decode after it has not changed the error class. A
		// repair that produced a new error (a quote broke the block
		// structure) is rolled back — the executor is honestly shown
		// the original cause.
		tryApply := func(fixed string, ok bool, note string) bool {
			if !ok {
				return false
			}
			if derr := yamlio.DecodeStrict([]byte(fixed), &Delta{}); derr != nil && !repairableDecode(derr.Error()) {
				return false
			}
			text = fixed
			notes = append(notes, note)
			return true
		}

		// (4) A scalar with ": " inside the value: the "mapping values
		// are not allowed" rejection — the value is taken into double
		// quotes, the single mechanical fix.
		if strings.Contains(msg, canondata.T("format.match.mapping-values")) {
			fixed, ok := quoteColonScalar(text, msg)
			if tryApply(fixed, ok, canondata.T("format.fix.quoted-scalar")) {
				continue
			}
			break
		}

		// (5) A closed mapping with a single entry where a list is
		// expected: a {…} value is wrapped into […].
		if strings.Contains(msg, "cannot unmarshal !!map into") {
			fixed, ok := wrapClosedMap(text, msg)
			if tryApply(fixed, ok, canondata.T("format.fix.wrapped-map")) {
				continue
			}
			break
		}

		// (5b) A comma inside an unquoted flow-map value tears the
		// value into entries (a segment after the comma without
		// "key:"): the torn segments are merged back into the value
		// they came from, and the value goes into double quotes.
		if decodeConfusedClass(msg) {
			if fixed, ok := quoteFlowComma(text, msg); ok {
				lineNo := ""
				if lm := lineErrRe.FindStringSubmatch(msg); lm != nil {
					lineNo = lm[1]
				}
				if tryApply(fixed, true, canondata.T("format.fix.flow-comma", canondata.M{"line": lineNo})) {
					continue
				}
			}
		}

		m := unknownFieldRepairRe.FindStringSubmatch(msg)
		if m == nil {
			break // outside the whitelist — a rejection as today
		}
		lineNo, field := m[1], m[2]
		var candidate string
		found := 0
		for known := range knownFieldSet {
			if known != field && editDistanceAtMost(field, known, 2) {
				candidate = known
				found++
			}
		}
		if found != 1 {
			break // not a single candidate — ambiguous, reject
		}
		lines = strings.Split(text, "\n")
		idx := -1
		if n, ok := atoiSafe(lineNo); ok && n >= 1 && n <= len(lines) {
			keyRe := regexp.MustCompile(`^(\s*)` + regexp.QuoteMeta(field) + `(\s*):`)
			if keyRe.MatchString(lines[n-1]) {
				idx = n - 1
			}
		}
		if idx < 0 {
			// The line number is relative to the nested decode fragment —
			// the repair is allowed only with a single occurrence of
			// the key.
			keyRe := regexp.MustCompile(`^(\s*)` + regexp.QuoteMeta(field) + `(\s*):`)
			for i, line := range lines {
				if keyRe.MatchString(line) {
					if idx >= 0 {
						idx = -2 // more than one occurrence — ambiguous
						break
					}
					idx = i
				}
			}
		}
		if idx < 0 {
			break
		}
		keyRe := regexp.MustCompile(`^(\s*)` + regexp.QuoteMeta(field) + `(\s*):`)
		lines[idx] = keyRe.ReplaceAllString(lines[idx], "${1}"+candidate+"${2}:")
		text = strings.Join(lines, "\n")
		noteIdx := len(notes)
		renames = append(renames, fieldRename{idx: idx, from: field, noteIdxs: []int{noteIdx}})
		notes = append(notes, canondata.T("format.fix.field-renamed",
			canondata.M{"from": field, "to": candidate, "line": fmt.Sprintf("%d", idx+1)}))
	}
	// A repair that did not repair the decode is not a repair: if after
	// all edits the decode still fails on an unknown field, the renames
	// are rolled back — the author's rejection needs their own field,
	// not an edit-distance candidate.
	if len(renames) > 0 {
		var d Delta
		if err := yamlio.DecodeStrict([]byte(text), &d); err != nil &&
			unknownFieldRepairRe.MatchString(err.Error()) {
			lines := strings.Split(text, "\n")
			drop := map[int]bool{}
			for _, r := range renames {
				keyRe := regexp.MustCompile(`^(\s*)[^\s:]+(\s*):`)
				if r.idx < len(lines) && keyRe.MatchString(lines[r.idx]) {
					lines[r.idx] = keyRe.ReplaceAllString(lines[r.idx], "${1}"+r.from+"${2}:")
				}
				for _, ni := range r.noteIdxs {
					drop[ni] = true
				}
			}
			text = strings.Join(lines, "\n")
			kept := notes[:0]
			for i, n := range notes {
				if !drop[i] {
					kept = append(kept, n)
				}
			}
			notes = kept
		}
	}
	return text, notes
}

// mapValueLineRe — a "key: value" line (with a possible list "- "):
// the prefix up to the value is separated, the value is the rest of
// the line.
var mapValueLineRe = regexp.MustCompile(`^(\s*-?\s*[\w.-]+:\s)(.+)$`)

// quoteColonScalar repairs a line with ": " inside the value of an
// unquoted scalar: the value is taken into double quotes. ok=false —
// the line did not fit the repair (already quoted/in a block, not
// found, ambiguous).
// repairableDecode — a decode error of the same repairable class:
// applying the edit is allowed (perhaps one more line of the same kind
// is needed). Any other class — the edit made it worse, roll back.
func repairableDecode(msg string) bool {
	return strings.Contains(msg, canondata.T("format.match.mapping-values")) ||
		strings.Contains(msg, "cannot unmarshal") ||
		unknownFieldRepairRe.FindStringSubmatch(msg) != nil
}

func quoteColonScalar(text, errMsg string) (string, bool) {
	m := lineErrRe.FindStringSubmatch(errMsg)
	if m == nil {
		return text, false
	}
	n, ok := atoiSafe(m[1])
	if !ok || n < 1 {
		return text, false
	}
	lines := strings.Split(text, "\n")
	if n > len(lines) {
		return text, false
	}
	mm := mapValueLineRe.FindStringSubmatch(lines[n-1])
	if mm == nil {
		return text, false
	}
	prefix, value := mm[1], strings.TrimRight(mm[2], " \t")
	if value == "" || !strings.Contains(value, ": ") {
		return text, false
	}
	if q := value[0]; q == '"' || q == '\'' || q == '{' || q == '[' {
		return text, false
	}
	lines[n-1] = prefix + doubleQuote(value)
	return strings.Join(lines, "\n"), true
}

// wrapClosedMap wraps a closed mapping {…} — the single mapping entry
// on the line — into a one-element list where the decode expects a
// list. ok=false — the value is not a closed one-line mapping.
func wrapClosedMap(text, errMsg string) (string, bool) {
	m := lineErrRe.FindStringSubmatch(errMsg)
	if m == nil {
		return text, false
	}
	n, ok := atoiSafe(m[1])
	if !ok || n < 1 {
		return text, false
	}
	lines := strings.Split(text, "\n")
	if n > len(lines) {
		return text, false
	}
	mm := mapValueLineRe.FindStringSubmatch(lines[n-1])
	if mm == nil {
		return text, false
	}
	prefix, value := mm[1], strings.TrimRight(mm[2], " \t")
	if !strings.HasPrefix(value, "{") || !strings.HasSuffix(value, "}") {
		return text, false
	}
	lines[n-1] = prefix + "[" + value + "]"
	return strings.Join(lines, "\n"), true
}

// quoteFlowComma repairs a line whose flow map carries a comma inside
// an unquoted value: the segments after the comma without "key:" are
// torn pieces of that value — they are merged back and the value goes
// into double quotes. One line, one flow map, no bracketed pieces;
// ok=false — the line does not fit the repair.
func quoteFlowComma(text, errMsg string) (string, bool) {
	m := lineErrRe.FindStringSubmatch(errMsg)
	if m == nil {
		return text, false
	}
	n, ok := atoiSafe(m[1])
	if !ok || n < 1 {
		return text, false
	}
	lines := strings.Split(text, "\n")
	if n > len(lines) {
		return text, false
	}
	line := lines[n-1]
	open := strings.Index(line, "{")
	closeIdx := strings.LastIndex(line, "}")
	if open < 0 || closeIdx <= open || strings.Count(line, "{") > 1 {
		return text, false
	}
	segs := splitFlowSegments(line[open+1 : closeIdx])
	j := -1
	for i, s := range segs {
		if !strings.Contains(s, ":") {
			j = i
			break
		}
	}
	if j < 1 {
		return text, false
	}
	kv := strings.SplitN(segs[j-1], ":", 2)
	if len(kv) != 2 {
		return text, false
	}
	key, value := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
	flowScalarUnsafe := func(v string) bool {
		if v == "" {
			return true
		}
		if q := v[0]; q == '"' || q == '\'' {
			return true
		}
		if q := v[len(v)-1]; q == '"' || q == '\'' {
			return true
		}
		return strings.ContainsAny(v, "[]{}")
	}
	if key == "" || flowScalarUnsafe(value) {
		return text, false
	}
	tail := []string{value}
	end := j - 1
	for i := j; i < len(segs); i++ {
		s := strings.TrimSpace(segs[i])
		if strings.Contains(s, ":") || flowScalarUnsafe(s) {
			break
		}
		tail = append(tail, s)
		end = i
	}
	merged := key + ": " + doubleQuote(strings.Join(tail, ", "))
	out := append(append([]string{}, segs[:j-1]...), merged)
	out = append(out, segs[end+1:]...)
	lines[n-1] = line[:open+1] + strings.Join(out, ", ") + line[closeIdx:]
	return strings.Join(lines, "\n"), true
}

func atoiSafe(s string) (int, bool) {
	n := 0
	if s == "" {
		return 0, false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
	}
	return n, true
}

// The seed channel's quoting trap, caught at intake. A seed written
// as a DOUBLE-QUOTED yaml scalar carrying \xNN escapes loses bytes
// silently: at the yaml layer such an escape names a Unicode
// codepoint (for values above 0x7F its UTF-8 encoding is two bytes),
// while the seed channel's own literal layer — where \xNN names one
// raw byte — never sees the escape. The seeded bytes then differ from
// what was written, and the difference is invisible post-decode (the
// backslash is gone). The scan below works on the RAW submission
// text, where the author's escapes still are; only genuinely
// byte-changing escapes fire — \xNN below 0x80 decodes at the yaml
// layer to the same single byte the literal layer would give, so an
// honest double-quoted ASCII escape stays legal.
//
// Two seed notations are scanned: the check table's one-line seeds
// (list items under `seeds:`, or a flow value on that line) and the
// scenario map form (`seed:` with `{path, content}` items). Block
// scalars are safe by construction (no escape decoding inside) and
// never fire.
package delta

import (
	"strings"
)

// SeedQuoteEscape scans raw submission text for the trap: a seed
// entry written as a double-quoted scalar whose escapes change
// bytes. Returns the 1-based line number and true when found.
func SeedQuoteEscape(text string) (int, bool) {
	lines := strings.Split(text, "\n")
	seedsIndent := -1 // inside a `seeds:` list block: items are deeper
	seedIndent := -1  // inside a `seed:` map block: items are deeper
	pathIndent := -1  // the map fields' own column (path:/content:)
	for i, line := range lines {
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		body := line[indent:]
		// a list item may itself carry the key (`- seeds:`, the
		// check-table row form); the key's column sits after the dash
		keyCol := indent
		keyBody := body
		if strings.HasPrefix(body, "-") {
			after := strings.TrimLeft(body[1:], " \t")
			keyCol = indent + (len(body) - 1 - len(after)) + 1
			keyBody = after
		}
		if seedsIndent >= 0 && (indent <= seedsIndent || !strings.HasPrefix(body, "-")) {
			seedsIndent = -1
		}
		if seedIndent >= 0 && indent <= seedIndent {
			seedIndent = -1
			pathIndent = -1
		}
		switch mappingKey(keyBody) {
		case "seeds":
			rest := strings.TrimLeft(keyBody[len("seeds:"):], " \t")
			if rest == "" {
				seedsIndent = indent
				continue
			}
			// flow form on the same line: escapes live in the quoted
			// values; one span scan sees them all
			if hasByteChangingEscape(rest) {
				return i + 1, true
			}
		case "seed":
			rest := strings.TrimLeft(keyBody[len("seed:"):], " \t")
			if rest == "" {
				seedIndent = indent
				pathIndent = -1
				continue
			}
			// a flow list of maps on the same line (`seed: [{…}]`)
			if strings.HasPrefix(rest, "[") && hasByteChangingEscape(rest) {
				return i + 1, true
			}
			continue
		case "path":
			if seedIndent >= 0 && pathIndent < 0 {
				pathIndent = keyCol
			}
		case "content":
			// the map's own field sits at the sibling column; deeper
			// lines are block-scalar bytes (safe, never decoded)
			if seedIndent >= 0 && (indent == pathIndent || (pathIndent < 0 && indent > seedIndent)) {
				rest := strings.TrimLeft(keyBody[len("content:"):], " \t")
				if strings.HasPrefix(rest, "\"") && hasByteChangingEscape(rest) {
					return i + 1, true
				}
			}
		}
		if seedIndent >= 0 && strings.HasPrefix(body, "-") && strings.Contains(body, "content:") &&
			hasByteChangingEscape(body) {
			// a one-line flow map item: `- {path: p, content: "…"}`
			return i + 1, true
		}
		if seedsIndent >= 0 {
			item := strings.TrimLeft(strings.TrimPrefix(body, "-"), " \t")
			if strings.HasPrefix(item, "\"") && hasByteChangingEscape(item) {
				return i + 1, true
			}
		}
	}
	return 0, false
}

// mappingKey — the `key:` head of a line inside the delta subset, ""
// when the line is not a mapping entry (a list item, a block scalar
// continuation, a comment).
func mappingKey(body string) string {
	if body == "" || body[0] == '-' || body[0] == '#' {
		return ""
	}
	i := strings.IndexByte(body, ':')
	if i <= 0 {
		return ""
	}
	key := strings.TrimRight(body[:i], " \t")
	if key == "" || strings.ContainsAny(key, " \t\"'[]{}") {
		return ""
	}
	return key
}

// hasByteChangingEscape — an escape that changes bytes lives in the
// span: a double-quoted yaml scalar (from the first quote to the
// last) carrying \xNN with NN >= 0x80 or \uXXXX/\UXXXXXXXX with a
// codepoint >= 0x80. Backslash runs keep their parity: an even run
// is literal backslashes, an odd run leaves the last one an escape
// starter.
func hasByteChangingEscape(s string) bool {
	first := strings.IndexByte(s, '"')
	if first < 0 {
		return false
	}
	last := strings.LastIndexByte(s, '"')
	if last <= first {
		return false
	}
	span := s[first+1 : last]
	for i := 0; i < len(span); {
		if span[i] != '\\' {
			i++
			continue
		}
		run := 0
		for i+run < len(span) && span[i+run] == '\\' {
			run++
		}
		if run%2 == 0 { // literal backslashes, not an escape
			i += run
			continue
		}
		j := i + run // the escaped character (run is odd: one starter)
		switch {
		case j < len(span) && span[j] == 'x':
			// \xNN names one byte; only NN >= 0x80 differs from what
			// the literal layer would give (below that the yaml layer
			// decodes the same single byte)
			if hexDigit(span, j+1) >= 8 {
				return true
			}
			i = j + 3
		case j < len(span) && (span[j] == 'u' || span[j] == 'U'):
			width := 4
			if span[j] == 'U' {
				width = 8
			}
			if codePointAt(span, j+1, width) >= 0x80 {
				return true
			}
			i = j + 1 + width
		default:
			i = j + 1
		}
	}
	return false
}

// hexDigit — the value of one hex digit at pos (0 when absent or not
// hex; the callers compare against thresholds, where the 0 of a
// non-hex char is the honest "no escape" answer).
func hexDigit(s string, pos int) int {
	if pos >= len(s) {
		return 0
	}
	switch c := s[pos]; {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return 0
}

// codePointAt — the value of width hex digits at pos (0 when any of
// them is not a hex digit — a malformed escape is not this trap).
func codePointAt(s string, pos, width int) int {
	v := 0
	for k := 0; k < width; k++ {
		if pos+k >= len(s) {
			return 0
		}
		c := s[pos+k]
		d := hexDigit(s, pos+k)
		if d == 0 && c != '0' {
			return 0
		}
		v = v<<4 | d
	}
	return v
}

// ExpectationQuoteEscape — the same trap on the expectation channels:
// out:/err: values, assertion value: fields, and the state: line's
// text tail (all decoded by the machine's own literal layer, where
// \xNN names one raw byte). A double-quoted scalar there loses bytes
// at the YAML layer exactly as a seed does. Scanned on the raw text
// of the kinds that carry these channels (checks, spec, assert) —
// `out:` of a conventions recipe is a path declaration, not a byte
// channel, and must not fire.
func ExpectationQuoteEscape(text string) (int, bool) {
	if !expectationBearingKind(text) {
		return 0, false
	}
	lines := strings.Split(text, "\n")
	stateIndent := -1 // inside a `state:` list block: items are deeper
	for i, line := range lines {
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		body := line[indent:]
		if stateIndent >= 0 && (indent <= stateIndent || !strings.HasPrefix(body, "-")) {
			stateIndent = -1
		}
		keyBody := body
		if strings.HasPrefix(body, "-") {
			after := strings.TrimLeft(body[1:], " \t")
			keyBody = after
		}
		key := mappingKey(keyBody)
		switch key {
		case "out", "err", "value":
			rest := strings.TrimLeft(keyBody[len(key)+1:], " \t")
			if strings.HasPrefix(rest, "\"") && hasByteChangingEscape(rest) {
				return i + 1, true
			}
		case "state":
			rest := strings.TrimLeft(keyBody[len("state:"):], " \t")
			if rest == "" {
				stateIndent = indent
				continue
			}
			// a flow list or an inline value on the same line
			if hasByteChangingEscape(rest) {
				return i + 1, true
			}
		}
		if stateIndent >= 0 {
			item := strings.TrimLeft(strings.TrimPrefix(body, "-"), " \t")
			if strings.HasPrefix(item, "\"") && hasByteChangingEscape(item) {
				return i + 1, true
			}
		}
	}
	return 0, false
}

// expectationBearingKind — the delta kind line names a kind whose
// raw text can carry the expectation channels; conventions is
// excluded on purpose (its out: is a path, not bytes).
func expectationBearingKind(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		if k, ok := strings.CutPrefix(t, "kind:"); ok {
			switch strings.TrimSpace(k) {
			case KindChecks, KindSpec, KindAssert:
				return true
			}
			return false
		}
		return false
	}
	return false
}

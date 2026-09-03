package yamlio

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// EscapeControls — control bytes and bytes that do not decode as
// valid UTF-8 become \xNN escapes (newline, tab and carriage return
// keep their mnemonics); every printable rune, the backslash
// included, passes through as is. The output is valid text by
// construction over any byte input, without losing the verbatim
// quality of ordinary values — the property every byte-strict text
// surface of the machine renders through (spec renders,
// measured-bytes reasons, run evidence).
func EscapeControls(s string) string { return escapeText(s, false) }

// EscapeHuman — the same discipline for living-words surfaces:
// printable unicode passes as is, control and undecodable bytes
// escape, and the backslash itself is escaped so no literal
// backslash can pose as an escape sequence.
func EscapeHuman(s string) string { return escapeText(s, true) }

func escapeText(s string, backslash bool) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			// an undecodable byte: escape the byte itself, never
			// pass it through (a multi-byte sequence that decodes
			// to the replacement rune legitimately is size > 1 and
			// falls through to the default branch untouched)
			fmt.Fprintf(&sb, "\\x%02X", s[i])
			i++
			continue
		}
		switch {
		case r == '\n':
			sb.WriteString("\\n")
		case r == '\t':
			sb.WriteString("\\t")
		case r == '\r':
			sb.WriteString("\\r")
		case r < 0x20 || r == 0x7F:
			fmt.Fprintf(&sb, "\\x%02X", r)
		case backslash && r == '\\':
			sb.WriteString("\\\\")
		default:
			sb.WriteString(s[i : i+size])
		}
		i += size
	}
	return sb.String()
}

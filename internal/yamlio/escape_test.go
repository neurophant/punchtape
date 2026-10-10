// Escape units: every byte-strict text surface of the machine
// renders through EscapeControls; living-words surfaces through
// EscapeHuman, where the backslash doubles so no literal backslash
// can pose as an escape sequence.
package yamlio

import (
	"testing"
)

func TestEscapeControls(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"mnemonics, control bytes, printable passthrough",
			"ab\n\t\rc\x01d\x7F\\ю", "ab\\n\\t\\rc\\x01d\\x7F\\ю"},
		{"an undecodable byte escapes", "a\xffb", "a\\xFFb"},
		{"a legitimately encoded replacement rune passes",
			"a\uFFFDb", "a\uFFFDb"},
		{"printable unicode passes", "snowman \u2603", "snowman \u2603"},
		{"empty input", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := EscapeControls(tc.in); got != tc.want {
				t.Fatalf("EscapeControls(%q):\n got %q\nwant %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestEscapeHuman(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"the backslash must double", `a\b`, `a\\b`},
		{"tab keeps its mnemonic", "x\ty", `x\ty`},
		{"newline keeps its mnemonic", "p\nq", `p\nq`},
		{"a control byte escapes", "q\x01", `q\x01`},
		{"an undecodable byte escapes", "q\xff", `q\xFF`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := EscapeHuman(tc.in); got != tc.want {
				t.Fatalf("EscapeHuman(%q):\n got %q\nwant %q", tc.in, got, tc.want)
			}
		})
	}
}

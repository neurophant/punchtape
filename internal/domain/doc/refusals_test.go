// The format's own refusal tails: every forbidden node kind of the
// deep armor answers its refusal, a broken indent in free text is
// refused WITH its source position, and the template's multi-line
// notes form decodes — the guess that cost both field hands one
// refusal each has nothing left to guess.
package doc

import (
	"strings"
	"testing"
)

func TestArmorRefusesForbiddenNodes(t *testing.T) {
	for _, tc := range []struct{ name, in string }{
		{"an alias", "a: *x\n"},
		{"a tag", "a: !!str 5\n"},
		{"a merge key", "<<: {b: 1}\n"},
		{"an anchored sequence item", "rows:\n    - &x 1\n"},
	} {
		if _, err := Decode([]byte(tc.in)); err == nil {
			t.Errorf("%s must be refused", tc.name)
		}
	}
}

func TestRefusalCarriesSourcePosition(t *testing.T) {
	_, err := Decode([]byte("behaviors:\n  - id: b1\n    name: n\n    notes: >\n   folded wrong indent\n"))
	if err == nil {
		t.Fatal("the broken document must be refused")
	}
	if !strings.Contains(err.Error(), "[5:") {
		t.Fatalf("the refusal must name the line: %v", err)
	}
}

func TestMultilineNotesDecode(t *testing.T) {
	d, err := Decode([]byte("notes:\n  - kind: work\n    text: a free-text note; a longer note continues\n      on the next line, indented deeper than its key\n"))
	if err != nil {
		t.Fatalf("the template's notes form must decode: %v", err)
	}
	if len(d.Notes) != 1 || !strings.Contains(d.Notes[0].Text, "on the next line") {
		t.Fatalf("the folded note must keep its text: %#v", d.Notes)
	}
}

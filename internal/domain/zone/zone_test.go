// The zone layout's battery: the path law's whole refusal family and
// the zone path spelling.
package zone

import "testing"

func TestValidateRel(t *testing.T) {
	for _, p := range []string{"a", "a/b.txt", "deep/dir/file.tar.gz", "x"} {
		if err := ValidateRel(p); err != nil {
			t.Fatalf("%q must pass: %v", p, err)
		}
	}
	for _, p := range []string{"", "/abs", "../out", "a/../b", "a//b", "./a", "a/.", "a\\b", "a\x00b"} {
		if err := ValidateRel(p); err == nil {
			t.Fatalf("%q must be refused", p)
		}
	}
}

func TestPathSpelling(t *testing.T) {
	if got := Path(StateFile); got != ".punchtape/state.md" {
		t.Fatalf("path spelling: %q", got)
	}
	if got := Path(JournalDir, "epoch-00001.yaml"); got != ".punchtape/journal/epoch-00001.yaml" {
		t.Fatalf("joined spelling: %q", got)
	}
}

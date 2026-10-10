// Bytes units: the digest family (cross-OS stable over line breaks)
// and the lenient reread of the machine's own caches.
package yamlio

import (
	"testing"
)

func TestDigestFamily(t *testing.T) {
	const sha256A = "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb"
	if got := Digest([]byte("a")); got != sha256A {
		t.Fatalf("Digest: %s", got)
	}
	if got := DigestShort([]byte("a")); got != sha256A[:16] {
		t.Fatalf("DigestShort: %s", got)
	}

	data := []byte("line1\nline2")
	if got := NormalizeEOL(data); &got[0] != &data[0] {
		t.Fatal("input without CR must pass through untouched")
	}
	if got := string(NormalizeEOL([]byte("a\r\nb\r\nc"))); got != "a\nb\nc" {
		t.Fatalf("CRLF: %q", got)
	}
	if got := string(NormalizeEOL([]byte("a\rb"))); got != "a\rb" {
		t.Fatalf("a lone CR must survive: %q", got)
	}
	if DigestNorm([]byte("x\r\ny")) != Digest([]byte("x\ny")) {
		t.Fatal("DigestNorm must hash normalized bytes")
	}
}

// Unmarshal is the reread of machine-written caches, not intake:
// unknown fields are tolerated, the first document wins, and only
// broken YAML refuses.
func TestUnmarshalLenient(t *testing.T) {
	var v struct {
		A int `yaml:"a"`
	}
	wantNoErr(t, "reread ok", Unmarshal([]byte("a: 1\nunknown: 2\n"), &v))
	if v.A != 1 {
		t.Fatalf("a decoded as %d", v.A)
	}
	wantErr(t, "malformed stream", Unmarshal([]byte("a: [1\n"), &v))
	wantNoErr(t, "trailing document tolerated", Unmarshal([]byte("a: 1\n---\nb: 2\n"), &v))
	if v.A != 1 {
		t.Fatalf("a drifted after the trailing document: %d", v.A)
	}
}

// Bytes units: the digest family.
package domain

import (
	"testing"
)

func TestDigest(t *testing.T) {
	const sha256A = "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb"
	if got := Digest([]byte("a")); got != sha256A {
		t.Fatalf("Digest: %s", got)
	}
	if len(Digest(nil)) != 64 {
		t.Fatal("the digest is 64 hex characters")
	}
}

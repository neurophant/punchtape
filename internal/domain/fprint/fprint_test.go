// The fingerprint laws' battery: glob matching, the scope
// fingerprint's order independence, the proof's move on every
// dependency's move.
package fprint

import "testing"

func TestMatchGlob(t *testing.T) {
	yes := [][2]string{
		{"src/**", "src/a"},
		{"src/**", "src/sub/deep/b"},
		{"**", "any/where/x"},
		{"**", "top"},
		{"out/*.txt", "out/built.txt"},
		{"a/*/b", "a/x/b"},
		{"*", "file"},
	}
	for _, c := range yes {
		if !MatchGlob(c[0], c[1]) {
			t.Fatalf("%q must match %q", c[0], c[1])
		}
	}
	no := [][2]string{
		{"src/**", "out/a"},
		{"out/*.txt", "out/sub/built.txt"},
		{"a/*/b", "a/x/y/b"},
		{"*.txt", "file.dat"},
		{"src", "src/a"},
	}
	for _, c := range no {
		if MatchGlob(c[0], c[1]) {
			t.Fatalf("%q must not match %q", c[0], c[1])
		}
	}
}

func TestScopeHashOrderIndependent(t *testing.T) {
	a := []FileDigest{{Path: "a", Digest: "1"}, {Path: "b", Digest: "2"}}
	b := []FileDigest{{Path: "b", Digest: "2"}, {Path: "a", Digest: "1"}}
	if ScopeHash(a) != ScopeHash(b) {
		t.Fatalf("scope hash must not see the input order")
	}
	if ScopeHash(a) == ScopeHash([]FileDigest{{Path: "a", Digest: "2"}, {Path: "b", Digest: "1"}}) {
		t.Fatalf("different content must move the hash")
	}
	if ScopeHash(a) == ScopeHash([]FileDigest{{Path: "a", Digest: "1"}, {Path: "b", Digest: "2"}, {Path: "c", Digest: "3"}}) {
		t.Fatalf("a new file must move the hash")
	}
	if ScopeHash([]FileDigest{{Path: "a", Digest: AbsentDigest}}) == ScopeHash([]FileDigest{{Path: "a", Digest: "1"}}) {
		t.Fatalf("absence is a different fact")
	}
}

func TestProofMovesWithItsDependencies(t *testing.T) {
	p := Proof([]byte("run: x"), "scope", "env")
	if p != Proof([]byte("run: x"), "scope", "env") {
		t.Fatalf("proof must be stable")
	}
	if p == Proof([]byte("run: y"), "scope", "env") {
		t.Fatalf("row bytes must move the proof")
	}
	if p == Proof([]byte("run: x"), "scope2", "env") {
		t.Fatalf("the scope hash must move the proof")
	}
	if p == Proof([]byte("run: x"), "scope", "env2") {
		t.Fatalf("the environment must move the proof")
	}
}

func TestMatchGlobEdges(t *testing.T) {
	if !MatchGlob("src/**", "src") {
		t.Fatal("the double star swallows zero segments")
	}
	if MatchGlob("src/**/b", "src/a/x/b") == false {
		t.Fatal("the double star swallows many")
	}
	if MatchGlob("", "x") || MatchGlob("a", "") {
		t.Fatal("empty sides never match")
	}
}

func TestMatchGlobTails(t *testing.T) {
	if MatchGlob("**/nope.txt", "a.txt") {
		t.Fatal("a non-matching double star must refuse")
	}
	if MatchGlob("a/b", "a") {
		t.Fatal("a pattern longer than the path must refuse")
	}
}

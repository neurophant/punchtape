package cas

import (
	"math/rand"
	"os"
	"testing"
)

// The property: a thousand random payloads of random sizes (0 B to
// 64 KB) — every address unique, the shards carry a flat share, and
// re-putting known bytes rewrites nothing.
func TestPutProperty(t *testing.T) {
	s := New(t.TempDir())
	r := rand.New(rand.NewSource(7))
	addrs := map[string]bool{}
	seen := map[string]string{} // payload bytes -> its address
	shards := map[string]int{}
	var kept [][]byte
	for i := 0; i < 1000; i++ {
		size := []int{0, 1, 64, 1024}[r.Intn(4)]
		if r.Intn(10) == 0 {
			size = 65536
		}
		data := make([]byte, size)
		r.Read(data)
		if len(kept) < 100 {
			kept = append(kept, data)
		}
		addr, err := s.Put(data)
		if err != nil {
			t.Fatalf("put %d: %v", i, err)
		}
		if prev, dup := seen[string(data)]; dup {
			if prev != addr {
				t.Fatalf("the same bytes answered two addresses at %d", i)
			}
			continue
		}
		if addrs[addr] {
			t.Fatalf("distinct bytes share an address at %d", i)
		}
		seen[string(data)] = addr
		addrs[addr] = true
		shards[addr[:2]]++
	}
	// flatness, softly: 1000 objects over 256 shards expect about
	// four each; a skewed digest would concentrate or clump.
	max, nonEmpty := 0, 0
	for _, n := range shards {
		if n > max {
			max = n
		}
		nonEmpty++
	}
	if max > 12 {
		t.Fatalf("one shard holds %d objects: the digest is skewed", max)
	}
	if nonEmpty < 234 {
		t.Fatalf("only %d of 256 shards used: the digest is skewed", nonEmpty)
	}
	// the repeat is silent: the kept bytes re-put leave the objects
	// exactly as they were
	mtimes := make([]os.FileInfo, len(kept))
	for i, data := range kept {
		addr, err := s.Put(data)
		if err != nil {
			t.Fatal(err)
		}
		st, err := os.Stat(s.path(addr))
		if err != nil {
			t.Fatal(err)
		}
		mtimes[i] = st
	}
	for i, data := range kept {
		addr, err := s.Put(data)
		if err != nil {
			t.Fatal(err)
		}
		st, err := os.Stat(s.path(addr))
		if err != nil {
			t.Fatal(err)
		}
		if !mtimes[i].ModTime().Equal(st.ModTime()) {
			t.Fatalf("repeat put rewrote %s", addr)
		}
	}
}

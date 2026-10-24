// The record codec's and the fold's battery: byte-exact round trips,
// the shaped-record law, the fold's merge semantics.
package journal

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/domain/doc"
)

func TestRecordRoundTripExact(t *testing.T) {
	rec := Record{Kind: KindBirth, Birth: &Birth{Addr: strings.Repeat("a", 64), Digest: strings.Repeat("b", 64)}}
	data := MarshalRecord(rec)
	var back Record
	if err := DecodeRecord(data, &back); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !bytes.Equal(MarshalRecord(back), data) {
		t.Fatalf("round trip must be byte-exact")
	}
	// A hand edit breaks the round trip: not a fact the machine wrote.
	tampered := bytes.Replace(data, []byte("birth"), []byte("burth"), 1)
	if err := DecodeRecord(tampered, new(Record)); err == nil {
		t.Fatalf("a tampered record must be refused")
	}
	if err := DecodeRecord(append(data, 'x'), new(Record)); err == nil {
		t.Fatalf("a record without the closing newline must be refused")
	}
}

func TestShapedRecords(t *testing.T) {
	bad := []Record{
		{Kind: "bogus"},
		{Kind: KindBirth},
		{Kind: KindRun, Run: &Run{N: 1}, Birth: &Birth{Addr: "a", Digest: "d"}},
	}
	for i, r := range bad {
		if err := DecodeRecord(MarshalRecord(r), new(Record)); err == nil {
			t.Fatalf("record %d must be refused as unshaped", i)
		}
	}
}

// objReader — the fold's object reader over a map.
func objReader(store map[string][]byte) func(string) ([]byte, error) {
	return func(addr string) ([]byte, error) {
		data, ok := store[addr]
		if !ok {
			return nil, fmt.Errorf("missing object %s", addr)
		}
		return data, nil
	}
}

func strp(s string) *string { return &s }

func TestFold(t *testing.T) {
	store := map[string][]byte{}
	put := func(b doc.Behavior) BehaviorOp {
		data := doc.MarshalBehavior(b)
		addr := fmt.Sprintf("%064d", len(store))
		store[addr] = data
		return BehaviorOp{ID: b.ID, Addr: addr}
	}
	one := put(doc.Behavior{ID: "a/one", Name: "One", Rows: []doc.Row{{Run: doc.Cmd{Line: "x"}}}})
	two := put(doc.Behavior{ID: "b/two", Name: "Two"})
	records := []Record{
		{Kind: KindBirth, Birth: &Birth{Addr: "w", Digest: "wd"}},
		{Kind: KindSubmission, Submission: &Submission{CorpusVersion: 1, Build: strp("make"), DocAddr: "d1", Added: []BehaviorOp{one, two}}},
		{Kind: KindRun, Run: &Run{N: 1, Rows: []RowFact{{Key: "a/one/0", Status: StatusGreen, Proof: "p1"}}}},
		{Kind: KindSubmission, Submission: &Submission{CorpusVersion: 2, DocAddr: "d2", Removed: []string{"a/one"}}},
	}
	st, err := Fold(records, objReader(store))
	if err != nil {
		t.Fatalf("fold: %v", err)
	}
	if !st.Born || st.GoalDigest != "wd" || st.CorpusVersion != 2 || st.Build != "make" {
		t.Fatalf("standing: %+v", st)
	}
	if _, alive := st.Behaviors["a/one"]; alive {
		t.Fatalf("removed behavior must be gone")
	}
	if _, alive := st.Behaviors["b/two"]; !alive {
		t.Fatalf("untouched behavior must stand")
	}
	if len(st.Order) != 1 || st.Order[0] != "b/two" {
		t.Fatalf("order keeps the living only: %v", st.Order)
	}
	if st.RunNumber != 1 || st.Rows["a/one/0"].Proof != "p1" {
		t.Fatalf("run facts: %+v", st.Rows)
	}
}

func TestFoldNamesTheMissingObject(t *testing.T) {
	rec := Record{Kind: KindSubmission, Submission: &Submission{
		CorpusVersion: 1, DocAddr: "d",
		Added: []BehaviorOp{{ID: "a/b", Addr: strings.Repeat("f", 64)}},
	}}
	if _, err := Fold([]Record{rec}, objReader(map[string][]byte{})); err == nil {
		t.Fatal("a missing object is an honest error")
	}
}

func TestDecodeMultiLineRefusalIsOneLine(t *testing.T) {
	err := DecodeRecord([]byte("kind: birth\nbirth:\n    addr: [\n"), new(Record))
	if err == nil || strings.Contains(err.Error(), "\n") {
		t.Fatalf("one line per refusal: %v", err)
	}
}

func TestOneLineHelper(t *testing.T) {
	if got := oneLine(fmt.Errorf("l1\nl2")); got == nil || strings.Contains(got.Error(), "\n") {
		t.Fatalf("first line only: %v", got)
	}
}

func TestDecodeRecordRefusesForeignSpelling(t *testing.T) {
	// Four-space indent decodes strict and shaped — but it is not the
	// machine's own spelling: the round trip catches the forgery.
	padded := []byte("kind: birth\nbirth:\n    addr: ab\n    digest: cd\n")
	if err := DecodeRecord(padded, new(Record)); err == nil {
		t.Fatal("a record that does not round-trip byte-exact must be refused")
	}
}

func TestFoldRefusesUndecodableBehaviorObject(t *testing.T) {
	addr := strings.Repeat("f", 64)
	rec := Record{Kind: KindSubmission, Submission: &Submission{
		CorpusVersion: 1, DocAddr: "d",
		Added: []BehaviorOp{{ID: "a/b", Addr: addr}},
	}}
	lying := objReader(map[string][]byte{addr: []byte("not a behavior: [\n")})
	if _, err := Fold([]Record{rec}, lying); err == nil {
		t.Fatal("an object that does not decode as a behavior must fail the fold")
	}
}

func TestBirthCarriesTheVersion(t *testing.T) {
	rec := Record{Kind: KindBirth, Birth: &Birth{
		Addr: strings.Repeat("a", 64), Digest: strings.Repeat("b", 64), Version: "9.9.9",
	}}
	var back Record
	if err := DecodeRecord(MarshalRecord(rec), &back); err != nil || back.Birth.Version != "9.9.9" {
		t.Fatalf("the birth's machine version must round-trip: %v %+v", err, back.Birth)
	}
	st, err := Fold([]Record{rec}, objReader(map[string][]byte{}))
	if err != nil || st.BirthVersion != "9.9.9" {
		t.Fatalf("the fold must expose the birth's machine version: %v %+v", err, st)
	}
	// A birth from before the field existed folds with the honest
	// empty origin.
	old := Record{Kind: KindBirth, Birth: &Birth{Addr: "x", Digest: "y"}}
	var backOld Record
	if err := DecodeRecord(MarshalRecord(old), &backOld); err != nil {
		t.Fatalf("the old birth shape must still round-trip: %v", err)
	}
	st, err = Fold([]Record{old}, objReader(map[string][]byte{}))
	if err != nil || st.BirthVersion != "" {
		t.Fatalf("an old birth carries no machine version: %v %+v", err, st)
	}
}

func TestRefusalRecordRoundTripAndFold(t *testing.T) {
	rec := Record{Kind: KindRefusal, Refusals: []doc.Refusal{
		{Where: "behaviors[0].id", Law: "law:document", Fact: "the id is empty"},
	}}
	var back Record
	if err := DecodeRecord(MarshalRecord(rec), &back); err != nil || len(back.Refusals) != 1 {
		t.Fatalf("the refusal fact must round-trip: %v %+v", err, back.Refusals)
	}
	st, err := Fold([]Record{rec, rec}, objReader(map[string][]byte{}))
	if err != nil || st.Refused != 2 {
		t.Fatalf("the fold must count every refused submission: %v %d", err, st.Refused)
	}
	// A refusal riding another payload is not a shape the machine
	// writes.
	mixed := Record{Kind: KindRefusal, Refusals: rec.Refusals, Run: &Run{N: 1}}
	if err := DecodeRecord(MarshalRecord(mixed), new(Record)); err == nil {
		t.Fatal("a refusal record with a second payload must be refused")
	}
}

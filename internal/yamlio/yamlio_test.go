// yamlio units: the five known format traps pinned (the armor
// must bite where the library stays silent), the strict-decode
// refusals, the dictionary form and the deterministic marshal.
package yamlio

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/neurophant/punchtape/internal/fsport"
)

type doc struct {
	Name  string   `yaml:"name"`
	Items []string `yaml:"items"`
	Num   int      `yaml:"num"`
}

func wantErr(t *testing.T, label string, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: the armor stayed silent", label)
	}
}

func wantNoErr(t *testing.T, label string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: refused (%v)", label, err)
	}
}

// Format trap 1 — Norway: the YAML 1.1 boolean aliases are strings
// ONLY in quotes; a bare one is a mistake to catch, because the
// dialect carries no booleans.
func TestTrapNorwayWords(t *testing.T) {
	for _, w := range []string{"no", "on", "yes", "off", "No", "TRUE", "False"} {
		var d doc
		wantErr(t, "bare "+w, DecodeStrict([]byte("name: "+w+"\nitems: []\nnum: 1\n"), &d))
	}
	var d doc
	wantNoErr(t, "quoted no", DecodeStrict([]byte("name: \"no\"\nitems: []\nnum: 1\n"), &d))
	if d.Name != "no" {
		t.Fatalf("quoted no decoded as %q", d.Name)
	}
}

// Format trap 2 — null versus the empty list: identity strings broke
// when one side wrote null and the other []. The decode side has ONE
// empty meaning (both spellings decode to a len-0 list); the marshal
// side keeps nil and empty distinct — nil is "not filled" (the null
// form) and [] is "explicitly emptied" (the list form): the
// partial-update semantics rides the difference.
func TestTrapNullVersusEmptyList(t *testing.T) {
	var viaNull, viaList doc
	wantNoErr(t, "null form", DecodeStrict([]byte("name: x\nitems:\nnum: 1\n"), &viaNull))
	wantNoErr(t, "list form", DecodeStrict([]byte("name: x\nitems: []\nnum: 1\n"), &viaList))
	if len(viaNull.Items) != 0 || len(viaList.Items) != 0 || viaNull.Name != viaList.Name || viaNull.Num != viaList.Num {
		t.Fatalf("null and [] differ: %+v vs %+v", viaNull, viaList)
	}
	out, err := Marshal(viaNull) // viaNull.Items is nil: the null form
	wantNoErr(t, "marshal nil", err)
	if !strings.Contains(string(out), "items: null") {
		t.Fatalf("a nil list marshals as the null form: %q", out)
	}
	filled := doc{Name: "x", Items: []string{}, Num: 1}
	out, err = Marshal(filled)
	wantNoErr(t, "marshal empty", err)
	if !strings.Contains(string(out), "items: []") {
		t.Fatalf("an empty list marshals as the list form: %q", out)
	}
	var back doc
	wantNoErr(t, "decode back", DecodeStrict(out, &back))
	if back.Items == nil || len(back.Items) != 0 {
		t.Fatalf("the empty list must decode to the empty non-nil list: %+v", back.Items)
	}
}

// Format trap 3 — anchors and aliases: refused even when the types
// line up and the library would resolve them silently.
func TestTrapAnchorsAndAliases(t *testing.T) {
	var d doc
	wantErr(t, "anchor+alias", DecodeStrict([]byte("name: &a x\nitems: *a\nnum: 1\n"), &d))
	wantErr(t, "anchor only", DecodeStrict([]byte("name: &a x\nitems: []\nnum: 1\n"), &d))
}

// Format trap 4 — duplicate keys: one key, one value; the library
// reports it with both positions, the armor passes it through.
func TestTrapDuplicateKeys(t *testing.T) {
	var d doc
	err := DecodeStrict([]byte("name: x\nname: y\nitems: []\nnum: 1\n"), &d)
	wantErr(t, "duplicate key", err)
	if !strings.Contains(err.Error(), "already defined") {
		t.Fatalf("not the duplicate refusal: %v", err)
	}
}

// Format trap 5 — trailing documents: the library drops every
// document after the first without a word; the armor refuses.
func TestTrapTrailingDocument(t *testing.T) {
	var d doc
	wantErr(t, "trailing doc", DecodeStrict([]byte("name: x\nitems: []\nnum: 1\n---\nname: y\n"), &d))
	var one doc
	wantNoErr(t, "single doc with lead separator", DecodeStrict([]byte("---\nname: x\nitems: []\nnum: 1\n"), &one))
}

// The strict refusals the library itself carries: unknown fields,
// tabs in indentation.
func TestStrictRefusals(t *testing.T) {
	var d doc
	err := DecodeStrict([]byte("name: x\nitems: []\nnum: 1\nextra: boom\n"), &d)
	wantErr(t, "unknown field", err)
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("not the unknown-field refusal: %v", err)
	}
	wantErr(t, "indent tab", DecodeStrict([]byte("name: x\n\titems: []\nnum: 1\n"), &d))
	wantErr(t, "tag", DecodeStrict([]byte("name: !!str x\nitems: []\nnum: 1\n"), &d))
}

// The dictionary form: canon keys ride their alphabet, an empty
// segment is a named error, block scalars keep their lines.
func TestDecodeDict(t *testing.T) {
	d, err := DecodeDict([]byte("error.no-instance: |\n  ERROR: no instance.\nfail: \"FAIL: {err}.\"\n"))
	wantNoErr(t, "dict", err)
	if d["error.no-instance"] != "ERROR: no instance." || d["fail"] != "FAIL: {err}." {
		t.Fatalf("dict decoded wrong: %#v", d)
	}
	_, err = DecodeDict([]byte("Bad_Key: x\n"))
	wantErr(t, "key alphabet", err)
	_, err = DecodeDict([]byte("empty: \"\"\n"))
	wantErr(t, "empty segment", err)
	_, err = DecodeDict([]byte("a: 1\na: 2\n"))
	wantErr(t, "dict duplicate", err)
}

// Determinism: the same data — the same bytes, map keys sorted.
func TestMarshalDeterministic(t *testing.T) {
	m1 := map[string]string{"zeta": "1", "alpha": "2", "mid": "3"}
	m2 := map[string]string{"mid": "3", "zeta": "1", "alpha": "2"}
	b1, err := Marshal(m1)
	wantNoErr(t, "marshal 1", err)
	b2, err := Marshal(m2)
	wantNoErr(t, "marshal 2", err)
	if string(b1) != string(b2) {
		t.Fatalf("marshal not deterministic:\n%q\nvs\n%q", b1, b2)
	}
	if !strings.Contains(string(b1), "alpha") || strings.Index(string(b1), "alpha") > strings.Index(string(b1), "zeta") {
		t.Fatalf("map keys not sorted: %q", b1)
	}
}

// Refusal plumbing: the branches behind the armor — a type-incompatible
// dict value, an unmarshalable value, the walk's inner loops (a trap
// word as a key, inside a list), an alias without an anchor, a
// single-pair document body, the nil oneLine and the write delegate.
func TestDecodeDictTypeRefusal(t *testing.T) {
	_, err := DecodeDict([]byte("a: [1, 2]\n"))
	wantErr(t, "list into string dict", err)
}

func TestMarshalUnmarshalable(t *testing.T) {
	_, err := Marshal(make(chan int))
	wantErr(t, "chan", err)
}

func TestWalkInnerRefusals(t *testing.T) {
	var d doc
	wantErr(t, "trap word as key", DecodeStrict([]byte("name: x\nitems: []\nnum: 1\nno: y\n"), &d))
	wantErr(t, "trap word in list", DecodeStrict([]byte("name: x\nitems:\n  - no\nnum: 1\n"), &d))
	wantErr(t, "alias without anchor", DecodeStrict([]byte("name: *ghost\nitems: []\nnum: 1\n"), &d))
	var nested map[string]any
	wantErr(t, "nested single-pair value", DecodeStrict([]byte("outer:\n  inner: no\n"), &nested))
	one, err := DecodeDict([]byte("solo: one\n"))
	wantNoErr(t, "single-pair doc", err)
	if one["solo"] != "one" {
		t.Fatalf("single-pair dict wrong: %#v", one)
	}
}

func TestOneLineNil(t *testing.T) {
	if oneLine(nil) != nil {
		t.Fatal("oneLine(nil) must stay nil")
	}
	if got := oneLine(errFake{}); got == nil || strings.Contains(got.Error(), "\n") {
		t.Fatalf("oneLine not flat: %v", got)
	}
}

type errFake struct{}

func (errFake) Error() string { return "line one\nline two" }

// writeFake — the minimal in-memory fsport.Port stand-in for the
// write delegation test (mock code lives in this test file only).
type writeFake struct {
	stored map[string][]byte
}

func (f *writeFake) Stat(string) (fs.FileInfo, error)      { return nil, fs.ErrNotExist }
func (f *writeFake) MkdirAll(string) error                 { return nil }
func (f *writeFake) ReadFile(string) ([]byte, error)       { return nil, fs.ErrNotExist }
func (f *writeFake) ReadDir(string) ([]fs.DirEntry, error) { return nil, fs.ErrNotExist }
func (f *writeFake) WriteAtomic(path string, data []byte) error {
	f.stored[path] = data
	return nil
}
func (f *writeFake) AppendFile(path string, data []byte) error   { return nil }
func (f *writeFake) MkdirTemp(string, string) (string, error)    { return "/tmp/mem", nil }
func (f *writeFake) WriteFile(string, []byte, fs.FileMode) error { return nil }
func (f *writeFake) Chmod(string, fs.FileMode) error             { return nil }
func (f *writeFake) RemoveAll(string) error                      { return nil }
func (f *writeFake) Walk(string, filepath.WalkFunc) error        { return nil }
func (f *writeFake) LockExclusive(string) (func(), error)        { return func() {}, nil }
func (f *writeFake) Remove(string) error                         { return nil }
func (f *writeFake) Lstat(string) (fs.FileInfo, error)           { return nil, fs.ErrNotExist }
func (f *writeFake) Readlink(string) (string, error)             { return "", fs.ErrNotExist }
func (f *writeFake) Symlink(target, link string) error           { return fs.ErrNotExist }
func (f *writeFake) TryLockExclusive(string) (func(), bool, error) {
	return func() {}, true, nil
}

func TestWriteAtomicDelegates(t *testing.T) {
	f := &writeFake{stored: map[string][]byte{}}
	old := fsport.Bind
	fsport.Bind = f
	t.Cleanup(func() { fsport.Bind = old })
	if err := WriteAtomic("/mem/x.yaml", []byte("a: 1\n")); err != nil {
		t.Fatalf("write refused: %v", err)
	}
	if string(f.stored["/mem/x.yaml"]) != "a: 1\n" {
		t.Fatalf("write not delegated: %q", f.stored["/mem/x.yaml"])
	}
}

// Roundtrip: decode(marshal(v)) == v.
func TestRoundtrip(t *testing.T) {
	orig := doc{Name: "quoted \"x\"", Items: []string{"a", "b"}, Num: 7}
	b, err := Marshal(orig)
	wantNoErr(t, "marshal struct", err)
	var back doc
	wantNoErr(t, "decode back", DecodeStrict(b, &back))
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("roundtrip drift: %+v vs %+v", orig, back)
	}
}

// manifestDoc — the document-dialect shape: yaml tags and the
// grammar's real booleans.
type manifestDoc struct {
	Kind     string      `yaml:"kind"`
	Accepted bool        `yaml:"accept-draft,omitempty"`
	Out      *string     `yaml:"out,omitempty"`
	Seed     manifestSed `yaml:"seed"`
}

type manifestSed struct {
	Path string `yaml:"path"`
}

func (m *manifestSed) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain manifestSed
	var p plain
	if err := unmarshal(&p); err != nil {
		return err
	}
	*m = manifestSed(p)
	if m.Path == "" {
		return errSeedForm
	}
	return nil
}

var errSeedForm = errors.New("seed wants exactly one source")

// The document dialect: booleans are real, plain trap-word strings
// decode as what they say, the structural armor still bites, and
// the custom form rules keep refusing through the native interface.
func TestDecodeDoc(t *testing.T) {
	var d manifestDoc
	wantNoErr(t, "doc dialect booleans", DecodeDoc([]byte(
		"kind: checks\naccept-draft: true\nout: \"\"\nseed:\n  path: files/a.txt\n"), &d))
	if !d.Accepted || d.Out == nil || *d.Out != "" || d.Seed.Path != "files/a.txt" {
		t.Fatalf("doc decode wrong: %+v", d)
	}
	var bare manifestDoc
	wantNoErr(t, "plain no stays a string", DecodeDoc([]byte("kind: no\nseed:\n  path: x\n"), &bare))
	if bare.Kind != "no" {
		t.Fatalf("plain no decoded as %q", bare.Kind)
	}
	var ref manifestDoc
	wantNoErr(t, "absent pointer stays nil", DecodeDoc([]byte("kind: checks\nseed:\n  path: x\n"), &ref))
	if ref.Out != nil {
		t.Fatalf("absent out decoded as %+v", *ref.Out)
	}
	wantErr(t, "unknown field", DecodeDoc([]byte("kind: x\nbogus: 1\nseed:\n  path: x\n"), new(manifestDoc)))
	wantErr(t, "anchor in doc dialect", DecodeDoc([]byte("kind: &a x\nseed:\n  path: x\n"), new(manifestDoc)))
	wantErr(t, "two documents", DecodeDoc([]byte("kind: x\nseed:\n  path: x\n---\nkind: y\n"), new(manifestDoc)))
	wantErr(t, "custom form refusal", DecodeDoc([]byte("kind: x\nseed:\n  path: \"\"\n"), new(manifestDoc)))
	var typed manifestDoc
	wantErr(t, "string into bool", DecodeDoc([]byte("kind: x\naccept-draft: sometimes\nseed:\n  path: x\n"), &typed))
}

// KindOf — the live-form probe: the kind scalar by AST, every
// broken or kindless shape is not a form.
func TestKindOf(t *testing.T) {
	k, ok := KindOf([]byte("kind: checks\nrows: []\n"))
	if !ok || k != "checks" {
		t.Fatalf("kind probe wrong: %q %v", k, ok)
	}
	if _, ok := KindOf([]byte("rows: []\n")); ok {
		t.Fatal("kindless doc is not a form")
	}
	if _, ok := KindOf([]byte("kind: 12\n")); ok {
		t.Fatal("non-string kind is not a form")
	}
	if _, ok := KindOf([]byte("kind: [a]\n")); ok {
		t.Fatal("list kind is not a form")
	}
	if _, ok := KindOf([]byte("- just\n- a\n- list\n")); ok {
		t.Fatal("non-mapping doc is not a form")
	}
	if _, ok := KindOf([]byte("kind: x\n\ttab: 1\n")); ok {
		t.Fatal("armor refusal is not a form")
	}
	if _, ok := KindOf([]byte("kind: [broken\n")); ok {
		t.Fatal("broken parse is not a form")
	}
	if _, ok := KindOf([]byte("kind: x\n---\nkind: y\n")); ok {
		t.Fatal("two documents are not a form")
	}
}

// The exactness law of Marshal: every byte of every string survives
// the marshal→decode round trip. Block scalars normalize CR away (a
// raw \r inside a literal block is a line break per the spec), so
// control-carrying strings ride exactScalar double-quoted scalars —
// the one style with a full escape set.
func TestMarshalExactBytes(t *testing.T) {
	cases := []string{
		"a\r\nb",        // CRLF: the translit-era pin
		"a\rb",          // a lone CR mid-string
		"trail\r",       // a trailing CR
		"\r",            // a bare CR
		"lead \r mid",   // a CR between words
		"tab\there",     // a tab
		"\x01ctl\x7f",   // controls and DEL
		"back\\slash\r", // the escape pair itself, beside a control
		"plain\nlines",  // multiline stays the literal block (exact)
		"\u0433\u0438",  // printable unicode passes unwrapped
	}
	for _, s := range cases {
		type doc struct {
			Text string `yaml:"text"`
		}
		b, err := Marshal(doc{Text: s})
		wantNoErr(t, "marshal "+fmt.Sprintf("%q", s), err)
		var back doc
		wantNoErr(t, "decode "+fmt.Sprintf("%q", s), DecodeDoc(b, &back))
		if back.Text != s {
			t.Fatalf("roundtrip lost bytes: %q became %q via\n%s", s, back.Text, b)
		}
		for _, ln := range strings.Split(strings.TrimRight(string(b), "\n"), "\n") {
			if ln == "---" || strings.HasPrefix(ln, "--- ") {
				t.Fatalf("column-0 separator inside a value:\n%s", b)
			}
		}
	}
	// map keys are identifiers: an unsafe key is a loud refusal
	if _, err := Marshal(map[string]int{"a\rb": 1}); err == nil {
		t.Fatal("a control-carrying map key must refuse loudly")
	}
	// the refusal propagates through every container degree
	if _, err := Marshal([]map[string]int{{"a\rb": 1}}); err == nil {
		t.Fatal("a bad key inside a slice must refuse loudly")
	}
	if _, err := Marshal(struct {
		M map[string]int `yaml:"m"`
	}{M: map[string]int{"a\rb": 1}}); err == nil {
		t.Fatal("a bad key inside a struct must refuse loudly")
	}
	if _, err := Marshal(&[]map[string]int{{"a\rb": 1}}); err == nil {
		t.Fatal("a bad key behind a pointer must refuse loudly")
	}
	if _, err := Marshal(map[string]map[string]int{"outer": {"a\rb": 1}}); err == nil {
		t.Fatal("a bad key inside a nested map value must refuse loudly")
	}
	// a type that renders itself rides untouched: the library
	// dispatches its method (the wrapper must not rebuild it)
	var self yamlioSelf
	self.inner.Text = "ok"
	ss, err := Marshal(self)
	wantNoErr(t, "self marshal", err)
	if !strings.Contains(string(ss), "selfy: ok") {
		t.Fatalf("the self-marshaling type renders itself: %q", ss)
	}
	mb, err := Marshal(map[string]int{"\xff": 1})
	wantNoErr(t, "sanitize map key marshal", err)
	if strings.Contains(string(mb), "\xff") || !strings.Contains(string(mb), "\ufffd") {
		t.Fatalf("a map key sanitizes: %q", mb)
	}
	// invalid UTF-8 has no byte-exact scalar spelling: deterministic
	// sanitization to U+FFFD (the storage layer carries text)
	type bin struct {
		Text string `yaml:"text"`
	}
	sb, err := Marshal(bin{Text: "\xff\xfe"})
	wantNoErr(t, "sanitize marshal", err)
	var sback bin
	wantNoErr(t, "sanitize decode", DecodeDoc(sb, &sback))
	if sback.Text != "\ufffd" {
		t.Fatalf("undecodable runs sanitize to one replacement: %q", sback.Text)
	}
}

// The exactness wrapper's walk: plain-data trees rebuild with their
// unsafe strings wrapped; unbuildable shapes pass through when they
// carry nothing unsafe and refuse loudly when they do; scalars ride
// as themselves.
func TestMarshalExactWalk(t *testing.T) {
	type inner struct {
		Note string `yaml:"note"`
	}
	type outer struct {
		Name  string            `yaml:"name"`
		Items []inner           `yaml:"items"`
		Tags  map[string]string `yaml:"tags"`
		When  *inner            `yaml:"when"`
		Any   interface{}       `yaml:"any"`
		N     int               `yaml:"n"`
		F     float64           `yaml:"f"`
		B     bool              `yaml:"b"`
	}
	v := outer{
		Name:  "x",
		Items: []inner{{Note: "a\rb"}},
		Tags:  map[string]string{"k": "tab\there"},
		When:  &inner{Note: "plain"},
		Any:   "ctrl\x01",
		N:     2, F: 1.5, B: true,
	}
	b, err := Marshal(v)
	wantNoErr(t, "walk marshal", err)
	var back outer
	wantNoErr(t, "walk decode", DecodeDoc(b, &back))
	if back.Items[0].Note != "a\rb" || back.Tags["k"] != "tab\there" ||
		back.When == nil || back.When.Note != "plain" || back.Any.(string) != "ctrl\x01" ||
		back.N != 2 || back.F != 1.5 || !back.B || back.Name != "x" {
		t.Fatalf("the walk roundtrip broke: %+v via\n%s", back, b)
	}
	// nested containers
	type nest struct {
		List []map[string][]string `yaml:"list"`
	}
	nb, err := Marshal(nest{List: []map[string][]string{{"m": {"a\rb"}}}})
	wantNoErr(t, "nest marshal", err)
	var nback nest
	wantNoErr(t, "nest decode", DecodeDoc(nb, &nback))
	if nback.List[0]["m"][0] != "a\rb" {
		t.Fatalf("nested exactness broke: %q", nback.List[0]["m"][0])
	}
	// time.Time passes through untouched (unexported fields, nothing to lose)
	type timed struct {
		At time.Time `yaml:"at"`
	}
	now := time.Date(2000, 1, 1, 4, 0, 0, 0, time.UTC)
	tb, err := Marshal(timed{At: now})
	wantNoErr(t, "time marshal", err)
	var tback timed
	wantNoErr(t, "time decode", DecodeDoc(tb, &tback))
	if !tback.At.Equal(now) {
		t.Fatalf("time passthrough broke: %v", tback.At)
	}
	// a struct mixing unexported fields with an unsafe string refuses loudly
	type mixed struct {
		Public string `yaml:"public"`
		hidden string
	}
	m := mixed{Public: "ok"}
	m.hidden = "a\rb"
	if _, err := Marshal(m); err == nil {
		t.Fatal("an unbuildable struct with unsafe bytes must refuse loudly")
	}
	// a struct mixing unexported fields WITHOUT unsafe bytes passes through
	type carrier struct {
		Public string `yaml:"public"`
		hidden string
	}
	cb, err := Marshal(carrier{Public: "fine", hidden: "secret"})
	wantNoErr(t, "carrier marshal", err)
	if strings.Contains(string(cb), "secret") {
		t.Fatalf("an unexported field leaked into the marshal:\n%s", cb)
	}
	var cback struct {
		Public string `yaml:"public"`
	}
	wantNoErr(t, "carrier decode", DecodeDoc(cb, &cback))
	if cback.Public != "fine" {
		t.Fatalf("carrier roundtrip broke: %q", cback.Public)
	}
	// the sanitization law rides every leaf kind the same way
	if b, err := Marshal([]string{"\xff"}); err != nil || !strings.Contains(string(b), "\ufffd") {
		t.Fatalf("an invalid-UTF-8 slice element sanitizes: %q %v", b, err)
	}
	if b, err := Marshal(map[string]string{"k": "\xff"}); err != nil || !strings.Contains(string(b), "\ufffd") {
		t.Fatalf("an invalid-UTF-8 map value sanitizes: %q %v", b, err)
	}
	type strukt struct {
		Text string `yaml:"text"`
	}
	if _, err := Marshal(strukt{Text: "\xff"}); err != nil {
		t.Fatalf("an invalid-UTF-8 struct field sanitizes: %v", err)
	}
	bad := "\xff"
	pb, err := Marshal(struct {
		Ptr *string `yaml:"ptr"`
	}{Ptr: &bad})
	wantNoErr(t, "sanitize ptr marshal", err)
	var pback struct {
		Ptr *string `yaml:"ptr"`
	}
	wantNoErr(t, "sanitize ptr decode", DecodeDoc(pb, &pback))
	if pback.Ptr == nil || *pback.Ptr != "\ufffd" {
		t.Fatalf("a string behind a pointer sanitizes: %+v", pback.Ptr)
	}
	if _, err := Marshal(make(chan int)); err == nil {
		t.Fatal("a non-data value must refuse loudly")
	}
	// nil leaves ride the null form
	type nullable struct {
		Ptr *inner `yaml:"ptr"`
		Any any    `yaml:"any"`
	}
	nl, err := Marshal(nullable{})
	wantNoErr(t, "nil marshal", err)
	var nlb struct {
		Ptr *inner `yaml:"ptr"`
		Any any    `yaml:"any"`
	}
	wantNoErr(t, "nil decode", DecodeDoc(nl, &nlb))
	if nlb.Ptr != nil || nlb.Any != nil {
		t.Fatalf("nil leaves broke: %+v", nlb)
	}
	if nb, err := Marshal(nil); err != nil || strings.TrimSpace(string(nb)) != "null" {
		t.Fatalf("a nil root marshals to null: %q %v", nb, err)
	}
	// a nil map stays null; an empty map is the empty mapping
	type withMaps struct {
		M map[string]string `yaml:"m,omitempty"`
	}
	mb, err := Marshal(withMaps{})
	wantNoErr(t, "nil map marshal", err)
	if strings.Contains(string(mb), "m:") {
		t.Fatalf("omitempty nil map rides: %q", mb)
	}
}

// yamlioSelf — a type carrying its own one-key render (the walk's
// self-marshaling passthrough pin).
type yamlioSelf struct {
	inner struct {
		Text string `yaml:"text"`
	}
}

func (s yamlioSelf) MarshalYAML() (interface{}, error) {
	return map[string]interface{}{"selfy": s.inner.Text}, nil
}

// The whitespace-only-line family (the root defect): a literal block
// scalar cannot carry a line that is only spaces — the YAML spec
// absorbs such a line's white space into the line break, and the
// lone "\n" renders as a clip-chomped empty body. Every such string
// must ride the exact double-quoted form, or the machine writes
// bytes it cannot read back (an instance-killing atom of a past series).
func TestMarshalWhitespaceOnlyLines(t *testing.T) {
	cases := []string{
		"\n",       // the registry episode's seed itself: a lone line feed
		"\n \n",    // an empty line and a space-only line
		"\n \n \n", // only empty/space-only lines
		"\n \nx",   // a space-only line above real text
		"x\n \n",   // a space-only line below real text
		"\n \nx\n", // both edges
		"\n\n \n",  // the space-only line last
		"x\n\n \n", // after a blank line
	}
	for _, s := range cases {
		type doc struct {
			Text string `yaml:"text"`
		}
		b, err := Marshal(doc{Text: s})
		wantNoErr(t, "marshal "+fmt.Sprintf("%q", s), err)
		var back doc
		wantNoErr(t, "decode "+fmt.Sprintf("%q", s), DecodeDoc(b, &back))
		if back.Text != s {
			t.Fatalf("roundtrip lost bytes: %q became %q via\n%s", s, back.Text, b)
		}
	}
}

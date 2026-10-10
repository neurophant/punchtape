package canon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/yamlio"
)

// --- command splitting ---

func TestSplitCommand(t *testing.T) {
	cases := []struct {
		name string
		line string
		want []string
		fail bool
	}{
		{"plain", "tool arg", []string{"tool", "arg"}, false},
		{"extra spaces", "tool  arg   x", []string{"tool", "arg", "x"}, false},
		{"tab separator", "tool\targ", []string{"tool", "arg"}, false},
		{"leading and trailing space", "  tool arg  ", []string{"tool", "arg"}, false},
		{"empty line", "", nil, false},
		{"one word", "tool", []string{"tool"}, false},
		{"double quoted arg", `sh -c "echo a b"`, []string{"sh", "-c", "echo a b"}, false},
		{"single quoted arg", `sh -c 'echo a b'`, []string{"sh", "-c", "echo a b"}, false},
		{"quotes inside the other quote", `sh -c 'echo "x y"'`, []string{"sh", "-c", `echo "x y"`}, false},
		{"empty quoted arg stays", `x "" y`, []string{"x", "", "y"}, false},
		{"only an empty quoted arg", `""`, []string{""}, false},
		{"quoted arg with spaces", `"a b" c`, []string{"a b", "c"}, false},
		{"adjacent quoted words", `"a""b" c`, []string{"ab", "c"}, false},
		{"unbalanced double quote", `echo "hi`, nil, true},
		{"unbalanced single quote", `echo 'hi`, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SplitCommand(tc.line)
			if tc.fail {
				if err == nil {
					t.Fatalf("must refuse: %v", got)
				}
				if !strings.Contains(err.Error(), "unbalanced quote") {
					t.Fatalf("the refusal names the quote: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("split: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("fields: %#v want %#v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("field %d: %q want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// --- the run command forms ---

func TestCmdDecode(t *testing.T) {
	decode := func(raw string) (Cmd, error) {
		var doc struct {
			Run Cmd `yaml:"run"`
		}
		err := yamlio.DecodeDoc([]byte(raw), &doc)
		return doc.Run, err
	}
	rc, err := decode("run: tool arg\n")
	if err != nil || rc.Line != "tool arg" || rc.IsArgv {
		t.Fatalf("scalar form: %v %+v", err, rc)
	}
	rc, err = decode("run: [tool, \"a b\"]\n")
	if err != nil || !rc.IsArgv || len(rc.Argv) != 2 || rc.Argv[1] != "a b" {
		t.Fatalf("argv form: %v %+v", err, rc)
	}
	rc, err = decode("run: []\n")
	if err != nil || !rc.IsArgv || len(rc.Argv) != 0 {
		t.Fatalf("empty argv decodes as the argv form: %v %+v", err, rc)
	}
	rc, err = decode("run:\n  - tool\n  - arg\n")
	if err != nil || !rc.IsArgv || len(rc.Argv) != 2 {
		t.Fatalf("block argv form: %v %+v", err, rc)
	}
	if _, err := decode("run: [1]\n"); err == nil {
		t.Fatal("a non-string list element must error")
	}
	if _, err := decode("run: {a: 1}\n"); err == nil {
		t.Fatal("an object is not a run form")
	}
	if _, err := decode("run: 5\n"); err == nil {
		t.Fatal("a number scalar is not a run form")
	}
}

func TestCmdFields(t *testing.T) {
	f, err := (Cmd{Argv: []string{"t", "x"}, IsArgv: true}).Fields()
	if err != nil || len(f) != 2 || f[0] != "t" {
		t.Fatalf("argv fields: %v %v", f, err)
	}
	if _, err := (Cmd{Argv: []string{}, IsArgv: true}).Fields(); err == nil {
		t.Fatal("an empty argv must error")
	}
	f, err = (Cmd{Line: "t a"}).Fields()
	if err != nil || len(f) != 2 {
		t.Fatalf("line fields: %v %v", f, err)
	}
}

func TestCmdForm(t *testing.T) {
	if got := (Cmd{Line: "x y"}).Form(); got != "x y" {
		t.Fatalf("scalar form render: %q", got)
	}
	if got := (Cmd{Argv: []string{"x"}, IsArgv: true}).Form(); got != "[x]" {
		t.Fatalf("argv form render: %q", got)
	}
	if got := (Cmd{Argv: []string{}, IsArgv: true}).Form(); got != "[]" {
		t.Fatalf("empty argv render: %q", got)
	}
	got := (Cmd{Argv: []string{"a b", `q"t`, "", "c,d", `a\b`, "e#f"}, IsArgv: true}).Form()
	want := `["a b", "q\"t", "", "c,d", a\b, "e#f"]`
	if got != want {
		t.Fatalf("argv quoting render:\n got %q\nwant %q", got, want)
	}
}

func TestCmdMarshalRoundTrip(t *testing.T) {
	forms := []Cmd{
		{Line: "tool arg"},
		{Line: `say "hi"`},
		{Argv: []string{"tool", "a b", "1"}, IsArgv: true},
		{Argv: []string{"x"}, IsArgv: true},
		{Line: "say \x01hi"}, // a control byte survives the exactness pass
	}
	for i, rc := range forms {
		data, err := yamlio.Marshal(struct {
			Run Cmd `yaml:"run"`
		}{rc})
		if err != nil {
			t.Fatalf("form %d marshal: %v", i, err)
		}
		var doc struct {
			Run Cmd `yaml:"run"`
		}
		if err := yamlio.DecodeDoc(data, &doc); err != nil {
			t.Fatalf("form %d re-decode: %v\n%s", i, err, data)
		}
		back := doc.Run
		if back.IsArgv != rc.IsArgv || back.Line != rc.Line || len(back.Argv) != len(rc.Argv) {
			t.Fatalf("form %d round trip: %+v -> %+v", i, rc, back)
		}
		for j := range back.Argv {
			if back.Argv[j] != rc.Argv[j] {
				t.Fatalf("form %d argv[%d]: %q want %q", i, j, back.Argv[j], rc.Argv[j])
			}
		}
	}
	// the scalar form marshals as the bare scalar
	data, err := yamlio.Marshal(struct {
		Run Cmd `yaml:"run"`
	}{Cmd{Line: "x"}})
	if err != nil || strings.TrimSpace(string(data)) != "run: x" {
		t.Fatalf("marshal scalar: %s %v", data, err)
	}
}

// --- text expectations ---

func TestTextExpectDecode(t *testing.T) {
	decode := func(raw string) (*TextExpect, error) {
		var doc struct {
			Out *TextExpect `yaml:"out"`
		}
		err := yamlio.DecodeDoc([]byte(raw), &doc)
		return doc.Out, err
	}
	te, err := decode("out: hello\n")
	if err != nil || te.Contains || te.Text != "hello" {
		t.Fatalf("exact form: %v %+v", err, te)
	}
	te, err = decode("out: \"(contains) abc\"\n")
	if err != nil || !te.Contains || te.Text != "abc" {
		t.Fatalf("contains form: %v %+v", err, te)
	}
	// any other spelling of the marker head is the exact form
	te, err = decode("out: \"(contains)abc\"\n")
	if err != nil || te.Contains || te.Text != "(contains)abc" {
		t.Fatalf("marker head without the space is exact: %v %+v", err, te)
	}
	te, err = decode("out: \"\"\n")
	if err != nil || te.Contains || te.Text != "" {
		t.Fatalf("explicit emptiness: %v %+v", err, te)
	}
	te, err = decode("out:\n")
	if err != nil || te != nil {
		t.Fatalf("a null value leaves the expectation absent: %v %+v", err, te)
	}
	// scalar coercion: numbers and booleans in their shortest spelling
	te, err = decode("out: 42\n")
	if err != nil || te.Text != "42" {
		t.Fatalf("int coercion: %v %+v", err, te)
	}
	te, err = decode("out: -7\n")
	if err != nil || te.Text != "-7" {
		t.Fatalf("negative int coercion: %v %+v", err, te)
	}
	te, err = decode("out: 3.5\n")
	if err != nil || te.Text != "3.5" {
		t.Fatalf("float coercion: %v %+v", err, te)
	}
	te, err = decode("out: 3.0\n")
	if err != nil || te.Text != "3" {
		t.Fatalf("integral float coercion: %v %+v", err, te)
	}
	te, err = decode("out: true\n")
	if err != nil || te.Text != "true" {
		t.Fatalf("bool coercion: %v %+v", err, te)
	}
	te, err = decode("out: \"42\"\n")
	if err != nil || te.Text != "42" {
		t.Fatalf("a quoted number stays text: %v %+v", err, te)
	}
	// refusals
	if _, err := decode("out: \"(contains) \"\n"); err == nil {
		t.Fatal("an empty fragment after the marker must refuse")
	}
	if _, err := decode("out: {a: 1}\n"); err == nil {
		t.Fatal("a mapping is not a scalar")
	}
	if _, err := decode("out: [a]\n"); err == nil {
		t.Fatal("a sequence is not a scalar")
	}
}

func TestTextExpectMarshalAndForm(t *testing.T) {
	data, err := yamlio.Marshal(struct {
		Out *TextExpect `yaml:"out"`
	}{&TextExpect{Contains: true, Text: "frag"}})
	if err != nil || strings.TrimSpace(string(data)) != `out: (contains) frag` {
		t.Fatalf("marshal contains: %s %v", data, err)
	}
	// nil marshals as the absent field
	data, err = yamlio.Marshal(struct {
		Out *TextExpect `yaml:"out,omitempty"`
	}{nil})
	if err != nil || strings.TrimSpace(string(data)) != "{}" {
		t.Fatalf("marshal nil: %s %v", data, err)
	}
	if got := (TextExpect{Contains: true, Text: "x"}).Form(); got != "(contains) x" {
		t.Fatalf("contains form render: %q", got)
	}
	if got := (TextExpect{Text: "x"}).Form(); got != "x" {
		t.Fatalf("exact form render: %q", got)
	}
}

// --- seeds ---

func TestSeedDecode(t *testing.T) {
	decode := func(raw string) ([]Seed, error) {
		var doc struct {
			Seeds []Seed `yaml:"seeds"`
		}
		err := yamlio.DecodeDoc([]byte(raw), &doc)
		return doc.Seeds, err
	}
	seeds, err := decode("seeds:\n  - path: a.txt\n    text: hi\n")
	if err != nil || len(seeds) != 1 || seeds[0].Path != "a.txt" || seeds[0].Text != "hi" || seeds[0].FromInput != "" {
		t.Fatalf("text form: %v %+v", err, seeds)
	}
	seeds, err = decode("seeds:\n  - path: big.bin\n    from-input: given.bin\n")
	if err != nil || len(seeds) != 1 || seeds[0].FromInput != "given.bin" || seeds[0].Text != "" {
		t.Fatalf("from-input form: %v %+v", err, seeds)
	}
	refusals := []struct {
		name string
		raw  string
	}{
		{"both forms", "seeds:\n  - path: a\n    text: x\n    from-input: b\n"},
		{"no source form", "seeds:\n  - path: a\n"},
		{"no path", "seeds:\n  - text: x\n"},
		{"no path with from-input", "seeds:\n  - from-input: x\n"},
		{"unknown field", "seeds:\n  - path: a\n    text: x\n    bogus: 1\n"},
		{"not a mapping", "seeds:\n  - plain-text\n"},
		{"path not a scalar", "seeds:\n  - path: [a]\n    text: x\n"},
		{"empty text is no form", "seeds:\n  - path: a\n    text: \"\"\n"},
	}
	for _, tc := range refusals {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := decode(tc.raw); err == nil {
				t.Fatal("must refuse")
			}
		})
	}
}

// --- state wants ---

func TestStateWantDecode(t *testing.T) {
	decode := func(raw string) ([]StateWant, error) {
		var doc struct {
			State []StateWant `yaml:"state"`
		}
		err := yamlio.DecodeDoc([]byte(raw), &doc)
		return doc.State, err
	}
	sts, err := decode("state:\n  - path: out.txt\n    text: expected\n")
	if err != nil || len(sts) != 1 || sts[0].Path != "out.txt" || sts[0].Text == nil || sts[0].Text.Text != "expected" {
		t.Fatalf("text form: %v %+v", err, sts)
	}
	sts, err = decode("state:\n  - path: out.bin\n    from-run: art.bin\n")
	if err != nil || len(sts) != 1 || sts[0].FromRun != "art.bin" || sts[0].Text != nil {
		t.Fatalf("from-run form: %v %+v", err, sts)
	}
	// the contains form rides the text expectation
	sts, err = decode("state:\n  - path: out.txt\n    text: \"(contains) frag\"\n")
	if err != nil || !sts[0].Text.Contains {
		t.Fatalf("contains state text: %v %+v", err, sts)
	}
	refusals := []struct {
		name string
		raw  string
	}{
		{"both forms", "state:\n  - path: a\n    text: x\n    from-run: b\n"},
		{"no expectation form", "state:\n  - path: a\n"},
		{"no path", "state:\n  - text: x\n"},
		{"unknown field", "state:\n  - path: a\n    text: x\n    bogus: 1\n"},
		{"not a mapping", "state:\n  - plain\n"},
		{"path not a scalar", "state:\n  - path: [a]\n    text: x\n"},
	}
	for _, tc := range refusals {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := decode(tc.raw); err == nil {
				t.Fatal("must refuse")
			}
		})
	}
}

// --- rows ---

func TestRowDecodeFull(t *testing.T) {
	raw := []byte(`run: ["./tool", "arg"]
seeds:
  - path: in.txt
    text: seed
out: "(contains) ok"
err: noise
rc: 3
state:
  - path: out.txt
    from-run: art.bin
bytes: file://sample.bin
timeout: 120
outcome: green
reason: because
env-fail: true
`)
	var row Row
	if err := yamlio.DecodeDoc(raw, &row); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !row.Run.IsArgv || row.Run.Argv[0] != "./tool" {
		t.Fatalf("run: %+v", row.Run)
	}
	if len(row.Seeds) != 1 || row.Seeds[0].Text != "seed" {
		t.Fatalf("seeds: %+v", row.Seeds)
	}
	if row.Out == nil || !row.Out.Contains || row.Out.Text != "ok" {
		t.Fatalf("out: %+v", row.Out)
	}
	if row.Err == nil || row.Err.Contains || row.Err.Text != "noise" {
		t.Fatalf("err: %+v", row.Err)
	}
	if row.Rc == nil || *row.Rc != 3 {
		t.Fatalf("rc: %+v", row.Rc)
	}
	if len(row.State) != 1 || row.State[0].FromRun != "art.bin" {
		t.Fatalf("state: %+v", row.State)
	}
	if row.Bytes != "file://sample.bin" || row.Timeout != 120 {
		t.Fatalf("bytes/timeout: %q %d", row.Bytes, row.Timeout)
	}
	if row.Outcome != RowGreen || row.Reason != "because" || !row.EnvFail {
		t.Fatalf("machine record: %+v", row)
	}
	// the row survives the canonical rewrite
	data, err := yamlio.Marshal(row)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back Row
	if err := yamlio.DecodeDoc(data, &back); err != nil {
		t.Fatalf("re-decode: %v\n%s", err, data)
	}
	if back.Timeout != row.Timeout || back.Outcome != row.Outcome || back.EnvFail != row.EnvFail {
		t.Fatalf("round trip: %+v", back)
	}
	if back.Out == nil || !back.Out.Contains || back.Out.Text != "ok" {
		t.Fatalf("round trip out: %+v", back.Out)
	}
}

func TestBytesRef(t *testing.T) {
	cases := []struct {
		bytes string
		want  string
	}{
		{"file://sample.bin", "sample.bin"},
		{"file://nested/dir/sample.bin", "nested/dir/sample.bin"},
		{"sample.bin", ""},
		{"file:/sample.bin", ""},
		{"", ""},
		{"http://x", ""},
	}
	for _, tc := range cases {
		if got := (Row{Bytes: tc.bytes}).BytesRef(); got != tc.want {
			t.Fatalf("BytesRef(%q): %q want %q", tc.bytes, got, tc.want)
		}
	}
}

// --- atoms ---

func TestAtomGreen(t *testing.T) {
	if (Atom{ID: "B1"}).Green() {
		t.Fatal("an atom without rows is not green")
	}
	green := Row{Outcome: RowGreen}
	if !(Atom{ID: "B1", Rows: []Row{green}}).Green() {
		t.Fatal("all rows green make the atom green")
	}
	if (Atom{ID: "B1", Rows: []Row{green, {Outcome: RowRed}}}).Green() {
		t.Fatal("one red row breaks the atom")
	}
	if (Atom{ID: "B1", Rows: []Row{green, {Outcome: RowUnknown}}}).Green() {
		t.Fatal("an unrun row breaks the atom")
	}
}

func TestRowLabel(t *testing.T) {
	if got := RowLabel("B1", 2); got != "B1/2" {
		t.Fatalf("label: %q", got)
	}
	if got := RowLabel("chain-passthrough", 1); got != "chain-passthrough/1" {
		t.Fatalf("label: %q", got)
	}
}

// --- stages ---

func TestStageOrder(t *testing.T) {
	want := []string{StageIntake, StageImplement, StageDeliver}
	if len(StageOrder) != len(want) {
		t.Fatalf("stage order: %v", StageOrder)
	}
	for i := range want {
		if StageOrder[i] != want[i] {
			t.Fatalf("stage order: %v want %v", StageOrder, want)
		}
	}
}

// --- the zone model ---

func TestZoneModel(t *testing.T) {
	cases := []struct {
		rel   string
		zone  string
		rule  string
		found bool
	}{
		{"canon/atoms/", ZoneCanon, WriteJournaled, true},
		{"canon/atoms/b1.yaml", ZoneCanon, WriteJournaled, true},
		{"state.yaml", ZoneService, WriteJournaled, true},
		{"checkpoints/0001.yaml", ZoneService, WriteJournaled, true},
		{"journal.yamll", ZoneService, WriteAppend, true},
		{"ledger.yamll", ZoneService, WriteAppend, true},
		{"cache/digest.bin", ZoneService, WriteRebuilt, true},
		{"inbox/current.txt", "", "", false},
		{"canon", "", "", false},     // the prefix without the slash
		{"spec.md", "", "", false},   // a render, not a zone
		{"AGENTS.md", "", "", false}, // the agent's territory
	}
	for _, tc := range cases {
		zone, ok := ZoneOf(tc.rel)
		if ok != tc.found || zone != tc.zone {
			t.Fatalf("ZoneOf(%q): %q %v want %q %v", tc.rel, zone, ok, tc.zone, tc.found)
		}
		rule, ok := ZoneWriteRule(tc.rel)
		if ok != tc.found || rule != tc.rule {
			t.Fatalf("ZoneWriteRule(%q): %q %v want %q %v", tc.rel, rule, ok, tc.rule, tc.found)
		}
		if JournaledPath(tc.rel) != (tc.found && tc.rule == WriteJournaled) {
			t.Fatalf("JournaledPath(%q) disagrees with the zone model", tc.rel)
		}
	}
	// the derived canon collection list — Init's skeleton mirrors it
	col := Collections()
	if len(col) != 1 || col[0] != "canon/atoms" {
		t.Fatalf("collections: %v", col)
	}
}

// --- the store (real filesystem, real ports) ---

func TestStoreInitAndOpen(t *testing.T) {
	dir := t.TempDir()
	s, err := Init(dir)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if s.Root() != Dir(dir) {
		t.Fatalf("root: %q want %q", s.Root(), Dir(dir))
	}
	// the skeleton: the canon collection, the service dirs, the state
	for _, rel := range []string{
		"state.yaml", "canon/atoms", "checkpoints", "cache", "inbox",
	} {
		if _, err := os.Stat(filepath.Join(Dir(dir), rel)); err != nil {
			t.Fatalf("skeleton %s: %v", rel, err)
		}
	}
	st, err := s.State()
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if st.SchemaVersion != SchemaVersion || st.Stage != StageIntake {
		t.Fatalf("fresh state: %+v", st)
	}
	if st.Counters["CKP"] != 0 {
		t.Fatalf("counters: %+v", st.Counters)
	}
	// a repeated init refuses: no silent overwrites
	if _, err := Init(dir); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("repeated init must refuse: %v", err)
	}
	// open finds it
	s2, exists, err := Open(dir)
	if err != nil || !exists || s2 == nil {
		t.Fatalf("open: %v %v", exists, err)
	}
	// an empty directory is a fact, not an error
	_, exists, err = Open(t.TempDir())
	if err != nil || exists {
		t.Fatalf("open on a virgin dir: %v %v", exists, err)
	}
}

func TestStoreStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s, err := Init(dir)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	wish := "make the echo tool sing"
	empty := []string{}
	st := State{
		SchemaVersion: SchemaVersion,
		Stage:         StageImplement,
		Wish:          &wish,
		Language:      "en",
		Build:         empty, // declared and cleared
		BuildIgnore:   []string{"__pycache__", "*.pyc"},
		Touched:       []string{"B1", "B2"},
		Counters:      map[string]int{"CKP": 3, "RUN": 7},
	}
	if err := s.PutState(st); err != nil {
		t.Fatalf("put: %v", err)
	}
	back, err := s.State()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if back.Stage != StageImplement || back.Wish == nil || *back.Wish != wish || back.Language != "en" {
		t.Fatalf("round trip head: %+v", back)
	}
	// the omitempty law of the stored state: a declared-empty build
	// canonicalizes to absent (nil) in the file form
	if back.Build != nil {
		t.Fatalf("an empty build canonicalizes to absent: %+v", back.Build)
	}
	if len(back.BuildIgnore) != 2 || back.BuildIgnore[0] != "__pycache__" {
		t.Fatalf("build-ignore: %+v", back.BuildIgnore)
	}
	if len(back.Touched) != 2 || back.Counters["RUN"] != 7 {
		t.Fatalf("touched/counters: %+v %+v", back.Touched, back.Counters)
	}
	// nil wish round trips as absent
	st.Wish, st.Stage = nil, StageDeliver
	if err := s.PutState(st); err != nil {
		t.Fatalf("put: %v", err)
	}
	if back, err = s.State(); err != nil || back.Wish != nil {
		t.Fatalf("nil wish: %v %+v", err, back)
	}
}

func TestStoreStateRefusals(t *testing.T) {
	dir := t.TempDir()
	s, err := Init(dir)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	// a foreign schema version is an honest refusal
	if err := s.PutState(State{SchemaVersion: SchemaVersion - 1, Stage: StageIntake}); err != nil {
		t.Fatalf("put: %v", err)
	}
	if _, err := s.State(); err == nil || !strings.Contains(err.Error(), "schema") {
		t.Fatalf("schema mismatch must refuse: %v", err)
	}
	// a broken state file refuses
	if err := os.WriteFile(filepath.Join(Dir(dir), "state.yaml"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.State(); err == nil {
		t.Fatal("a broken state file must refuse")
	}
	// a foreign field refuses (strict decode)
	body := "schema-version: 8\nstage: intake\ncounters: {}\nbogus: 1\n"
	if err := os.WriteFile(filepath.Join(Dir(dir), "state.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.State(); err == nil {
		t.Fatal("a foreign field must refuse")
	}
	// a missing state file is a failure
	if err := os.Remove(filepath.Join(Dir(dir), "state.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.State(); err == nil {
		t.Fatal("a missing state file must fail the read")
	}
}

func TestStoreAtoms(t *testing.T) {
	dir := t.TempDir()
	s, err := Init(dir)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	// an empty corpus reads empty
	if atoms, err := s.Atoms(); err != nil || len(atoms) != 0 {
		t.Fatalf("empty atoms: %v %+v", err, atoms)
	}
	// two atoms written out of id order read back sorted
	b2 := Atom{ID: "B2", Name: "second", Prose: "p", Rows: []Row{{Run: Cmd{Line: "tool"}, Outcome: RowGreen}}}
	b1 := Atom{ID: "A1", Name: "first", Prose: "p", Rows: []Row{{Run: Cmd{Line: "tool"}, Outcome: RowRed}}}
	for name, a := range map[string]Atom{"z-b2.yaml": b2, "a-b1.yaml": b1} {
		data, err := yamlio.Marshal(a)
		if err != nil {
			t.Fatalf("marshal %s: %v", name, err)
		}
		if err := yamlio.WriteAtomic(filepath.Join(Dir(dir), "canon", "atoms", name), data); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	atoms, err := s.Atoms()
	if err != nil {
		t.Fatalf("atoms: %v", err)
	}
	if len(atoms) != 2 || atoms[0].ID != "A1" || atoms[1].ID != "B2" {
		t.Fatalf("atoms must read sorted by id: %+v", atoms)
	}
	if atoms[0].Green() || !atoms[1].Green() {
		t.Fatalf("green state rides the file: %+v", atoms)
	}
	// non-yaml files and directories are not corpus entries
	if err := os.WriteFile(filepath.Join(Dir(dir), "canon", "atoms", "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(Dir(dir), "canon", "atoms", "sub.yaml"), 0o755); err != nil {
		t.Fatal(err)
	}
	if atoms, err = s.Atoms(); err != nil || len(atoms) != 2 {
		t.Fatalf("noise must not change the corpus: %v %+v", err, atoms)
	}
	// a broken atom file is an honest refusal
	if err := os.WriteFile(filepath.Join(Dir(dir), "canon", "atoms", "broken.yaml"), []byte("id: ["), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Atoms(); err == nil {
		t.Fatal("a broken atom file must refuse the read")
	}
	// an unreadable corpus entry is an honest refusal
	if err := os.Symlink("nowhere", filepath.Join(Dir(dir), "canon", "atoms", "dangling.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Atoms(); err == nil {
		t.Fatal("a dangling corpus entry must refuse the read")
	}
	// an unreadable corpus directory is an honest refusal (not a quiet
	// empty corpus): restore before the temp cleanup
	atomsDir := filepath.Join(Dir(dir), "canon", "atoms")
	if err := os.Chmod(atomsDir, 0o000); err == nil {
		t.Cleanup(func() { _ = os.Chmod(atomsDir, 0o755) })
		if _, err := s.Atoms(); err == nil {
			t.Fatal("an unreadable corpus directory must refuse")
		}
		if err := os.Chmod(atomsDir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

// --- the exact-block closers (fault branches over the real FS) ---

func TestStoreInitFaults(t *testing.T) {
	dir := t.TempDir()
	// a FILE where a directory must be created -> MkdirAll fails
	block := filepath.Join(dir, "blocker")
	if err := os.WriteFile(block, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Init(filepath.Join(block, "under")); err == nil {
		t.Fatal("mkdir fault swallowed")
	}
	// an unwritable .punchtape (state.yaml absent, dir read-only) ->
	// the skeleton stands, PutState's atomic write fails
	dir2 := t.TempDir()
	hidden := filepath.Join(dir2, ".punchtape")
	if err := os.MkdirAll(hidden, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(hidden, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(hidden, 0o755) })
	if _, err := Init(dir2); err == nil {
		t.Fatal("putstate fault swallowed")
	}
}

func TestStoreOpenENOTDIR(t *testing.T) {
	file := filepath.Join(t.TempDir(), "plain")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Open(filepath.Join(file, "x")); err == nil {
		t.Fatal("ENOTDIR swallowed")
	}
}

func TestEachFileFaults(t *testing.T) {
	root := t.TempDir()
	s := &Store{root: filepath.Join(root, "no-such")}
	if _, err := s.Atoms(); err != nil {
		t.Fatalf("missing atoms dir must be empty, got %v", err)
	}
	// a broken symlink named *.yaml -> the entry read fails
	s2 := &Store{root: root}
	if err := os.MkdirAll(filepath.Join(root, "canon", "atoms"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/nonexistent-target", filepath.Join(root, "canon", "atoms", "weird.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, err := s2.Atoms(); err == nil {
		t.Fatal("unreadable entry swallowed")
	}
}

func TestAtomDigestOf(t *testing.T) {
	a := Atom{ID: "B1", Name: "n", Prose: "p", Rows: []Row{{}}}
	d1 := a.DigestOf()
	if d1 == "" {
		t.Fatal("empty digest")
	}
	a.Digest = "stale"
	if got := a.DigestOf(); got != d1 {
		t.Fatal("digest is not self-excluding")
	}
}

func TestScalarTextNil(t *testing.T) {
	if s, err := scalarText(nil); err != nil || s != "" {
		t.Fatalf("nil: %q %v", s, err)
	}
	if _, err := scalarText([]string{}); err == nil {
		t.Fatal("non-scalar accepted")
	}
}

func TestZoneDirsUnknown(t *testing.T) {
	if zoneDirs("bogus-zone") != nil {
		t.Fatal("unknown zone has dirs")
	}
}

// The self-damage guard: a grammar-valid seed whose
// text is one line feed serialized into an atom the machine itself
// could not decode back — every verb of the instance died on the one
// atom. The atom record must survive ANY grammar-valid seed and
// expectation byte-exactly: marshal→decode→marshal is the identity.
func TestAtomNewlineSeedRoundTrip(t *testing.T) {
	row := Row{
		Run:   Cmd{Line: "cat seed.txt"},
		Seeds: []Seed{{Path: "seed.txt", Text: "\n"}},
		Out:   &TextExpect{Text: "\n"},
	}
	atom := Atom{ID: "b1", Name: "one-line-feed", Prose: "prose", Rows: []Row{row}}
	data, err := yamlio.Marshal(atom)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back Atom
	if err := yamlio.DecodeDoc(data, &back); err != nil {
		t.Fatalf("the machine wrote an atom it cannot read back: %v\n%s", err, data)
	}
	if len(back.Rows) != 1 || len(back.Rows[0].Seeds) != 1 || back.Rows[0].Seeds[0].Text != "\n" {
		t.Fatalf("seed lost: %+v", back.Rows[0].Seeds)
	}
	if back.Rows[0].Out == nil || back.Rows[0].Out.Text != "\n" {
		t.Fatalf("expectation lost: %+v", back.Rows[0].Out)
	}
	again, err := yamlio.Marshal(back)
	if err != nil || string(again) != string(data) {
		t.Fatalf("re-marshal drifted: %q vs %q (%v)", again, data, err)
	}
}

// The loose fragment form: "(contains-loose) X" is the
// contains form carrying the executor's explicit mark that the
// fragment is short on purpose — decode, marshal and render keep the
// spelling.
func TestTextExpectLooseForm(t *testing.T) {
	decode := func(raw string) (*TextExpect, error) {
		var doc struct {
			Out *TextExpect `yaml:"out"`
		}
		err := yamlio.DecodeDoc([]byte(raw), &doc)
		return doc.Out, err
	}
	e, err := decode("out: \"(contains-loose) frag\"\n")
	if err != nil || e == nil || !e.Contains || !e.Loose || e.Text != "frag" {
		t.Fatalf("loose decode: %v %+v", err, e)
	}
	e2, err := decode("out: \"(contains) frag\"\n")
	if err != nil || e2.Loose {
		t.Fatalf("plain contains stays unmarked: %v %+v", err, e2)
	}
	back, err := e.MarshalYAML()
	if err != nil || back != exactTestScalar("(contains-loose) frag") {
		t.Fatalf("loose marshal: %v %v", back, err)
	}
	if got := (TextExpect{Contains: true, Loose: true, Text: "x"}).Form(); got != "(contains-loose) x" {
		t.Fatalf("loose form: %q", got)
	}
	if _, err := decode("out: \"(contains-loose) \"\n"); err == nil {
		t.Fatal("an empty loose fragment must refuse")
	}
}

// exactTestScalar — the wrapped scalar the marshal product carries.
func exactTestScalar(s string) any { return s }

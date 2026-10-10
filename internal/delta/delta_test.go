package delta

import (
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/canon"
)

// --- kind routing by stage ---

func TestKnownKind(t *testing.T) {
	for _, kind := range []string{KindSpec, KindRun} {
		if !KnownKind(kind) {
			t.Fatalf("%s must be a known kind", kind)
		}
	}
	for _, kind := range []string{"", "wish", "spec ", "RUN", "checks", "spek"} {
		if KnownKind(kind) {
			t.Fatalf("%q must not be a known kind", kind)
		}
	}
}

func TestKindAllowedOnStage(t *testing.T) {
	stages := []string{canon.StageIntake, canon.StageImplement, canon.StageDeliver}
	// expected allowance per kind over the stage list above
	want := map[string][]bool{
		KindSpec: {true, true, true},
		KindRun:  {false, true, false},
	}
	for kind, expected := range want {
		for i, stage := range stages {
			if got := KindAllowedOnStage(kind, stage); got != expected[i] {
				t.Fatalf("%s at %s: got %v, want %v", kind, stage, got, expected[i])
			}
		}
	}
	// an unknown kind is allowed nowhere; no stage list needs it
	for _, kind := range []string{"wish", "", "no-such-kind"} {
		for _, stage := range stages {
			if KindAllowedOnStage(kind, stage) {
				t.Fatalf("%q must be allowed nowhere", kind)
			}
		}
	}
	// a foreign stage allows nothing
	if KindAllowedOnStage(KindSpec, "review") || KindAllowedOnStage(KindRun, "") {
		t.Fatal("a foreign stage allows no kind")
	}
}

// --- decode of the one document shape ---

func TestDecodeSpecFull(t *testing.T) {
	raw := []byte(`kind: spec
build: [make, build]
build-ignore: [__pycache__, "*.pyc"]
behaviors:
  - id: B1
    name: chain-passthrough
    prose: |
      Bytes go through the chain as read.
    rows:
      - run: ["app", "relay", "a", "b"]
        seeds:
          - path: in.txt
            text: |
              first
              second
          - path: big.bin
            from-input: given.bin
        out: "(contains) first"
        err: "(contains) usage:"
        rc: 0
        state:
          - path: out.txt
            text: first
          - path: result.bin
            from-run: cmd-artifact.bin
        bytes: file://big-ref.bin
        timeout: 60
  - id: B2
    name: quiet-exit
    prose: |
      Silence on success.
    rows:
      - run: app --quiet
        out: ""
        rc: 0
`)
	d, err := Decode(raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if d.Kind != KindSpec {
		t.Fatalf("kind: %q", d.Kind)
	}
	if d.Ratify {
		t.Fatal("spec must not carry ratify from a document without it")
	}
	if len(d.Build) != 2 || d.Build[0] != "make" || d.Build[1] != "build" {
		t.Fatalf("build: %v", d.Build)
	}
	if len(d.BuildIgnore) != 2 || d.BuildIgnore[1] != "*.pyc" {
		t.Fatalf("build-ignore: %v", d.BuildIgnore)
	}
	if len(d.Behaviors) != 2 {
		t.Fatalf("behaviors: %d", len(d.Behaviors))
	}
	b1 := d.Behaviors[0]
	if b1.ID != "B1" || b1.Name != "chain-passthrough" || b1.Prose != "Bytes go through the chain as read.\n" {
		t.Fatalf("behavior head: %+v", b1)
	}
	if len(b1.Rows) != 1 {
		t.Fatalf("rows: %d", len(b1.Rows))
	}
	row := b1.Rows[0]
	if !row.Run.IsArgv || len(row.Run.Argv) != 4 || row.Run.Argv[1] != "relay" {
		t.Fatalf("run argv: %+v", row.Run)
	}
	if len(row.Seeds) != 2 {
		t.Fatalf("seeds: %d", len(row.Seeds))
	}
	if row.Seeds[0].Path != "in.txt" || row.Seeds[0].Text != "first\nsecond\n" || row.Seeds[0].FromInput != "" {
		t.Fatalf("text seed: %+v", row.Seeds[0])
	}
	if row.Seeds[1].Path != "big.bin" || row.Seeds[1].FromInput != "given.bin" || row.Seeds[1].Text != "" {
		t.Fatalf("from-input seed: %+v", row.Seeds[1])
	}
	if row.Out == nil || !row.Out.Contains || row.Out.Text != "first" {
		t.Fatalf("out: %+v", row.Out)
	}
	if row.Err == nil || !row.Err.Contains || row.Err.Text != "usage:" {
		t.Fatalf("err: %+v", row.Err)
	}
	if row.Rc == nil || *row.Rc != 0 {
		t.Fatalf("rc: %+v", row.Rc)
	}
	if len(row.State) != 2 {
		t.Fatalf("state: %d", len(row.State))
	}
	if row.State[0].Path != "out.txt" || row.State[0].Text == nil || row.State[0].Text.Text != "first" {
		t.Fatalf("state text: %+v", row.State[0])
	}
	if row.State[1].Path != "result.bin" || row.State[1].FromRun != "cmd-artifact.bin" {
		t.Fatalf("state from-run: %+v", row.State[1])
	}
	if row.Bytes != "file://big-ref.bin" {
		t.Fatalf("bytes: %q", row.Bytes)
	}
	if row.Timeout != 60 {
		t.Fatalf("timeout: %d", row.Timeout)
	}
	// the second behavior rides the line form
	if f, err := d.Behaviors[1].Rows[0].Run.Fields(); err != nil || len(f) != 2 {
		t.Fatalf("line form fields: %v %v", f, err)
	}
	if out := d.Behaviors[1].Rows[0].Out; out == nil || out.Contains || out.Text != "" {
		t.Fatalf("explicit empty out: %+v", out)
	}
}

func TestDecodeRun(t *testing.T) {
	raw := []byte("kind: run\nratify: true\n")
	d, err := Decode(raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if d.Kind != KindRun || !d.Ratify {
		t.Fatalf("run head: %+v", d)
	}
	// a bare re-run: no behaviors, no build — the document is the kind head
	raw = []byte("kind: run\n")
	d, err = Decode(raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(d.Behaviors) != 0 || d.Ratify {
		t.Fatalf("bare re-run: %+v", d)
	}
}

func TestDecodeAnswers(t *testing.T) {
	raw := []byte("kind: run\nanswers:\n  - q: Q1\n    choice: a\n  - q: Q2\n    choice: b\n")
	d, err := Decode(raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(d.Answers) != 2 || d.Answers[0].Q != "Q1" || d.Answers[1].Choice != "b" {
		t.Fatalf("answers: %+v", d.Answers)
	}
}

func TestDecodeMinimalSpec(t *testing.T) {
	d, err := Decode([]byte("kind: spec\n"))
	if err != nil {
		t.Fatalf("a bare kind head is decodable (form rules are the engine's): %v", err)
	}
	if d.Kind != KindSpec || d.Build != nil || d.Behaviors != nil {
		t.Fatalf("minimal spec: %+v", d)
	}
}

func TestDecodeBlockAndFlowSpellings(t *testing.T) {
	flow := []byte(`kind: spec {nope: 1}`)
	// flow mapping at the document head is refused by the document dialect
	if _, err := Decode(flow); err == nil {
		t.Fatal("the document dialect refuses flow-mapping heads")
	}
	block := []byte("kind: run\nbuild:\n  - make\n  - all\n")
	d, err := Decode(block)
	if err != nil {
		t.Fatalf("block build: %v", err)
	}
	if len(d.Build) != 2 || d.Build[0] != "make" {
		t.Fatalf("block build: %v", d.Build)
	}
}

// --- the refusal family ---

func TestDecodeRefusals(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string // a fragment the refusal must carry
	}{
		{"unknown kind far", "kind: zzzzzzzz\n", `"zzzzzzzz"`},
		{"unknown kind near miss", "kind: spek\n", `"spek"`},
		{"missing kind", "build: [make]\n", `""`},
		{"kind not a scalar", "kind:\n  x: 1\n", ""},
		{"foreign field", "kind: spec\nbogus: 1\n", ""},
		{"a sequence document", "- just\n- a list\n", ""},
		{"a scalar document", "kind\n", ""},
		{"a trailing second document", "kind: spec\n---\nkind: run\n", ""},
		{"duplicate key", "kind: spec\nkind: run\n", ""},
		{"wrong field type", "kind: spec\nbuild: make\n", ""},
		{"behaviors not a list", "kind: spec\nbehaviors: B1\n", ""},
		{"answers not a list", "kind: run\nanswers: Q1\n", ""},
		{"ratify not a bool", "kind: run\nratify: yes-please\n", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := Decode([]byte(tc.raw))
			if err == nil {
				t.Fatalf("must refuse: %+v", d)
			}
			if d != nil {
				t.Fatal("a refused decode hands back no document")
			}
			if tc.want != "" && !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("the refusal must carry %q: %v", tc.want, err)
			}
		})
	}
}

func TestDecodeUnknownKindHint(t *testing.T) {
	// a near miss carries the corrected hint — the only surviving hint
	_, err := Decode([]byte("kind: spek\n"))
	if err == nil {
		t.Fatal("must refuse")
	}
	if !strings.Contains(err.Error(), "kind: spec") {
		t.Fatalf("the near miss must suggest kind: spec: %v", err)
	}
	// the run near miss suggests run
	_, err = Decode([]byte("kind: runc\n"))
	if err == nil {
		t.Fatal("must refuse")
	}
	if !strings.Contains(err.Error(), "kind: run") {
		t.Fatalf("the near miss must suggest kind: run: %v", err)
	}
	// a far miss carries no suggestion text
	_, err = Decode([]byte("kind: zzzzzzzz\n"))
	if err == nil {
		t.Fatal("must refuse")
	}
	if strings.Contains(err.Error(), "e.g.") {
		t.Fatalf("a far miss carries no hint: %v", err)
	}
}

// --- kind suggestions ---

func TestSuggestKind(t *testing.T) {
	if _, ok := SuggestKind(""); ok {
		t.Fatal("an empty kind has no suggestion")
	}
	if _, ok := SuggestKind("spec"); ok {
		t.Fatal("a known kind needs no suggestion")
	}
	if _, ok := SuggestKind("  SPEC  "); ok {
		t.Fatal("a known kind in any spelling needs no suggestion")
	}
	if _, ok := SuggestKind(" Run\t"); ok {
		t.Fatal("a known kind with whitespace needs no suggestion")
	}
	hint, ok := SuggestKind("spek")
	if !ok || !strings.Contains(hint, "spec") {
		t.Fatalf("a near miss must suggest: %q %v", hint, ok)
	}
	hint, ok = SuggestKind("rn")
	if !ok || !strings.Contains(hint, "run") {
		t.Fatalf("a deletion near miss must suggest: %q %v", hint, ok)
	}
	if _, ok := SuggestKind("zzzzzzzz"); ok {
		t.Fatal("a far miss has no suggestion")
	}
	// equidistant to spec and run: ambiguous, no suggestion
	if _, ok := SuggestKind("spun"); ok {
		t.Fatal("an ambiguous near miss must not suggest")
	}
}

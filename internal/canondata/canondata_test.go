package canondata

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/neurophant/punchtape/internal/yamlio"
)

// freshState swaps every package registry for an empty one and
// restores the production init state when the test ends. Loader and
// validator units run over synthetic trees without touching the
// embedded delivery data.
func freshState(t *testing.T) {
	t.Helper()
	savedTexts, savedLimits := texts, limits
	savedTemplates := templates
	savedStages := stageList
	savedLint, savedEnv := lintRules, envDiagList
	savedHeal := healCatalog
	texts, limits = map[string]string{}, map[string]int{}
	templates = map[string][]byte{}
	stageList = nil
	lintRules, envDiagList = nil, nil
	healCatalog = nil
	t.Cleanup(func() {
		texts, limits = savedTexts, savedLimits
		templates = savedTemplates
		stageList = savedStages
		lintRules, envDiagList = savedLint, savedEnv
		healCatalog = savedHeal
	})
}

// mustPanic fails the test unless fn panics.
func mustPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s: expected a panic", name)
		}
	}()
	fn()
}

func mapFS(files map[string]string) fstest.MapFS {
	out := fstest.MapFS{}
	for name, content := range files {
		out["data/"+name] = &fstest.MapFile{Data: []byte(content)}
	}
	return out
}

// --- the served registries of the delivered tree ---

func TestTexts(t *testing.T) {
	if got := T("format.error.unknown-kind", M{"kind": `"spek"`}); !strings.Contains(got, "unknown kind") || !strings.Contains(got, "spek") {
		t.Fatalf("T with params: %q", got)
	}
	if got := T("cli.error.no-verb", M{"verb": `"wat"`}); !strings.Contains(got, "wat") {
		t.Fatalf("T substitutes named params: %q", got)
	}
	if T("format.error.unknown-kind") == T("format.error.ratify-not-run") {
		t.Fatal("distinct keys must render distinct texts")
	}
	mustPanic(t, "unknown key", func() { T("no.such.key") })

	// TFor: a language variant wins when present, the base key
	// otherwise; the protocol language never suffixes.
	freshState(t)
	texts["k"] = "base"
	texts["k.xx"] = "translated"
	if got := TFor("xx", "k"); got != "translated" {
		t.Fatalf("TFor language variant: %q", got)
	}
	if got := TFor("en", "k"); got != "base" {
		t.Fatalf("TFor en: %q", got)
	}
	if got := TFor("", "k"); got != "base" {
		t.Fatalf("TFor empty language: %q", got)
	}
	if got := TFor("yy", "k"); got != "base" {
		t.Fatalf("TFor without a variant: %q", got)
	}
	if got := TFor("xx", "k", M{"p": "v"}); got != "translated" {
		t.Fatalf("TFor passes params through: %q", got)
	}
}

func TestLimits(t *testing.T) {
	if got := Limit("check.timeout-default-s"); got != 60 {
		t.Fatalf("the row timeout default: %d", got)
	}
	if got := Limit("check.timeout-max-s"); got != 300 {
		t.Fatalf("the row timeout ceiling: %d", got)
	}
	mustPanic(t, "unknown limit", func() { Limit("no.such.limit") })
}

func TestCatalogGetters(t *testing.T) {
	if got := Heal("idempotent-resubmit"); got == "" {
		t.Fatal("Heal must find the known detection")
	}
	if got := Heal("no-such-heal"); got != "" {
		t.Fatalf("Heal absence must be an honest empty reply: %q", got)
	}
	if _, ok := LintRuleByID("fixture.path-relative"); !ok {
		t.Fatal("LintRuleByID must find the known rule")
	}
	if _, ok := LintRuleByID("no.such.rule"); ok {
		t.Fatal("LintRuleByID must report absence")
	}
}

func TestLintMessage(t *testing.T) {
	got := LintMessage("fixture.path-relative", M{"path": "x/y"})
	if !strings.Contains(got, "x/y") {
		t.Fatalf("LintMessage renders the registry text with params: %q", got)
	}
	mustPanic(t, "unknown lint rule", func() { LintMessage("no.such.rule") })
}

func TestEnvDiag(t *testing.T) {
	got := EnvDiag("exec-not-found", M{"command": "tool"})
	if !strings.Contains(got, "tool") || !strings.Contains(got, "FIX:") {
		t.Fatalf("EnvDiag: %q", got)
	}
	got = EnvDiag("command-timeout", M{"sec": "1", "command": "sleep 5"})
	if !strings.Contains(got, "timed out after 1s") {
		t.Fatalf("EnvDiag substitutes every param: %q", got)
	}
	mustPanic(t, "unknown env diagnosis", func() { EnvDiag("no.such.diag") })
}

func TestStages(t *testing.T) {
	ids := StageIDs()
	want := []string{"intake", "implement", "deliver"}
	if len(ids) != len(want) {
		t.Fatalf("stage ids: %v", ids)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("stage ids: %v want %v", ids, want)
		}
	}
	byID := StagesByID()
	if len(byID) != len(ids) {
		t.Fatal("stage maps must agree in size")
	}
	for _, id := range ids {
		if _, ok := byID[id]; !ok {
			t.Fatalf("stage %s missing from the map", id)
		}
	}
	// deliver reopens on the next spec version: the exit gate is the
	// reopen check, the goto is the work stage (the living case)
	if byID["deliver"].Exit.Gate != "spec-resubmitted" || byID["deliver"].Exit.Goto != "implement" {
		t.Fatalf("deliver must reopen on a resubmitted spec: %+v", byID["deliver"].Exit)
	}
	// every non-terminal exit leads to a known stage
	for _, id := range ids {
		s := byID[id]
		if s.Exit.Gate == "" {
			continue
		}
		if _, ok := byID[s.Exit.Goto]; !ok {
			t.Fatalf("stage exit %s: unknown target %s", s.ID, s.Exit.Goto)
		}
	}
}

// --- synthetic trees: the loader's every refusal ---

func TestLoadFSBrokenTrees(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
	}{
		{"no data dir", nil},
		{"broken limits", map[string]string{"limits.yaml": "{"}},
		{"broken stages", map[string]string{"stages.yaml": "{"}},
		{"broken lint", map[string]string{"lint.yaml": "{"}},
		{"broken envdiag", map[string]string{"envdiag.yaml": "{"}},
		{"broken heal", map[string]string{"heal.yaml": "{"}},
		{"unknown top-level part", map[string]string{"custom.yaml": "a: 1"}},
		{"a stray format among texts", map[string]string{"texts/a.zip": "a: 1"}},
		{"empty namespace", map[string]string{"texts/.yaml": "key: v"}},
		{"a zone outside ref", map[string]string{"texts/ns/leaf.md": "text"}},
		{"a dictionary under ref", map[string]string{"texts/ref/doc/one.yaml": "a: 1"}},
		{"a non-md document under ref", map[string]string{"texts/ref/doc/one.txt": "doc"}},
		{"document key path with double dots", map[string]string{"texts/ref/a..b.md": "doc"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			freshState(t)
			if err := loadFS(mapFS(tc.files)); err == nil {
				t.Fatalf("expected a load error for %v", tc.files)
			}
		})
	}

	t.Run("minimal clean tree", func(t *testing.T) {
		freshState(t)
		files := map[string]string{
			"texts/ns.yaml":     "key: |2-\n  text\n",
			"templates/one.txt": "tpl-bytes",
		}
		if err := loadFS(mapFS(files)); err != nil {
			t.Fatalf("a minimal consistent tree must load: %v", err)
		}
		if T("ns.key") != "text" {
			t.Fatalf("the segment must load byte-exact: %q", T("ns.key"))
		}
		if string(TemplateBytes("one.txt")) != "tpl-bytes" {
			t.Fatalf("the template must load byte-exact: %q", TemplateBytes("one.txt"))
		}
	})

	t.Run("dotted keys carry locale segments", func(t *testing.T) {
		freshState(t)
		files := map[string]string{"texts/ns.yaml": "leaf.ru: |2-\n  перевод\n"}
		if err := loadFS(mapFS(files)); err != nil {
			t.Fatalf("a dictionary with a dotted key must load: %v", err)
		}
		if T("ns.leaf.ru") != "перевод" {
			t.Fatalf("the dotted key must resolve: %q", T("ns.leaf.ru"))
		}
	})

	t.Run("dictionary faults are load errors", func(t *testing.T) {
		freshState(t)
		for name, body := range map[string]string{
			"key outside the alphabet":   "Bad_Key: x",
			"empty segment":              "key: \"\"",
			"duplicate key":              "a: one\na: two",
			"a list value in the dict":   "a: [1]",
			"not a dictionary":           "- just\n- a list",
			"a scalar document":          "x",
			"an anchor in the dict":      "a: &x v",
			"a trailing second document": "a: x\n---\nb: y",
		} {
			if err := loadFS(mapFS(map[string]string{"texts/ns.yaml": body})); err == nil {
				t.Fatalf("%s must refuse", name)
			}
		}
		// the edge law rides the decoder: a block's trailing newline
		// is the separator, not content
		if err := loadFS(mapFS(map[string]string{"texts/ns.yaml": "key: |\n  text\n"})); err != nil {
			t.Fatalf("the edge-newline form loads: %v", err)
		}
		if T("ns.key") != "text" {
			t.Fatalf("the edge-trimmed segment: %q", T("ns.key"))
		}
	})

	t.Run("an empty dictionary is a legal state", func(t *testing.T) {
		freshState(t)
		if err := loadFS(mapFS(map[string]string{"texts/ns.yaml": "{}"})); err != nil {
			t.Fatalf("an empty dictionary loads empty: %v", err)
		}
	})

	t.Run("documents load under ref only", func(t *testing.T) {
		freshState(t)
		if err := loadFS(mapFS(map[string]string{"texts/ref/doc/one.md": "doc"})); err != nil {
			t.Fatalf("a document under ref/ must load: %v", err)
		}
		if T("ref.doc.one") != "doc" {
			t.Fatalf("the document bytes must be the text: %q", T("ref.doc.one"))
		}
	})

	t.Run("TemplateSet answers a prefix slice", func(t *testing.T) {
		freshState(t)
		files := map[string]string{
			"texts/ns.yaml":                 "key: |2-\n  text\n",
			"templates/kind.yaml":           "kind: wish",
			"templates/kind.files/stub.txt": "hint",
		}
		if err := loadFS(mapFS(files)); err != nil {
			t.Fatalf("the tree loads: %v", err)
		}
		set := TemplateSet("kind.files/")
		if len(set) != 1 || string(set["kind.files/stub.txt"]) != "hint" {
			t.Fatalf("the prefix slice: %+v", set)
		}
		if len(TemplateSet("nope/")) != 0 {
			t.Fatal("a foreign prefix answers empty")
		}
		mustPanic(t, "unknown template", func() { TemplateBytes("no-such.yaml") })
	})

	t.Run("missing text and template roots load empty", func(t *testing.T) {
		freshState(t)
		if err := loadFS(mapFS(map[string]string{"limits.yaml": `{}`})); err != nil {
			t.Fatalf("an empty tree is a legal state: %v", err)
		}
	})

	t.Run("mustLoad panics on a broken tree", func(t *testing.T) {
		freshState(t)
		mustPanic(t, "mustLoad", func() { mustLoad(mapFS(map[string]string{"limits.yaml": "{"})) })
	})
}

// failFS — a synthetic tree whose reads fail at named paths: the
// loader's read-error arms (an I/O fault mid-load) are reachable
// only over a filesystem that can actually fail.
type failFS struct {
	fs.FS
	failRead map[string]bool
	failDir  string
}

func (f failFS) ReadFile(name string) ([]byte, error) {
	if f.failRead[name] {
		return nil, fs.ErrInvalid
	}
	return fs.ReadFile(f.FS, name)
}

func (f failFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name == f.failDir {
		return nil, fs.ErrInvalid
	}
	return fs.ReadDir(f.FS, name)
}

// failOn wraps a map tree with read faults at the named data paths.
func failOn(files map[string]string, failRead []string, failDir string) fs.FS {
	set := map[string]bool{}
	for _, name := range failRead {
		set["data/"+name] = true
	}
	return failFS{FS: mapFS(files), failRead: set, failDir: "data/" + failDir}
}

func TestLoadFSReadFaults(t *testing.T) {
	base := func() map[string]string {
		return map[string]string{
			"limits.yaml":       `{}`,
			"texts/a.yaml":      "x: |2-\n  text\n",
			"texts/ref/d.md":    "doc",
			"templates/one.txt": "tpl",
		}
	}
	cases := []struct {
		name     string
		failRead []string
		failDir  string
	}{
		{"top-level part unreadable", []string{"limits.yaml"}, ""},
		{"text file unreadable", []string{"texts/a.yaml"}, ""},
		{"ref document unreadable", []string{"texts/ref/d.md"}, ""},
		{"template file unreadable", []string{"templates/one.txt"}, ""},
		{"text root unreadable", nil, "texts"},
		{"template root unreadable", nil, "templates"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			freshState(t)
			if err := loadFS(failOn(base(), tc.failRead, tc.failDir)); err == nil {
				t.Fatal("a read fault must surface as a load error")
			}
		})
	}
}

// --- the cross-registry invariants ---

func TestValidateStructuralUnits(t *testing.T) {
	cases := []struct {
		name   string
		mutate func()
	}{
		{"duplicate stage", func() {
			stageList = []Stage{{ID: "s1"}, {ID: "s1"}}
		}},
		{"stage exit to an unknown stage", func() {
			stageList = []Stage{{ID: "s1", Exit: StageExit{Gate: "g", Goto: "zz"}}}
		}},
		{"duplicate lint rule", func() {
			lintRules = []LintRule{{ID: "r"}, {ID: "r"}}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			freshState(t)
			tc.mutate()
			if err := validateStructuralErr(); err == nil {
				t.Fatal("expected a structural error")
			}
			mustPanic(t, tc.name, validateStructural)
		})
	}
	t.Run("clean registries", func(t *testing.T) {
		freshState(t)
		stageList = []Stage{
			{ID: "s1", Exit: StageExit{Gate: "g", Goto: "s1"}},
			{ID: "s2", Exit: StageExit{}}, // a terminal stage is legal
		}
		lintRules = []LintRule{{ID: "r"}}
		if err := validateStructuralErr(); err != nil {
			t.Fatalf("consistent registries must pass: %v", err)
		}
		validateStructural() // the panic wrapper on a healthy state
	})
}

// --- the delivery tree is the truth ---

// TestEmbeddedTreeIsTheTruth — the permanent invariant of the
// container: every dictionary segment renders byte for byte as parsed
// from the embedded file (T(key) is the segment, edges clean), every
// ref/ document is the file bytes verbatim, and every template is the
// embedded bytes. No loader transformation may ever stand between the
// source tree and T().
func TestEmbeddedTreeIsTheTruth(t *testing.T) {
	segs, docs := 0, 0
	err := fs.WalkDir(dataFS, "data/texts", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, rerr := fs.ReadFile(dataFS, p)
		if rerr != nil {
			return rerr
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, "data/texts"), "/")
		if strings.Contains(rel, "/") {
			// a document: the file bytes are the text
			key := strings.ReplaceAll(strings.TrimSuffix(rel, ".md"), "/", ".")
			if T(key) != string(raw) {
				t.Fatalf("document %s: T() is not the file bytes", rel)
			}
			docs++
			return nil
		}
		ns := strings.TrimSuffix(rel, ".yaml")
		dict, derr := yamlio.DecodeDict(raw)
		if derr != nil {
			t.Fatalf("dictionary %s: %v", rel, derr)
		}
		for key, val := range dict {
			if T(ns+"."+key) != val {
				t.Fatalf("segment %s.%s: T() is not the parsed bytes", ns, key)
			}
			if strings.HasPrefix(val, "\n") || strings.HasSuffix(val, "\n") {
				t.Fatalf("segment %s.%s: an edge newline", ns, key)
			}
			segs++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if segs < 150 {
		t.Fatalf("the delivery dictionary is unexpectedly small: %d segments", segs)
	}
	if docs < 5 {
		t.Fatalf("the delivery documents are unexpectedly few: %d files", docs)
	}

	t.Run("locale param parity", func(t *testing.T) {
		// A locale variant addressing a parameter the caller does not
		// pass prints the literal brace token: the variant's parameter
		// set must equal its base's — no extra, none dropped.
		params := func(s string) map[string]bool {
			out := map[string]bool{}
			for _, part := range strings.Split(s, "{") {
				if i := strings.Index(part, "}"); i > 0 {
					out[part[:i]] = true
				}
			}
			return out
		}
		all := map[string]map[string]bool{}
		err := fs.WalkDir(dataFS, "data/texts", func(p string, d fs.DirEntry, werr error) error {
			if werr != nil || d.IsDir() || strings.Contains(strings.TrimPrefix(p, "data/texts/"), "/") {
				return werr
			}
			raw, rerr := fs.ReadFile(dataFS, p)
			if rerr != nil {
				return rerr
			}
			ns := strings.TrimSuffix(filepath.Base(p), ".yaml")
			dict, derr := yamlio.DecodeDict(raw)
			if derr != nil {
				return derr
			}
			for key, val := range dict {
				all[ns+"."+key] = params(val)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk: %v", err)
		}
		for key, want := range all {
			for _, loc := range []string{".ru", ".zh"} {
				if !strings.HasSuffix(key, loc) {
					continue
				}
				got, ok := all[strings.TrimSuffix(key, loc)]
				if !ok {
					t.Fatalf("locale %s has no base segment", key)
				}
				for name := range got {
					if !want[name] {
						t.Fatalf("locale %s drops the base parameter {%s}", key, name)
					}
				}
				for name := range want {
					if !got[name] {
						t.Fatalf("locale %s addresses {%s} the base does not carry", key, name)
					}
				}
			}
		}
	})

	tcount := 0
	err = fs.WalkDir(dataFS, "data/templates", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, "data/templates"), "/")
		raw, err := fs.ReadFile(dataFS, p)
		if err != nil {
			return err
		}
		if string(TemplateBytes(rel)) != string(raw) {
			t.Fatalf("template %s: not the embedded bytes", rel)
		}
		tcount++
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if tcount == 0 {
		t.Fatal("the delivery template tree is empty")
	}
	mustPanic(t, "unknown template", func() { TemplateBytes("no-such.yaml") })
}

// TestDiskTreeIsTheEmbeddedTree — the source tree on disk and the
// embedded copy are one truth: every file under data/ on disk exists
// in the embed with the same bytes, and vice versa. A file the embed
// skips (a hidden or underscored name) or a stale build is caught.
func TestDiskTreeIsTheEmbeddedTree(t *testing.T) {
	// The data tree by its source location: the coverage battery runs
	// test binaries from a scratch directory, not the package dir.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("no source location for the data tree")
	}
	dataDir := filepath.Join(filepath.Dir(thisFile), "data")
	disk := map[string]string{}
	err := filepath.WalkDir(dataDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		rel, rerr := filepath.Rel(dataDir, p)
		if rerr != nil {
			return rerr
		}
		disk["data/"+filepath.ToSlash(rel)] = string(raw)
		return nil
	})
	if err != nil {
		t.Fatalf("disk walk: %v", err)
	}
	if len(disk) == 0 {
		t.Fatal("the on-disk data tree is empty — wrong working directory?")
	}
	embedded := map[string]string{}
	err = fs.WalkDir(dataFS, "data", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, rerr := fs.ReadFile(dataFS, p)
		if rerr != nil {
			return rerr
		}
		embedded[filepath.ToSlash(p)] = string(raw)
		return nil
	})
	if err != nil {
		t.Fatalf("embed walk: %v", err)
	}
	for p, want := range disk {
		got, ok := embedded[p]
		if !ok {
			t.Fatalf("%s exists on disk but not in the embed", p)
		}
		if got != want {
			t.Fatalf("%s: disk and embedded bytes differ", p)
		}
	}
	for p := range embedded {
		if _, ok := disk[p]; !ok {
			t.Fatalf("%s exists in the embed but not on disk", p)
		}
	}
}

func TestSubstituteNoParams(t *testing.T) {
	if got := substitute("plain {x} text"); got != "plain {x} text" {
		t.Fatalf("no-params substitution changed the text: %q", got)
	}
}

// The negative-obligation witness canon: the
// grammar document carries the canonical witness shapes - the
// machine's own tree gate for write prohibitions, the traced-empty
// pattern for network and read scope - so a "must not happen" of a
// brief always has a copyable row form.
func TestRefGrammarWitnessCanon(t *testing.T) {
	g := T("ref.doc.grammar")
	for _, marker := range []string{
		"Negative obligations (witness rows)",
		"the machine's own gate is the witness",
		"trace=network",
		"trace=openat",
		"unshare -n",
	} {
		if !strings.Contains(g, marker) {
			t.Fatalf("the grammar doc lost the witness canon marker %q", marker)
		}
	}
}

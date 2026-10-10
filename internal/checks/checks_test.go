package checks

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/execport"
	"github.com/neurophant/punchtape/internal/fsport"
)

// The checks unit runs over the real filesystem and the real process
// port: the row is a black box executed by the operating system, and
// the honest test of that is the operating system itself. Commands are
// small /bin/sh scripts; the tree is a fresh t.TempDir() per test.

// shRow — one row whose command is a /bin/sh script.
func shRow(script string) canon.Row {
	return canon.Row{Run: canon.Cmd{Argv: []string{"sh", "-c", script}, IsArgv: true}}
}

// wantOut — a row with an exact stdout expectation.
func wantOut(script, out string) canon.Row {
	row := shRow(script)
	row.Out = &canon.TextExpect{Text: out}
	return row
}

func mustWrite(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil { // no umask surprises
		t.Fatal(err)
	}
}

func runRow(t *testing.T, row canon.Row, env RunEnv) Result {
	t.Helper()
	return Run(t.TempDir(), "T/1", row, env)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// --- green rows ---

func TestRunGreen(t *testing.T) {
	cases := []struct {
		name string
		row  canon.Row
		env  RunEnv
	}{
		{"out exact", wantOut("echo hi", "hi"), RunEnv{}},
		{"out exact with own newline", wantOut("echo hi", "hi\n"), RunEnv{}},
		{"out EOL tolerance want bare", wantOut("printf hi", "hi"), RunEnv{}},
		{"out EOL tolerance want newline", wantOut("printf hi", "hi\n"), RunEnv{}},
		{"out empty stream", wantOut("true", ""), RunEnv{}},
		{"out contains", shRow("echo hello fragment world"), RunEnv{}},
		{"line form command", canon.Row{Run: canon.Cmd{Line: "echo line-form"}}, RunEnv{}},
	}
	cases[5].row.Out = &canon.TextExpect{Contains: true, Text: "fragment"}
	cases[6].row.Out = &canon.TextExpect{Text: "line-form"}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := runRow(t, tc.row, tc.env)
			if !res.Green {
				t.Fatalf("must be green: %s", res.Reason)
			}
			if res.Env {
				t.Fatal("a green row is never an env failure")
			}
			if res.Label != "T/1" || res.Exit == nil || *res.Exit != 0 || res.OutDigest == "" {
				t.Fatalf("the measured facts must carry the exit and the output digest: %+v", res)
			}
		})
	}
}

func TestRunGreenStreamsAndRc(t *testing.T) {
	// stderr exact and contains
	row := shRow("echo noise-line 1>&2; echo boom 1>&2")
	row.Err = &canon.TextExpect{Contains: true, Text: "boom"}
	res := runRow(t, row, RunEnv{})
	if !res.Green {
		t.Fatalf("err contains: %s", res.Reason)
	}

	row = shRow("echo boom 1>&2")
	row.Err = &canon.TextExpect{Text: "boom"}
	if res = runRow(t, row, RunEnv{}); !res.Green {
		t.Fatalf("err exact: %s", res.Reason)
	}

	// a non-zero rc is a green row when declared
	row = shRow("exit 3")
	rc := 3
	row.Rc = &rc
	if res = runRow(t, row, RunEnv{}); !res.Green {
		t.Fatalf("declared rc: %s", res.Reason)
	}
	if res.Exit == nil || *res.Exit != 3 {
		t.Fatalf("the fact carries the measured rc: %+v", res)
	}
}

func TestRunGreenState(t *testing.T) {
	dir := t.TempDir()
	row := shRow("echo hello > out.txt")
	row.State = []canon.StateWant{{Path: "out.txt", Text: &canon.TextExpect{Text: "hello"}}}
	if res := Run(dir, "S/1", row, RunEnv{}); !res.Green {
		t.Fatalf("state text: %s", res.Reason)
	}
	// contains form rides the state text too
	row.State = []canon.StateWant{{Path: "out.txt", Text: &canon.TextExpect{Contains: true, Text: "ell"}}}
	if res := Run(dir, "S/2", row, RunEnv{}); !res.Green {
		t.Fatalf("state contains: %s", res.Reason)
	}

	// from-run: byte equality with a file the run produced
	dir = t.TempDir()
	row = shRow("printf same > a.bin; printf same > b.bin")
	row.State = []canon.StateWant{{Path: "a.bin", FromRun: "b.bin"}}
	if res := Run(dir, "S/3", row, RunEnv{}); !res.Green {
		t.Fatalf("state from-run: %s", res.Reason)
	}
}

func TestRunGreenBytesSample(t *testing.T) {
	inbox := t.TempDir()
	mustWrite(t, filepath.Join(inbox, "sample.bin"), "bytes-out\n", 0o644)
	row := shRow("echo bytes-out")
	row.Bytes = "file://sample.bin"
	res := Run(t.TempDir(), "B/1", row, RunEnv{InboxDir: inbox})
	if !res.Green {
		t.Fatalf("bytes sample: %s", res.Reason)
	}
	// the sample wins over the exact form only when both hold; here the
	// exact form diverges and the row is red
	row.Out = &canon.TextExpect{Text: "not-the-sample"}
	if res = Run(t.TempDir(), "B/2", row, RunEnv{InboxDir: inbox}); res.Green {
		t.Fatal("a diverging exact form stays red")
	}
}

// --- argv resolution over the live tree ---

func TestLaunchForms(t *testing.T) {
	t.Run("bare non-executable tree file gets the launch grant and it folds back", func(t *testing.T) {
		dir := t.TempDir()
		tool := filepath.Join(dir, "tool")
		mustWrite(t, tool, "#!/bin/sh\necho tool-ran\n", 0o644)
		row := canon.Row{Run: canon.Cmd{Line: "tool"}}
		row.Out = &canon.TextExpect{Text: "tool-ran"}
		res := Run(dir, "L/1", row, RunEnv{})
		if !res.Green {
			t.Fatalf("grant launch: %s", res.Reason)
		}
		info, err := os.Stat(tool)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o644 {
			t.Fatalf("the owner's mode must return: %v", info.Mode().Perm())
		}
	})

	t.Run("bare executable tree file launches without a grant", func(t *testing.T) {
		dir := t.TempDir()
		tool := filepath.Join(dir, "tool")
		mustWrite(t, tool, "#!/bin/sh\necho already-exec\n", 0o755)
		row := canon.Row{Run: canon.Cmd{Line: "tool"}}
		row.Out = &canon.TextExpect{Text: "already-exec"}
		if res := Run(dir, "L/2", row, RunEnv{}); !res.Green {
			t.Fatalf("executable launch: %s", res.Reason)
		}
		info, err := os.Stat(tool)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o755 {
			t.Fatalf("the mode must not move: %v", info.Mode().Perm())
		}
	})

	t.Run("a name with a slash launches as is", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
			t.Fatal(err)
		}
		mustWrite(t, filepath.Join(dir, "sub", "tool"), "#!/bin/sh\necho from-sub\n", 0o755)
		row := canon.Row{Run: canon.Cmd{Line: "./sub/tool"}}
		row.Out = &canon.TextExpect{Text: "from-sub"}
		if res := Run(dir, "L/3", row, RunEnv{}); !res.Green {
			t.Fatalf("slash launch: %s", res.Reason)
		}
	})

	t.Run("a bare name absent from the tree goes to PATH", func(t *testing.T) {
		row := canon.Row{Run: canon.Cmd{Line: "echo path-launch"}}
		row.Out = &canon.TextExpect{Text: "path-launch"}
		if res := runRow(t, row, RunEnv{}); !res.Green {
			t.Fatalf("PATH launch: %s", res.Reason)
		}
	})

	t.Run("a directory in the tree does not capture a bare name", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "echo"), 0o755); err != nil {
			t.Fatal(err)
		}
		row := canon.Row{Run: canon.Cmd{Line: "echo still-path"}}
		row.Out = &canon.TextExpect{Text: "still-path"}
		if res := Run(dir, "L/4", row, RunEnv{}); !res.Green {
			t.Fatalf("a directory is not an entry point: %s", res.Reason)
		}
	})
}

// --- red rows (expectation divergences) ---

func TestRunReds(t *testing.T) {
	cases := []struct {
		name  string
		row   canon.Row
		env   RunEnv
		marks []string
	}{
		{
			"rc mismatch", shRow("exit 1"), RunEnv{},
			[]string{"exit code", "expected 0"},
		},
		{
			"out not equal", wantOut("echo aa", "bb"), RunEnv{},
			[]string{"not equal to", "stdout"},
		},
		{
			"two trailing newlines diverge with a repair proposal", wantOut("echo hi", "hi\n\n"), RunEnv{},
			[]string{"not equal to", "measured value for the repair"},
		},
		{
			"contains missing", func() canon.Row {
				r := shRow("echo other")
				r.Out = &canon.TextExpect{Contains: true, Text: "fragment"}
				return r
			}(), RunEnv{},
			[]string{"does not contain"},
		},
		{
			"err diverges", func() canon.Row {
				r := shRow("echo noise 1>&2")
				r.Err = &canon.TextExpect{Text: "other-noise"}
				return r
			}(), RunEnv{},
			[]string{"stderr", "not equal to"},
		},
		{
			"state file missing", func() canon.Row {
				r := shRow("true")
				r.State = []canon.StateWant{{Path: "gone.txt", Text: &canon.TextExpect{Text: "x"}}}
				return r
			}(), RunEnv{},
			[]string{"missing after the row"},
		},
		{
			"state text not equal", func() canon.Row {
				r := shRow("echo real > out.txt")
				r.State = []canon.StateWant{{Path: "out.txt", Text: &canon.TextExpect{Text: "wanted"}}}
				return r
			}(), RunEnv{},
			[]string{"not equal"},
		},
		{
			"state from-run source missing", func() canon.Row {
				r := shRow("true")
				r.State = []canon.StateWant{{Path: "a.bin", FromRun: "src.bin"}}
				return r
			}(), RunEnv{},
			[]string{"missing after the row", "src.bin"},
		},
		{
			"state from-run target missing", func() canon.Row {
				r := shRow("printf x > src.bin")
				r.State = []canon.StateWant{{Path: "a.bin", FromRun: "src.bin"}}
				return r
			}(), RunEnv{},
			[]string{"missing after the row", "a.bin"},
		},
		{
			"state from-run bytes differ", func() canon.Row {
				r := shRow("printf aaa > a.bin; printf bbb > b.bin")
				r.State = []canon.StateWant{{Path: "a.bin", FromRun: "b.bin"}}
				return r
			}(), RunEnv{},
			[]string{"differs from"},
		},
		{
			"explicit empty against real output", wantOut("echo x", ""), RunEnv{},
			[]string{"not equal to", "<empty>"},
		},
		{
			"contains missing against an empty stream", func() canon.Row {
				r := shRow("true")
				r.Out = &canon.TextExpect{Contains: true, Text: "fragment"}
				return r
			}(), RunEnv{},
			[]string{"does not contain", "<empty>"},
		},
		{
			"bytes sample missing", func() canon.Row {
				r := wantOut("echo x", "x")
				r.Bytes = "file://vanish.bin"
				return r
			}(), RunEnv{InboxDir: t.TempDir()},
			[]string{"vanished from the inbox"},
		},
		{
			"bytes sample differs", func() canon.Row {
				r := wantOut("echo other", "other")
				r.Bytes = "file://sample.bin"
				return r
			}(), RunEnv{InboxDir: t.TempDir()},
			nil, // the inbox file is written below
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := tc.env
			if tc.name == "bytes sample differs" {
				mustWrite(t, filepath.Join(env.InboxDir, "sample.bin"), "sample\n", 0o644)
			}
			res := runRow(t, tc.row, env)
			if res.Green {
				t.Fatal("must be red")
			}
			if res.Env {
				t.Fatal("an expectation divergence is not an env failure")
			}
			if res.Reason == "" {
				t.Fatal("a red row carries a reason line")
			}
			if strings.Contains(res.Reason, "\n") {
				t.Fatalf("the reason must stay one line: %q", res.Reason)
			}
			for _, mark := range tc.marks {
				if !strings.Contains(res.Reason, mark) {
					t.Fatalf("the reason must carry %q: %q", mark, res.Reason)
				}
			}
		})
	}
}

func TestRunRedLongOutputStaysOneLine(t *testing.T) {
	row := wantOut("yes xxxxxxxxxxxxxxxxxxxxxxxxxxx | head -c 400", "nope")
	res := runRow(t, row, RunEnv{})
	if res.Green {
		t.Fatal("must be red")
	}
	if strings.Contains(res.Reason, "\n") {
		t.Fatalf("the reason must stay one line over long output: %q", res.Reason)
	}
	if !strings.HasSuffix(res.Reason, "...\"") && !strings.Contains(res.Reason, "...") {
		t.Fatalf("long output must be cut with an ellipsis: %q", res.Reason)
	}
}

// --- environment failures ---

func TestRunEnvFailures(t *testing.T) {
	cases := []struct {
		name  string
		row   canon.Row
		env   RunEnv
		marks []string
	}{
		{
			"command not installed", canon.Row{Run: canon.Cmd{Line: "no-such-tool-xyz"}}, RunEnv{},
			[]string{"not installed"},
		},
		{
			"slash launch without the exec bit", canon.Row{Run: canon.Cmd{Line: "./plain"}}, RunEnv{},
			[]string{"no permission to execute"},
		},
		{
			"not an executable format", canon.Row{Run: canon.Cmd{Line: "./not-a-script"}}, RunEnv{},
			[]string{"did not start"},
		},
		{
			"unbalanced quote in the line", canon.Row{Run: canon.Cmd{Line: `echo "hi`}}, RunEnv{},
			[]string{"unbalanced quote"},
		},
		{
			"empty argv", canon.Row{Run: canon.Cmd{Argv: []string{}, IsArgv: true}}, RunEnv{},
			[]string{"argv list is empty"},
		},
		{
			"from-input seed missing", func() canon.Row {
				r := shRow("cat big.bin")
				r.Seeds = []canon.Seed{{Path: "big.bin", FromInput: "given.bin"}}
				return r
			}(), RunEnv{InputDir: t.TempDir()},
			[]string{"from-input"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := tc.env
			if tc.name == "slash launch without the exec bit" || tc.name == "not an executable format" {
				dir := t.TempDir()
				mustWrite(t, filepath.Join(dir, "plain"), "#!/bin/sh\necho x\n", 0o644)
				mustWrite(t, filepath.Join(dir, "not-a-script"), "no shebang, no binary\n", 0o755)
				res := Run(dir, "E/1", tc.row, env)
				inspectEnvRed(t, res, tc.marks)
				return
			}
			inspectEnvRed(t, runRow(t, tc.row, env), tc.marks)
		})
	}
}

func inspectEnvRed(t *testing.T, res Result, marks []string) {
	t.Helper()
	if res.Green {
		t.Fatal("must be red")
	}
	if !res.Env {
		t.Fatalf("must be an env failure: %q", res.Reason)
	}
	for _, mark := range marks {
		if !strings.Contains(res.Reason, mark) {
			t.Fatalf("the reason must carry %q: %q", mark, res.Reason)
		}
	}
}

// --- seeds ---

func TestSeeds(t *testing.T) {
	t.Run("a text seed lives only for the row", func(t *testing.T) {
		dir := t.TempDir()
		row := shRow("cat in.txt")
		row.Seeds = []canon.Seed{{Path: "in.txt", Text: "seed-body\n"}}
		row.Out = &canon.TextExpect{Text: "seed-body"}
		if res := Run(dir, "SD/1", row, RunEnv{}); !res.Green {
			t.Fatalf("seeded row: %s", res.Reason)
		}
		if _, err := os.Stat(filepath.Join(dir, "in.txt")); !os.IsNotExist(err) {
			t.Fatalf("the seed must self-clean: %v", err)
		}
	})

	t.Run("a seed over a tree file restores the original", func(t *testing.T) {
		dir := t.TempDir()
		keep := filepath.Join(dir, "keep.txt")
		mustWrite(t, keep, "original\n", 0o644)
		row := shRow("cat keep.txt")
		row.Seeds = []canon.Seed{{Path: "keep.txt", Text: "seeded\n"}}
		row.Out = &canon.TextExpect{Text: "seeded"}
		if res := Run(dir, "SD/2", row, RunEnv{}); !res.Green {
			t.Fatalf("seed over tree file: %s", res.Reason)
		}
		if got := readFile(t, keep); got != "original\n" {
			t.Fatalf("the tree file must return: %q", got)
		}
		info, err := os.Stat(keep)
		if err != nil || info.Mode().Perm() != 0o644 {
			t.Fatalf("a rewrite never moves permissions: %v %v", info, err)
		}
	})

	t.Run("a seed in a fresh nested directory self-cleans", func(t *testing.T) {
		dir := t.TempDir()
		row := shRow("cat deep/nest/f.txt")
		row.Seeds = []canon.Seed{{Path: "deep/nest/f.txt", Text: "nested\n"}}
		row.Out = &canon.TextExpect{Text: "nested"}
		if res := Run(dir, "SD/3", row, RunEnv{}); !res.Green {
			t.Fatalf("nested seed: %s", res.Reason)
		}
		if _, err := os.Stat(filepath.Join(dir, "deep", "nest", "f.txt")); !os.IsNotExist(err) {
			t.Fatalf("the nested seed must self-clean: %v", err)
		}
	})

	t.Run("the machine-created directories vanish with the seed", func(t *testing.T) {
		dir := t.TempDir()
		row := shRow("cat rec/one/meta.yaml")
		row.Seeds = []canon.Seed{{Path: "rec/one/meta.yaml", Text: "m: 1\n"}}
		row.Out = &canon.TextExpect{Text: "m: 1"}
		if res := Run(dir, "SD/6", row, RunEnv{}); !res.Green {
			t.Fatalf("dir-seeded row: %s", res.Reason)
		}
		for _, d := range []string{filepath.Join(dir, "rec", "one"), filepath.Join(dir, "rec")} {
			if _, err := os.Stat(d); !os.IsNotExist(err) {
				t.Fatalf("the seed's directory must self-clean with the file: %v", err)
			}
		}
	})

	t.Run("a directory the tree already had survives the seed", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "rec"), 0o755); err != nil {
			t.Fatal(err)
		}
		mustWrite(t, filepath.Join(dir, "rec", "keep.txt"), "keep\n", 0o644)
		row := shRow("cat rec/s.txt")
		row.Seeds = []canon.Seed{{Path: "rec/s.txt", Text: "s\n"}}
		row.Out = &canon.TextExpect{Text: "s"}
		if res := Run(dir, "SD/7", row, RunEnv{}); !res.Green {
			t.Fatalf("seed in an existing dir: %s", res.Reason)
		}
		if _, err := os.Stat(filepath.Join(dir, "rec", "keep.txt")); err != nil {
			t.Fatalf("the pre-existing neighbor must stay: %v", err)
		}
		if _, err := os.Stat(filepath.Join(dir, "rec", "s.txt")); !os.IsNotExist(err) {
			t.Fatalf("the seed must self-clean: %v", err)
		}
	})

	t.Run("a product file keeps the seed's directory behind", func(t *testing.T) {
		dir := t.TempDir()
		row := shRow("echo wrote > rec/extra.txt; cat rec/f.txt")
		row.Seeds = []canon.Seed{{Path: "rec/f.txt", Text: "f\n"}}
		row.Out = &canon.TextExpect{Text: "f"}
		if res := Run(dir, "SD/8", row, RunEnv{}); !res.Green {
			t.Fatalf("seed with a product write: %s", res.Reason)
		}
		if got := readFile(t, filepath.Join(dir, "rec", "extra.txt")); got != "wrote\n" {
			t.Fatalf("the product's file must stay: %q", got)
		}
	})

	t.Run("a red row still cleans its seeds", func(t *testing.T) {
		dir := t.TempDir()
		row := shRow("exit 1")
		row.Seeds = []canon.Seed{{Path: "in.txt", Text: "x"}}
		if res := Run(dir, "SD/4", row, RunEnv{}); res.Green {
			t.Fatal("must be red")
		}
		if _, err := os.Stat(filepath.Join(dir, "in.txt")); !os.IsNotExist(err) {
			t.Fatalf("cleanup rides the red row too: %v", err)
		}
	})

	t.Run("from-input copies the input bytes file-to-file", func(t *testing.T) {
		input := t.TempDir()
		if err := os.MkdirAll(filepath.Join(input, "data"), 0o755); err != nil {
			t.Fatal(err)
		}
		mustWrite(t, filepath.Join(input, "data", "big.bin"), "input-bytes\x00\x01tail", 0o644)
		dir := t.TempDir()
		row := shRow("cat big.bin")
		row.Seeds = []canon.Seed{{Path: "big.bin", FromInput: "data/big.bin"}}
		row.Out = &canon.TextExpect{Text: "input-bytes\x00\x01tail"}
		if res := Run(dir, "SD/5", row, RunEnv{InputDir: input}); !res.Green {
			t.Fatalf("from-input copy: %s", res.Reason)
		}
		if _, err := os.Stat(filepath.Join(dir, "big.bin")); !os.IsNotExist(err) {
			t.Fatalf("the copied seed must self-clean: %v", err)
		}
	})
}

// --- the timeout ceiling ---

func TestRunTimeout(t *testing.T) {
	row := shRow("sleep 5")
	row.Timeout = 1
	started := time.Now()
	res := runRow(t, row, RunEnv{})
	if res.Green {
		t.Fatal("a hang must be red")
	}
	if !strings.Contains(res.Reason, "timed out after 1s") {
		t.Fatalf("the timeout diagnosis: %q", res.Reason)
	}
	if !strings.Contains(res.Reason, "sleep 5") {
		t.Fatalf("the diagnosis names the command: %q", res.Reason)
	}
	if res.Env {
		t.Fatal("a hang is product behavior, not the environment")
	}
	if elapsed := time.Since(started); elapsed.Seconds() > 4 {
		t.Fatalf("the ceiling must bite early: %v", elapsed)
	}
}

// --- the environment failure classifier ---

func TestEnvClass(t *testing.T) {
	if got := EnvClass(nil); got != "" {
		t.Fatalf("no error is no class: %q", got)
	}
	if got := EnvClass(errors.New("plain")); got != "" {
		t.Fatalf("a plain error is not the environment: %q", got)
	}
	if got := EnvClass(&exec.Error{Name: "x", Err: exec.ErrNotFound}); got != "exec-not-found" {
		t.Fatalf("not found: %q", got)
	}
	if got := EnvClass(&fs.PathError{Op: "exec", Path: "x", Err: fs.ErrNotExist}); got != "exec-not-found" {
		t.Fatalf("a not-exist path error: %q", got)
	}
	if got := EnvClass(&fs.PathError{Op: "exec", Path: "x", Err: fs.ErrPermission}); got != "exec-permission" {
		t.Fatalf("permission: %q", got)
	}
}

// --- the exact-block closers: fault branches over a faulting fs port ---

// faultFS wraps the OS port with precise failure knobs (embedding
// gives every method; the knobs override theirs).
type faultFS struct {
	fsport.OS
	chmodErr  error
	mkdirErr  error
	writeErr  error
	removeErr error
}

func (f faultFS) Chmod(string, os.FileMode) error             { return f.chmodErr }
func (f faultFS) MkdirAll(string) error                       { return f.mkdirErr }
func (f faultFS) WriteFile(string, []byte, os.FileMode) error { return f.writeErr }
func (f faultFS) Remove(string) error                         { return f.removeErr }

// faultProc — a scripted process.
type faultProc struct{ out execport.Outcome }

func (p faultProc) Wait(time.Duration) execport.Outcome { return p.out }
func (p faultProc) KillGroup()                          {}

// faultExec wraps the OS process port; procFn scripts Start.
type faultExec struct {
	execport.OS
	procFn func(argv []string) (execport.Process, error)
}

func (f faultExec) Start(argv []string, _ string, _ *string) (execport.Process, error) {
	if f.procFn == nil {
		return execport.OS{}.Start(argv, f.dirFor(argv), nil)
	}
	return f.procFn(argv)
}

// StartOpts — Start's disciplines: a plain launch rides the script, a
// mode launch is the real OS one (the network-less namespace is the
// thing under test).
func (f faultExec) StartOpts(argv []string, _ string, _ *string, opts execport.LaunchOpts) (execport.Process, error) {
	if opts.Netns || f.procFn == nil {
		return execport.OS{}.StartOpts(argv, f.dirFor(argv), nil, opts)
	}
	return f.procFn(argv)
}

func (f faultExec) dirFor([]string) string { return "" }

func bindFaults(t *testing.T, fs fsport.Port, ex faultExec) {
	t.Helper()
	oldFS, oldProc := fsport.Bind, execport.Bind
	fsport.Bind, execport.Bind = fs, ex
	t.Cleanup(func() { fsport.Bind, execport.Bind = oldFS, oldProc })
}

func TestRowRunnerFaultBranches(t *testing.T) {
	dir := t.TempDir()
	t.Run("resolve chmod fault", func(t *testing.T) {
		// the file exists, is not executable: the launch grant path
		if err := os.WriteFile(filepath.Join(dir, "somefile"), []byte("#!/bin/sh\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		bindFaults(t, faultFS{chmodErr: errIO2}, faultExec{})
		r := newRowRunner(dir)
		if _, err := r.resolve("somefile"); err == nil {
			t.Fatal("chmod fault swallowed")
		}
	})
	t.Run("invoke resolve fault", func(t *testing.T) {
		bindFaults(t, faultFS{chmodErr: errIO2}, faultExec{})
		if _, reason, env := newRowRunner(dir).invoke([]string{"somefile"}, 5, execport.LaunchOpts{}); reason == "" || !env {
			t.Fatalf("resolve fault: %q %v", reason, env)
		}
	})
	t.Run("invoke wait fault", func(t *testing.T) {
		bindFaults(t, faultFS{}, faultExec{procFn: func([]string) (execport.Process, error) {
			return faultProc{out: execport.Outcome{WaitErr: "boom"}}, nil
		}})
		if _, reason, env := newRowRunner(dir).invoke([]string{"true"}, 5, execport.LaunchOpts{}); !strings.Contains(reason, "abnormally") || !env {
			t.Fatalf("wait fault: %q %v", reason, env)
		}
	})
}

func TestSeedFaultBranches(t *testing.T) {
	dir := t.TempDir()
	t.Run("apply mkdir fault", func(t *testing.T) {
		bindFaults(t, faultFS{mkdirErr: errIO2}, faultExec{})
		if _, err := applySeed(dir, []canon.Seed{{Path: "a.txt", Text: "x"}}, RunEnv{}); err == nil {
			t.Fatal("mkdir fault swallowed")
		}
	})
	t.Run("apply write fault", func(t *testing.T) {
		bindFaults(t, faultFS{writeErr: errIO2}, faultExec{})
		if _, err := applySeed(dir, []canon.Seed{{Path: "b.txt", Text: "x"}}, RunEnv{}); err == nil {
			t.Fatal("write fault swallowed")
		}
	})
	t.Run("cleanup restore fault", func(t *testing.T) {
		// a mark over an existing file whose restore fails
		path := filepath.Join(dir, "c.txt")
		if err := os.WriteFile(path, []byte("orig"), 0o644); err != nil {
			t.Fatal(err)
		}
		bindFaults(t, faultFS{writeErr: errIO2}, faultExec{})
		left := cleanupSeeds(dir, []seedMark{{path: path, existed: true, orig: []byte("orig")}})
		if len(left) != 1 {
			t.Fatalf("restore fault: %v", left)
		}
	})
	t.Run("cleanup remove fault", func(t *testing.T) {
		bindFaults(t, faultFS{removeErr: errIO2}, faultExec{})
		left := cleanupSeeds(dir, []seedMark{{path: filepath.Join(dir, "d.txt")}})
		if len(left) != 1 {
			t.Fatalf("remove fault: %v", left)
		}
	})
	t.Run("run reports a leftover seed as env red", func(t *testing.T) {
		bindFaults(t, faultFS{removeErr: errIO2}, faultExec{})
		row := canon.Row{Run: canon.Cmd{Line: "true"},
			Seeds: []canon.Seed{{Path: "e.txt", Text: "x"}}}
		res := Run(dir, "B1/1", row, RunEnv{})
		if res.Green || !res.Env || !strings.Contains(res.Reason, "did not clean up") {
			t.Fatalf("leftover: %+v", res)
		}
	})
}

var errIO2 = errors.New("io failure")

func TestTrimPartialEscapeShapes(t *testing.T) {
	if got := trimPartialEscape("abc\\x4"); got != "abc" {
		t.Fatalf("cut stump: %q", got)
	}
	if got := trimPartialEscape("abc\\x41"); got != "abc\\x41" {
		t.Fatalf("whole escape: %q", got)
	}
	if got := trimPartialEscape("plain"); got != "plain" {
		t.Fatalf("no escape: %q", got)
	}
}

func TestFirstLineOfMulti(t *testing.T) {
	if got := firstLineOf(errMulti{}); got != "first" {
		t.Fatalf("first line: %q", got)
	}
}

type errMulti struct{}

func (errMulti) Error() string { return "first\nsecond" }

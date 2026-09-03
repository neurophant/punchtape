// Package checks — execution of generated checks. A black box: a row
// lives in a clean temporary directory, its commands are an arbitrary
// argv (a surface entry point, a utility, a shell — all the same to
// the machine); the outcome is observed only through universal
// observables: streams, the exit code, files. The machine knows
// nothing about the product or the environment.
package checks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"

	"gopkg.in/yaml.v3"
)

// Result — the outcome of one check's run: green/red and one reason
// line for the red. Env — the red is an environment or launch
// failure (not an expectation divergence): such reds do not eat the
// attempt budget — they are not product behavior.
type Result struct {
	CheckID string
	Green   bool
	Reason  string
	Env     bool
}

// Mode — the run mode: normal, or a probe with a broken surface.
type Mode int

const (
	// Normal — a normal run.
	Normal Mode = iota
	// ProbeBrokenSurface — a probe: every surface entry point the row
	// called is replaced with a stub that always fails. A check that
	// stays green in such a run checks nothing — it does not run the
	// declared path; a row that touched no surface entry point proves
	// nothing either.
	ProbeBrokenSurface
)

// Run executes a check in an isolated temporary directory. The
// surface is what the operator declared as the build recipe (the
// artifact: a file or a whole tree — the machine carries the declared
// into the run directory as is), or, without a declaration, the file
// named command[0], submitted into the project directly. The
// operator decides the surface's form — the machine assumes nothing
// about it. An unbuilt surface is an honest red. The pre setup
// commands run before the action, each must exit with code 0 — state
// setup by a series of calls. A normal run doubles: the whole
// sequence executes twice in two clean directories and the
// observations must match — restart/transience is the machine's job;
// the executor does not need to write anything about it. artifact —
// the declared artifact's relative path in the project (empty — the
// direct submission mode).
func Run(workdir, artifact string, check canon.Check, mode Mode) Result {
	return runWithAdvice(workdir, artifact, check, mode, false)
}

// RunNoCode — Run with the env advice routed for a not-yet-delivered
// product: at zero code "install the tool" misleads (there is nothing
// to install against); the advice says deliver the code.
func RunNoCode(workdir, artifact string, check canon.Check, mode Mode) Result {
	return runWithAdvice(workdir, artifact, check, mode, true)
}

func runWithAdvice(workdir, artifact string, check canon.Check, mode Mode, noCode bool) Result {
	if mode == ProbeBrokenSurface {
		return runOnce(workdir, artifact, check, true, noCode).Result
	}
	// Run lock: one instance's check executors are serialized (a
	// background rehearsal and live gates run one suite — the resource
	// occupied by a scenario is single). The wait is bounded: a live
	// caller first marks the rehearsal for stop; it exits after the
	// current check.
	release, err := acquireRunLock(workdir)
	if err != nil {
		return failEnv(check.ID, canondata.T("check.fail.run-lock", canondata.M{"err": err.Error()}))
	}
	defer release()
	// Both transience probes — in the SAME run directory path (the
	// directory is cleaned between probes): the run path is the
	// machine's choice, and a tool that prints its working directory
	// must not become "non-deterministic" because of that choice.
	// Random bytes of the behavior itself are still caught by the
	// comparison.
	tmp, err := os.MkdirTemp("", "check-*")
	if err != nil {
		return failEnv(check.ID, canondata.T("check.fail.run-dir", canondata.M{"err": err.Error()}))
	}
	defer os.RemoveAll(tmp)
	first := runInDir(tmp, workdir, artifact, check, false, noCode)
	if !first.Green {
		// The last captured red run is kept in the instance's rebuilt
		// cache (run evidence): why fragment re-plays the measured
		// bytes of each currently red assertion, so the hand reads
		// what the product actually printed without a probe
		// submission. The write is bounded (excerpt discipline) and
		// safe to lose — the cache is not a source of truth.
		if first.ev != nil {
			_ = persistRunEvidence(workdir, check.ID, first.ev)
		}
		return first.Result
	}
	if err := cleanDir(tmp); err != nil {
		return failEnv(check.ID, canondata.T("check.fail.run-dir", canondata.M{"err": err.Error()}))
	}
	second := runInDir(tmp, workdir, artifact, check, false, noCode)
	if !second.Green {
		return second.Result
	}
	if diff, obs := observationDiff(first.obs, second.obs, volatileSet(check)); diff != "" {
		return fail(check.ID, canondata.T("check.fail.not-transient", canondata.M{
			"diff": diff, "proposal": volatileProposal(obs),
		}))
	}
	return Result{CheckID: check.ID, Green: true}
}

// cleanDir — clean out a directory's contents, the path itself
// stays: the second transience probe starts in the same clean
// directory.
func cleanDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

// volatileSet — the volatility declared by the scenario
// (when.volatile): streams (stdout/stderr) and relative state file
// paths whose exact bytes the brief declares unpredictable.
// Transience skips exactly what is declared; the undeclared stays
// byte-strict.
func volatileSet(check canon.Check) map[string]bool {
	var vol []string
	if check.When != nil {
		vol = check.When.Volatile
	}
	set := make(map[string]bool, len(vol))
	for _, v := range vol {
		set[v] = true
	}
	return set
}

// runOutcome — the observations of one run of the sequence: the
// action's streams and code, a full snapshot of the directory's files
// (without the run tooling).
type runOutcome struct {
	exit   int
	stdout string
	stderr string
	files  map[string]string
}

// runResult — a run outcome with observations (the observations only
// matter for green: a red has already named its reason).
type runResult struct {
	Result
	obs runOutcome
	// ev — the captured evidence of an assertion red: what the last
	// run measured when it reddened (nil for green and environment
	// reds).
	ev *RunEvidence
}

// RunEvidence — the measured bytes of the last red run of a check:
// the failing assertion with its measured actual, the invocation's
// exit and both streams, all in the one-line excerpt discipline.
// Kept per check id in the rebuilt cache zone; a lost file is an
// honest "no captured run", never a wrong answer.
type RunEvidence struct {
	At       time.Time `yaml:"at"`
	Scenario string    `yaml:"scenario"`
	Command  string    `yaml:"command"`
	Exit     int       `yaml:"exit"`
	Stdout   string    `yaml:"stdout"`
	Stderr   string    `yaml:"stderr"`
	Assert   string    `yaml:"assert"`
	Want     string    `yaml:"want"`
	Actual   string    `yaml:"actual"`
	Path     string    `yaml:"path,omitempty"`
}

// runOnce — one run of a check in its own clean directory: seed,
// setup commands, action, assertions. broken=true replaces surface
// entry points with stubs.
func runOnce(workdir, artifact string, check canon.Check, broken, noCode bool) runResult {
	release, err := acquireRunLock(workdir)
	if err != nil {
		return runResult{Result: failEnv(check.ID, canondata.T("check.fail.run-lock",
			canondata.M{"err": err.Error()}))}
	}
	defer release()
	tmp, err := os.MkdirTemp("", "check-*")
	if err != nil {
		return runResult{Result: failEnv(check.ID, canondata.T("check.fail.run-dir",
			canondata.M{"err": err.Error()}))}
	}
	defer os.RemoveAll(tmp)
	return runInDir(tmp, workdir, artifact, check, broken, noCode)
}

// runInDir — one run of a check in the given run directory dir
// (assumed clean): the shared path of the two transience probes
// lives here.
func runInDir(dir, workdir, artifact string, check canon.Check, broken, noCode bool) runResult {
	if err := applySeed(dir, workdir, check.Seed, check.Materials); err != nil {
		return runResult{Result: failEnv(check.ID, err.Error())}
	}
	if check.When == nil {
		return runResult{Result: failEnv(check.ID, canondata.T("check.fail.no-surface"))}
	}
	r := newResolver(workdir, artifact, dir, broken, noCode)
	// Surface names named by the table are brought into the run before
	// the commands: references to the product from inside argv (a
	// shell) find the files; a name without a file in the project has
	// nothing to bring here — a PATH launch.
	for _, name := range check.SurfaceNames {
		_, _ = r.resolve(name)
	}
	// The world freezes before reading observations: everything the
	// row's command left alive (a shell — its child servers) is taken
	// down by the process group; survivors would hold resources and
	// corrupt neighboring rows. OS level, zero product knowledge.
	defer r.killGroups()

	// Setup commands: each must exit with code 0 — state is built by
	// a series of calls; a setup failure honestly reddens the check.
	for _, pre := range check.Pre {
		out, reason, startFail := r.invoke(pre, nil, check.When.TimeoutSec)
		if reason != "" {
			return runResult{Result: failClass(check.ID, canondata.T("check.fail.setup", canondata.M{
				"command": strings.Join(pre, " "), "reason": reason,
				"argv": strings.Join(pre, "]["),
			}), startFail)}
		}
		if out.exit != 0 {
			return runResult{Result: fail(check.ID, canondata.T("check.fail.setup-exit", canondata.M{
				"command": strings.Join(pre, " "), "exit": fmt.Sprintf("%d", out.exit),
				"argv": strings.Join(pre, "]["),
			}))}
		}
	}

	obs, reason, startFail := r.invoke(check.When.Command, check.When.Stdin, check.When.TimeoutSec)
	if reason != "" {
		return runResult{Result: failClass(check.ID, reason, startFail)}
	}
	r.killGroups()
	for _, a := range check.Then {
		reason, actual := assertObservation(dir, a, obs.stdout, obs.stderr, obs.exit)
		if reason != "" {
			// A terminator-shaped divergence (the pinned bytes vs the
			// measured ones differ by EXACTLY one trailing newline)
			// has a computable repair: the red reason carries the
			// copy-ready update-assertions op with the MEASURED value
			// — a PROPOSAL; the executor submits it, the red stays red
			// until their own submission changes it.
			if op, ok := terminatorProposal(check.Scenario, a, obs.stdout, obs.stderr); ok {
				reason += canondata.T("check.assert.terminator-proposal", canondata.M{"op": op})
			}
			invocation := strings.Join(check.When.Command, " ")
			ev := &RunEvidence{
				At:       time.Now().UTC(),
				Scenario: check.Scenario,
				Command:  invocation,
				Exit:     obs.exit,
				Stdout:   excerpt(obs.stdout),
				Stderr:   excerpt(obs.stderr),
				Assert:   a.Observation + " " + a.Condition,
				Want:     excerpt(a.Value),
				Actual:   excerpt(actual),
				Path:     a.Path,
			}
			return runResult{Result: fail(check.ID, canondata.T("check.fail.observed", canondata.M{
				"reason": reason, "command": invocation, "exit": fmt.Sprintf("%d", obs.exit),
				"stdout": excerpt(obs.stdout), "stderr": excerpt(obs.stderr),
			})), ev: ev}
		}
	}
	files, serr := snapshotFiles(dir, r.materialized, volatileSet(check))
	if serr != nil {
		return runResult{Result: failEnv(check.ID, serr.Error())}
	}
	obs.files = files
	return runResult{Result: Result{CheckID: check.ID, Green: true}, obs: obs}
}

// terminatorProposal — the copy-ready repair op when the divergence
// is terminator-shaped: want and measured differ by EXACTLY one
// trailing newline on a pinned stream. Byte-shape predicate, zero
// task knowledge; the value rides in single quotes with the
// channel's own \xNN escapes (the expectation literal layer decodes
// them; the YAML layer passes single-quoted bytes through).
func terminatorProposal(scenario string, a canon.Assertion, stdout, stderr string) (string, bool) {
	if a.Condition != "equals" || (a.Observation != "stdout" && a.Observation != "stderr") {
		return "", false
	}
	actual := stdout
	if a.Observation == "stderr" {
		actual = stderr
	}
	if !(actual == a.Value+"\n" || a.Value == actual+"\n") {
		return "", false
	}
	return fmt.Sprintf(
		"update-assertions: {scenario: %s, ops: [{match: {observation: %s, condition: equals, value: %s}, set: {observation: %s, condition: equals, value: %s}}]}",
		scenario, a.Observation, yamlSingleQuoted(a.Value), a.Observation, yamlSingleQuoted(actual)), true
}

// yamlSingleQuoted — the value in single quotes with the byte
// escapes the expectation channel itself decodes (\xNN; a literal
// backslash is \x5C); inner single quotes double.
func yamlSingleQuoted(v string) string {
	var b strings.Builder
	b.WriteByte('\'')
	for i := 0; i < len(v); i++ {
		c := v[i]
		switch {
		case c == '\'':
			b.WriteString("''")
		case c == '\\':
			b.WriteString("\\x5C")
		case c == '\n':
			b.WriteString("\\x0A")
		case c == '\r':
			b.WriteString("\\x0D")
		case c == '\t':
			b.WriteString("\\x09")
		case c >= 0x20 && c < 0x7F:
			b.WriteByte(c)
		default:
			fmt.Fprintf(&b, "\\x%02X", c)
		}
	}
	b.WriteByte('\'')
	return b.String()
}

// applySeed — seeding a row's bytes: inline content is written as
// files, project materials are carried byte-for-byte under the same
// paths. Carrying bytes is the machine's only file operation on the
// run world; it knows nothing about the content.
func applySeed(dir, projectRoot string, seed []canon.Seed, materials []canon.Material) error {
	for _, s := range seed {
		path := filepath.Join(dir, filepath.FromSlash(s.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("%s", canondata.T("check.fail.seed", canondata.M{"path": s.Path, "err": err.Error()}))
		}
		if err := os.WriteFile(path, []byte(s.Content), 0o644); err != nil {
			return fmt.Errorf("%s", canondata.T("check.fail.seed", canondata.M{"path": s.Path, "err": err.Error()}))
		}
	}
	for _, m := range materials {
		src := filepath.Join(projectRoot, filepath.FromSlash(m.From))
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("%s", canondata.T("check.fail.material", canondata.M{"from": m.From, "err": err.Error()}))
		}
		dst := filepath.Join(dir, filepath.FromSlash(m.From))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("%s", canondata.T("check.fail.material", canondata.M{"from": m.From, "err": err.Error()}))
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return fmt.Errorf("%s", canondata.T("check.fail.material", canondata.M{"from": m.From, "err": err.Error()}))
		}
	}
	return nil
}

// resolver — resolving a row's command[0] in the run directory.
// Order: what is already in the directory (seed, material, an
// earlier row command's output), a surface entry (a file inside the
// artifact tree, a file artifact, a project file by name without a
// recipe); a name without an entry has nothing to look up here — it
// goes to the environment as is: a regular launch, the environment
// finds the path. A "product entry point", a "shell" and a "utility"
// are all just argv to the machine; there is no command dictionary
// or knowledge about them. broken=true — a probe: every actually
// existing surface entry the row called is replaced with a stub; the
// probe does not touch PATH names.
type resolver struct {
	workdir      string
	artifact     string
	dir          string
	broken       bool
	noCode       bool
	materialized map[string]bool
	groups       []*exec.Cmd
}

func newResolver(workdir, artifact, dir string, broken, noCode bool) *resolver {
	return &resolver{
		workdir: workdir, artifact: artifact,
		dir: dir, broken: broken, noCode: noCode,
		materialized: map[string]bool{},
	}
}

// resolve returns the launch path for command[0] or the name as is
// (a PATH launch by the environment).
func (r *resolver) resolve(name string) (string, error) {
	entry := filepath.Join(r.dir, filepath.FromSlash(name))
	if _, err := os.Stat(entry); err == nil {
		return entry, nil
	}
	root := name
	if r.artifact != "" {
		root = r.artifact
	}
	src := filepath.Join(r.workdir, filepath.FromSlash(root))
	info, err := os.Stat(src)
	if err != nil {
		return name, nil // PATH launch: the environment knows its tools
	}
	if r.broken {
		// The probe replaces only the entries a normal run would have
		// materialized for this name; a name without an entry is PATH —
		// the probe does not touch it.
		if surfaceHasEntry(info, src, root, name) {
			if err := r.writeStub(name); err != nil {
				return "", err
			}
			return entry, nil
		}
		return name, nil
	}
	if info.IsDir() {
		files, err := copyTree(src, filepath.Join(r.dir, filepath.FromSlash(root)), filepath.ToSlash(root))
		if err != nil {
			return "", fmt.Errorf("surface %s: %w", name, err)
		}
		for path := range files {
			r.materialized[path] = true
		}
	} else {
		if err := copyFileKeepMode(src, filepath.Join(r.dir, filepath.FromSlash(root))); err != nil {
			return "", fmt.Errorf("surface %s: %w", name, err)
		}
		r.materialized[filepath.ToSlash(root)] = true
	}
	ei, err := os.Stat(entry)
	if err != nil || ei.IsDir() {
		// The name is not among the surface entries (or the entry is a
		// directory): this is not a surface entry point; the environment
		// launches the name by PATH.
		return name, nil
	}
	// The right to execute is a property of the entry, not the code:
	// a copy of an entry in the run directory is always executable.
	if err := os.Chmod(entry, ei.Mode().Perm()|0o111); err != nil {
		return "", fmt.Errorf("surface %s: %w", name, err)
	}
	return entry, nil
}

// surfaceHasEntry — a check without carrying: would a normal run
// have materialized a surface file entry for the name — a project
// file by name, a file artifact under its own name, or a file inside
// the artifact tree. A name equal to the tree root and directories
// are not entries.
func surfaceHasEntry(info os.FileInfo, src, root, name string) bool {
	if !info.IsDir() {
		return name == root
	}
	if root == name {
		return false
	}
	ei, err := os.Stat(filepath.Join(src, filepath.FromSlash(name)))
	return err == nil && !ei.IsDir()
}

// writeStub — an entry stub for the probe: empty output, exit code 1.
// The stub's interpreter is sh from the environment's PATH (the same
// basis as a PATH launch); without sh in PATH the file is empty —
// the launch fails, which is what a broken surface needs.
func (r *resolver) writeStub(name string) error {
	entry := filepath.Join(r.dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(entry), 0o755); err != nil {
		return err
	}
	body := []byte{}
	if sh, err := exec.LookPath("sh"); err == nil {
		body = []byte("#!" + sh + "\nexit 1\n")
	}
	if err := os.WriteFile(entry, body, 0o755); err != nil {
		return err
	}
	r.materialized[filepath.ToSlash(name)] = true
	return nil
}

// killGroups removes everything alive left by the row's commands:
// each command ran in its own process group — surviving children (a
// shell with servers) do not outlive the row.
func (r *resolver) killGroups() {
	for _, cmd := range r.groups {
		killProcessGroup(cmd)
	}
	r.groups = nil
}

// invoke — one call of a row command with a timeout; the reason is
// non-empty when the call did not happen (timeout, process death).
// startFail — an environment failure: the process did not start or
// an environment class; a launch timeout is not the environment (the
// behavior may have hung; the attempt stays an attempt).
func (r *resolver) invoke(command []string, stdin *string, timeoutSec int) (runOutcome, string, bool) {
	argv0, err := r.resolve(command[0])
	if err != nil {
		return runOutcome{}, err.Error(), true
	}
	cmd := exec.Command(argv0, command[1:]...)
	cmd.Dir = r.dir
	// Its own process group: the command's descendants (servers,
	// workers, what a shell raised) are taken down by the group —
	// they hold no resources after the row.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	r.groups = append(r.groups, cmd)
	if stdin != nil {
		cmd.Stdin = strings.NewReader(*stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	timeout := time.Duration(timeoutSec) * time.Second
	if timeout <= 0 {
		timeout = time.Duration(canondata.Limit("check.timeout-default-s")) * time.Second
	}
	invocation := strings.Join(command, " ")
	if err := cmd.Start(); err != nil {
		r.groups = r.groups[:len(r.groups)-1]
		if class := EnvClass(err); class != "" {
			return runOutcome{}, canondata.EnvDiagAt(class, r.noCode, canondata.M{
				"command": invocation,
			}), true
		}
		return runOutcome{}, canondata.T("check.fail.start", canondata.M{
			"command": invocation, "err": err.Error(),
		}), true
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var exitErr error
	select {
	case err := <-done:
		exitErr = err
	case <-time.After(timeout):
		killProcessGroup(cmd)
		return runOutcome{}, canondata.EnvDiag("command-timeout", canondata.M{
			"sec": fmt.Sprintf("%d", timeoutSec), "command": invocation,
		}), false
	}
	out := runOutcome{exit: 0, stdout: stdout.String(), stderr: stderr.String()}
	if exitErr != nil {
		if ee, ok := exitErr.(*exec.ExitError); ok {
			out.exit = ee.ExitCode()
		} else {
			return runOutcome{}, canondata.T("check.fail.wait", canondata.M{
				"err": exitErr.Error(), "command": invocation,
			}), true
		}
	}
	return out, "", false
}

// killProcessGroup removes the call entirely: the direct process and
// all its descendants — and WAITS for the group's actual death: a
// kill signal is not death, a dying member still holds what it held
// (a port, a lock in /tmp). The two determinism probes of one row are
// strictly serial: what the first probe held must be released before
// the second begins. The wait is bounded; a group that refuses to die
// within it is left to the OS.
func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	pgid := cmd.Process.Pid
	_ = syscall.Kill(-pgid, syscall.SIGKILL)
	_ = cmd.Process.Kill()
	deadline := time.Now().Add(1 * time.Second)
	for {
		if err := syscall.Kill(-pgid, 0); err != nil {
			return // the group is gone
		}
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
}

// firstStreamDiff — the first differing line of two streams with its
// line number: the red reason carries the clue of which exact line
// to compare. Long output is cut to a one-line budget.
func firstStreamDiff(a, b string) string {
	as, bs := strings.Split(a, "\n"), strings.Split(b, "\n")
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		var x, y string
		if i < len(as) {
			x = as[i]
		}
		if i < len(bs) {
			y = bs[i]
		}
		if x != y {
			return fmt.Sprintf("line %d: %q vs %q", i+1, excerpt(x), excerpt(y))
		}
	}
	return ""
}

// observationDiff — the first divergence of the observations of two
// clean runs. volatile — the unpredictable observations declared by
// the scenario (when.volatile): stream names and state file paths
// whose exact bytes the brief declares random; the content
// reconciliation skips them. File presence is always strict: the
// file sets of the two runs must match — "sometimes writes, sometimes
// not" is not cured by volatility.
// volatileProposal — the copy-ready volatile line for the differing
// observation: a PROPOSAL the executor submits (pre-first-green it
// is free; post-code on proven rows it stages for a human
// assertion). The machine never declares volatile by itself.
func volatileProposal(obs string) string {
	return fmt.Sprintf("volatile: [%s]", obs)
}

func observationDiff(a, b runOutcome, volatile map[string]bool) (string, string) {
	switch {
	case a.exit != b.exit:
		return canondata.T("check.diff.exit", canondata.M{
			"first": fmt.Sprintf("%d", a.exit), "second": fmt.Sprintf("%d", b.exit)}), "exit-code"
	case !volatile["stdout"] && a.stdout != b.stdout:
		return canondata.T("check.diff.stdout", canondata.M{"diff": firstStreamDiff(a.stdout, b.stdout)}), "stdout"
	case !volatile["stderr"] && a.stderr != b.stderr:
		return canondata.T("check.diff.stderr", canondata.M{"diff": firstStreamDiff(a.stderr, b.stderr)}), "stderr"
	case len(a.files) != len(b.files):
		return canondata.T("check.diff.files-count", canondata.M{
			"first": fmt.Sprintf("%d", len(a.files)), "second": fmt.Sprintf("%d", len(b.files))}), ""
	}
	paths := make([]string, 0, len(a.files))
	for p := range a.files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		other, ok := b.files[p]
		if !ok {
			return canondata.T("check.diff.file-one-run", canondata.M{"path": p}), ""
		}
		if volatile[p] {
			continue
		}
		if string(yamlio.NormalizeEOL([]byte(a.files[p]))) != string(yamlio.NormalizeEOL([]byte(other))) {
			return canondata.T("check.diff.file-differs", canondata.M{"path": p}), p
		}
	}
	return "", ""
}

// Captured — the observable facts of one command run: codes, streams
// and the files that changed after the seed.
type Captured struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Files    map[string]string
}

// Capture runs a command in a clean temporary directory with a seed
// and returns the observable — facts, not a judgment. Determinism —
// two runs and a comparison — is the caller's concern.
func Capture(workdir, artifact string, seed []canon.Seed, command []string, timeoutSec int) (Captured, error) {
	return CaptureSeq(workdir, artifact, seed, nil, nil, command, timeoutSec)
}

// CaptureSeq — the same with materials and setup commands before
// the action: scenarios with a series of calls are observed whole.
func CaptureSeq(workdir, artifact string, seed []canon.Seed, materials []canon.Material, pre [][]string, command []string, timeoutSec int) (Captured, error) {
	var out Captured
	if len(command) == 0 {
		return out, fmt.Errorf("%s", canondata.T("check.error.capture-empty"))
	}
	if timeoutSec <= 0 {
		timeoutSec = 10
	}
	release, err := acquireRunLock(workdir)
	if err != nil {
		return out, fmt.Errorf("%s", canondata.T("check.fail.run-lock", canondata.M{"err": err.Error()}))
	}
	defer release()
	tmp, err := os.MkdirTemp("", "probe-*")
	if err != nil {
		return out, fmt.Errorf("probe dir: %w", err)
	}
	defer os.RemoveAll(tmp)
	if err := applySeed(tmp, workdir, seed, materials); err != nil {
		return out, err
	}
	before, err := snapshotFiles(tmp, nil, nil)
	if err != nil {
		return out, err
	}
	r := newResolver(workdir, artifact, tmp, false, false)
	defer r.killGroups()
	for _, setup := range pre {
		if _, reason, _ := r.invoke(setup, nil, timeoutSec); reason != "" {
			return out, fmt.Errorf("%s", canondata.T("check.fail.setup", canondata.M{
				"command": strings.Join(setup, " "), "reason": reason,
			}))
		}
	}
	out2, reason, _ := r.invoke(command, nil, timeoutSec)
	if reason != "" {
		return out, fmt.Errorf("%s", reason)
	}
	// The world is frozen before the snapshot: everything alive left
	// by the commands is taken down by process groups.
	r.killGroups()
	out.ExitCode, out.Stdout, out.Stderr = out2.exit, out2.stdout, out2.stderr
	after, err := snapshotFiles(tmp, r.materialized, nil)
	if err != nil {
		return out, err
	}
	out.Files = map[string]string{}
	for path, content := range after {
		if old, seeded := before[path]; seeded && old == content {
			continue
		}
		out.Files[path] = content
	}
	return out, nil
}

// runEvidencePath — the run evidence file of the instance: the
// rebuilt cache zone (not a source of truth, safe to delete).
func runEvidencePath(workdir string) string {
	return filepath.Join(canon.Dir(workdir), "cache", "run-evidence.yamll")
}

// persistRunEvidence — merge one check's captured red run into the
// evidence file: read-modify-write under the run lock (all runs
// serialize on it), one entry per check id — the LAST red run.
func persistRunEvidence(workdir, checkID string, ev *RunEvidence) error {
	all := map[string]*RunEvidence{}
	if raw, err := os.ReadFile(runEvidencePath(workdir)); err == nil {
		_ = yaml.Unmarshal(raw, &all)
	}
	all[checkID] = ev
	data, err := yaml.Marshal(all)
	if err != nil {
		return err
	}
	return yamlio.WriteAtomic(runEvidencePath(workdir), data)
}

// LoadRunEvidence — the captured red run of one check id (nil — none
// captured: green checks and environment reds leave nothing).
func LoadRunEvidence(workdir, checkID string) *RunEvidence {
	all := map[string]*RunEvidence{}
	raw, err := os.ReadFile(runEvidencePath(workdir))
	if err != nil {
		return nil
	}
	if err := yaml.Unmarshal(raw, &all); err != nil {
		return nil
	}
	return all[checkID]
}

// snapshotFiles — a map relative path → content for the directory's
// regular files. materialized — the set of files brought by the
// surface: run tooling, not a product observation. volatile — the
// unpredictable paths declared by the scenario: their presence is
// always strict, the bytes are not read at all (the comparison skips
// them anyway; reading gigabytes behind the comparison is not
// hidden). The total read volume is bounded by a delivery threshold:
// the snapshot must live in memory; heavy product artifacts honestly
// go red with a diagnosis instead of taking the machine down.
func snapshotFiles(dir string, materialized, volatile map[string]bool) (map[string]string, error) {
	out := map[string]string{}
	budget := int64(canondata.Limit("snapshot.bytes-max"))
	var total int64
	walkErr := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !info.Mode().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if materialized[rel] {
			return nil
		}
		if volatile[rel] {
			out[rel] = ""
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		total += int64(len(data))
		if total > budget {
			return fmt.Errorf("%s", canondata.T("check.fail.snapshot-budget", canondata.M{
				"max": strconv.FormatInt(budget, 10), "path": rel,
			}))
		}
		out[rel] = string(data)
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return out, nil
}

// excerpt — captured output in one line: newlines collapse, long
// output is cut. The red reason must stay one line.
func excerpt(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", "; "))
	s = yamlio.EscapeControls(s)
	if s == "" {
		return "<empty>"
	}
	if len(s) > 160 {
		return trimPartialEscape(s[:160]) + "..."
	}
	return `"` + s + `"`
}

// trimPartialEscape — a truncated window must not end inside a \xNN
// escape: the stump of a cut sequence reads as garbage bytes. A
// complete sequence (or no sequence) at the cut is left as is.
func trimPartialEscape(s string) string {
	i := strings.LastIndex(s, "\\x")
	if i >= 0 && len(s)-i < 4 {
		return s[:i]
	}
	return s
}

// copyTree copies a directory tree preserving relative paths and
// file permissions; returns the set of carried files (slash paths).
func copyTree(src, dst, root string) (map[string]bool, error) {
	materialized := map[string]bool{}
	err := filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return nil // non-regular files are not carried by the surface
		}
		if err := copyFileKeepMode(p, target); err != nil {
			return err
		}
		materialized[filepath.ToSlash(filepath.Join(root, rel))] = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	return materialized, nil
}

// copyFileKeepMode copies a file preserving permissions; parent
// directories are created along the way.
func copyFileKeepMode(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode().Perm())
}

// assertObservation returns an empty reason when the assertion
// holds, otherwise one reason line with the fact and the measured
// value of the observation (empty for the existence-only classes —
// there is nothing measured to show). Text file comparisons run on
// normalized copies (cross-OS; bytes-equals — exact bytes per the
// declared contract.
func assertObservation(tmp string, a canon.Assertion, stdout, stderr string, exitCode int) (string, string) {
	var actual string
	normFile := func(v string) string {
		if a.Observation != "file" {
			return v
		}
		return string(yamlio.NormalizeEOL([]byte(v)))
	}
	switch a.Observation {
	case "stdout":
		actual = stdout
	case "stderr":
		actual = stderr
	case "exit-code":
		actual = fmt.Sprintf("%d", exitCode)
	case "file":
		data, err := os.ReadFile(filepath.Join(tmp, filepath.FromSlash(a.Path)))
		if err != nil {
			if a.Condition == "not-exists" || a.Condition == "absent" {
				return "", ""
			}
			if a.Condition == "exists" {
				return canondata.T("check.assert.file-missing", canondata.M{"path": a.Path}), ""
			}
			return fmt.Sprintf("file %s: %s", a.Path, err), ""
		}
		if a.Condition == "exists" {
			return "", ""
		}
		if a.Condition == "not-exists" || a.Condition == "absent" {
			return canondata.T("check.assert.file-exists", canondata.M{"path": a.Path}), ""
		}
		actual = string(data)
	default:
		return canondata.T("check.assert.observation-unknown", canondata.M{"observation": a.Observation}), ""
	}
	switch a.Condition {
	case "contains":
		if !strings.Contains(normFile(actual), normFile(a.Value)) {
			return canondata.T("check.assert.not-contains", canondata.M{
				"observation": a.Observation, "value": fmt.Sprintf("%q", a.Value), "actual": measured(a.Observation, actual),
			}), actual
		}
	case "absent":
		// Streams: absent ≡ equals "" — the output must be empty.
		// Files are handled above (the file is absent); a file with
		// absent does not reach here.
		if strings.TrimSpace(actual) != "" {
			return canondata.T("check.assert.not-empty", canondata.M{
				"observation": a.Observation, "actual": measured(a.Observation, actual),
			}), actual
		}
	case "fails":
		// exit-code: any non-zero code; no specific code is pinned.
		if actual == "0" {
			return canondata.T("check.assert.zero", canondata.M{"observation": a.Observation}), actual
		}
	case "equals":
		// Byte-for-byte against the pin as stored: a stream's trailing
		// terminator is part of the pinned bytes, not formatting to be
		// forgiven. The pin rides verbatim from the table; a product
		// differing by one byte is red. (contains is the lenient
		// channel; json-equals the structural one.)
		wantEq, gotEq := normFile(a.Value), normFile(actual)
		if wantEq != gotEq {
			return canondata.T("check.assert.not-equal", canondata.M{
				"observation": a.Observation, "value": fmt.Sprintf("%q", a.Value), "actual": measured(a.Observation, actual),
				"diff": divergenceLine(wantEq, gotEq),
			}), actual
		}
	case "json-equals":
		if reason := assertJSONEquals(a, actual); reason != "" {
			return reason, actual
		}
	case "bytes-equals":
		// The file's exact bytes: no trimming — trailing newlines and
		// other byte invariants of textual states are checked as is.
		if actual != a.Value {
			return canondata.T("check.assert.bytes-differ", canondata.M{
				"path": a.Path, "actual": measured(a.Observation, actual), "diff": divergenceLine(a.Value, actual),
			}), actual
		}
	case "json-contains":
		if reason := assertJSONContains(a, actual); reason != "" {
			return reason, actual
		}
	case "exists", "not-exists":
		// handled in the file branch
	default:
		return canondata.T("check.assert.condition-unknown", canondata.M{"condition": a.Condition}), ""
	}
	return "", ""
}

// assertJSONEquals — semantic equality: both sides are parsed as
// JSON and compared structurally. Key order, whitespace and number
// notation (1 vs 1.0) do not matter; array element order does: it is
// part of the state, not the formatting.
// measured — the raw measured bytes for an observation's red reason:
// file observations carry the excerpt budget, streams their brief.
func measured(observation, actual string) string {
	if observation == "file" {
		return excerpt(actual)
	}
	return brief(actual)
}

func assertJSONEquals(a canon.Assertion, actual string) string {
	want, err := parseJSON(a.Value)
	if err != nil {
		return canondata.T("check.assert.json-want-invalid", canondata.M{
			"observation": a.Observation, "err": err.Error(),
		})
	}
	got, err := parseJSON(actual)
	if err != nil {
		return canondata.T("check.assert.json-invalid", canondata.M{
			"observation": a.Observation, "err": err.Error(), "actual": measured(a.Observation, actual),
		})
	}
	if !reflect.DeepEqual(got, want) {
		return canondata.T("check.assert.json-not-equal", canondata.M{
			"observation": a.Observation, "want": canonicalJSON(want), "actual": brief(canonicalJSON(got)),
		})
	}
	return ""
}

// assertJSONContains — semantic containment: the actual value is
// parsed as JSON, so is the wanted one. For an array — a match of at
// least one element; for an object element a subset of its fields
// suffices: the spec names what must be present and does not pay for
// knowing others' extra fields. For the root object — containment
// of all the wanted's fields.
func assertJSONContains(a canon.Assertion, actual string) string {
	want, err := parseJSON(a.Value)
	if err != nil {
		return canondata.T("check.assert.jsonc-want-invalid", canondata.M{
			"observation": a.Observation, "err": err.Error(),
		})
	}
	got, err := parseJSON(actual)
	if err != nil {
		return canondata.T("check.assert.json-invalid", canondata.M{
			"observation": a.Observation, "err": err.Error(), "actual": measured(a.Observation, actual),
		})
	}
	if !jsonContains(got, want) {
		return canondata.T("check.assert.json-not-contains", canondata.M{
			"observation": a.Observation, "want": canonicalJSON(want), "actual": brief(canonicalJSON(got)),
		})
	}
	return ""
}

// parseJSON parses JSON text into a tree of standard values.
func parseJSON(s string) (any, error) {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, err
	}
	return v, nil
}

// jsonContains — the wanted tree is contained in the actual one: a
// scalar — an exact match; an object — all its fields present and
// equal (subset); an actual array — any element matching by the same
// rules; a wanted array inside an array — an exact match only.
func jsonContains(got, want any) bool {
	switch g := got.(type) {
	case []any:
		if _, wantIsArr := want.([]any); wantIsArr {
			return reflect.DeepEqual(got, want)
		}
		for _, el := range g {
			if jsonContains(el, want) {
				return true
			}
		}
		return false
	case map[string]any:
		wantMap, wantIsMap := want.(map[string]any)
		if !wantIsMap {
			return false
		}
		for k, v := range wantMap {
			other, ok := g[k]
			if !ok || !reflect.DeepEqual(other, v) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(got, want)
	}
}

// canonicalJSON prints a value in one line: the red reason must
// stay one line.
func canonicalJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "<unprintable>"
	}
	return string(data)
}

// divergenceLine — the place of the first divergence for exact
// comparisons: both sides' lengths and a window of bytes around the
// first difference. A one-byte difference at the tail must not be
// invisible behind a head cut.
func divergenceLine(want, got string) string {
	at := 0
	for at < len(want) && at < len(got) && want[at] == got[at] {
		at++
	}
	window := func(s string) string {
		if s == "" {
			return "<empty>"
		}
		lo, hi := at-16, at+24
		if lo < 0 {
			lo = 0
		}
		if hi > len(s) {
			hi = len(s)
		}
		return fmt.Sprintf("%q", s[lo:hi])
	}
	return fmt.Sprintf("want %d bytes, got %d, first difference at byte %d: want %s, got %s",
		len(want), len(got), at, window(want), window(got))
}

// brief squeezes a fact into a readable line for feedback. Escaped
// the same way as every text surface: the line stays valid text over
// any measured bytes.
func brief(s string) string {
	s = yamlio.EscapeControls(strings.TrimSpace(s))
	if len(s) > 80 {
		return trimPartialEscape(s[:80]) + "…"
	}
	if s == "" {
		return "<empty>"
	}
	return s
}

func fail(id, reason string) Result {
	return Result{CheckID: id, Green: false, Reason: reason}
}

// failEnv — a red for an environment/launch failure: not an
// expectation divergence; does not eat the attempt budget.
func failEnv(id, reason string) Result {
	return Result{CheckID: id, Green: false, Reason: reason, Env: true}
}

// failClass — a red with a class by source: the start did not
// happen — the environment; a timeout and everything that reached
// observation — behavior.
func failClass(id, reason string, env bool) Result {
	if env {
		return failEnv(id, reason)
	}
	return fail(id, reason)
}

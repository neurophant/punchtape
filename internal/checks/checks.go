// Package checks — execution of check rows. A black box: a row runs
// in the LIVE working tree (the engine builds before the rows when
// the spec declares a build, and checks the tree's digest identity
// around the run); its command is an arbitrary argv (a product entry
// point, a utility, a shell — all the same to the machine); the
// outcome is observed only through universal observables: streams,
// the exit code, files. The machine knows nothing about the product
// or the environment.
package checks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/execport"
	"github.com/neurophant/punchtape/internal/fsport"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// RunEnv — the row's external byte sources: the instance's input
// directory (from-input seeds copy their bytes from there,
// file-to-file) and the submission's inbox directory (the bytes
// sample of the row's file:// reference).
type RunEnv struct {
	InputDir string
	InboxDir string
}

// Result — the outcome of one row's run: green/red and one reason
// line for the red. Env — the red is an environment or launch
// failure (not an expectation divergence): such reds do not eat the
// attempt budget — they are not product behavior. Exit — the
// observed exit code of the launched command (nil when it never
// launched); OutDigest — the digest of the observed stdout (empty
// when it never launched): the measured facts the row's memo and
// the report stand on.
type Result struct {
	Label     string
	Green     bool
	Reason    string
	Env       bool
	Exit      *int
	OutDigest string
}

// Run executes one check row in the live tree dir: the seed bytes
// are written into the tree (state setup) and leave it again when
// the row ends — between rows seeds do not live; the command's
// observations are matched against the row's expectations. One row —
// one run in one tree: determinism is a property of the recorded
// expectations, not a comparison of repeats. The two exceptions by
// head: the repeat-run witness — the idempotency family whose MEANING
// is the second run over the leftovers — and the staged-bytes probe
// (witness), whose bytes arrive through the door and run under the
// row's procedural modes (the launch mode rides the launch itself;
// the tree mode is the caller's own gate).
func Run(dir string, label string, row canon.Row, env RunEnv) (res Result) {
	marks, err := applySeed(dir, row.Seeds, env)
	if err != nil {
		return failEnv(label, err.Error())
	}
	// Seeds do not outlive their row: a fresh seed path is removed, a
	// seed over a tree file restores the file's original bytes —
	// whatever this row's outcome. A leftover is an honest env red
	// (the next row would see a world this row left behind).
	defer func() {
		if left := cleanupSeeds(dir, marks); len(left) > 0 {
			res = failEnv(label, canondata.T("check.fail.seed-left", canondata.M{
				"paths": strings.Join(left, ", "),
			}))
		}
	}()
	fields, err := row.CommandFields()
	if err != nil {
		return failEnv(label, err.Error())
	}
	if row.Repeat != nil {
		return repeatRun(dir, label, row, fields)
	}
	opts := execport.LaunchOpts{}
	if row.Launch == canon.LaunchModeNonet {
		opts.Netns = true
	}
	if row.Witness != nil {
		return witnessRun(dir, label, row, fields, env, opts)
	}
	r := newRowRunner(dir)
	// The world freezes before reading observations: everything the
	// row's command left alive (a shell — its child servers) is taken
	// down by the process group; survivors would hold resources and
	// corrupt neighboring rows. OS level, zero product knowledge.
	// The launch grants fold back with it: a file the machine made
	// executable to launch returns to its owner's mode.
	defer r.finish()

	timeout := row.Timeout
	if timeout <= 0 {
		timeout = canondata.Limit("check.timeout-default-s")
	}
	obs, reason, startFail := r.invoke(fields, timeout, opts)
	if reason != "" {
		return failClass(label, reason, startFail)
	}
	r.finish()
	exit := obs.exit
	facts := Result{Exit: &exit, OutDigest: yamlio.DigestShort([]byte(obs.stdout))}
	if red := assertRow(dir, label, row, obs, env); red != "" {
		res := fail(label, red)
		res.Exit, res.OutDigest = facts.Exit, facts.OutDigest
		return res
	}
	return Result{Label: label, Green: true, Exit: facts.Exit, OutDigest: facts.OutDigest}
}

// witnessRun — the staged-bytes probe: the probe's bytes ride the
// door (a file next to the submission in the inbox); the machine
// stages them into the live tree for the row's duration — executable,
// because a probe is a program — runs them under the row's procedural
// launch mode and compares rc/stdout with the row's want. The machine
// knows nothing about what the probe checks: the staged bytes are the
// checker, the row's expectations are the want. The probe leaves the
// tree when the row ends (a leftover is an honest env red, the same
// law as the seeds').
func witnessRun(dir, label string, row canon.Row, fields []string, env RunEnv, opts execport.LaunchOpts) (res Result) {
	ref := row.ProbeRef()
	if ref == "" {
		return failEnv(label, canondata.T("check.fail.probe-form", canondata.M{
			"given": row.Witness.Probe,
		}))
	}
	data, err := fsport.Bind.ReadFile(filepath.Join(env.InboxDir, filepath.FromSlash(ref)))
	if err != nil {
		return failEnv(label, canondata.T("check.fail.probe-missing", canondata.M{
			"path": ref,
		}))
	}
	// stage the probe into the tree: the bytes land at the probe's own
	// name, executable for the row's duration; a tree file the name
	// collided with is restored byte for byte when the row ends
	path := filepath.Join(dir, filepath.FromSlash(ref))
	if err := fsport.Bind.MkdirAll(filepath.Dir(path)); err != nil {
		return failEnv(label, canondata.T("check.fail.probe-stage", canondata.M{
			"path": ref, "err": firstLineOf(err),
		}))
	}
	orig, existed := []byte(nil), false
	if o, rerr := fsport.Bind.ReadFile(path); rerr == nil {
		orig, existed = o, true
	}
	if err := fsport.Bind.WriteFile(path, data, 0o755); err != nil {
		return failEnv(label, canondata.T("check.fail.probe-stage", canondata.M{
			"path": ref, "err": firstLineOf(err),
		}))
	}
	defer func() {
		if existed {
			if err := fsport.Bind.WriteFile(path, orig, 0o644); err != nil {
				res = failEnv(label, canondata.T("check.fail.probe-left", canondata.M{
					"paths": ref,
				}))
			}
			return
		}
		if err := fsport.Bind.Remove(path); err != nil && !os.IsNotExist(err) {
			res = failEnv(label, canondata.T("check.fail.probe-left", canondata.M{
				"paths": ref,
			}))
		}
	}()
	r := newRowRunner(dir)
	defer r.finish()
	timeout := row.Timeout
	if timeout <= 0 {
		timeout = canondata.Limit("check.timeout-default-s")
	}
	obs, reason, startFail := r.invoke(fields, timeout, opts)
	if reason != "" {
		return failClass(label, reason, startFail)
	}
	r.finish()
	exit := obs.exit
	facts := Result{Exit: &exit, OutDigest: yamlio.DigestShort([]byte(obs.stdout))}
	wantRc := 0
	if row.Rc != nil {
		wantRc = *row.Rc
	}
	if obs.exit != wantRc {
		reason := canondata.T("check.assert.rc", canondata.M{
			"want": fmt.Sprintf("%d", wantRc), "got": fmt.Sprintf("%d", obs.exit),
		})
		res := fail(label, reason)
		res.Exit, res.OutDigest = facts.Exit, facts.OutDigest
		return res
	}
	if row.Out != nil {
		if reason, proposal := compareText(row.Out, obs.stdout); reason != "" {
			res := fail(label, streamRed("stdout", row.Out, obs.stdout, reason, proposal))
			res.Exit, res.OutDigest = facts.Exit, facts.OutDigest
			return res
		}
	}
	return Result{Label: label, Green: true, Exit: facts.Exit, OutDigest: facts.OutDigest}
}

// repeatRun — the repeat-run witness: the command runs TWICE in
// immediate succession over the same tree (the seeds stand for both,
// the first run's leftovers are the second run's world — exactly the
// idempotency the long form imitated with shell chains). Green is
// agreement: both runs exit 0 and print byte-identical stdout (the
// row's out, when spelled, demands its text both times). Divergence
// is a measured red quoting both answers.
func repeatRun(dir, label string, row canon.Row, fields []string) Result {
	r := newRowRunner(dir)
	defer r.finish()
	timeout := row.Timeout
	if timeout <= 0 {
		timeout = canondata.Limit("check.timeout-default-s")
	}
	first, reason, startFail := r.invoke(fields, timeout, execport.LaunchOpts{})
	if reason != "" {
		return failClass(label, reason, startFail)
	}
	second, reason2, startFail2 := r.invoke(fields, timeout, execport.LaunchOpts{})
	if reason2 != "" {
		return failClass(label, reason2, startFail2)
	}
	exit := second.exit
	facts := Result{Exit: &exit, OutDigest: yamlio.DigestShort([]byte(second.stdout))}
	if first.exit != 0 || second.exit != 0 {
		reason := canondata.T("check.assert.rc", canondata.M{
			"want": "0", "got": fmt.Sprintf("%d/%d", first.exit, second.exit),
		})
		res := fail(label, reason)
		res.Exit, res.OutDigest = facts.Exit, facts.OutDigest
		return res
	}
	if first.stdout != second.stdout {
		res := fail(label, canondata.T("check.fail.repeat-diverge", canondata.M{
			"first": excerpt(first.stdout), "second": excerpt(second.stdout),
		}))
		res.Exit, res.OutDigest = facts.Exit, facts.OutDigest
		return res
	}
	if row.Out != nil {
		if reason, proposal := compareText(row.Out, second.stdout); reason != "" {
			res := fail(label, streamRed("stdout", row.Out, second.stdout, reason, proposal))
			res.Exit, res.OutDigest = facts.Exit, facts.OutDigest
			return res
		}
	}
	return Result{Label: label, Green: true, Exit: facts.Exit, OutDigest: facts.OutDigest}
}

// runOutcome — the observations of the row's run: the streams and
// the exit code.
type runOutcome struct {
	exit   int
	stdout string
	stderr string
}

// rowRunner — one row's command invocation inside the live tree.
// argv[0] resolution: a name with a slash is a path into the tree
// (as is); a bare name that exists as a file in the tree is launched
// through its tree path (and made executable for the row's duration
// — the right to execute is a property of the record, not of the
// bytes); anything else goes to the environment as is — a PATH
// launch. A "product entry point", a "shell" and a "utility" are all
// just argv to the machine; there is no command dictionary or
// knowledge about them.
type rowRunner struct {
	dir    string
	groups []execport.Process
	grants []modeGrant
}

// modeGrant — one launch grant to fold back after the row: the file's
// path and the mode its owner gave it. Mode bits are outside the
// digest identity (bytes are the tree's identity), but the tree is
// the executor's territory — the machine leaves it as it found it.
type modeGrant struct {
	path string
	mode os.FileMode
}

func newRowRunner(dir string) *rowRunner {
	return &rowRunner{dir: dir}
}

// resolve returns the launch form for command[0].
func (r *rowRunner) resolve(name string) (string, error) {
	if strings.Contains(name, "/") {
		return name, nil
	}
	entry := filepath.Join(r.dir, filepath.FromSlash(name))
	info, err := fsport.Bind.Stat(entry)
	if err != nil || info.IsDir() {
		return name, nil // PATH launch: the environment knows its tools
	}
	if info.Mode()&0o111 == 0 {
		// The right to execute is a property of the launch, not of the
		// bytes: the machine grants it for the row's duration and
		// folds it back when the row ends.
		if err := fsport.Bind.Chmod(entry, info.Mode().Perm()|0o111); err != nil {
			return "", err
		}
		r.grants = append(r.grants, modeGrant{path: entry, mode: info.Mode().Perm()})
	}
	// A relative launch: the run directory path is the machine's
	// choice — a tool that prints its working directory must not
	// become "non-deterministic" because of that choice.
	return "./" + name, nil
}

// killGroups removes everything alive left by the row's command: it
// ran in its own process group — surviving children (a shell with
// servers) do not outlive the row.
func (r *rowRunner) killGroups() {
	for _, proc := range r.groups {
		proc.KillGroup()
	}
	r.groups = nil
}

// finish — the row's world folds back: the process groups die, the
// launch grants return to their owner's modes.
func (r *rowRunner) finish() {
	r.killGroups()
	r.restoreModes()
}

// restoreModes — the launch grants in reverse order (a path granted
// twice returns to the same mode; idempotent by construction).
func (r *rowRunner) restoreModes() {
	for i := len(r.grants) - 1; i >= 0; i-- {
		g := r.grants[i]
		_ = fsport.Bind.Chmod(g.path, g.mode)
	}
	r.grants = nil
}

// invoke — the row's command with its timeout and launch disciplines;
// the reason is non-empty when the call did not happen (timeout,
// process death). startFail — an environment failure: the process did
// not start or an environment class; a launch timeout is not the
// environment (the behavior may have hung; the attempt stays an
// attempt).
func (r *rowRunner) invoke(command []string, timeoutSec int, opts execport.LaunchOpts) (runOutcome, string, bool) {
	argv0, err := r.resolve(command[0])
	if err != nil {
		return runOutcome{}, err.Error(), true
	}
	invocation := strings.Join(command, " ")
	// The process runs in its own process group (the port's
	// discipline): the command's descendants are taken down by the
	// group together with it.
	proc, err := execport.Bind.StartOpts(append([]string{argv0}, command[1:]...), r.dir, nil, opts)
	if err != nil {
		if class := EnvClass(err); class != "" {
			return runOutcome{}, canondata.EnvDiag(class, canondata.M{
				"command": invocation,
			}), true
		}
		return runOutcome{}, canondata.T("check.fail.start", canondata.M{
			"command": invocation, "err": err.Error(),
		}), true
	}
	r.groups = append(r.groups, proc)
	out := proc.Wait(time.Duration(timeoutSec) * time.Second)
	if out.TimedOut {
		return runOutcome{}, canondata.EnvDiag("command-timeout", canondata.M{
			"sec": fmt.Sprintf("%d", timeoutSec), "command": invocation,
		}), false
	}
	if out.WaitErr != "" {
		return runOutcome{}, canondata.T("check.fail.wait", canondata.M{
			"err": out.WaitErr, "command": invocation,
		}), true
	}
	return runOutcome{exit: out.Exit, stdout: out.Stdout, stderr: out.Stderr}, "", false
}

// assertRow — the row's expectations against the observations: an
// empty string when everything holds, otherwise the first failing
// reason line (with the measured value — a proposal the executor can
// copy when the divergence is terminator-shaped).
func assertRow(dir, label string, row canon.Row, obs runOutcome, env RunEnv) string {
	wantRc := 0
	if row.Rc != nil {
		wantRc = *row.Rc
	}
	if obs.exit != wantRc {
		return canondata.T("check.assert.rc", canondata.M{
			"want": fmt.Sprintf("%d", wantRc), "got": fmt.Sprintf("%d", obs.exit),
		})
	}
	for _, pair := range []struct {
		name   string
		want   *canon.TextExpect
		actual string
	}{
		{"stdout", row.Out, obs.stdout},
		{"stderr", row.Err, obs.stderr},
	} {
		if pair.want == nil {
			continue
		}
		if reason, proposal := compareText(pair.want, pair.actual); reason != "" {
			return streamRed(pair.name, pair.want, pair.actual, reason, proposal)
		}
	}
	for _, st := range row.State {
		if st.Text != nil {
			data, err := fsport.Bind.ReadFile(filepath.Join(dir, filepath.FromSlash(st.Path)))
			if err != nil {
				return canondata.T("check.assert.file-missing", canondata.M{"path": st.Path})
			}
			if reason, _ := compareText(st.Text, string(data)); reason != "" {
				return canondata.T("check.assert.state-not-equal", canondata.M{
					"path": st.Path, "want": quotedScalar(st.Text.Form()),
					"reason": reason,
				})
			}
			continue
		}
		want, err := fsport.Bind.ReadFile(filepath.Join(dir, filepath.FromSlash(st.FromRun)))
		if err != nil {
			return canondata.T("check.assert.file-missing", canondata.M{"path": st.FromRun})
		}
		got, err := fsport.Bind.ReadFile(filepath.Join(dir, filepath.FromSlash(st.Path)))
		if err != nil {
			return canondata.T("check.assert.file-missing", canondata.M{"path": st.Path})
		}
		if string(want) != string(got) {
			return canondata.T("check.assert.bytes-differ", canondata.M{
				"path": st.Path, "other": st.FromRun,
				"diff": divergenceLine(string(want), string(got)),
			})
		}
	}
	if ref := row.BytesRef(); ref != "" {
		sample, err := fsport.Bind.ReadFile(filepath.Join(env.InboxDir, filepath.FromSlash(ref)))
		if err != nil {
			return canondata.T("check.assert.sample-missing", canondata.M{"path": ref})
		}
		if string(sample) != obs.stdout {
			return canondata.T("check.assert.bytes-differ", canondata.M{
				"path": "stdout", "other": ref,
				"diff": divergenceLine(string(sample), obs.stdout),
			})
		}
	}
	return ""
}

// streamRed — the red line of a stream expectation: the reason with
// the measured value; a terminator-shaped divergence carries the
// copy-ready expectation value with the MEASURED bytes — a PROPOSAL;
// the executor repairs the row with it, the red stays red until
// their own submission changes it.
func streamRed(name string, want *canon.TextExpect, actual, reason, proposal string) string {
	if proposal != "" {
		reason += canondata.T("check.assert.terminator-proposal", canondata.M{"value": proposal})
	}
	return canondata.T("check.fail.observed", canondata.M{
		"reason": reason,
		"stream": name,
		"want":   quotedScalar(want.Form()),
		"stdout": excerpt(actual),
	})
}

// compareText — one text expectation against the actual text: exact
// (EOL-tolerant: exactly one trailing newline is ignored on either
// side) or a fragment that must occur. The proposal is non-empty
// when the divergence is exactly one trailing newline.
func compareText(want *canon.TextExpect, actual string) (reason, proposal string) {
	if want.Contains {
		if !strings.Contains(actual, want.Text) {
			return canondata.T("check.assert.not-contains", canondata.M{
				"value": quotedScalar(want.Text), "actual": brief(actual),
			}), ""
		}
		return "", ""
	}
	w, a := trimOneNL(want.Text), trimOneNL(actual)
	if w == a {
		return "", ""
	}
	if want.Text == actual+"\n" || want.Text+"\n" == actual {
		return canondata.T("check.assert.not-equal", canondata.M{
			"value": quotedScalar(want.Text), "actual": brief(actual),
			"diff": divergenceLine(want.Text, actual),
		}), quotedScalar(actual)
	}
	return canondata.T("check.assert.not-equal", canondata.M{
		"value": quotedScalar(want.Text), "actual": brief(actual),
		"diff": divergenceLine(want.Text, actual),
	}), ""
}

// trimOneNL — at most one trailing newline away.
func trimOneNL(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s[:len(s)-1]
	}
	return s
}

// quotedScalar — the value as a quoted string scalar with the byte
// discipline of the machine's renders: a byte-exact picture in one
// line, whatever the bytes are.
func quotedScalar(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	return `"` + yamlio.EscapeControls(v) + `"`
}

// seedMark — one seed's footprint on the live tree: the path written
// and, when that path was a tree file, its original bytes (a seed
// over a tree file honestly replaces it for the row's duration — the
// cleanup restores the tree; the live world has no copy to throw
// away), plus the directories the machine created to place the seed
// (removed again when the row ends and they are empty — the seed
// leaves nothing behind, not even an empty folder a product might
// read as a candidate of its own).
type seedMark struct {
	path    string
	existed bool
	orig    []byte
	dirs    []string
}

// applySeed — seeding a row's files into the live tree: inline text
// is written as bytes; a from-input seed copies the named input
// file's bytes file-to-file. The marks carry what the cleanup must
// restore. Writing bytes is the machine's only file operation on the
// run world; it knows nothing about the content.
func applySeed(dir string, seeds []canon.Seed, env RunEnv) ([]seedMark, error) {
	marks := make([]seedMark, 0, len(seeds))
	for _, s := range seeds {
		var content []byte
		if s.FromInput != "" {
			data, err := fsport.Bind.ReadFile(filepath.Join(env.InputDir, filepath.FromSlash(s.FromInput)))
			if err != nil {
				return nil, fmt.Errorf("%s", canondata.T("check.fail.seed-input", canondata.M{
					"path": s.FromInput, "err": firstLineOf(err),
				}))
			}
			content = data
		} else {
			content = []byte(s.Text)
		}
		path := filepath.Join(dir, filepath.FromSlash(s.Path))
		parent := filepath.Dir(path)
		// the directories that do not exist yet are the machine's own
		// footprint; the walk starts at the seed's parent and goes
		// upward, so the list is deep-first - removed in this order
		var created []string
		for p := parent; len(p) > len(dir); p = filepath.Dir(p) {
			if _, err := fsport.Bind.Stat(p); err != nil {
				created = append(created, p)
				continue
			}
			break
		}
		if err := fsport.Bind.MkdirAll(parent); err != nil {
			return nil, fmt.Errorf("%s", canondata.T("check.fail.seed", canondata.M{"path": s.Path, "err": err.Error()}))
		}
		mark := seedMark{path: path, dirs: created}
		if orig, rerr := fsport.Bind.ReadFile(path); rerr == nil {
			mark.existed, mark.orig = true, orig
		}
		if err := fsport.Bind.WriteFile(path, content, 0o644); err != nil {
			return nil, fmt.Errorf("%s", canondata.T("check.fail.seed", canondata.M{"path": s.Path, "err": err.Error()}))
		}
		marks = append(marks, mark)
	}
	return marks, nil
}

// cleanupSeeds — the seeds' self-removal after the row: a path the
// tree never had is removed; a path that replaced a tree file is
// restored to its original bytes (mode untouched — a rewrite never
// changes an existing file's permissions). The return lists the
// seed paths that would not leave — an honest env red for the row:
// the next row must not run in this row's leftovers.
func cleanupSeeds(dir string, marks []seedMark) []string {
	var left []string
	rel := func(path string) string {
		return strings.TrimPrefix(filepath.ToSlash(path), filepath.ToSlash(dir)+"/")
	}
	// all seed files leave FIRST: a created directory can only go
	// when no sibling seed still lives in it (per-mark removal lost
	// exactly there - the first mark's Remove(tpl) found the second
	// mark's file still inside)
	var created []string
	seen := map[string]bool{}
	for _, m := range marks {
		if m.existed {
			if err := fsport.Bind.WriteFile(m.path, m.orig, 0o644); err != nil {
				left = append(left, rel(m.path))
			}
		} else if err := fsport.Bind.Remove(m.path); err != nil && !os.IsNotExist(err) {
			left = append(left, rel(m.path))
		}
		for _, d := range m.dirs {
			if !seen[d] {
				seen[d] = true
				created = append(created, d)
			}
		}
	}
	// the machine-created directories go deep-first (the collection
	// walks from each seed's parent upward); a non-empty one (the
	// product wrote there) stays - the tree control sees those files,
	// the seed's own footprint is still gone
	for _, d := range created {
		_ = fsport.Bind.Remove(d)
	}
	return left
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

// firstLineOf — an error's first line.
func firstLineOf(err error) string {
	line := err.Error()
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	return strings.TrimSpace(line)
}

func fail(label, reason string) Result {
	return Result{Label: label, Green: false, Reason: reason}
}

// failEnv — a red for an environment/launch failure: not an
// expectation divergence; does not eat the attempt budget.
func failEnv(label, reason string) Result {
	return Result{Label: label, Green: false, Reason: reason, Env: true}
}

// failClass — a red with a class by source: the start did not
// happen — the environment; a timeout and everything that reached
// observation — behavior.
func failClass(label, reason string, env bool) Result {
	if env {
		return failEnv(label, reason)
	}
	return fail(label, reason)
}

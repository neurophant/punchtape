// Background rehearsal of the full suite: after a code/fix
// submission the machine runs all instance checks against the digest
// of the current surface while the hand prepares the next submission;
// the ready report is given in the next submission's reply (the
// surface unchanged — the outcomes are valid). The run is
// deterministic and isolated (each check gets its own clean
// directory), the time goes to the machine wall of the ledger, not
// into the hand's ceremony. The report lives in the cache, not the
// canon: verdicts and transactions do not depend on the background
// run's race.
package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/checks"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// BgCheckOut — the outcome of one background-run check.
type BgCheckOut struct {
	ID       string `yaml:"id"`
	Scenario string `yaml:"scenario"`
	Outcome  string `yaml:"outcome"`
	Reason   string `yaml:"reason,omitempty"`
	// Env — a red of environment/launch failure (not an expectation
	// mismatch): such outcomes eat no attempt budget.
	Env bool `yaml:"env,omitempty"`
}

// BgReport — the report of a full-suite background run.
type BgReport struct {
	Digest string       `yaml:"digest"`
	WallMs int64        `yaml:"wall-ms"`
	Checks []BgCheckOut `yaml:"checks"`
}

// bgSuiteDir — the directory of background-run reports in the instance cache.
func bgSuiteDir(workdir string) string {
	return filepath.Join(workdir, ".punchtape", "cache", "bg-suite")
}

// bgStopPath — the rehearsal stop marker: a live run (submission
// gates, probes) asks the background rehearsal to yield — it exits
// after the current check without starting the next one. The
// rehearsal is advisory: its report gates nothing; yielding to the
// live run is the right order.
func bgStopPath(workdir string) string {
	return filepath.Join(workdir, ".punchtape", "cache", "bg-stop")
}

// stopRehearsal marks the background rehearsal for stopping (to
// yield to a live run).
func stopRehearsal(workdir string) {
	_ = os.MkdirAll(filepath.Dir(bgStopPath(workdir)), 0o755)
	_ = os.WriteFile(bgStopPath(workdir), []byte("stop"), 0o644)
}

// resumeRehearsal removes the stop marker — the live run has finished.
func resumeRehearsal(workdir string) {
	_ = os.Remove(bgStopPath(workdir))
}

// rehearsalStopped — whether the stop marker is present.
func rehearsalStopped(workdir string) bool {
	_, err := os.Stat(bgStopPath(workdir))
	return err == nil
}

// surfaceDigest — the digest of the rehearsal's current state: the
// surface (the declared artifact — a file or tree — or the files by
// command names), THE RECIPE DIGEST (the surface's truth depends on
// how it is built: a conventions change without an artifact change
// changes the truth) and the canonical definitions of all checks. A
// change in any term invalidates the previous report.
func surfaceDigest(e *Engine, all []canon.Check) (string, bool) {
	names := map[string]bool{}
	for _, c := range all {
		if c.When != nil && len(c.When.Command) > 0 {
			names[c.When.Command[0]] = true
		}
	}
	if len(names) == 0 {
		return "", false
	}
	digest := ""
	if artifact := e.surfaceArtifact(); artifact != "" {
		d, ok := e.pathDigest(artifact)
		if !ok {
			return "", false // no surface — nothing to rehearse
		}
		digest += artifact + "\x00" + d + "\x00"
	} else {
		for _, name := range sortedKeys(names) {
			d, ok := e.digestOf(name)
			if !ok {
				return "", false // no surface — nothing to rehearse
			}
			digest += name + "\x00" + d + "\x00"
		}
	}
	digest += "conventions\x00" + e.Store.ConventionsDigest() + "\x00"
	sorted := make([]canon.Check, len(all))
	copy(sorted, all)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	for _, c := range sorted {
		data, err := yamlio.Marshal(c)
		if err != nil {
			return "", false
		}
		digest += c.ID + "\x00" + yamlio.Digest(data) + "\x00"
	}
	return yamlio.Digest([]byte(digest)), true
}

// backfillCheckProven — legacy normalization applied wherever raw
// checks are digested or compared: without the stored proven field a
// current green is proven as a fact; a current red without history is
// conservatively protected (the unknown is not silently weakened).
// The signature of a check set must not depend on which side of a
// load it was computed on.
func backfillCheckProven(checks []canon.Check) []canon.Check {
	out := make([]canon.Check, len(checks))
	copy(out, checks)
	for i := range out {
		if out[i].Outcome == canon.CheckGreen {
			out[i].Proven = true
		}
	}
	return out
}

func backfillCheckProvenOne(c canon.Check) canon.Check {
	if c.Outcome == canon.CheckGreen {
		c.Proven = true
	}
	return c
}

// pathDigest — the digest of a surface file or tree (a sorted
// walk): an artifact of any shape folds into one signature. The
// path is slash-style from the project root; reads go through the
// digest cache with newline normalization.
func (e *Engine) pathDigest(relRoot string) (string, bool) {
	abs := filepath.Join(e.Workdir, filepath.FromSlash(relRoot))
	info, err := os.Stat(abs)
	if err != nil {
		return "", false
	}
	if !info.IsDir() {
		return e.digestOf(relRoot)
	}
	// The machine's own service zone is not part of the surface: an
	// artifact covering the project root must not carry verification
	// state in the digest — otherwise drift on every transaction with
	// an unchanged product. The zone is excluded only for the root
	// artifact: a nested directory with the same name is product
	// data, the machine does not touch it.
	var serviceDir string
	if relRoot == "." {
		serviceDir = canon.Dir(e.Workdir)
	}
	rels := []string{}
	_ = filepath.Walk(abs, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if p == serviceDir && fi.IsDir() {
			return filepath.SkipDir
		}
		if !fi.Mode().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(abs, p)
		if err != nil {
			return nil
		}
		// Machine renders at the root of the root artifact are the
		// same service zone: the machine recognizes ITS OWN records
		// by its first-line marker (the render writers' policy); a
		// file with the same name without the marker is product data,
		// the digest carries it honestly. Nested files with the same
		// names are always product.
		if relRoot == "." && filepath.ToSlash(filepath.Dir(rel)) == "." && machineOwnedRender(filepath.Join(abs, rel)) {
			return nil
		}
		rels = append(rels, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(rels)
	acc := ""
	for _, rel := range rels {
		d, ok := e.digestOf(relRoot + "/" + rel)
		if !ok {
			return "", false
		}
		acc += rel + "\x00" + d + "\x00"
	}
	return yamlio.Digest([]byte(acc)), true
}

// machineOwnedRender — the file carries the machine-written marker
// of one of the machine's renders (spec, summary, changelog, agent
// reference): the writer with this marker regenerates the file, a
// foreign one without the marker is not touched — the same ownership
// boundary the renders are served by.
func machineOwnedRender(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	head := make([]byte, 96)
	n, _ := f.Read(head)
	line := string(head[:n])
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	for _, marker := range []string{specMarker, summaryMarker, changelogMarker, refMarker} {
		if line == marker {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// RunBackgroundSuite — the background process entry point: runs all
// instance checks against the current surface and atomically writes
// the report under the digest. The digest is reconciled before and
// after: the surface changed — the report is not written (stale by
// definition). The background process writes no journal states — the
// writer lock is not taken (no race with the hand's verbs); an
// unfinished transaction — the rehearsal is cancelled: the state is
// still being built. The stop marker is cleaned at start (a lingering
// route after a live-run failure) and checked before every check: a
// live run outweighs the rehearsal, the current check is finished,
// the next one is not started.
func RunBackgroundSuite(workdir string) error {
	store, exists, err := canon.Open(workdir)
	if err != nil || !exists {
		return err
	}
	resumeRehearsal(workdir)
	if _, err := os.Stat(filepath.Join(store.Root(), "cache", "pending-transaction")); err == nil {
		return nil
	}
	e, err := openEngine(workdir, store)
	if err != nil {
		return err
	}
	all, err := e.Store.Checks()
	if err != nil {
		return err
	}
	d1, ok := surfaceDigest(e, all)
	if !ok {
		return nil
	}
	started := time.Now()
	report := BgReport{Digest: d1}
	for _, c := range all {
		if rehearsalStopped(workdir) {
			return nil
		}
		res := checks.Run(e.Workdir, e.surfaceArtifact(), c, checks.Normal)
		out := canon.CheckRed
		reason := res.Reason
		if res.Green {
			out, reason = canon.CheckGreen, ""
		}
		report.Checks = append(report.Checks, BgCheckOut{
			ID: c.ID, Scenario: c.Scenario, Outcome: out, Reason: reason, Env: res.Env,
		})
	}
	report.WallMs = time.Since(started).Milliseconds()
	d2, ok := surfaceDigest(e, all)
	if !ok || d1 != d2 {
		return nil
	}
	data, err := yamlio.Marshal(report)
	if err != nil {
		return err
	}
	data = append([]byte(canon.CacheHeader()+"\n"), data...)
	if err := os.MkdirAll(bgSuiteDir(workdir), 0o755); err != nil {
		return err
	}
	return yamlio.WriteAtomic(filepath.Join(bgSuiteDir(workdir), d1+".yamll"), data)
}

// LoadBgReport — the ready report under the digest, if there is one.
func LoadBgReport(workdir, digest string) (*BgReport, bool) {
	if digest == "" {
		return nil, false
	}
	data, err := os.ReadFile(filepath.Join(bgSuiteDir(workdir), digest+".yamll"))
	if err != nil {
		return nil, false
	}
	var r BgReport
	if err := yamlio.DecodeStrict(data, &r); err != nil || r.Digest != digest {
		return nil, false
	}
	return &r, true
}

// SpawnBackgroundSuite starts the background rehearsal as a separate
// detached process: the submission's reply is already assembled, the
// hand works on the next one — the machine runs the suite in
// parallel. Zombie reaping is done by the init of the environment
// the machine runs in.
func SpawnBackgroundSuite(workdir string) {
	self, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(self, "bg-suite", workdir)
	cmd.Dir = workdir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err == nil {
		go func() { _ = cmd.Wait() }()
	}
}

// bgReplayOutcome — a check's outcome from the ready report, if the
// report covers it; ok=false — the check is not in the report, run
// it live.
func (r *BgReport) bgReplayOutcome(check canon.Check) (green bool, reason string, env bool, ok bool) {
	for _, c := range r.Checks {
		if c.ID == check.ID {
			return c.Outcome == canon.CheckGreen, c.Reason, c.Env, true
		}
	}
	return false, "", false, false
}

// bgSummaryLine — the reply line about the background rehearsal.
func (r *BgReport) bgSummaryLine() string {
	green := 0
	for _, c := range r.Checks {
		if c.Outcome == canon.CheckGreen {
			green++
		}
	}
	return canondata.T("bgsuite.summary-line", canondata.M{
		"total":  fmt.Sprintf("%d", len(r.Checks)),
		"green":  fmt.Sprintf("%d", green),
		"red":    fmt.Sprintf("%d", len(r.Checks)-green),
		"digest": r.Digest[:12],
		"wall":   fmt.Sprintf("%d", r.WallMs),
	})
}

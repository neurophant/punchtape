// Suite auto-extension: from every executable scenario the machine
// deterministically derives variants — empty state, operation repeat
// (restart), zero and negation of every numeric argument, neighbors
// of every numeric argument, removal of every seed key — a budget of
// at most eight per scenario. Derivatives are not spec scenarios:
// they do not enter the canon, traceability and the scenario count
// do not count them, they are marked with a separate line in the
// verdict. The derivative invariant is robustness: clean success or
// failure (exit code 0/1).
package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/checks"
	"github.com/neurophant/punchtape/internal/compiler"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// DerivedOut — the outcome of one derivative.
type DerivedOut struct {
	Parent string `yaml:"parent"`
	Kind   string `yaml:"kind"`
	Green  bool   `yaml:"green"`
	Reason string `yaml:"reason,omitempty"`
}

// DerivedReport — the auto-extension report; cached under the instance digest.
type DerivedReport struct {
	Digest string       `yaml:"digest"`
	Outs   []DerivedOut `yaml:"outs"`
}

// robustRun — the robustness invariant of one command: the
// scenario's setup commands (pre) and the seed go before it.
func robustRun(workdir, artifact string, seed []canon.Seed, materials []canon.Material, pre [][]string, command []string, timeoutSec int) (bool, string) {
	cap, err := checks.CaptureSeq(workdir, artifact, seed, materials, pre, command, timeoutSec)
	if err != nil {
		return false, err.Error()
	}
	if cap.ExitCode != 0 && cap.ExitCode != 1 {
		return false, canondata.T("derive.reason.exit", canondata.M{"code": fmt.Sprintf("%d", cap.ExitCode)})
	}
	return true, ""
}

// deriveVariants — the deterministic list of one scenario's
// derivatives in a fixed order, cut by the budget.
func deriveVariants(sc canon.Scenario) []struct {
	kind string
	seed []canon.Seed
	cmd  []string
} {
	if sc.When == nil || len(sc.When.Command) == 0 {
		return nil
	}
	timeout := sc.When.TimeoutSec
	_ = timeout
	var out []struct {
		kind string
		seed []canon.Seed
		cmd  []string
	}
	add := func(kind string, seed []canon.Seed, cmd []string) {
		out = append(out, struct {
			kind string
			seed []canon.Seed
			cmd  []string
		}{kind, seed, cmd})
	}
	cmd := sc.When.Command
	// empty state
	add(canondata.T("derive.kind.empty-state"), nil, cmd)
	// operation repeat: a restart right after it — the state survives
	// the second launch without a crash
	add(canondata.T("derive.kind.restart-rerun"), sc.Seed, cmd)
	// numeric arguments: zero, negation, boundary neighbors
	for i := 1; i < len(cmd); i++ {
		if _, err := strconv.Atoi(cmd[i]); err != nil {
			continue
		}
		v, _ := strconv.Atoi(cmd[i])
		with := func(x int) []string {
			c := append([]string{}, cmd...)
			c[i] = strconv.Itoa(x)
			return c
		}
		add(canondata.T("derive.kind.zero-arg", canondata.M{"num": fmt.Sprintf("%d", i)}), sc.Seed, with(0))
		add(canondata.T("derive.kind.neg-arg", canondata.M{"num": fmt.Sprintf("%d", i)}), sc.Seed, with(-v))
		add(canondata.T("derive.kind.edge-minus", canondata.M{"num": fmt.Sprintf("%d", i)}), sc.Seed, with(v-1))
		add(canondata.T("derive.kind.edge-plus", canondata.M{"num": fmt.Sprintf("%d", i)}), sc.Seed, with(v+1))
	}
	if len(out) > canondata.Limit("derive.budget") {
		out = out[:canondata.Limit("derive.budget")]
	}
	return out
}

// runDerived — runs the derivatives of all the instance's executable scenarios.
func runDerived(e *Engine, scns []canon.Scenario, digest string) *DerivedReport {
	rep := &DerivedReport{Digest: digest}
	ordered := make([]canon.Scenario, len(scns))
	copy(ordered, scns)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	for _, sc := range ordered {
		if compiler.DegreeOf(sc).Level != compiler.DegreeExecutable {
			continue
		}
		timeout := 10
		if sc.When != nil && sc.When.TimeoutSec > 0 {
			timeout = sc.When.TimeoutSec
		}
		for _, v := range deriveVariants(sc) {
			started := time.Now()
			green, reason := robustRun(e.Workdir, e.surfaceArtifact(), v.seed, sc.Materials, sc.Pre, v.cmd, timeout)
			if v.kind == canondata.T("derive.kind.restart-rerun") {
				// restart: the first run seeds the state, the second must
				// pass over it without a crash
				cap1, err := checks.CaptureSeq(e.Workdir, e.surfaceArtifact(), v.seed, sc.Materials, sc.Pre, v.cmd, timeout)
				if err == nil {
					seed2 := []canon.Seed{}
					paths := make([]string, 0, len(cap1.Files))
					for p := range cap1.Files {
						paths = append(paths, p)
					}
					sort.Strings(paths)
					for _, p := range paths {
						content := cap1.Files[p]
						seed2 = append(seed2, canon.Seed{
							Path: p, Content: content,
						})
					}
					green, reason = robustRun(e.Workdir, e.surfaceArtifact(), seed2, sc.Materials, sc.Pre, v.cmd, timeout)
				} else {
					green, reason = false, err.Error()
				}
			}
			// Derivative runs are honest verification work:
			// the verdict latency line subtracts them from the submission wall.
			e.lastSubmitRunsMs += time.Since(started).Milliseconds()
			rep.Outs = append(rep.Outs, DerivedOut{
				Parent: sc.ID, Kind: v.kind, Green: green, Reason: reason,
			})
		}
	}
	return rep
}

// derivedLine — the verdict line about auto-extension (the cache is
// under the instance digest; instances without a surface are not
// extended).
func (e *Engine) derivedLine(state canon.State) string {
	if state.Stage != canon.StageDeliver {
		return canondata.T("derive.line.not-run")
	}
	scns, err := e.Store.Scenarios()
	if err != nil {
		return canondata.T("derive.line.unavailable")
	}
	chk, err := e.Store.Checks()
	if err != nil {
		return canondata.T("derive.line.unavailable")
	}
	d, ok := surfaceDigest(e, chk)
	if !ok {
		return canondata.T("derive.line.none")
	}
	return loadOrRunDerived(e, scns, d).verdictLine()
}

// loadOrRunDerived — the auto-extension report under the instance
// digest: cache before the run, the run writes the cache.
func loadOrRunDerived(e *Engine, scns []canon.Scenario, d string) *DerivedReport {
	cachePath := filepath.Join(e.Store.Root(), "cache", "derived", d+".yamll")
	if data, err := os.ReadFile(cachePath); err == nil {
		var rep DerivedReport
		if yamlio.DecodeStrict(data, &rep) == nil && rep.Digest == d {
			return &rep
		}
	}
	rep := runDerived(e, scns, d)
	if data, err := yamlio.Marshal(rep); err == nil {
		_ = os.MkdirAll(filepath.Dir(cachePath), 0o755)
		_ = yamlio.WriteAtomic(cachePath, data)
	}
	return rep
}

func (r *DerivedReport) verdictLine() string {
	green, red, _ := r.counts()
	return canondata.T("derive.line.verdict", canondata.M{
		"total":  fmt.Sprintf("%d", len(r.Outs)),
		"budget": fmt.Sprintf("%d", canondata.Limit("derive.budget")),
		"green":  fmt.Sprintf("%d", green),
		"red":    fmt.Sprintf("%d", red),
	})
}

// counts — the report's green/red/total derivative runs; the verdict
// and the human-readable summary count the same numbers.
func (r *DerivedReport) counts() (green, red, total int) {
	for _, o := range r.Outs {
		if o.Green {
			green++
		} else {
			red++
		}
	}
	return green, red, len(r.Outs)
}

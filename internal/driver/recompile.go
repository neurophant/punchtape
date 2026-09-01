// The recompile service mode: an automated double compilation gate.
// One spec — two independent implementations: the machine restores
// the canon's spec into a delta, seeds a fresh instance through the
// regular submissions and spins its loop with a second executor
// (zero shared memory — the executor substitution pattern). The gate
// is green when both compilations pass the same suite. The
// cross-model axis (two different models at once) is the owner's
// separate decision.
package driver

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/engine"
	"github.com/neurophant/punchtape/internal/ledger"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Recompile runs the gate: a fresh instance receives the same spec
// (the machine assembles it from the canon as a delta — the single
// write form), a second executor implements it independently, the
// suites are compared by scenario. The outcome is written to the
// original's ledger as a gate event; the original's journal gets a
// minimal transaction — the verdict refreezes with the gate line.
// Return code: 0 — gate green, 1 — red, 2 — failure.
func Recompile(workdir, shape, executor string, limit int, stdout io.Writer) int {
	eng, exists, err := openWithRetry(workdir)
	if err != nil {
		fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail", canondata.M{"err": err.Error()}))
		return 2
	}
	if !exists {
		fmt.Fprintln(stdout, canondata.T("driver.fail.no-instance"))
		return 2
	}
	state, err := eng.Store.State()
	if err != nil {
		fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail", canondata.M{"err": err.Error()}))
		return 2
	}
	if state.Stage != canon.StageDeliver {
		fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail.stage",
			canondata.M{"stage": state.Stage}))
		return 2
	}

	// A fresh twin instance: the same spec, zero shared memory with
	// the original.
	freshDir := strings.TrimSuffix(workdir, "/") + ".recompile"
	if err := os.RemoveAll(freshDir); err != nil {
		fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail", canondata.M{"err": err.Error()}))
		return 2
	}
	if err := seedFreshInstance(freshDir, eng); err != nil {
		fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail.seed", canondata.M{"err": err.Error()}))
		return 2
	}

	fmt.Fprintf(stdout, "%s\n", canondata.T("driver.recompile.arm", canondata.M{"dir": freshDir}))
	if code := Run(freshDir, shape, executor, limit, io.Discard); code != 0 {
		if err := recordGate(eng, freshDir, false, canondata.T("driver.recompile.no-verdict")); err != nil {
			fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail.ledger", canondata.M{"err": err.Error()}))
			return 2
		}
		fmt.Fprintln(stdout, canondata.T("driver.recompile.gate-red"))
		return 1
	}

	green, reason, err := compareSuites(eng.Workdir, freshDir)
	if err != nil {
		fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail", canondata.M{"err": err.Error()}))
		return 2
	}
	outcome := "RED"
	if green {
		outcome = "GREEN"
	}
	if err := recordGate(eng, freshDir, green, reason); err != nil {
		fmt.Fprintf(stdout, "%s\n", canondata.T("driver.fail.ledger", canondata.M{"err": err.Error()}))
		return 2
	}
	fmt.Fprintf(stdout, "%s\n", canondata.T("driver.recompile.gate-outcome",
		canondata.M{"outcome": outcome, "reason": reason}))
	if !green {
		return 1
	}
	return 0
}

// seedFreshInstance seeds a fresh instance with the same spec:
// intent — a machine fact (clarifications live in the table cycle
// and are not reproduced in recompilation: the choices are already
// baked into the final spec); fixture declarations and the spec
// itself are restored from the canon as deltas.
func seedFreshInstance(dir string, eng *engine.Engine) error {
	fresh, err := engine.Init(dir)
	if err != nil {
		return err
	}
	head := eng.HeadDigest()
	submit := func(d any) error {
		data, err := yamlio.Marshal(d)
		if err != nil {
			return err
		}
		reply, code := fresh.SubmitAuto(data)
		if code != 0 {
			return fmt.Errorf("rejected: %s", firstLineOf(reply))
		}
		return nil
	}
	if err := submit(delta.Intent{
		Kind:          "intent",
		SubmissionKey: "intent-recompile-" + head,
		Text:          canondata.T("driver.recompile.intent-text"),
	}); err != nil {
		return fmt.Errorf("intent: %w", err)
	}
	sp, err := specFromCanon(eng)
	if err != nil {
		return err
	}
	if err := submit(sp); err != nil {
		return fmt.Errorf("spec: %w", err)
	}
	return nil
}

// specFromCanon restores a spec delta from the canon: every
// requirement and scenario as an add operation, in stable order.
func specFromCanon(eng *engine.Engine) (delta.Spec, error) {
	reqs, err := eng.Store.Requirements()
	if err != nil {
		return delta.Spec{}, err
	}
	scns, err := eng.Store.Scenarios()
	if err != nil {
		return delta.Spec{}, err
	}
	sort.Slice(reqs, func(i, j int) bool { return reqs[i].ID < reqs[j].ID })
	sort.Slice(scns, func(i, j int) bool { return scns[i].ID < scns[j].ID })
	sp := delta.Spec{Kind: "spec", SubmissionKey: "spec-recompile-" + eng.HeadDigest()}
	for _, r := range reqs {
		deps := r.Dependencies
		if deps == nil {
			deps = []string{}
		}
		sp.Operations = append(sp.Operations, delta.SpecOperation{AddRequirement: &delta.AddRequirement{
			ID: r.ID, Formulation: r.Formulation, Dependencies: deps,
		}})
	}
	for _, sc := range scns {
		seed := sc.Seed
		if seed == nil {
			seed = []canon.Seed{}
		}
		materials := sc.Materials
		if materials == nil {
			materials = []canon.Material{}
		}
		then := sc.Then
		if then == nil {
			then = []canon.Assertion{}
		}
		sp.Operations = append(sp.Operations, delta.SpecOperation{AddScenario: &delta.AddScenario{
			ID: sc.ID, Requirement: sc.Requirement, Summary: sc.Summary,
			Seed: seed, Materials: materials, Pre: sc.Pre,
			When: sc.When, Then: then, Prose: sc.Prose,
		}})
	}
	return sp, nil
}

// compareSuites compares the instances' suites by scenario: the
// scenario sets match and every check is green in both instances.
func compareSuites(originalDir, freshDir string) (bool, string, error) {
	origEng, _, err := openWithRetry(originalDir)
	if err != nil {
		return false, "", err
	}
	freshEng, exists, err := openWithRetry(freshDir)
	if err != nil || !exists {
		return false, "", fmt.Errorf("%s", canondata.T("driver.recompile.instance-missing"))
	}
	orig := outcomesByScenario(origEng)
	fresh := outcomesByScenario(freshEng)
	if len(orig) == 0 {
		return false, canondata.T("driver.recompile.suite-empty"), nil
	}
	var missing, differing []string
	for scn := range orig {
		if _, ok := fresh[scn]; !ok {
			missing = append(missing, scn)
			continue
		}
		if orig[scn] != canon.CheckGreen || fresh[scn] != canon.CheckGreen {
			differing = append(differing, scn)
		}
	}
	sort.Strings(missing)
	sort.Strings(differing)
	switch {
	case len(missing) > 0:
		return false, canondata.T("driver.recompile.lost-scenarios",
			canondata.M{"scenarios": strings.Join(missing, ", ")}), nil
	case len(differing) > 0:
		return false, canondata.T("driver.recompile.outcomes-differ",
			canondata.M{"scenarios": strings.Join(differing, ", ")}), nil
	}
	return true, canondata.T("driver.recompile.both-green",
		canondata.M{"count": fmt.Sprintf("%d", len(orig))}), nil
}

func outcomesByScenario(eng *engine.Engine) map[string]string {
	out := map[string]string{}
	chk, err := eng.Store.Checks()
	if err != nil {
		return out
	}
	for _, c := range chk {
		out[c.Scenario] = c.Outcome
	}
	return out
}

// recordGate records the gate outcome in the original: a ledger
// event and a minimal journal transaction (the head moves — the
// verdict refreezes with the gate line).
func recordGate(eng *engine.Engine, freshDir string, green bool, reason string) error {
	freshEng, _, err := openWithRetry(freshDir)
	if err != nil {
		return err
	}
	key := "double-compile-" + freshEng.HeadDigest()
	le := ledger.NewEntry("gate", canon.StageDeliver)
	le.Details = map[string]string{
		"name":    "double-compile",
		"outcome": boolGreenOf(green),
		"reason":  firstLineOf(reason),
	}
	if err := eng.Ledger.Append(le); err != nil {
		return err
	}
	if err := eng.CommitRunTransaction(key); err != nil {
		return err
	}
	// The journal head has moved — the verdict refreezes with the
	// gate line: the file in the human's channel is always fresh, not
	// awaiting a recompile.
	_, err = eng.RenderVerdict()
	return err
}

func boolGreenOf(b bool) string {
	if b {
		return "green"
	}
	return "red"
}

func firstLineOf(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

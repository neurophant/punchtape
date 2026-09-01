// Package compiler — spec compiler: expands acceptance scenarios into
// named checks, computes a degree for every scenario and maintains
// scenario↔check traceability. Generated checks are a no-hand-edit
// zone: they are regenerated from the spec automatically.
package compiler

import (
	"fmt"
	"sort"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Scenario degrees.
const (
	DegreeExecutable = "executable"
	DegreeSkeleton   = "skeleton"
	DegreeProse      = "prose"
)

// Degree — a scenario's degree with a named reason when it is
// below executable. The degree is the compiler's product: computed,
// not stored.
type Degree struct {
	Level  string
	Reason string
}

// DegreeOf computes a scenario's degree: prose, an unfinished
// skeleton (no call or no expectations), or executable.
func DegreeOf(sc canon.Scenario) Degree {
	if sc.Prose != nil {
		return Degree{Level: DegreeProse, Reason: *sc.Prose}
	}
	if sc.When == nil {
		return Degree{Level: DegreeSkeleton, Reason: canondata.T("compiler.degree.no-when")}
	}
	if len(sc.Then) == 0 {
		return Degree{Level: DegreeSkeleton, Reason: canondata.T("compiler.degree.then-empty")}
	}
	return Degree{Level: DegreeExecutable}
}

// Compile expands an executable scenario into a check. The check's
// identifier inherits the scenario's number: a one-to-one link; the
// traceability anchor is the scenario identifier in the check
// itself, not in the code.
func Compile(sc canon.Scenario) (canon.Check, error) {
	if !strings.HasPrefix(sc.ID, "SCN-") {
		return canon.Check{}, fmt.Errorf("%s", canondata.T("compiler.error.scenario-id",
			canondata.M{"id": fmt.Sprintf("%q", sc.ID)}))
	}
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
	return canon.Check{
		ID:           "TST-" + strings.TrimPrefix(sc.ID, "SCN-"),
		Scenario:     sc.ID,
		SourceDigest: scenarioDigest(sc),
		Seed:         seed,
		Materials:    materials,
		Pre:          sc.Pre,
		When:         sc.When,
		SurfaceNames: surfaceNamesOf(sc),
		Then:         then,
		Outcome:      canon.CheckUnknown,
	}, nil
}

// surfaceNamesOf — entry point names: the scenario's field (the
// table writes all its rows' command[0] there), otherwise the
// command[0] of the action and the setup commands; deduplicated in
// stable order.
func surfaceNamesOf(sc canon.Scenario) []string {
	if len(sc.SurfaceNames) > 0 {
		return sc.SurfaceNames
	}
	seen := map[string]bool{}
	var out []string
	add := func(cmd []string) {
		if len(cmd) > 0 && !seen[cmd[0]] {
			seen[cmd[0]] = true
			out = append(out, cmd[0])
		}
	}
	for _, pre := range sc.Pre {
		add(pre)
	}
	if sc.When != nil {
		add(sc.When.Command)
	}
	sort.Strings(out)
	return out
}

// scenarioDigest — the digest of a scenario's canonical YAML: the
// regeneration source. Scenario changed — digest changed — the
// check is regenerated; untouched checks are not rewritten.
func scenarioDigest(sc canon.Scenario) string {
	data, err := yamlio.Marshal(sc)
	if err != nil {
		return ""
	}
	return yamlio.Digest(data)
}

// Plan — the desired set of checks after compiling all executable
// scenarios, and what to delete as orphaned.
type Plan struct {
	Put    []canon.Check // write (new and changed)
	Keep   []canon.Check // untouched: digest matched
	Remove []string      // orphaned check identifiers
}

// Sync compares the desired check set with the existing one: the
// untouched stays bit-for-bit, the touched is regenerated, the
// orphaned is deleted. Regeneration touches only what changed.
func Sync(scns []canon.Scenario, existing []canon.Check) (Plan, error) {
	var plan Plan
	existingByID := map[string]canon.Check{}
	for _, c := range existing {
		existingByID[c.ID] = c
	}
	wantIDs := map[string]bool{}
	for _, sc := range scns {
		d := DegreeOf(sc)
		if d.Level != DegreeExecutable {
			continue
		}
		check, err := Compile(sc)
		if err != nil {
			return plan, err
		}
		wantIDs[check.ID] = true
		old, ok := existingByID[check.ID]
		if ok && old.SourceDigest == check.SourceDigest {
			plan.Keep = append(plan.Keep, old)
			continue
		}
		// the last run's outcome is carried over to the regenerated
		// check: the source did not change — the outcome is honest;
		// provenness (green at least once) is carried by the same
		// rule — the expectations' protection from post-code
		// weakening survives cosmetics
		if ok {
			check.Outcome = old.Outcome
			check.Reason = old.Reason
			check.Proven = old.Proven
		}
		plan.Put = append(plan.Put, check)
	}
	for _, c := range existing {
		if !wantIDs[c.ID] {
			plan.Remove = append(plan.Remove, c.ID)
		}
	}
	return plan, nil
}

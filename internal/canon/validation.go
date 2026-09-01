package canon

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
)

// Allowed assertion condition values for text observations.
var assertionConditions = map[string]bool{
	"contains": true, "equals": true, "absent": true,
	"json-equals": true, "json-contains": true,
}

// Conditions that require valid JSON in value: semantic comparisons.
var jsonConditions = map[string]bool{
	"json-equals": true, "json-contains": true,
}

// RelativePathOK — path portability: a data path must be relative and
// stay inside the project root — absolute paths, escapes outward
// (..), home references (~), drive letters and backslashes break
// moving an instance between environments. An empty path is allowed:
// absence is not a stub.
func RelativePathOK(p string) bool {
	if p == "" {
		return true
	}
	if strings.HasPrefix(p, "/") || strings.HasPrefix(p, "\\") || strings.HasPrefix(p, "~") {
		return false
	}
	if strings.Contains(p, "..") || strings.Contains(p, "\\") {
		return false
	}
	if len(p) >= 2 && p[1] == ':' &&
		((p[0] >= 'a' && p[0] <= 'z') || (p[0] >= 'A' && p[0] <= 'Z')) {
		return false
	}
	return true
}

// ValidateDetail — the boundary's detail minimum: a filled block is
// complete — all four behavior classes are named; each class either
// by the requirement's scenarios or by an explicit "no" with a
// non-empty reason. A scenario must exist and belong to the
// requirement.
func ValidateDetail(d *RequirementDetail, reqID string, scenarioReq map[string]string) error {
	if d == nil {
		return nil
	}
	for _, class := range []struct {
		name  string
		value *DetailClass
	}{
		{"invariants", d.Invariants},
		{"edge-cases", d.EdgeCases},
		{"failure-matrix", d.FailureMatrix},
		{"concurrency", d.Concurrency},
	} {
		if class.value == nil {
			return fmt.Errorf("%s", canondata.LintMessage("detail.class-missing", canondata.M{
				"requirement": reqID, "class": class.name,
			}))
		}
		c := class.value
		switch {
		case len(c.Scenarios) > 0 && c.Absent != "":
			return fmt.Errorf("%s", canondata.LintMessage("detail.class-both", canondata.M{
				"requirement": reqID, "class": class.name,
			}))
		case len(c.Scenarios) > 0:
			for _, sid := range c.Scenarios {
				owner, ok := scenarioReq[sid]
				if !ok || owner != reqID {
					return fmt.Errorf("%s", canondata.LintMessage("detail.class-scenario", canondata.M{
						"requirement": reqID, "class": class.name, "scenario": sid,
					}))
				}
			}
		case c.Absent != "":
		default:
			return fmt.Errorf("%s", canondata.LintMessage("detail.class-empty", canondata.M{
				"requirement": reqID, "class": class.name,
			}))
		}
	}
	return nil
}

// ValidateSeed — seed and material paths are relative: the files live
// inside the run instance, a material's source — inside the project;
// an absolute path is not portable.
func ValidateSeed(seed []Seed, materials []Material) error {
	for _, sd := range seed {
		if !RelativePathOK(sd.Path) {
			return fmt.Errorf("%s", canondata.LintMessage("fixture.path-relative", canondata.M{
				"fixture": "seed", "action": "write-file", "path": sd.Path,
			}))
		}
	}
	for _, m := range materials {
		if !RelativePathOK(m.From) {
			return fmt.Errorf("%s", canondata.LintMessage("fixture.path-relative", canondata.M{
				"fixture": "seed", "action": "copy-file", "path": m.From,
			}))
		}
	}
	return nil
}

// ValidateAssertion validates an assertion-observation: the
// observation + condition pair is allowed, for file — a path, for
// semantic conditions — parseable JSON in value. The absent
// condition is first-class: it takes no value (for streams it means
// "the stream is empty", for a file — "the file is absent"); a value
// with absent is a lint error, not a silent degenerate check.
func ValidateAssertion(a Assertion) error {
	switch a.Observation {
	case "stdout", "stderr":
		if !assertionConditions[a.Condition] {
			return fmt.Errorf("%s", canondata.LintMessage("assertion.condition-text", canondata.M{
				"observation": a.Observation, "condition": fmt.Sprintf("%q", a.Condition),
			}))
		}
	case "exit-code":
		switch a.Condition {
		case "equals", "fails":
		default:
			return fmt.Errorf("%s", canondata.LintMessage("assertion.condition-exit",
				canondata.M{"condition": fmt.Sprintf("%q", a.Condition)}))
		}
	case "file":
		switch a.Condition {
		case "exists", "not-exists":
			// no value needed
		case "contains", "equals", "absent", "json-equals", "json-contains", "bytes-equals":
			if a.Path == "" {
				return fmt.Errorf("%s", canondata.LintMessage("assertion.path-required",
					canondata.M{"condition": fmt.Sprintf("%q", a.Condition)}))
			}
		default:
			return fmt.Errorf("%s", canondata.LintMessage("assertion.condition-file",
				canondata.M{"condition": fmt.Sprintf("%q", a.Condition)}))
		}
	default:
		return fmt.Errorf("%s", canondata.LintMessage("assertion.observation-unknown",
			canondata.M{"observation": fmt.Sprintf("%q", a.Observation)}))
	}
	if a.Condition == "absent" && a.Value != "" {
		return fmt.Errorf("%s", canondata.LintMessage("assertion.absent-value"))
	}
	if a.Condition == "fails" && a.Value != "" {
		return fmt.Errorf("%s", canondata.LintMessage("assertion.fails-value"))
	}
	if !RelativePathOK(a.Path) {
		return fmt.Errorf("%s", canondata.LintMessage("assertion.path-relative", canondata.M{
			"path": a.Path,
		}))
	}
	if jsonConditions[a.Condition] {
		var v any
		if err := json.Unmarshal([]byte(a.Value), &v); err != nil {
			return fmt.Errorf("%s", canondata.LintMessage("assertion.json-invalid", canondata.M{
				"condition": a.Condition, "err": err.Error(),
			}))
		}
	}
	return nil
}

// ValidateWhen validates a surface call: in v1 there is one surface —
// cli; the command is non-empty, the timeout positive, the declared
// volatility — names of observed streams (stdout/stderr) or relative
// state file paths (random bytes declared by the brief: identifiers,
// timestamps; file presence stays strict).
func ValidateWhen(w *When) error {
	if w == nil {
		return fmt.Errorf("%s", canondata.LintMessage("when.required"))
	}
	if w.Surface != "cli" {
		return fmt.Errorf("%s", canondata.LintMessage("when.surface",
			canondata.M{"surface": fmt.Sprintf("%q", w.Surface)}))
	}
	if len(w.Command) == 0 {
		return fmt.Errorf("%s", canondata.LintMessage("when.command"))
	}
	if w.TimeoutSec <= 0 {
		return fmt.Errorf("%s", canondata.LintMessage("when.timeout"))
	}
	for _, v := range w.Volatile {
		if v != "stdout" && v != "stderr" && (v == "" || !RelativePathOK(v)) {
			return fmt.Errorf("%s", canondata.LintMessage("when.volatile",
				canondata.M{"value": v}))
		}
	}
	return nil
}

// ScenarioReady — a scenario is ready for compilation into a check:
// either prose (the unmeasurable remainder) or a full when with
// pinned assertions. An unfinished scenario (skeleton) returns false.
func ScenarioReady(sc Scenario) bool {
	if sc.Prose != nil {
		return true
	}
	return sc.When != nil && len(sc.Then) > 0
}

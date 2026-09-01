// Executor routing: configuration decides which command fills which
// slot. Replacing the configuration is replacing a file, not a
// rebuild: the policy lives in the file and in the ledger, not in
// the code.
package driver

import (
	"fmt"
	"os"

	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/engine"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// Slot kinds known to routing: the expected delta kind of the given
// slot. Clarifications are not routed: questions and answers ride
// with checks-table row submissions; the slot has no separate human
// point.
var routableKinds = map[string]bool{
	"intent": true, "draft": true, "spec": true, "assert": true, "code": true, "fix": true,
}

// Routing — executor configuration: default is mandatory, kinds
// override by slot kind, facets — by card facet (more specific than
// kind). An empty map — no routing, everything in default.
func (r *Routing) HasKind(kind string) bool {
	return r != nil && r.Kinds != nil && r.Kinds[kind] != ""
}

type Routing struct {
	Default string            `yaml:"default"`
	Kinds   map[string]string `yaml:"kinds,omitempty"`
	Facets  map[string]string `yaml:"facets,omitempty"`
}

// LoadRouting reads the configuration from a file; the decode is
// strict, an unknown field or slot kind is an error.
func LoadRouting(path string) (*Routing, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("routing %s: %w", path, err)
	}
	var r Routing
	if err := yamlio.DecodeStrict(data, &r); err != nil {
		return nil, fmt.Errorf("routing %s: %w", path, err)
	}
	if r.Default == "" {
		return nil, fmt.Errorf("%s", canondata.T("driver.routing.default-empty",
			canondata.M{"path": path}))
	}
	for kind := range r.Kinds {
		if !routableKinds[kind] {
			return nil, fmt.Errorf("%s", canondata.T("driver.routing.bad-kind", canondata.M{
				"path": path, "kind": fmt.Sprintf("%q", kind),
			}))
		}
	}
	return &r, nil
}

// Route — the executor command for a slot: facet over kind, kind
// over the default.
func (r *Routing) Route(info engine.SlotInfo) string {
	if info.Facet != "" {
		if cmd, ok := r.Facets[info.Facet]; ok {
			return cmd
		}
	}
	if cmd, ok := r.Kinds[info.Kind]; ok {
		return cmd
	}
	return r.Default
}

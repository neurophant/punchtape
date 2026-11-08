// Package canon — the machine's one source: an embedded file that
// holds the base contract with its Russian and Chinese mirrors, the
// template, the reference with its own mirrors, the example, the
// validation↔law parity map and the environment rule. The zone's
// contract.md, template.yaml and reference.md are generated from
// it; the mirrors are data for the humans who read those languages
// — the zone carries the English pages alone. A canon that would
// lie about the machine fails the build (the self-validation
// battery runs the example against the real machine). The English
// contract is the base and rules.
package canon

import (
	_ "embed" // the canon source rides the binary
	"strings"
	"sync"

	"github.com/goccy/go-yaml"
)

//go:embed canon.yaml
var source []byte

// Source — the canon's whole shape.
type Source struct {
	Contract    string   `yaml:"contract"`
	ContractRU  string   `yaml:"contract-ru"`
	ContractZH  string   `yaml:"contract-zh"`
	Template    string   `yaml:"template"`
	Reference   string   `yaml:"reference"`
	ReferenceRU string   `yaml:"reference-ru"`
	ReferenceZH string   `yaml:"reference-zh"`
	Example     string   `yaml:"example"`
	LawMap      []LawRef `yaml:"law-map"`
	Env         []string `yaml:"env"`
}

// LawRef — one validation rule bound to the law the contract prints.
type LawRef struct {
	Rule string `yaml:"rule"`
	Law  string `yaml:"law"`
}

// The canon's one parse: the source is embedded and immutable, the
// parsed shape is shared and read-only — every consumer (CleanEnv's
// iteration, LawIDs' reads, the zone pages' writes of these very
// bytes) mutates nothing.
var (
	once   sync.Once
	cached Source
)

// Load — the canon, parsed once. A canon that does not parse is a
// build failure: the machine refuses to carry a broken source — the
// first call carries the panic, every caller sees it.
func Load() Source {
	once.Do(func() {
		if err := yaml.UnmarshalWithOptions(source, &cached, yaml.Strict()); err != nil {
			panic("canon: the source does not parse: " + err.Error())
		}
	})
	return cached
}

// LawIDs — the law ids the contract prints, the one source every
// law citation resolves against: no other spelling resolves. The
// ids parse from the contract's own law headers — the canon is the
// one truth, and nothing beside it can drift.
func LawIDs() []string {
	var out []string
	for _, line := range strings.Split(Load().Contract, "\n") {
		if !strings.HasPrefix(line, "law:") {
			continue
		}
		id := line
		if cut := strings.IndexAny(line, " \t"); cut >= 0 {
			id = line[:cut]
		}
		out = append(out, id)
	}
	return out
}

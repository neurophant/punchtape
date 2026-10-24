// Package canon — the machine's one source: an embedded file that
// holds the base contract with its Russian and Chinese mirrors, the
// template, the example, the validation↔law parity map and the
// environment rule. The zone's contract files and template.yaml are
// generated from it; a canon that would lie about the machine fails
// the build (the self-validation battery runs the example against
// the real machine). The English contract is the base and rules; the
// mirrors are data for the humans who read those languages.
package canon

import (
	_ "embed" // the canon source rides the binary

	"github.com/goccy/go-yaml"
)

//go:embed canon.yaml
var source []byte

// Source — the canon's whole shape.
type Source struct {
	Contract   string   `yaml:"contract"`
	ContractRU string   `yaml:"contract-ru"`
	ContractZH string   `yaml:"contract-zh"`
	Template   string   `yaml:"template"`
	Example    string   `yaml:"example"`
	LawMap     []LawRef `yaml:"law-map"`
	Env        []string `yaml:"env"`
}

// LawRef — one validation rule bound to the law the contract prints.
type LawRef struct {
	Rule string `yaml:"rule"`
	Law  string `yaml:"law"`
}

// Load — the canon, parsed once. A canon that does not parse is a
// build failure: the machine refuses to carry a broken source.
func Load() Source {
	var s Source
	if err := yaml.UnmarshalWithOptions(source, &s, yaml.Strict()); err != nil {
		panic("canon: the source does not parse: " + err.Error())
	}
	return s
}

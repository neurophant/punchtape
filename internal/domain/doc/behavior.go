package doc

// The behavior's own bytes: one behavior marshals alone — the
// journal stores each behavior as one content-addressed object, and
// the corpus fold decodes them back one by one. The bytes are the
// same canonical spelling the whole document uses.

import "github.com/goccy/go-yaml"

// MarshalBehavior — one behavior's canonical bytes.
func MarshalBehavior(b Behavior) []byte {
	return Canonical(b)
}

// MarshalRow — one row's canonical bytes: the row's own identity in
// every proof fingerprint.
func MarshalRow(r Row) []byte {
	return Canonical(r)
}

// DecodeBehavior — one behavior from its canonical bytes, strict.
func DecodeBehavior(data []byte) (Behavior, error) {
	if err := armor(data); err != nil {
		return Behavior{}, err
	}
	var b Behavior
	if err := yaml.UnmarshalWithOptions(data, &b, yaml.Strict()); err != nil {
		return Behavior{}, oneLine(err)
	}
	return b, nil
}

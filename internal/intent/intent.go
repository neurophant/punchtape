// Package intent — the everyday-phrase router: a delivery data
// table (intents.yaml) maps an everyday phrase to a machine verb. A
// row fires when all its keywords are among the phrase's tokens
// (case-insensitive, punctuation stripped) — the mechanics are
// language-independent. The default is the wish (the "wish →
// verdict" loop); matches into different targets — ambiguity: the
// human resolves the fork via a menu, the machine does not. The
// table grows with new data rows, without code.
package intent

import (
	"sort"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
)

// Target — a routing target: a verb and a topic (for why).
type Target struct {
	Verb  string
	Topic string
}

// Key — the target's canonical key.
func (t Target) Key() string {
	if t.Topic == "" {
		return t.Verb
	}
	return t.Verb + " " + t.Topic
}

// Wish — the default: a phrase becomes a product wish.
var Wish = Target{Verb: "wish"}

// Route — phrase routing: one target, the wish default, or
// ambiguity (matches into different targets). The menu material is
// the match targets in stable order; the default does not take
// part in the menu — it is already reachable by silence, by
// repeating the plain phrase.
func Route(phrase string) (target Target, ambiguous bool, matched []Target) {
	tokens := phraseTokens(phrase)
	if len(tokens) == 0 {
		return Wish, false, nil
	}
	seen := map[string]bool{}
	var targets []Target
	for _, row := range canondata.Intents() {
		if !rowMatches(tokens, row.Keys) {
			continue
		}
		t := Target{Verb: row.Verb, Topic: row.Topic}
		if seen[t.Key()] {
			continue
		}
		seen[t.Key()] = true
		targets = append(targets, t)
	}
	switch len(targets) {
	case 0:
		return Wish, false, nil
	case 1:
		return targets[0], false, targets
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].Key() < targets[j].Key() })
	return Wish, true, targets
}

// rowMatches — all of the row's keys are among the phrase's tokens.
func rowMatches(tokens []string, keys []string) bool {
	set := map[string]bool{}
	for _, t := range tokens {
		set[t] = true
	}
	for _, k := range keys {
		if !set[strings.ToLower(k)] {
			return false
		}
	}
	return true
}

// phraseTokens — a phrase's tokens: split by spaces, lower case,
// edge punctuation stripped; deterministic and without dictionaries.
func phraseTokens(phrase string) []string {
	var out []string
	for _, tok := range strings.Fields(phrase) {
		tok = strings.ToLower(strings.Trim(tok, ".,!?;:()[]\"'`"))
		if tok != "" {
			out = append(out, tok)
		}
	}
	return out
}

// Label — a human-readable target label for the menu: the surface
// text is canon data; the code only picks the key.
func (t Target) Label() string {
	switch t.Verb {
	case "next":
		return canondata.T("cli.intent.label.next")
	case "why":
		topic := t.Topic
		if topic == "" {
			topic = "topics"
		}
		return canondata.T("cli.intent.label.why", canondata.M{"topic": topic})
	}
	return canondata.T("cli.intent.label.wish")
}

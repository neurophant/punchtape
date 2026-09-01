package delta

import "github.com/neurophant/punchtape/internal/canondata"

// Skeleton — the canonical skeleton of the correct form for a delta
// kind. The machine teaches its own format by construction: an error
// line plus a skeleton. Every field is named, placeholders carry the
// type; dictionaries of allowed values are folded into one-line
// comments (the full form lives in the instance's AGENTS.md
// reference). Semantics is neither repaired nor proposed — only the
// form. Skeleton texts are delivery canon data; the code picks the
// key by delta kind.
func Skeleton(kind string) string {
	switch kind {
	case KindIntent:
		return canondata.T("format.skeleton.intent")
	case KindLint:
		return canondata.T("format.skeleton.lint")
	case KindClarify:
		return canondata.T("format.skeleton.clarify")
	case KindChecks:
		return canondata.T("format.skeleton.checks")
	case KindSpec:
		return canondata.T("format.skeleton.spec")
	case KindAssert:
		return canondata.T("format.skeleton.assert")
	case KindSplit:
		return canondata.T("format.skeleton.split")
	case KindCode:
		return canondata.T("format.skeleton.code")
	case KindFix:
		return canondata.T("format.skeleton.fix")
	case KindProbe:
		return canondata.T("format.skeleton.probe")
	case KindConventions:
		return canondata.T("format.skeleton.conventions")
	case KindAmend:
		return canondata.T("format.skeleton.amend")
	case KindFeature:
		return canondata.T("format.skeleton.feature")
	case KindDecision:
		return canondata.T("format.skeleton.decision")
	case KindKb:
		return canondata.T("format.skeleton.kb")
	default:
		return canondata.T("format.skeleton.default")
	}
}

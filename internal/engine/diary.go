// The self-observation diary: the machine's pains are a computed view
// of the ledger. Pain facts are recorded as events at the moment of
// discovery (append-only is inherited from the ledger), separate
// storage would be a second bookkeeping; duplicates = frequency =
// priority. The pain catalog is delivery canon data: the diagnosis
// and workaround of each pain (a workaround is mandatory, an empty
// one is rendered by the finding itself). Entries are machine facts
// and pains only: there are and can be no model opinions here (the
// step observer is forbidden forever).
package engine

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
)

// painFact — one catalog pain: how many times it was observed and
// when last.
type painFact struct {
	id    string
	count int
	last  string
}

// collectPains — a mechanical projection of the ledger into catalog
// pains. A loop is a repeat of the same red suite reason in DIFFERENT
// submission windows (a new submission happened between the repeats):
// identical reasons within one run are the norm (several checks of
// one submission share a reason), not a loop.
func (e *Engine) collectPains() []painFact {
	counts := map[string]int{}
	lasts := map[string]string{}
	note := func(id, at string) {
		counts[id]++
		if at > lasts[id] {
			lasts[id] = at
		}
	}
	window := 0
	redWindow := map[string]int{}
	for _, en := range e.Ledger.All() {
		switch en.Event {
		case "apply":
			window++
		case "verb":
			if en.Details["name"] != "submit" {
				continue
			}
			at := en.At.UTC().Format("2006-01-02 15:04")
			if en.Details["outcome"] == "rejected" {
				note("rejected-delta", at)
			}
			if en.Details["idempotent"] == "yes" {
				note("idempotent-resubmit", at)
			}
		case "gate":
			if en.Details["name"] != "suite" || en.Details["outcome"] == "green" || en.Details["reason"] == "" {
				continue
			}
			at := en.At.UTC().Format("2006-01-02 15:04")
			reason := en.Details["reason"]
			if first, seen := redWindow[reason]; seen && first < window {
				// a reason repeat after a new submission is a repair loop
				note("red-loop", at)
			}
			redWindow[reason] = window
		case "quarantine":
			note("quarantine", en.At.UTC().Format("2006-01-02 15:04"))
		case "question":
			if en.Details["kind"] == "amend-assertion" {
				note("amend-needed", en.At.UTC().Format("2006-01-02 15:04"))
			}
		}
	}
	out := make([]painFact, 0, len(counts))
	for id, n := range counts {
		out = append(out, painFact{id: id, count: n, last: lasts[id]})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].count != out[j].count {
			return out[i].count > out[j].count
		}
		return out[i].id < out[j].id
	})
	return out
}

// WhyDiary — the diary surface: what hurt, how many times, what to
// do while it is open. A pain outside the catalog is rendered as an
// honest line without a diagnosis; an empty catalog workaround is
// rendered by the finding itself (knowledge without action is also
// value).
func (e *Engine) WhyDiary() string {
	pains := e.collectPains()
	if len(pains) == 0 {
		return canondata.T("why.diary.none")
	}
	var sb strings.Builder
	sb.WriteString(canondata.T("why.diary.head"))
	sb.WriteString("\n")
	for _, p := range pains {
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.diary.item", canondata.M{
			"id": p.id, "count": strconv.Itoa(p.count), "last": p.last,
		}))
		entry, known := canondata.Pain(p.id)
		if !known {
			fmt.Fprintf(&sb, "%s\n", canondata.T("why.diary.item-finding"))
			continue
		}
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.diary.item-what", canondata.M{
			"diagnosis": strings.TrimSpace(entry.Diagnosis),
		}))
		workaround := strings.TrimSpace(entry.Workaround)
		if workaround == "" {
			fmt.Fprintf(&sb, "%s\n", canondata.T("why.diary.item-finding"))
			continue
		}
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.diary.item-open", canondata.M{
			"workaround": workaround,
		}))
	}
	return strings.TrimRight(sb.String(), "\n")
}

// WhyRetro — the diary retro: pains ranked by frequency (dedup is
// already done by the diary collector — duplicates = frequency),
// each shows its workaround and a call to record the cure decision
// as a decision delta; decisions without a record stay open. The
// cheapness of a cure is not invented: the price of a fix is unknown
// until measured, the line says so honestly. Process metrics are the
// same machine facts.
func (e *Engine) WhyRetro() string {
	pains := e.collectPains()
	if len(pains) == 0 {
		return canondata.T("why.retro.none")
	}
	total := 0
	for _, p := range pains {
		total += p.count
	}
	var sb strings.Builder
	sb.WriteString(canondata.T("why.retro.head"))
	sb.WriteString("\n")
	for _, p := range pains {
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.retro.item", canondata.M{
			"id": p.id, "count": strconv.Itoa(p.count),
		}))
		entry, known := canondata.Pain(p.id)
		if known {
			fmt.Fprintf(&sb, "%s\n", canondata.T("why.retro.item-what", canondata.M{
				"diagnosis": strings.TrimSpace(entry.Diagnosis),
			}))
			if workaround := strings.TrimSpace(entry.Workaround); workaround != "" {
				fmt.Fprintf(&sb, "%s\n", canondata.T("why.retro.item-open", canondata.M{
					"workaround": workaround,
				}))
			}
		} else {
			fmt.Fprintf(&sb, "%s\n", canondata.T("why.retro.item-finding"))
		}
		fmt.Fprintf(&sb, "%s\n", canondata.T("why.retro.item-cost"))
	}
	fmt.Fprintf(&sb, "%s", canondata.T("why.retro.total", canondata.M{
		"pains": strconv.Itoa(len(pains)), "observations": strconv.Itoa(total),
	}))
	return strings.TrimRight(sb.String(), "\n")
}

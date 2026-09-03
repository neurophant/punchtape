// The instance's knowledge about the product: the Decision Log with
// the forbidden zone and the "when X do Y" kb reference — the
// curatorial zone of the passport, written by deltas through the
// validator (human prose, but not by hand into files), read by the
// machine: kb advises the machine's own question about the build
// recipe, everything is rendered by computed why views (there are no
// second sources of truth).
package engine

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// applyFeature — the author's description of a feature of the human
// spec: requirement prose or ratification by the operator. The
// requirement must exist; new text removes the former ratification;
// ratification without text is also an operation (the existing prose
// is affirmed). One submission may carry the texts of several or all
// features (entries:) — validated as a whole, applied as one
// transaction. Knowledge, not behavior: it opens no cycle and does
// not touch checks.
func (b *builder) applyFeature(d *delta.Feature) error {
	if d == nil {
		return rejection("feature", canondata.T("submit.reject.empty-delta"))
	}
	if len(d.Entries) > 0 {
		if strings.TrimSpace(d.Requirement) != "" || strings.TrimSpace(d.Text) != "" || d.Ratify {
			return rejection("feature", canondata.T("submit.reject.feature-two-forms"))
		}
		return b.applyFeatureBatch(d.Entries)
	}
	return b.applyFeatureOne(d.Requirement, d.Text, d.Ratify)
}

// applyFeatureBatch — the entries form: the whole human spec in one
// submission. Every entry is validated before any is applied — a
// batch is one transaction, a partial write is not a batch.
func (b *builder) applyFeatureBatch(entries []delta.FeatureEntry) error {
	seen := make(map[string]bool, len(entries))
	for _, en := range entries {
		if strings.TrimSpace(en.Requirement) == "" {
			return rejection("feature", canondata.T("submit.reject.feature-req-missing"))
		}
		if seen[en.Requirement] {
			return rejection("feature", canondata.T("submit.reject.feature-dup", canondata.M{
				"id": en.Requirement,
			}))
		}
		seen[en.Requirement] = true
		if _, ok := b.reqs[en.Requirement]; !ok {
			return rejection("feature", canondata.T("submit.reject.feature-req-not-found", canondata.M{
				"id": en.Requirement,
			}))
		}
		if strings.TrimSpace(en.Text) == "" && !en.Ratify {
			return rejection("feature", canondata.T("submit.reject.feature-empty"))
		}
	}
	for _, en := range entries {
		if err := b.applyFeatureOne(en.Requirement, en.Text, en.Ratify); err != nil {
			return err
		}
	}
	return nil
}

// applyFeatureOne — the single-feature core: prose or ratification
// for one existing requirement.
func (b *builder) applyFeatureOne(requirement, text string, ratify bool) error {
	if strings.TrimSpace(requirement) == "" {
		return rejection("feature", canondata.T("submit.reject.feature-req-missing"))
	}
	req, ok := b.reqs[requirement]
	if !ok {
		return rejection("feature", canondata.T("submit.reject.feature-req-not-found", canondata.M{
			"id": requirement,
		}))
	}
	switch {
	case strings.TrimSpace(text) != "":
		req.SpecDoc = &canon.FeatureDoc{Text: strings.TrimSpace(text)}
	case ratify:
		if req.SpecDoc == nil {
			return rejection("feature", canondata.T("submit.reject.feature-no-doc", canondata.M{
				"id": requirement,
			}))
		}
		req.SpecDoc.Ratified = true
	default:
		return rejection("feature", canondata.T("submit.reject.feature-empty"))
	}
	return b.putReq(req)
}

// applyDecision — an entry into the Decision Log and the forbidden
// zone: exactly one operation, the prose must be non-empty — there
// are no silent entries. The curatorial zone files are written in
// full (the list is data).
func (b *builder) applyDecision(d *delta.DecisionDelta) error {
	if d == nil {
		return rejection("decision", canondata.T("submit.reject.empty-delta"))
	}
	filled := 0
	if d.AddDecision != nil {
		filled++
		if strings.TrimSpace(d.AddDecision.Decision) == "" ||
			strings.TrimSpace(d.AddDecision.Rationale) == "" {
			return rejection("decision", canondata.T("submit.reject.decision-empty"))
		}
		next := b.state.Counters["DEC"] + 1
		if next < 1 {
			next = 1
		}
		b.state.Counters["DEC"] = next
		decID := fmt.Sprintf("DEC-%03d", next)
		b.knowledgeNotes = append(b.knowledgeNotes, canondata.T("submit.reply.dec-id", canondata.M{
			"id": decID,
		}))
		b.decisions = append(b.decisions, canon.Decision{
			ID:        decID,
			Decision:  strings.TrimSpace(d.AddDecision.Decision),
			Rationale: strings.TrimSpace(d.AddDecision.Rationale),
			At:        time.Now().UTC().Truncate(time.Millisecond),
		})
	}
	if d.UpdateDecision != nil {
		filled++
		if strings.TrimSpace(d.UpdateDecision.Decision) == "" &&
			strings.TrimSpace(d.UpdateDecision.Rationale) == "" {
			return rejection("decision", canondata.T("submit.reject.decision-update-empty"))
		}
		found := false
		for i := range b.decisions {
			if b.decisions[i].ID == d.UpdateDecision.ID {
				if v := strings.TrimSpace(d.UpdateDecision.Decision); v != "" {
					b.decisions[i].Decision = v
				}
				if v := strings.TrimSpace(d.UpdateDecision.Rationale); v != "" {
					b.decisions[i].Rationale = v
				}
				found = true
				break
			}
		}
		if !found {
			return rejection("decision", canondata.T("submit.reject.decision-not-found", canondata.M{
				"id": d.UpdateDecision.ID,
			}))
		}
	}
	if d.AddOutOfScope != nil {
		filled++
		if strings.TrimSpace(d.AddOutOfScope.What) == "" ||
			strings.TrimSpace(d.AddOutOfScope.Why) == "" {
			return rejection("decision", canondata.T("submit.reject.outofscope-empty"))
		}
		for _, ex := range b.outofscope {
			if ex.What == strings.TrimSpace(d.AddOutOfScope.What) {
				return rejection("decision", canondata.T("submit.reject.outofscope-taken", canondata.M{
					"what": ex.What,
				}))
			}
		}
		b.outofscope = append(b.outofscope, canon.OutOfScopeItem{
			What: strings.TrimSpace(d.AddOutOfScope.What),
			Why:  strings.TrimSpace(d.AddOutOfScope.Why),
		})
		sort.Slice(b.outofscope, func(i, j int) bool { return b.outofscope[i].What < b.outofscope[j].What })
	}
	if d.RemoveOutOfScope != nil {
		filled++
		kept := b.outofscope[:0]
		found := false
		for _, ex := range b.outofscope {
			if ex.What == strings.TrimSpace(d.RemoveOutOfScope.What) {
				found = true
				continue
			}
			kept = append(kept, ex)
		}
		if !found {
			return rejection("decision", canondata.T("submit.reject.outofscope-not-found", canondata.M{
				"what": d.RemoveOutOfScope.What,
			}))
		}
		b.outofscope = kept
	}
	if filled != 1 {
		return rejection("decision", canondata.T("submit.reject.decision-one-op"))
	}
	return b.putKnowledge()
}

// applyKb — the knowledge reference: an addition with a mandatory
// source (trigger tokens non-empty) or a removal; supersedes
// deactivates the named entry — the supersession chain is visible in
// the summary.
func (b *builder) applyKb(k *delta.KbDelta) error {
	if k == nil {
		return rejection("kb", canondata.T("submit.reject.empty-delta"))
	}
	filled := 0
	if k.Add != nil {
		filled++
		if strings.TrimSpace(k.Add.Do) == "" || strings.TrimSpace(k.Add.Source) == "" {
			return rejection("kb", canondata.T("submit.reject.kb-empty"))
		}
		if len(k.Add.When) == 0 {
			return rejection("kb", canondata.T("submit.reject.kb-when"))
		}
		if k.Add.Supersedes != "" && !b.kbKnown(k.Add.Supersedes) {
			return rejection("kb", canondata.T("submit.reject.kb-supersedes", canondata.M{
				"id": k.Add.Supersedes,
			}))
		}
		key := strings.Join(k.Add.When, "\x00") + "\x00" + strings.TrimSpace(k.Add.Do)
		kbID := "KB-" + yamlio.DigestShort([]byte(key))
		b.knowledgeNotes = append(b.knowledgeNotes, canondata.T("submit.reply.kb-id", canondata.M{
			"id": kbID,
		}))
		b.kb = append(b.kb, canon.KbEntry{
			ID:         kbID,
			When:       k.Add.When,
			Do:         strings.TrimSpace(k.Add.Do),
			Source:     strings.TrimSpace(k.Add.Source),
			Supersedes: k.Add.Supersedes,
		})
	}
	if k.Update != nil {
		filled++
		if len(k.Update.When) == 0 && strings.TrimSpace(k.Update.Do) == "" &&
			strings.TrimSpace(k.Update.Source) == "" {
			return rejection("kb", canondata.T("submit.reject.kb-update-empty"))
		}
		found := false
		for i := range b.kb {
			if b.kb[i].ID != k.Update.ID {
				continue
			}
			if len(k.Update.When) > 0 {
				b.kb[i].When = k.Update.When
			}
			if v := strings.TrimSpace(k.Update.Do); v != "" {
				b.kb[i].Do = v
			}
			if v := strings.TrimSpace(k.Update.Source); v != "" {
				b.kb[i].Source = v
			}
			found = true
			break
		}
		if !found {
			return rejection("kb", canondata.T("submit.reject.kb-not-found", canondata.M{
				"id": k.Update.ID,
			}))
		}
	}
	if k.Remove != nil {
		filled++
		kept := b.kb[:0]
		found := false
		for _, en := range b.kb {
			if en.ID == k.Remove.ID {
				found = true
				continue
			}
			kept = append(kept, en)
		}
		if !found {
			return rejection("kb", canondata.T("submit.reject.kb-not-found", canondata.M{
				"id": k.Remove.ID,
			}))
		}
		b.kb = kept
	}
	if filled != 1 {
		return rejection("kb", canondata.T("submit.reject.decision-one-op"))
	}
	return b.putKnowledge()
}

func (b *builder) kbKnown(id string) bool {
	for _, en := range b.kb {
		if en.ID == id {
			return true
		}
	}
	return false
}

// putKnowledge — the effects of writing curatorial-zone knowledge:
// three files written in full (the list is data, the manifest header
// inside the effect).
func (b *builder) putKnowledge() error {
	if data, err := yamlio.Marshal(b.decisions); err == nil {
		b.effects = append(b.effects, journal.Effect{
			Path: canon.DecisionsFile, Content: machineFile(data),
		})
	}
	if data, err := yamlio.Marshal(b.outofscope); err == nil {
		b.effects = append(b.effects, journal.Effect{
			Path: canon.OutOfScopeFile, Content: machineFile(data),
		})
	}
	if data, err := yamlio.Marshal(b.kb); err == nil {
		b.effects = append(b.effects, journal.Effect{
			Path: canon.KbFile, Content: machineFile(data),
		})
	}
	return nil
}

// kbSupersededBy — identifiers of entries deactivated by newer ones.
func kbSupersededBy(entries []canon.KbEntry) map[string]bool {
	out := map[string]bool{}
	for _, en := range entries {
		if en.Supersedes != "" {
			out[en.Supersedes] = true
		}
	}
	return out
}

// kbAdvice — active knowledge for a context: all when tokens are
// present in the context — a deterministic match (the same mechanic
// as the intent router). The machine consults it before its own
// question about the build recipe; the advice found is shown in the
// question block.
func kbAdvice(entries []canon.KbEntry, contextTokens map[string]bool) *canon.KbEntry {
	dead := kbSupersededBy(entries)
	var hit *canon.KbEntry
	for i, en := range entries {
		if dead[en.ID] {
			continue
		}
		match := true
		for _, tok := range en.When {
			if !contextTokens[strings.ToLower(tok)] {
				match = false
				break
			}
		}
		if match {
			hit = &entries[i]
			break
		}
	}
	return hit
}

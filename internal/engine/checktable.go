// The checks table is the executor's interface to the spec. A
// table row: a seed + a sequence of commands + the expectations of
// the last command. The machine compiles each row into a canon
// scenario, groups rows by the wish's command families into
// requirements, expands scenarios into checks with the compiler, and
// keeps the scenario↔check trace automatically. A row with the same
// seed and the same sequence of commands replaces the previous row's
// expectations — repair is row-by-row, identifiers stay invisible to
// the hand.
package engine

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// reqFormulationOf — the formulation of a family's machine requirement
// in the language of the wish (the feature name in a human spec is
// living words): the exact string by which the requirement is found
// again (stable across submissions, without extra state; the
// instance's language is unchanged since the intent submission).
func reqFormulationOf(lang, family string) string {
	return canondata.TFor(lang, "checktable.req-formulation", canondata.M{"family": family})
}

// isTableRequirement — the requirement was created by the checks
// table: the anchor tail of the formulation in the instance's
// language.
func isTableRequirement(lang string, r canon.Requirement) bool {
	return strings.HasSuffix(r.Formulation, canondata.TFor(lang, "checktable.req-anchor"))
}

// noteSeedQuotes — a seed value wrapped in single quotes rides as
// bytes (the YAML level has already consumed its own quoting): the
// machine names it in the reply instead of leaving a silent false
// red on the quote bytes.
func (b *builder) noteSeedQuotes(what string, seed []canon.Seed) {
	for _, s := range seed {
		if len(s.Content) >= 2 && s.Content[0] == '\'' && s.Content[len(s.Content)-1] == '\'' {
			b.rowNotes = append(b.rowNotes, canondata.T("checktable.note.seed-quotes",
				canondata.M{"what": what, "path": s.Path}))
		}
	}
}

// rowKey — a table row's identity: seed + materials + setup +
// sequence of commands. The same means the same row, the expectations
// are replaced.
func rowKey(seed []canon.Seed, materials []canon.Material, pre [][]string, when []string) string {
	data, err := yamlio.Marshal(struct {
		Seed      []canon.Seed     `yaml:"seed"`
		Materials []canon.Material `yaml:"materials,omitempty"`
		Pre       [][]string       `yaml:"pre"`
		When      []string         `yaml:"when"`
	}{seed, materials, pre, when})
	if err != nil {
		return ""
	}
	return string(data)
}

// scenarioKey — the same identity for an existing canon scenario.
func scenarioKey(sc canon.Scenario) string {
	if sc.When == nil {
		return ""
	}
	return rowKey(sc.Seed, sc.Materials, sc.Pre, sc.When.Command)
}

// familyOf — a row's family: the longest match of the command's
// leading literals against the wish's families; outside any match,
// the command's first token (its own family, also legitimate
// behavior).
func familyOf(cmd []string, fams []string) string {
	best := ""
	for _, f := range fams {
		lits := strings.Fields(f)
		if len(cmd) < len(lits) {
			continue
		}
		match := true
		for i, l := range lits {
			if cmd[i] != l {
				match = false
				break
			}
		}
		if match && len(f) > len(best) {
			best = f
		}
	}
	if best != "" {
		return best
	}
	return cmd[0]
}

// flatRunRow — the number of run elements when there are at least two
// and all are single words (spaceless scalars): the signature of a
// flattened argv ("run: [tool, arg]" instead of one element
// "tool arg"). 0 — no signature (fewer than two, argv form, or
// commands with arguments).
func flatRunRow(run []delta.RunCommand) int {
	if len(run) < 2 {
		return 0
	}
	for _, cmd := range run {
		if cmd.IsArgv || strings.ContainsAny(cmd.Line, " \t") {
			return 0
		}
	}
	return len(run)
}

// rowIdentity — the parsed parts of a row's identity: seed, materials,
// setup, action. Shared by table application and the detection of
// replacements of proven rows — identity is computed the same way
// everywhere.
func (b *builder) rowIdentity(row delta.CheckRow) (seed []canon.Seed, materials []canon.Material, pre [][]string, when []string, err error) {
	seed, err = b.parseSeeds(row.Seeds, "")
	if err != nil {
		return
	}
	materials, err = parseMaterials(row.Materials, "")
	if err != nil {
		return
	}
	for j, cmd := range row.Run {
		fields, perr := cmd.Fields()
		if perr != nil {
			err = fmt.Errorf("run[%d]: %s", j+1, perr.Error())
			return
		}
		if len(fields) == 0 {
			err = fmt.Errorf("run[%d]: %s", j+1, canondata.T("checktable.error.run-empty",
				canondata.M{"num": fmt.Sprintf("%d", j+1)}))
			return
		}
		pre = append(pre, fields)
	}
	if len(pre) == 0 {
		err = fmt.Errorf("%s", canondata.T("checktable.error.run-required"))
		return
	}
	when = pre[len(pre)-1]
	pre = pre[:len(pre)-1]
	return
}

// parseSeeds — one-line seeds "<path> <content>": literal escapes
// (\n, \t, \r, \\, \xNN) are decoded — a multi-line file is seeded
// from one line. The second form is "<path> @<64hex>": the content
// is taken from the canon by digest (the FILES WRITTEN mechanism),
// no byte transfer needed; the resolved bytes give the same row
// identity as the full text.
func (b *builder) parseSeeds(lines []string, what string) ([]canon.Seed, error) {
	var ops []canon.Seed
	for i, line := range lines {
		// The leading indent is trimmed, the value's tail is untouchable:
		// authorial bytes (YAML quotes and blocks carry real newlines)
		// are the seed's content, trimming the tail changed the meaning.
		val := strings.TrimLeft(line, " \t")
		j := strings.IndexByte(val, ' ')
		if j <= 0 || strings.TrimSpace(val[j+1:]) == "" {
			return nil, rejection(what, canondata.T("checktable.error.seeds", canondata.M{
				"what": what, "num": fmt.Sprintf("%d", i+1),
			}))
		}
		path, rest := val[:j], strings.TrimLeft(val[j+1:], " \t")
		if content, handled, err := b.seedByDigestRef(path, rest, what, i); handled {
			if err != nil {
				return nil, err
			}
			ops = append(ops, canon.Seed{Path: path, Content: content})
			continue
		}
		ops = append(ops, canon.Seed{
			Path:    path,
			Content: yamlio.DecodeLiteral(rest),
		})
	}
	return ops, nil
}

// seedByDigestRef — a seed value of the form "@<64hex>": bytes from
// the journal by digest. handled=false — the value is not a reference
// (a plain literal); handled=true with err — the reference is broken
// (by format or an unknown digest), a refusal is mandatory.
func (b *builder) seedByDigestRef(path, ref, what string, idx int) (content string, handled bool, err error) {
	if len(ref) != 65 || ref[0] != '@' {
		return "", false, nil
	}
	digest := ref[1:]
	if !validCanonDigest(digest) {
		return "", true, rejection(what, canondata.T("submit.reject.seed-digest-format", canondata.M{
			"path": path, "digest": digest,
		}))
	}
	got, ok := b.e.canonContentByDigest(digest)
	if !ok {
		return "", true, rejection(what, canondata.T("checktable.error.seed-digest", canondata.M{
			"what": what, "num": fmt.Sprintf("%d", idx+1), "path": path, "digest": digest,
		}))
	}
	return got, true, nil
}

// parseMaterials — a row's materials: project file paths, the bytes
// are copied into the run directory under the same paths. Binary
// inputs are not transferred by hand — the machine carries them
// itself; about the nature of the bytes it knows nothing.
func parseMaterials(lines []string, what string) ([]canon.Material, error) {
	var ops []canon.Material
	for i, m := range lines {
		p := strings.TrimSpace(m)
		if p == "" || strings.ContainsAny(p, " \t") || !canon.RelativePathOK(p) {
			return nil, rejection(what, canondata.T("checktable.error.material", canondata.M{
				"num": fmt.Sprintf("%d", i+1),
			}))
		}
		ops = append(ops, canon.Material{From: p})
	}
	return ops, nil
}

// compactStreamCond — the out/err value of a compact row: the prefix
// "(contains) " means the contains condition (a fragment occurring in
// the stream), otherwise — exact byte equality.
func compactStreamCond(v string) (string, string) {
	if strings.HasPrefix(v, "(contains) ") {
		return "contains", v[len("(contains) "):]
	}
	return "equals", v
}

// rowTimeoutOr — the row's declared ceiling or the delivery default
// (limits.yaml check.timeout-default-s); the code keeps no constant
// of its own (the limits mirror rule).
func rowTimeoutOr(declared *int) int {
	if declared != nil && *declared > 0 {
		return *declared
	}
	return canondata.Limit("check.timeout-default-s")
}

// rowExpectations — the last command's expectations of a row into
// typed assertions: exit code (default 0), exact stdout/stderr,
// structural assertions about state files.
func rowExpectations(row delta.CheckRow) ([]canon.Assertion, string, error) {
	var then []canon.Assertion
	var shorts []string
	rc := 0
	if row.Rc != nil {
		rc = *row.Rc
	}
	then = append(then, canon.Assertion{
		Observation: "exit-code", Condition: "equals", Value: fmt.Sprintf("%d", rc),
	})
	shorts = append(shorts, fmt.Sprintf("rc %d", rc))
	// The "(contains) " prefix on out/err is the contains condition:
	// exact bytes are not available to a row (test-harness timings,
	// random identifiers), the row pins a fragment occurrence. A silent
	// literal-trap is excluded: the prefix is either recognized, or the
	// value is honestly exact (literal parentheses are written by the
	// full spec form).
	if row.Out != nil {
		cond, val := compactStreamCond(*row.Out)
		if err := yamlio.ValidateExpectLiteral(val); err != nil {
			return nil, "", rejection("out", err.Error())
		}
		out := yamlio.DecodeHexOnly(val)
		then = append(then, canon.Assertion{Observation: "stdout", Condition: cond, Value: out})
		shorts = append(shorts, fmt.Sprintf("out %s %q", cond, firstLineOf(out)))
	}
	if row.Err != nil {
		cond, val := compactStreamCond(*row.Err)
		if err := yamlio.ValidateExpectLiteral(val); err != nil {
			return nil, "", rejection("err", err.Error())
		}
		errv := yamlio.DecodeHexOnly(val)
		then = append(then, canon.Assertion{Observation: "stderr", Condition: cond, Value: errv})
		shorts = append(shorts, fmt.Sprintf("err %s %q", cond, firstLineOf(errv)))
	}
	// The grammar of a state row: "<path> absent" — no file;
	// "<path> text <bytes>" — exact bytes (escapes \n, \t, \\);
	// "<path> <json>" — structural JSON comparison; anything else —
	// exact bytes. The byte form covers textual state files (newlines,
	// trailing \n, zero-byte subtleties). The value's tail is not
	// trimmed: real newlines from YAML quotes and blocks are the
	// expectation's authorial bytes.
	for i, line := range row.State {
		val := strings.TrimLeft(line, " \t")
		j := strings.IndexByte(val, ' ')
		if j <= 0 || strings.TrimSpace(val[j+1:]) == "" {
			return nil, "", rejection("state", canondata.T("checktable.error.state", canondata.M{
				"num": fmt.Sprintf("%d", i+1),
			}))
		}
		path, want := val[:j], strings.TrimLeft(val[j+1:], " \t")
		switch {
		case strings.TrimRight(want, " \t\r\n") == "absent":
			then = append(then, canon.Assertion{
				Observation: "file", Condition: "not-exists", Path: path,
			})
			shorts = append(shorts, "state "+path+" absent")
		case strings.HasPrefix(want, "text "):
			bytes := yamlio.DecodeLiteral(want[len("text "):])
			then = append(then, canon.Assertion{
				Observation: "file", Condition: "bytes-equals", Path: path, Value: bytes,
			})
			shorts = append(shorts, "state "+path+" bytes")
		default:
			var probe any
			if err := json.Unmarshal([]byte(want), &probe); err == nil {
				then = append(then, canon.Assertion{
					Observation: "file", Condition: "json-equals", Path: path, Value: want,
				})
			} else {
				then = append(then, canon.Assertion{
					Observation: "file", Condition: "bytes-equals", Path: path,
					Value: yamlio.DecodeLiteral(want),
				})
			}
			shorts = append(shorts, "state "+path)
		}
	}
	return then, strings.Join(shorts, ", "), nil
}

// applyChecks — intake of the checks table: compiling rows into
// scenarios, grouping by families into requirements, row-by-row
// repair of matching rows. On implement — growth and repair only
// (the composition is not rebuilt), on deliver — a reopen by the same
// instance (a rollback to acceptance compilation).
func (b *builder) applyChecks(ck *delta.Checks) error {
	// Rows are mandatory, except for a pure answer to an open batch
	// (answers without rows) and accepting a draft (accept-draft).
	if ck == nil || (len(ck.Rows) == 0 && len(ck.Answers) == 0 && !ck.AcceptDraft) {
		return rejection("checks", canondata.T("checktable.error.rows-empty"))
	}
	if b.state.Intent == nil {
		return rejection("checks", canondata.T("checktable.error.no-intent"))
	}
	// The seed channel's quoting trap is caught at intake, from the
	// raw submission text (post-decode the evidence is gone): a
	// double-quoted seed value with byte-changing escapes silently
	// reseeds different bytes. A refusal with a recipe, not an
	// eternal md5 fight.
	if b.rawInput != "" {
		if line, trap := delta.SeedQuoteEscape(b.rawInput); trap {
			return rejection("seeds", canondata.T("checktable.error.seed-quote-escape",
				canondata.M{"line": fmt.Sprintf("%d", line)}))
		}
		if line, trap := delta.ExpectationQuoteEscape(b.rawInput); trap {
			return rejection("expectations", canondata.T("checktable.error.expectation-quote-escape",
				canondata.M{"line": fmt.Sprintf("%d", line)}))
		}
	}
	// Once code has started, the expectations of already-run checks are
	// changed only by human assertion — but row by row, not by the
	// whole submission: the fate of additions does not depend on the
	// fate of replacements (a row that proved nothing is repaired
	// freely; an addition touches nothing proven). A submission with
	// replacements of proven rows holds ONLY those rows: free rows are
	// applied by the same transaction, the held part (the delta
	// verbatim) awaits the human. On deliver, replacements of proven
	// rows are a change request (a cycle rollback, the application
	// plan in the assertion block); pure growth is free: new rows do
	// not touch green proofs, the cycle rolls back to acceptance and
	// the new rows are run by the same submission.
	if b.e.codeStarted() && !b.assertedAmend && len(ck.Rows) > 0 {
		heldIdx, changes := b.amendTriggerRows(ck)
		if len(heldIdx) > 0 {
			if len(heldIdx) == len(ck.Rows) {
				// nothing to hold besides the replacements — the submission in full
				if b.state.Stage == canon.StageDeliver {
					b.stageChangeRequest(delta.KindChecks, b.deltaKey, "", changes)
				} else {
					b.stageAmend(delta.KindChecks, b.deltaKey, "", changes, nil)
				}
				return nil
			}
			held, free := splitRows(ck.Rows, heldIdx)
			raw, err := yamlio.Marshal(delta.Checks{
				Kind: delta.KindChecks, SubmissionKey: b.deltaKey, Rows: held,
			})
			if err != nil {
				return fmt.Errorf("held delta: %w", err)
			}
			// The change request plan is not attached: the bulk effects
			// (cycle rollback, card rebuild) were already produced by
			// the free rows of this same submission — only the
			// replacement is asked.
			b.stageAmend(delta.KindChecks, b.deltaKey, string(raw), changes, nil)
			// free rows are applied now; answers/questions keep
			// riding with the table as usual
			ck.Rows = free
		}
	}

	// The clarification cycle lives inside the answer to the table:
	// an open batch is resolved by this same submission (explicit
	// answers, silence — defaults except conflicts, with the ByDefault
	// marker); a new batch rides in the same package and
	// returns to the hand in the submission's answer — the stage holds
	// until the next submission, there is no separate transfer. A
	// replay of the held delta (human assertion) is an internal
	// machine submission: it does not enact the human's silence, an
	// open batch does not count as a conversation (the silence budget
	// is spent only by the calling side's submissions). Closedness is
	// measured BEFORE resolution: a batch closed by this same
	// submission is not a "closed batch" for its answers.
	batchClosedOnEntry := b.state.Lint != nil && len(b.state.Lint.Findings) > 0 && b.state.Clarified
	if b.state.Lint != nil && len(b.state.Lint.Findings) > 0 && !b.state.Clarified && !b.assertedAmend {
		if err := b.resolveLintBatch(ck.Answers); err != nil {
			return err
		}
		// A ledger "choice" entry — only on the batch's full closure:
		// the attention meter sees the human's choice on the batch,
		// not intermediate submissions (a conflict can hold a batch
		// open across many submissions).
		b.resolvedBatch = b.state.Clarified
	}
	// An explicit answer to a closed batch is not silently lost: a
	// default recorded by silence is upgraded to the human's explicit
	// choice; choices already explicit are untouchable — the answer is
	// named a notification.
	if batchClosedOnEntry && len(ck.Answers) > 0 {
		if err := b.upgradeSilentAnswers(ck.Answers); err != nil {
			return err
		}
	}
	if len(ck.Questions) > 0 {
		if err := validateLintFindings(ck.Questions); err != nil {
			return err
		}
		// A new batch — a new dialogue: the clarification list starts
		// over (past choices remain in the journal); question
		// identifiers are numbered from Q1 in every batch, without the
		// reset past records would block the new batch's answers.
		b.state.Clarifications = nil
		findings := make([]canon.LintFinding, len(ck.Questions))
		copy(findings, ck.Questions)
		b.state.Lint = &canon.LintResult{Findings: findings}
		b.state.Clarified = false
		b.openedQuestions = len(findings)
	}
	// Accepting a draft without edits: the hand's review confirmation,
	// no rows — the canon already carries the draft table (the
	// questions above could have been resolved by this same
	// submission).
	if len(ck.Rows) == 0 && ck.AcceptDraft {
		return nil
	}

	// The wish's families — against the surface names from the table's rows.
	names := map[string]bool{}
	for _, row := range ck.Rows {
		for _, cmd := range row.Run {
			fields, err := cmd.Fields()
			if err == nil && len(fields) > 0 {
				names[fields[0]] = true
			}
		}
	}
	fams := commandFamilies(*b.state.Intent, names)
	// The whole table's surface names — for every scenario of the
	// submission: the runner brings them into every run so that
	// references to the product from inside commands (the shell) find
	// the files. The union of this submission's names and all existing
	// scenarios: a row added later sees the same names as the old
	// rows.
	for _, sc := range b.scns {
		if sc.When != nil && len(sc.When.Command) > 0 {
			names[sc.When.Command[0]] = true
		}
		for _, n := range sc.SurfaceNames {
			names[n] = true
		}
	}
	tableSurfaceNames := sortedKeys(names)

	// One intention — one scenario: a table with two or more distinct
	// surfaces (command[0]) is a candidate for splitting into
	// different tasks. Candidates are only names NOT provided by the
	// environment: a PATH name (shell, row utility) is resolved by the
	// environment on the same grounds as a run and is not a product;
	// a product is an artifact record or a submission file. The
	// machine opens a choice question (silence applies the recommended
	// "one product" — not a blocker); the question is asked once per
	// instance (the QC marker in choice records).
	var candidates []string
	for _, n := range sortedKeys(names) {
		if _, err := exec.LookPath(n); err != nil {
			candidates = append(candidates, n)
		}
	}
	if len(candidates) >= 2 {
		b.openCompositeQuestion(candidates)
	}

	// Cross-row validation of setups: a setup command that in another
	// row OF THIS SAME SUBMISSION WITH THE SAME SEED AND MATERIALS is
	// expected with a non-zero code is structurally impassable (a
	// setup must exit 0, the composition is frozen after code starts)
	// — caught at table intake, not by an eternal red afterwards. The
	// context is the same identity as the row identity
	// (seed+materials+commands): the same command with a different
	// seed is different behavior, not a collision.
	type failKey struct {
		state string
		argv  string
	}
	stateKey := func(row delta.CheckRow) string {
		seed, materials, _, _, err := b.rowIdentity(row)
		if err != nil {
			return "\x00broken"
		}
		parts := make([]string, 0, len(seed)+len(materials))
		for _, s := range seed {
			parts = append(parts, s.Path+"\x1e"+s.Content)
		}
		for _, m := range materials {
			parts = append(parts, "m\x1e"+m.From)
		}
		return strings.Join(parts, "\x1d")
	}
	failing := map[failKey]int{} // state+argv → row number
	for i, row := range ck.Rows {
		rc := 0
		if row.Rc != nil {
			rc = *row.Rc
		}
		if rc == 0 || len(row.Run) == 0 {
			continue
		}
		fields, err := row.Run[len(row.Run)-1].Fields()
		if err != nil || len(fields) == 0 {
			continue
		}
		key := failKey{state: stateKey(row), argv: strings.Join(fields, "\x1f")}
		if _, dup := failing[key]; !dup {
			failing[key] = i + 1
		}
	}
	for i, row := range ck.Rows {
		if len(row.Run) < 2 {
			continue
		}
		state := stateKey(row)
		for j, cmd := range row.Run[:len(row.Run)-1] {
			fields, err := cmd.Fields()
			if err != nil || len(fields) == 0 {
				continue
			}
			if src, bad := failing[failKey{state: state, argv: strings.Join(fields, "\x1f")}]; bad {
				return rejection(fmt.Sprintf("rows[%d]", i+1),
					canondata.T("checktable.error.setup-expected-fail", canondata.M{
						"run": fmt.Sprintf("%d", j+1), "row": fmt.Sprintf("%d", src),
						"command": strings.Join(fields, " "),
					}))
			}
		}
	}

	// The identities of existing scenarios — for row-by-row replacement.
	existing := map[string]string{} // key → scenario id
	// The same identity without materials: a row with the same seed
	// and commands but a different material composition is NOT a
	// replacement (materials are part of the identity), the machine
	// names such an addition explicitly — a silent duplicate in the
	// canon is impossible.
	byShape := map[string][]string{} // key without materials → scenario ids
	for _, sc := range b.scns {
		if k := scenarioKey(sc); k != "" {
			existing[k] = sc.ID
			shape := rowKey(sc.Seed, nil, sc.Pre, sc.When.Command)
			byShape[shape] = append(byShape[shape], sc.ID)
		}
	}
	// The authorial repair keys of existing scenarios: a never-green
	// row is replaced under its key even with different seeds and
	// run — the repair of a broken row must not orphan the broken
	// original. First holder wins, deterministically by id.
	byKey := map[string]string{} // authorial key → scenario id
	for _, id := range sortedIDs(b.scns) {
		if k := b.scns[id].Key; k != "" {
			if _, taken := byKey[k]; !taken {
				byKey[k] = id
			}
		}
	}

	newScenarioIDs := []string{}
	var matRows, matNews, matOlds []string
	for i, row := range ck.Rows {
		what := fmt.Sprintf("rows[%d]", i+1)
		if len(row.Run) == 0 {
			return rejection(what, canondata.T("checktable.error.run-required"))
		}
		seed, materials, pre, when, err := b.rowIdentity(row)
		if err != nil {
			return rejection(what, err.Error())
		}
		// A flattened argv: every run position is a WHOLE command;
		// the signature of single-word elements is a named finding in
		// the answer, garbage scenarios do not stay silent.
		if n := flatRunRow(row.Run); n > 0 {
			b.rowNotes = append(b.rowNotes, canondata.T("checktable.note.run-elements",
				canondata.M{"row": fmt.Sprintf("%d", i+1), "count": fmt.Sprintf("%d", n)}))
		}
		b.noteSeedQuotes(what, seed)
		if seed == nil {
			seed = []canon.Seed{}
		}
		then, short, err := rowExpectations(row)
		if err != nil {
			return err
		}

		sc := canon.Scenario{
			Key:          row.Key,
			Seed:         seed,
			Materials:    materials,
			Pre:          pre,
			SurfaceNames: tableSurfaceNames,
			When: &canon.When{
				Surface:    "cli",
				Command:    when,
				TimeoutSec: rowTimeoutOr(row.TimeoutSec),
				Volatile:   row.Volatile,
			},
			Then:    then,
			Summary: fmt.Sprintf("%s → %s", strings.Join(when, " "), short),
		}

		// Row identity FIRST, the requirement after: a row replacing an
		// existing one — the same seed, materials, and the same
		// sequence, or its authorial key on a never-green row — KEEPS
		// the scenario's requirement binding. The binding is the
		// executor's re-cut (update-scenario is the rebinding channel);
		// re-deriving it from the command spelling on repair re-drafted
		// per-family placeholders and silently moved rows — proven ones
		// included — off the re-cut features, leaving the real
		// requirements without a single scenario.
		if id, ok := existing[rowKey(seed, materials, pre, when)]; ok {
			sc.ID = id
		} else if row.Key != "" {
			if id, ok := byKey[row.Key]; ok {
				// An asserted amend replay replaces the row in place:
				// the human assertion HAS resolved the protection this
				// hold was staged for — inheriting the ID keeps the
				// scenario (and its requirement binding) instead of
				// spawning a twin on a fresh family draft.
				if !b.rowProtected(b.scns[id]) || b.assertedAmend {
					sc.ID = id
					delete(byKey, row.Key)
					b.rowNotes = append(b.rowNotes, canondata.T("checktable.note.key-replaced",
						canondata.M{"row": fmt.Sprintf("%d", i+1), "scn": id, "key": row.Key}))
				} else {
					// A proven row keeps the change-request channel: what
					// reaches here (the hold let it pass — identical
					// expectations, a different run) is honestly a new
					// row, the machine names the key collision.
					b.rowNotes = append(b.rowNotes, canondata.T("checktable.note.key-held",
						canondata.M{"row": fmt.Sprintf("%d", i+1), "scn": id, "key": row.Key}))
				}
			}
		}
		if sc.ID != "" {
			if req, ok := b.reqs[b.scns[sc.ID].Requirement]; ok && req.ID != "" {
				sc.Requirement = req.ID
			}
		}
		// A genuinely NEW row scaffolds onto the per-family draft: the
		// table's authoring window (the canon's re-cut contract); the
		// re-cut into living words is the executor's, the machine only
		// gives every row a requirement to hang on.
		if sc.Requirement == "" {
			family := familyOf(when, fams)
			reqID := ""
			for _, r := range b.reqs {
				if r.Formulation == reqFormulationOf(b.state.Language, family) {
					reqID = r.ID
					break
				}
			}
			if reqID == "" {
				next := b.state.Counters["REQ"]
				if next < 1 {
					next = 1
				}
				reqID = fmt.Sprintf("REQ-%03d", next)
				b.state.Counters["REQ"] = next + 1
				if err := b.putReq(canon.Requirement{
					ID: reqID, Status: canon.ReqAccepted,
					Dependencies: []string{}, Formulation: reqFormulationOf(b.state.Language, family),
				}); err != nil {
					return err
				}
			}
			sc.Requirement = reqID
		}
		if sc.ID == "" {
			next := b.state.Counters["SCN"]
			if next < 1 {
				next = 1
			}
			sc.ID = fmt.Sprintf("SCN-%03d", next)
			b.state.Counters["SCN"] = next + 1
			existing[rowKey(seed, materials, pre, when)] = sc.ID
			newScenarioIDs = append(newScenarioIDs, sc.ID)
			// An addition indistinguishable from a replacement to the
			// eye (the same seed and commands, a different material
			// composition) is a named finding in the answer: the old
			// rows remain in the canon.
			if olds := byShape[rowKey(seed, nil, pre, when)]; len(olds) > 0 {
				matRows = append(matRows, fmt.Sprintf("%d", i+1))
				matNews = append(matNews, sc.ID)
				for _, old := range olds {
					if !slices.Contains(matOlds, old) {
						matOlds = append(matOlds, old)
					}
				}
			}
		}
		if err := b.validateScenario(sc); err != nil {
			return err
		}
		if err := b.putScn(sc); err != nil {
			return err
		}
	}

	// A note on twin additions: one line per submission — all rows
	// whose shape matched existing ones without materials.
	if len(matRows) > 0 {
		b.rowNotes = append(b.rowNotes, canondata.T("checktable.note.materials-identity",
			canondata.M{
				"rows": strings.Join(matRows, ", "),
				"news": strings.Join(matNews, ", "),
				"olds": strings.Join(matOlds, ", "),
			}))
	}

	// Orphaned machine requirements of families do not remain: a row
	// without a scenario does not exist, and neither does a
	// requirement without a row.
	scnsByReq := map[string]int{}
	for _, sc := range b.scns {
		scnsByReq[sc.Requirement]++
	}
	for _, id := range sortedIDs(b.reqs) {
		if scnsByReq[id] == 0 && isTableRequirement(b.state.Language, b.reqs[id]) {
			delete(b.reqs, id)
			b.effects = append(b.effects, journal.Effect{
				Path: canonFile("requirements", id), Delete: true,
			})
			b.touch(id)
		}
	}
	// The spec IR holds after EVERY channel that touches bindings, the
	// table channel included: a requirement left without scenarios is
	// a poisoned state — its only repair routes (rebind, detail
	// rewrite) are themselves refused against it. The machine may not
	// write such a state, whatever path led here.
	if err := b.checkSpecIR(); err != nil {
		return err
	}

	// Growth on implement: new rows go into the card that owns the
	// state files; replacing an existing row does not change the
	// composition.
	if b.state.Stage == canon.StageImplement {
		for _, id := range newScenarioIDs {
			if err := b.attachGrownScenario(id); err != nil {
				return err
			}
		}
	}
	// Reopen: a table on deliver rolls the cycle back to acceptance
	// compilation by the same instance — the change passes through
	// the same stages anew.
	if b.state.Stage == canon.StageDeliver {
		b.state.Stage = canon.StageACCompile
		b.refreshCard()
		b.reopenedCycle = true
	}
	return nil
}

// renderCheckTable — the machine render of an observation set as
// table rows (backfill on deliver proposes a table, not GWT).
func renderCheckTable(rows []delta.CheckRow, keyNum int) string {
	var sb strings.Builder
	sb.WriteString(canondata.T("checktable.render.heredoc"))
	sb.WriteString("\n")
	sb.WriteString("kind: checks\n")
	fmt.Fprintf(&sb, "submission-key: checks-%03d\n", keyNum)
	sb.WriteString("rows:\n")
	for _, r := range rows {
		if r.Key != "" {
			fmt.Fprintf(&sb, "  key: %s\n", r.Key)
		}
		if len(r.Seeds) > 0 {
			sb.WriteString("  seeds:\n")
			for _, s := range r.Seeds {
				fmt.Fprintf(&sb, "    - %s\n", s)
			}
		}
		sb.WriteString("  run:\n")
		for _, c := range r.Run {
			fmt.Fprintf(&sb, "    - %s\n", c.YAML())
		}
		if r.Out != nil {
			fmt.Fprintf(&sb, "  out: %s\n", yamlScalarOf(*r.Out))
		}
		if r.Err != nil {
			fmt.Fprintf(&sb, "  err: %s\n", yamlScalarOf(*r.Err))
		}
		if r.Rc != nil {
			fmt.Fprintf(&sb, "  rc: %d\n", *r.Rc)
		}
		if len(r.State) > 0 {
			sb.WriteString("  state:\n")
			for _, s := range r.State {
				fmt.Fprintf(&sb, "    - %s\n", s)
			}
		}
	}
	sb.WriteString(canondata.T("checktable.render.submit"))
	return sb.String()
}

// yamlScalarOf — a one-line value as a YAML scalar: short without
// newlines — quoted, multi-line — a literal block with chomping by
// the value's actual tail (clip/keep/strip): the proposed row is
// resubmitted byte-for-byte, trailing newlines are part of the
// expectation, not formatting.
func yamlScalarOf(s string) string {
	if !strings.Contains(s, "\n") {
		return fmt.Sprintf("%q", s)
	}
	var sb strings.Builder
	body := s
	switch {
	case strings.HasSuffix(s, "\n\n"):
		sb.WriteString("|+") // keep: all trailing newlines
		body = strings.TrimSuffix(s, "\n")
	case strings.HasSuffix(s, "\n"):
		sb.WriteString("|") // clip: one trailing newline
		body = strings.TrimSuffix(s, "\n")
	default:
		sb.WriteString("|-") // strip: no trailing newline
	}
	for _, line := range strings.Split(body, "\n") {
		sb.WriteString("\n      ")
		if line != "" {
			sb.WriteString(line)
		}
	}
	return sb.String()
}

// compositeQuestionID — the pseudo-identifier of the machine's
// compositeness question: outside the Q<n> numbering of hand batches,
// the answer is looked up by it.
const compositeQuestionID = "QC"

// openCompositeQuestion — the machine question "one product or
// separate tasks": appended to an open batch as a finding with
// options and a recommendation; one already decided (a QC choice
// record) is not asked again.
func (b *builder) openCompositeQuestion(names []string) {
	for _, c := range b.state.Clarifications {
		if c.Question == compositeQuestionID {
			return
		}
	}
	for _, f := range openFindings(b.state) {
		if f.ID == compositeQuestionID {
			return
		}
	}
	if b.state.Lint == nil {
		b.state.Lint = &canon.LintResult{}
	}
	b.state.Lint.Findings = append(b.state.Lint.Findings, canon.LintFinding{
		ID:        compositeQuestionID,
		Dimension: canon.LintAmbiguous,
		Finding:   canondata.T("composite.finding", canondata.M{"names": strings.Join(names, ", ")}),
		Question:  canondata.T("composite.question"),
		Options: []canon.LintOption{
			// the default is the machine's recommendation: one product
			{ID: "a", Label: canondata.T("composite.one.label"), Scope: canondata.T("composite.one.scope"), Default: true},
			{ID: "b", Label: canondata.T("composite.split.label"), Scope: canondata.T("composite.split.scope")},
		},
	})
	b.state.Clarified = false
	b.openedQuestions++
}

// splitRows — splitting a submission's rows by hold indexes: free
// rows are applied now, held rows await the human with their own
// delta.
func splitRows(rows []delta.CheckRow, heldIdx []int) (held, free []delta.CheckRow) {
	heldSet := map[int]bool{}
	for _, i := range heldIdx {
		heldSet[i] = true
	}
	for i, r := range rows {
		if heldSet[i] {
			held = append(held, r)
		} else {
			free = append(free, r)
		}
	}
	return held, free
}

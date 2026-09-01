// The product passport: the boundary registry and environment
// conventions. A boundary = a requirement; the machine keeps the
// records (spec synchronization and gate runs), the anchors — code
// file digests, the spec digest, and the journal transaction — are
// collected and verified by the machine. Freshness is by the
// scope-hash of content (git is not read); editing the conventions
// changes their digest and thereby invalidates the reconciliation of
// all boundaries.
package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/neurophant/punchtape/internal/canon"
	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/checks"
	"github.com/neurophant/punchtape/internal/delta"
	"github.com/neurophant/punchtape/internal/journal"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// toolOutcome — the outcome of a static gate command from the
// conventions: the name, green/red, the red reason in one line, and
// whether the gate was applied. An undeclared recipe means the gate
// was not applied: green, but it did not check.
type toolOutcome struct {
	Name    string
	Green   bool
	Reason  string
	Applied bool
}

// applyConventions — validation and application of a conventions
// declaration: the section is replaced in full (the last submission
// is the truth). The build must place {out}; the placeholders are
// only {out} and {name}.
func (b *builder) applyConventions(c *delta.Conventions) error {
	if c == nil || (c.Build == nil && c.Types == nil && c.Lint == nil) {
		return rejection("conventions", canondata.LintMessage("conventions.empty"))
	}
	conv := canon.Conventions{SchemaVersion: canon.SchemaVersion}
	var err error
	if conv.Build, err = recipeOf("build", c.Build, true); err != nil {
		return err
	}
	if conv.Types, err = recipeOf("types", c.Types, false); err != nil {
		return err
	}
	if conv.Lint, err = recipeOf("lint", c.Lint, false); err != nil {
		return err
	}
	data, err := yamlio.Marshal(conv)
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	b.effects = append(b.effects, journal.Effect{
		Path: canon.ConventionsFile, Content: machineFile(data),
	})
	// The gate lease is proof about a specific recipe: changing the
	// types/lint recipe devalues it (the green runs belonged to the
	// old commands) — the gate returns to duty and pays the lease
	// anew. Retiring without a recipe change lives as before.
	if old, err := b.e.Store.Conventions(); err == nil && old != nil {
		recipesChanged := map[string]bool{}
		if !recipeEqual(old.Types, conv.Types) {
			recipesChanged["types"] = true
		}
		if !recipeEqual(old.Lint, conv.Lint) {
			recipesChanged["lint"] = true
		}
		if len(recipesChanged) > 0 {
			kept := b.state.RetiredGates[:0]
			for _, g := range b.state.RetiredGates {
				if !recipesChanged[g] {
					kept = append(kept, g)
				}
			}
			b.state.RetiredGates = kept
		}
	}
	return nil
}

// recipeEqual — recipes match in commands and artifact.
func recipeEqual(a, r *canon.Recipe) bool {
	if a == nil || r == nil {
		return a == nil && r == nil
	}
	if a.Out != r.Out || len(a.Commands) != len(r.Commands) {
		return false
	}
	for i := range a.Commands {
		if len(a.Commands[i]) != len(r.Commands[i]) {
			return false
		}
		for j := range a.Commands[i] {
			if a.Commands[i][j] != r.Commands[i][j] {
				return false
			}
		}
	}
	return true
}

// recipeOf validates a recipe: commands non-empty, arguments
// non-empty, placeholders only {out}/{name}; the artifact path (if
// declared) is relative, inside the project. A build without {out}
// and without a declared artifact does not know where the surface
// is — a refusal.
func recipeOf(section string, d *delta.RecipeDecl, needsOut bool) (*canon.Recipe, error) {
	if d == nil {
		return nil, nil
	}
	mk := func(what string, i int) canondata.M {
		return canondata.M{"section": section, "index": strconv.Itoa(i), "ph": what}
	}
	if len(d.Commands) == 0 {
		return nil, rejection("conventions",
			canondata.LintMessage("conventions.commands-empty", canondata.M{"section": section}))
	}
	if d.Out != "" && (filepath.IsAbs(d.Out) || d.Out != filepath.ToSlash(filepath.Clean(filepath.FromSlash(d.Out))) ||
		strings.Contains(d.Out, "..") || strings.Contains(d.Out, "\\")) {
		return nil, rejection("conventions",
			canondata.LintMessage("conventions.out-path", canondata.M{
				"section": section, "out": d.Out,
			}))
	}
	out := d.Out != ""
	for i, cmd := range d.Commands {
		if len(cmd) == 0 {
			return nil, rejection("conventions",
				canondata.LintMessage("conventions.command-empty", mk("", i)))
		}
		for _, tok := range cmd {
			if tok == "" {
				return nil, rejection("conventions",
					canondata.LintMessage("conventions.command-empty", mk("", i)))
			}
			rest := stripPlaceholders(tok)
			if rest == "" && (tok == "{out}" || tok == "{name}") {
				if tok == "{out}" {
					out = true
				}
				continue
			}
			if strings.ContainsAny(rest, "{}") {
				return nil, rejection("conventions",
					canondata.LintMessage("conventions.placeholder", mk(unknownPlaceholder(tok), i)))
			}
			out = out || strings.Contains(tok, "{out}")
		}
	}
	if needsOut && !out {
		return nil, rejection("conventions", canondata.LintMessage("conventions.out"))
	}
	return &canon.Recipe{Out: d.Out, Commands: d.Commands}, nil
}

// stripPlaceholders removes the known placeholders from a token.
func stripPlaceholders(tok string) string {
	tok = strings.ReplaceAll(tok, "{out}", "")
	return strings.ReplaceAll(tok, "{name}", "")
}

// unknownPlaceholder extracts the first {...} fragment of a token for the refusal.
func unknownPlaceholder(tok string) string {
	if i := strings.IndexByte(tok, '{'); i >= 0 {
		if j := strings.IndexByte(tok[i:], '}'); j > 0 {
			return tok[i : i+j+1]
		}
	}
	return tok
}

// placeholderContext — a recipe's substitutions for diagnosing a red:
// what {name}/{out} meant in this run. The hand does not see the
// surface's inner mechanics — a red recipe must name its
// substitutions, otherwise guessing the command blind costs red cycles.
func placeholderContext(rec *canon.Recipe, name, outPath string) string {
	usedName, usedOut := false, false
	for _, cmd := range rec.Commands {
		for _, tok := range cmd {
			usedName = usedName || strings.Contains(tok, "{name}")
			usedOut = usedOut || strings.Contains(tok, "{out}")
		}
	}
	var parts []string
	if usedName {
		parts = append(parts, "{name}="+name)
	}
	if usedOut {
		parts = append(parts, "{out}="+outPath)
	}
	return strings.Join(parts, ", ")
}

// runPlainCommand — running a recipe command without a shell from the
// project root with a timeout: a non-zero code or timeout is a
// one-line error. An environment failure (command not installed, no
// permission, timeout) carries a diagnosis and a FIX hint from the
// catalog.
func runPlainCommand(workdir string, argv []string, timeoutSec int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = workdir
	raw, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	if ctx.Err() == context.DeadlineExceeded {
		return errors.New(canondata.EnvDiag("command-timeout", canondata.M{
			"sec": strconv.Itoa(timeoutSec), "command": strings.Join(argv, " "),
		}))
	}
	if class := checks.EnvClass(err); class != "" {
		return errors.New(canondata.EnvDiag(class, canondata.M{
			"command": strings.Join(argv, " "),
		}))
	}
	text := onelineCmd(string(raw))
	if text == "" {
		text = err.Error()
	}
	return errors.New(text)
}

// onelineCmd squeezes a command's output into one line: newlines
// become "; ", the tail after 400 characters is dropped.
func onelineCmd(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", "; "))
	if len(s) > 400 {
		s = s[:400] + "…"
	}
	return s
}

// runBuildRecipe runs the surface build recipe: commands in order
// with {out}/{name} substitution; the artifact is a file or
// directory, the path is declared by the operator (Out) or equals the
// command name. The execute permission of a file artifact is a
// property of the record, not of the code. A missing artifact after
// the recipe is an honest error.
func (e *Engine) runBuildRecipe(rec *canon.Recipe, name string) error {
	outRel := rec.Out
	if outRel == "" {
		outRel = name
	}
	outPath := filepath.Join(e.Workdir, filepath.FromSlash(outRel))
	for _, cmd := range rec.Commands {
		argv := make([]string, len(cmd))
		for i, tok := range cmd {
			argv[i] = strings.ReplaceAll(tok, "{name}", name)
			argv[i] = strings.ReplaceAll(argv[i], "{out}", outPath)
		}
		if err := runPlainCommand(e.Workdir, argv, canondata.Limit("conventions.build-timeout-s")); err != nil {
			if ctx := placeholderContext(rec, name, outPath); ctx != "" {
				return fmt.Errorf("%s (%s)", err, ctx)
			}
			return err
		}
	}
	info, err := os.Stat(outPath)
	if err != nil {
		return errors.New(canondata.EnvDiag("artifact-missing", canondata.M{"out": outRel}))
	}
	if info.IsDir() {
		return nil // a directory artifact: the surface tree is assembled as is
	}
	if info.Mode()&0o111 == 0 {
		return os.Chmod(outPath, info.Mode()|0o111)
	}
	return nil
}

// gateFromRecipe — a static gate from the conventions: an undeclared
// recipe is not applied (green, but it did not check); a declared one
// runs all the commands from the project root with the same
// placeholder substitution as the build ({name} is the surface name,
// {out} is the artifact path), the first red is the reason.
func gateFromRecipe(workdir, name, surfaceName string, rec *canon.Recipe, outPath string, timeoutSec int) toolOutcome {
	if rec == nil {
		return toolOutcome{Name: name, Green: true}
	}
	for _, cmd := range rec.Commands {
		argv := make([]string, len(cmd))
		for i, tok := range cmd {
			argv[i] = strings.ReplaceAll(tok, "{name}", surfaceName)
			argv[i] = strings.ReplaceAll(argv[i], "{out}", outPath)
		}
		if err := runPlainCommand(workdir, argv, timeoutSec); err != nil {
			reason := err.Error()
			if ctx := placeholderContext(rec, surfaceName, outPath); ctx != "" {
				reason = reason + " (" + ctx + ")"
			}
			return toolOutcome{Name: name, Applied: true, Green: false, Reason: reason}
		}
	}
	return toolOutcome{Name: name, Applied: true, Green: true}
}

// boundaryFileName — a boundary record's file: req-001.yaml.
func boundaryFileName(id string) string {
	return boundariesDirPath + strings.ToLower(id) + ".yaml"
}

// passportSync — synchronizing the boundary registry with the spec:
// a new requirement gets a record (claimed, origin — the wish), a
// vanished requirement takes its record away, a changed boundary
// spec is claimed anew (the proof was against a different spec).
// The domain is the scenarios' command families, the digest is over
// the canonical YAML of the requirement and its scenarios.
func (b *builder) passportSync() error {
	scnsByReq := map[string][]canon.Scenario{}
	for _, sc := range b.scns {
		scnsByReq[sc.Requirement] = append(scnsByReq[sc.Requirement], sc)
	}
	for id := range b.boundaries {
		if _, ok := b.reqs[id]; ok {
			continue
		}
		delete(b.boundaries, id)
		b.effects = append(b.effects, journal.Effect{Path: boundaryFileName(id), Delete: true})
	}
	for id, req := range b.reqs {
		digest := boundaryDigest(req, scnsByReq[id])
		domain := boundaryDomain(scnsByReq[id])
		old, ok := b.boundaries[id]
		if !ok {
			// The boundary's origin: the wish produced the spec before
			// code — wish; the code came first (the boundary claimed over
			// the existing) — backfill, a formalization of the survivor.
			origin := canon.OriginWish
			if b.e.codeStarted() {
				origin = canon.OriginBackfill
			}
			if err := b.putBoundary(canon.Boundary{
				ID: id, Status: canon.BoundaryClaimed, Origin: origin,
				Domain: domain, Digest: digest,
			}); err != nil {
				return err
			}
			continue
		}
		if old.Digest == digest && equalStrings(old.Domain, domain) {
			continue
		}
		// A changed spec — a re-claim: the previous
		// proof was against a different spec; the transition target is
		// ladder data (drift → claimed).
		driftTo, _ := canondata.TrustTarget("drift")
		old.Digest = digest
		old.Domain = domain
		old.Status = driftTo
		old.Verdict = nil
		if err := b.putBoundary(old); err != nil {
			return err
		}
	}
	return nil
}

// putBoundary — writing a passport boundary record: optional lists are
// written by absence, not by stubs; the header-manifest is inside
// the effect.
func (b *builder) putBoundary(bd canon.Boundary) error {
	data, err := yamlio.Marshal(bd)
	if err != nil {
		return fmt.Errorf("system failure: %w", err)
	}
	b.boundaries[bd.ID] = bd
	b.effects = append(b.effects, journal.Effect{
		Path: boundaryFileName(bd.ID), Content: machineFile(data),
	})
	return nil
}

// boundaryDigest — the digest of a boundary's spec: the canonical YAML
// of the requirement and all its scenarios.
func boundaryDigest(req canon.Requirement, scns []canon.Scenario) string {
	var sb strings.Builder
	if data, err := yamlio.Marshal(req); err == nil {
		sb.Write(data)
	}
	ids := make([]string, 0, len(scns))
	byID := map[string]canon.Scenario{}
	for _, sc := range scns {
		ids = append(ids, sc.ID)
		byID[sc.ID] = sc
	}
	sort.Strings(ids)
	for _, id := range ids {
		if data, err := yamlio.Marshal(byID[id]); err == nil {
			sb.Write(data)
		}
	}
	return yamlio.Digest([]byte(sb.String()))
}

// boundaryDomain — a boundary's command families: the first tokens of
// the executable scenarios' commands, unique and sorted.
func boundaryDomain(scns []canon.Scenario) []string {
	seen := map[string]bool{}
	for _, sc := range scns {
		if sc.When == nil || len(sc.When.Command) == 0 {
			continue
		}
		seen[sc.When.Command[0]] = true
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// boundaryScope — boundary code files from the journal: slice cards
// give the files of their (exclusive) ownership; a whole-product card
// makes all code files shared, the attribution honestly names the
// coarseness reason.
func (e *Engine) boundaryScope(cards []canon.Card, scns []canon.Scenario) (map[string][]string, map[string]string) {
	scnsByReq := map[string][]canon.Scenario{}
	for _, sc := range scns {
		scnsByReq[sc.Requirement] = append(scnsByReq[sc.Requirement], sc)
	}
	whole := false
	for _, c := range cards {
		if len(c.Files) == 0 {
			whole = true
		}
	}
	shared := e.codeFilesOfJournal()
	scope := map[string][]string{}
	notes := map[string]string{}
	for reqID := range scnsByReq {
		if whole {
			scope[reqID] = shared
			notes[reqID] = canondata.T("passport.note.whole-product")
			continue
		}
		var files []string
		for _, c := range cards {
			if !cardCoversAny(c, scnsByReq[reqID]) {
				continue
			}
			files = mergeUnique(files, c.Files)
		}
		sort.Strings(files)
		scope[reqID] = files
	}
	return scope, notes
}

// cardCoversAny — whether a card covers one of the scenarios.
func cardCoversAny(c canon.Card, scns []canon.Scenario) bool {
	for _, sid := range c.Scenarios {
		for _, sc := range scns {
			if sc.ID == sid {
				return true
			}
		}
	}
	return false
}

// codeFilesOfJournal — code files ever written by code/fix
// submissions: a fold of the journal's effects, the last Write wins.
func (e *Engine) codeFilesOfJournal() []string {
	files := map[string]bool{}
	for _, en := range e.Journal.All() {
		if en.DeltaKind != deltaKindCode && en.DeltaKind != deltaKindFix {
			continue
		}
		for _, ef := range en.Effects {
			if strings.HasPrefix(filepath.ToSlash(ef.Path), ".punchtape/") {
				continue
			}
			if ef.Delete {
				delete(files, ef.Path)
				continue
			}
			files[ef.Path] = true
		}
	}
	out := make([]string, 0, len(files))
	for p := range files {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// scopeHash — freshness by content: sha256 over sorted
// "path + content digest" pairs; a missing file enters with a missing
// marker and is named separately — fail-safe, git is not read. The
// digests go through a cache with newline normalization: the same
// semantics on any OS, unchanged files are not re-read.
func (e *Engine) scopeHash(files []string) (string, []string) {
	sorted := append([]string(nil), files...)
	sort.Strings(sorted)
	var sb strings.Builder
	var missing []string
	for _, f := range sorted {
		d, ok := e.digestOf(f)
		if !ok {
			missing = append(missing, f)
			fmt.Fprintf(&sb, "%s: missing\n", f)
			continue
		}
		fmt.Fprintf(&sb, "%s: %s\n", f, d)
	}
	return yamlio.Digest([]byte(sb.String())), missing
}

// passportProof — machine trust transitions after a gate run: the
// transition targets are named by ladder data (trust.yaml,
// green-run → verified, drift → claimed; above verified there are no
// proofs — validity and convergence are ascribed elsewhere). Drift
// (code files or conventions changed against the verdict) demotes the
// status to claimed; a green card proves its scenarios' boundaries —
// verified against the verdict's scope-hash, the conventions digest,
// and the anchor transaction. The scope and the attribution
// coarseness reason are fixed in the record by the same transaction.
func (b *builder) passportProof(card canon.Card, tx string, allGreen bool) error {
	scns := make([]canon.Scenario, 0, len(b.scns))
	for _, id := range sortedIDs(b.scns) {
		scns = append(scns, b.scns[id])
	}
	cards := make([]canon.Card, 0, len(b.cards))
	for _, id := range sortedIDs(b.cards) {
		cards = append(cards, b.cards[id])
	}
	scope, notes := b.e.boundaryScope(cards, scns)
	convDigest := b.e.Store.ConventionsDigest()
	provenTo, _ := canondata.TrustTarget("green-run")
	driftTo, _ := canondata.TrustTarget("drift")
	for id, bd := range b.boundaries {
		files := scope[id]
		hash, _ := b.e.scopeHash(files)
		drifted := bd.Verdict == nil ||
			bd.Verdict.ScopeHash != hash ||
			bd.Verdict.ConventionsDigest != convDigest
		proven := allGreen && cardCoversAny(card, scnsOfReq(scns, id))
		switch {
		case proven:
			bd.Status = provenTo
			bd.Verdict = &canon.BoundaryVerdict{
				ScopeHash: hash, ConventionsDigest: convDigest, JournalTx: tx,
			}
		case drifted:
			bd.Status = driftTo
			bd.Verdict = nil
		default:
			continue
		}
		bd.ScopeFiles = files
		bd.Attribution = notes[id]
		if err := b.putBoundary(bd); err != nil {
			return err
		}
	}
	return nil
}

// scnsOfReq — a requirement's scenarios from a slice.
func scnsOfReq(scns []canon.Scenario, reqID string) []canon.Scenario {
	var out []canon.Scenario
	for _, sc := range scns {
		if sc.Requirement == reqID {
			out = append(out, sc)
		}
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// missingSurfaces — material for the build-recipe question at an empty
// start. The rows are an arbitrary argv: a name without a file in the
// project can be an environment tool (PATH), the machine does not
// know this and does not decide — the question is raised only when
// NOT ONE name named by the table is found in the project and there
// is no recipe: only then does the surface truly not exist in any
// form.
func (e *Engine) missingSurfaces(card canon.Card) []string {
	if conv, err := e.Store.Conventions(); err == nil && conv != nil && conv.Build != nil && conv.Build.Out != "" {
		return nil
	}
	scns, err := e.Store.Scenarios()
	if err != nil {
		return nil
	}
	byID := map[string]canon.Scenario{}
	for _, sc := range scns {
		byID[sc.ID] = sc
	}
	seen := map[string]bool{}
	var out []string
	for _, scnID := range card.Scenarios {
		sc, ok := byID[scnID]
		if !ok || sc.When == nil || len(sc.When.Command) == 0 {
			continue
		}
		name := sc.When.Command[0]
		if seen[name] {
			continue
		}
		seen[name] = true
		if _, err := os.Stat(filepath.Join(e.Workdir, name)); err != nil {
			out = append(out, name)
		} else {
			return nil // at least one name resolves from the project — the surface exists
		}
	}
	sort.Strings(out)
	return out
}

// cardHasRunChecks — whether a card has had runs: at least one of its
// checks left unknown. A card without runs — no code yet, it is too
// early to run the gates (they would paint "does not build" over
// unstarted work).
func (e *Engine) cardHasRunChecks(card canon.Card) bool {
	checks, err := e.Store.Checks()
	if err != nil {
		return false
	}
	own := map[string]bool{}
	for _, scnID := range card.Scenarios {
		own[scnID] = true
	}
	for _, c := range checks {
		if own[c.Scenario] && c.Outcome != canon.CheckUnknown {
			return true
		}
	}
	return false
}

// workingCard — the working card: the first active one, else the first
// red; empty — no cards (as the next slot chooses).
func (e *Engine) workingCard() string {
	cards, err := e.Store.Cards()
	if err != nil {
		return ""
	}
	for _, c := range cards {
		if c.Status == canon.CardActive {
			return c.ID
		}
	}
	for _, c := range cards {
		if c.Status == canon.CardRed {
			return c.ID
		}
	}
	if len(cards) > 0 {
		return cards[0].ID
	}
	return ""
}

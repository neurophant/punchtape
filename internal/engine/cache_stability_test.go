// The cache-stability law tests: the table's
// render is INSERTION-STABLE between versions — a replacement edits
// its row's line in place, an added row lands at the END of its
// table — so the platform's prefix cache re-reads the same bytes
// (the 5.4x cheaper repeat). The slot never renders the machine's
// own run timestamps (a timestamp per measurement would break the
// prefix on every run); report --full keeps them — there they are
// legal.
package engine

import (
	"strconv"
	"strings"
	"testing"

	"github.com/neurophant/punchtape/internal/execport"
)

// outcome0 — the scripted green outcome of an echo command.
func outcome0(out string) execport.Outcome {
	return execport.Outcome{Exit: 0, Stdout: out + "\n"}
}

// addRowPatch — one add item: the row lands at its table's END.
func addRowPatch(atom string, r pEcoRow) string {
	return "kind: spec\npatch:\n  add:\n    - atom: " + atom + "\n      row:\n        " +
		strings.TrimPrefix(r.body("        "), "        ")
}

// rowLines — the report's row-line block: the contiguous tail of the
// render, one line per row, opened by each row's label. The line
// above the block must be the count header — the extraction guard.
func rowLines(t *testing.T, text string, want int) []string {
	t.Helper()
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(lines) < want+1 {
		t.Fatalf("the report is shorter than its %d rows:\n%s", want, text)
	}
	block := lines[len(lines)-want:]
	if head := lines[len(lines)-want-1]; !strings.HasPrefix(head, "rows (") {
		t.Fatalf("the line above the row block must be the count header, got %q:\n%s", head, text)
	}
	return block
}

// labelOf — one row line's label ("T1/10 [never] ..." — "T1/10").
func labelOf(line string) string { return strings.SplitN(line, " ", 2)[0] }

// commonPrefixLen — the length of the longest common byte prefix.
func commonPrefixLen(a, b string) int {
	n := 0
	for n < len(a) && n < len(b) && a[n] == b[n] {
		n++
	}
	return n
}

// The prefix-extension law: two consecutive spec versions
// — v3 replaces ONE row and adds one row — and the v3 render is a
// byte prefix-extension of the v2 render's row body: everything
// before the replaced row's line is byte-equal, the replacement
// edits its own line in place (same label, same position), the rows
// after it re-render byte-equal, and the added row rides LAST. The
// only line above the block that may move is the count header (one
// short line — the table body is the cache surface).
func TestRenderPrefixExtension(t *testing.T) {
	stack, eng := newInst(t)
	model := &pEcoModel{atoms: []*pEcoAtom{{
		id: "T1", name: "core", prose: "The core behavior.",
		rows: []pEcoRow{
			{cmd: "v1", out: "v1"}, {cmd: "v2", out: "v2"}, {cmd: "v3", out: "v3"},
			{cmd: "v4", out: "v4"}, {cmd: "v5", out: "v5"},
		},
	}}}
	for _, a := range model.atoms {
		for _, r := range a.rows {
			stack.proc.plan["echo "+r.cmd] = outcome0(r.out)
		}
	}
	first := strings.Replace(model.fullDoc(), "kind: spec\nfull-replace: true\n", "kind: spec\n", 1)
	if reply, code := submit(t, stack, eng, first); code != Accepted {
		t.Fatalf("first spec: %d %s", code, reply)
	}

	// v2: replace row 4 in place + add row 6 at the end — one patch
	fix4 := pEcoRow{cmd: "fix4", out: "fix4"}
	stack.proc.plan["echo "+fix4.cmd] = outcome0(fix4.out)
	r6 := model.newRow()
	stack.proc.plan["echo "+r6.cmd] = outcome0(r6.out)
	model.atom("T1").rows[3] = fix4
	model.atom("T1").rows = append(model.atom("T1").rows, r6)
	patchV2 := "kind: spec\npatch:\n  replace:\n    T1/4:\n      " +
		strings.TrimPrefix(fix4.body("      "), "      ") +
		"  add:\n    - atom: T1\n      row:\n        " +
		strings.TrimPrefix(r6.body("        "), "        ")
	if reply, code := submit(t, stack, eng, patchV2); code != Accepted {
		t.Fatalf("patch v2: %d %s", code, reply)
	}
	text2, code := eng.Report("", true)
	if code != Accepted {
		t.Fatalf("report v2: %d %s", code, text2)
	}

	// v3: replace row 2 in place + add row 7 at the end
	fix2 := pEcoRow{cmd: "fix2", out: "fix2"}
	stack.proc.plan["echo "+fix2.cmd] = outcome0(fix2.out)
	r7 := model.newRow()
	stack.proc.plan["echo "+r7.cmd] = outcome0(r7.out)
	model.atom("T1").rows[1] = fix2
	model.atom("T1").rows = append(model.atom("T1").rows, r7)
	patchV3 := "kind: spec\npatch:\n  replace:\n    T1/2:\n      " +
		strings.TrimPrefix(fix2.body("      "), "      ") +
		"  add:\n    - atom: T1\n      row:\n        " +
		strings.TrimPrefix(r7.body("        "), "        ")
	if reply, code := submit(t, stack, eng, patchV3); code != Accepted {
		t.Fatalf("patch v3: %d %s", code, reply)
	}
	text3, code := eng.Report("", true)
	if code != Accepted {
		t.Fatalf("report v3: %d %s", code, text3)
	}

	body2, body3 := rowLines(t, text2, 6), rowLines(t, text3, 7)
	if len(body3) != len(body2)+1 {
		t.Fatalf("v3 must carry exactly one more row line: %d vs %d", len(body3), len(body2))
	}
	// exactly ONE line differs among v2's own positions — the replaced
	// row's own line (same label, same position: the edit is in place)
	diffs, at := 0, -1
	for i := range body2 {
		if body2[i] != body3[i] {
			diffs, at = diffs+1, i
		}
	}
	if diffs != 1 || at != 1 {
		t.Fatalf("the replacement must edit exactly its own line (row 2, index 1): diffs=%d at=%d\nv2:\n%s\nv3:\n%s",
			diffs, at, strings.Join(body2, "\n"), strings.Join(body3, "\n"))
	}
	if labelOf(body2[1]) != "T1/2" || labelOf(body3[1]) != "T1/2" {
		t.Fatalf("the edited line must keep its address:\n%s\n%s", body2[1], body3[1])
	}
	// the added row rides LAST; the old tail line is unchanged one
	// position up
	if labelOf(body3[len(body3)-1]) != "T1/7" {
		t.Fatalf("the added row must be the table's last line:\n%s", body3[len(body3)-1])
	}
	if body2[len(body2)-1] != body3[len(body2)-1] {
		t.Fatalf("the rows after the edit must re-render byte-equal:\n%s\n%s",
			body2[len(body2)-1], body3[len(body2)-1])
	}
	// the byte law: the common byte prefix covers everything before
	// the edited line, and the whole suffix after it matches too
	joined2, joined3 := strings.Join(body2, "\n"), strings.Join(body3, "\n")
	if got := commonPrefixLen(joined2, joined3); got < len(strings.Join(body2[:1], "\n")) {
		t.Fatalf("the byte prefix broke before the edited line: %d", got)
	}
	if strings.Join(body2[2:], "\n") != strings.Join(body3[2:6], "\n") {
		t.Fatal("the rows after the edited line diverged")
	}
}

// The slot is the surface the hand re-reads every move: the machine's
// own run timestamps (ran-at) never render there — a timestamp per
// measurement would break the byte prefix on every run.
// report --full keeps them: there they are legal.
func TestSlotNeverRendersRanAt(t *testing.T) {
	stack, eng := newInst(t)
	doc := "kind: spec\nbehaviors:\n  - id: B1\n    name: sample\n    prose: |\n      One paragraph.\n    rows:\n      - run: \"echo one\"\n        out: \"one\"\n        rc: 0\n        form: \"test corpus: the pinned surface\"\n      - run: \"echo two\"\n        out: \"two\"\n        rc: 0\n        form: \"test corpus: the pinned surface\"\n"
	stack.proc.plan["echo one"] = outcome0("one")
	stack.proc.plan["echo two"] = outcome0("two")
	reply, code := submit(t, stack, eng, doc)
	mustAccept(t, reply, code)
	reply, code = submit(t, stack, eng, "kind: run\n")
	mustAccept(t, reply, code, "suite=green")
	atoms, err := eng.Store.Atoms()
	if err != nil {
		t.Fatal(err)
	}
	ranAt := ""
	for _, a := range atoms {
		for _, r := range a.Rows {
			if r.RanAt != "" {
				ranAt = r.RanAt
			}
		}
	}
	if ranAt == "" {
		t.Fatal("the green rows carry no ran-at to pin against")
	}
	slot, err := eng.Next()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(slot, ranAt) {
		t.Fatalf("the slot renders the run timestamp %q:\n%s", ranAt, slot)
	}
	full, code := eng.Report("", true)
	if code != Accepted || !strings.Contains(full, ranAt) {
		t.Fatalf("report --full must keep ran-at (legal there): %d", code)
	}
}

// The table order is insertion-stable, never lexicographic: with
// eleven rows the render reads T1/1..T1/11 numerically — B1/9 rides
// BEFORE B1/10 (lexicographic order would put B1/10 right after
// B1/1), and rows added later ride at the end.
func TestTableOrderInsertionStable(t *testing.T) {
	stack, eng := newInst(t)
	var rows []pEcoRow
	for i := 1; i <= 9; i++ {
		r := pEcoRow{cmd: "s" + strconv.Itoa(i), out: "s" + strconv.Itoa(i)}
		rows = append(rows, r)
		stack.proc.plan["echo "+r.cmd] = outcome0(r.out)
	}
	model := &pEcoModel{atoms: []*pEcoAtom{{
		id: "T1", name: "core", prose: "The core behavior.", rows: rows,
	}}}
	first := strings.Replace(model.fullDoc(), "kind: spec\nfull-replace: true\n", "kind: spec\n", 1)
	if reply, code := submit(t, stack, eng, first); code != Accepted {
		t.Fatalf("first spec: %d %s", code, reply)
	}
	// two later adds: T1/10, T1/11 — the END of the table
	for i := 10; i <= 11; i++ {
		r := model.newRow()
		stack.proc.plan["echo "+r.cmd] = outcome0(r.out)
		model.atom("T1").rows = append(model.atom("T1").rows, r)
		if reply, code := submit(t, stack, eng, addRowPatch("T1", r)); code != Accepted {
			t.Fatalf("add %d: %d %s", i, code, reply)
		}
	}
	text, code := eng.Report("", true)
	if code != Accepted {
		t.Fatalf("report: %d %s", code, text)
	}
	body := rowLines(t, text, 11)
	for i, line := range body {
		if labelOf(line) != "T1/"+strconv.Itoa(i+1) {
			t.Fatalf("row %d rides out of table order (want T1/%d):\n%s",
				i+1, i+1, strings.Join(body, "\n"))
		}
	}
}

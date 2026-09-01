// Contextual slot diffs: re-reading a slot with an unchanged
// instance renders changed blocks in full and replaces unchanged ones
// with "block unchanged" anchors carrying a digest; cards fold the
// same way, item by item. ACTION is the one block never anchored.
// The full printout goes to the first read of an instance or an
// explicit request (next --full); every later render serves the
// contextual diff. The "block → digest of the last show" map lives
// in the instance cache; the diff digest is written
// to the ledger — verifiable.
package engine

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
	"github.com/neurophant/punchtape/internal/yamlio"
)

// slotBlockLabel — a top-level slot line: a block header.
var slotBlockLabel = regexp.MustCompile(`^[A-Z][A-Z0-9 ()/+-]*:`)

// splitSlotBlocks splits the slot text into named blocks by
// top-level headers; the definition is deterministic.
func splitSlotBlocks(text string) []struct {
	Name  string
	Lines []string
} {
	var blocks []struct {
		Name  string
		Lines []string
	}
	var cur *struct {
		Name  string
		Lines []string
	}
	for _, line := range strings.Split(text, "\n") {
		if slotBlockLabel.MatchString(line) {
			name := strings.SplitN(line, ":", 2)[0]
			blocks = append(blocks, struct {
				Name  string
				Lines []string
			}{Name: name})
			cur = &blocks[len(blocks)-1]
		}
		if cur == nil {
			blocks = append(blocks, struct {
				Name  string
				Lines []string
			}{Name: "head"})
			cur = &blocks[len(blocks)-1]
		}
		cur.Lines = append(cur.Lines, line)
	}
	return blocks
}

// compactSlot replaces unchanged blocks with anchors, and inside
// changed list-blocks folds unchanged ITEMS (context cards, red
// list lines): a growing list changes the block on every submission —
// a block anchor does not help, an item anchor saves bytes
// proportional to the unchanged tail. Returns the text, the number
// of changed and anchored blocks, and the diff digest.
func compactSlot(full string, prev map[string]string) (string, int, int, string) {
	var sb strings.Builder
	changed, anchored := 0, 0
	var diff []string
	for _, b := range splitSlotBlocks(full) {
		digest := yamlio.Digest([]byte(strings.Join(b.Lines, "\n")))
		// ACTION is never anchored: the next step is the one line the
		// hand must see on every read, an anchor would hide it behind
		// a digest.
		if b.Name == "ACTION" {
			compactBlockItems(b.Name, b.Lines, prev, &sb)
			changed++
			diff = append(diff, b.Name+"=new:"+digest[:12])
			continue
		}
		if d, ok := prev[b.Name]; ok && d == digest {
			sb.WriteString(canondata.T("compact.anchor", canondata.M{
				"name": b.Name, "digest": digest[:12],
			}))
			sb.WriteString("\n")
			anchored++
			diff = append(diff, b.Name+"=same:"+digest[:12])
			continue
		}
		compactBlockItems(b.Name, b.Lines, prev, &sb)
		changed++
		diff = append(diff, b.Name+"=new:"+digest[:12])
	}
	return sb.String(), changed, anchored, yamlio.Digest([]byte(strings.Join(diff, "\n")))
}

// itemIDPatterns — how to get the stable identifier of a list item:
// a context card is named by its "id: SCN-n" line, a red list line
// by its "TST-n".
var (
	contextItemID = regexp.MustCompile(`(?m)^id: (SCN-[0-9]+)\s*$`)
	listItemID    = regexp.MustCompile(`^\s+(TST-[0-9]+)`)
)

// compactBlockItems prints the block, folding consecutive unchanged
// items into an anchor with a count. Returns the number folded.
func compactBlockItems(name string, lines []string, prev map[string]string, sb *strings.Builder) int {
	items, ok := splitBlockItems(name, lines)
	if !ok {
		for _, l := range lines {
			sb.WriteString(l)
			sb.WriteString("\n")
		}
		return 0
	}
	same := 0
	run := 0 // the current run of unchanged items
	flushRun := func() {
		if run > 0 {
			sb.WriteString(canondata.T("compact.item-anchor", canondata.M{
				"count": strconv.Itoa(run), "name": name,
			}))
			sb.WriteString("\n")
			same += run
			run = 0
		}
	}
	for _, it := range items {
		if it.id == "" {
			flushRun()
			sb.WriteString(it.text)
			continue
		}
		key := name + "#" + it.id
		d := yamlio.Digest([]byte(it.text))
		if prev[key] == d {
			run++
			continue
		}
		flushRun()
		sb.WriteString(it.text)
	}
	flushRun()
	return same
}

// blockItem — a list-block item with a stable identifier.
type blockItem struct {
	id   string
	text string
}

// splitBlockItems splits a list block into items with identifiers.
// ok=false — the block is not a list (printed in full). The block
// header and the closing service lines ("...and N more") are printed
// as is.
func splitBlockItems(name string, lines []string) ([]blockItem, bool) {
	switch {
	case name == "CONTEXT":
		var items []blockItem
		var cur []string
		flush := func() {
			if len(cur) == 0 {
				return
			}
			text := strings.Join(cur, "\n") + "\n"
			if m := contextItemID.FindStringSubmatch(text); m != nil {
				items = append(items, blockItem{id: m[1], text: text})
			} else {
				items = append(items, blockItem{id: "", text: text})
			}
		}
		for _, l := range lines {
			if l == "---" {
				flush()
				cur = []string{"---"}
				continue
			}
			cur = append(cur, l)
		}
		flush()
		// Compaction makes sense when more than one card is
		// identified; nameless fragments (the block header, prose)
		// are printed as is — they do not become anchors.
		named := 0
		for _, it := range items {
			if it.id != "" {
				named++
			}
		}
		if named > 1 {
			return items, true
		}
		return nil, false
	case strings.HasPrefix(name, "RED LIST"), strings.HasPrefix(name, "GATE REDS"), strings.HasPrefix(name, "PROBE REDS"):
		var items []blockItem
		for _, l := range lines {
			if m := listItemID.FindStringSubmatch(l); m != nil {
				items = append(items, blockItem{id: m[1], text: l + "\n"})
			}
		}
		if len(items) > 1 {
			return items, true
		}
		return nil, false
	}
	return nil, false
}

// slotBlocksPath — the shown-blocks map in the instance cache.
func slotBlocksPath(e *Engine) string {
	return filepath.Join(e.Store.Root(), "cache", "slot-blocks.yamll")
}

// loadShownBlocks reads the shown-blocks map.
func loadShownBlocks(e *Engine) map[string]string {
	out := map[string]string{}
	if data, err := os.ReadFile(slotBlocksPath(e)); err == nil {
		_ = yamlio.DecodeStrict(data, &out)
	}
	return out
}

// storeShownBlocks records the map of blocks AND ITEMS of list
// blocks shown by the full render: item anchors are compared
// against it.
func storeShownBlocks(e *Engine, full string) {
	blocks := splitSlotBlocks(full)
	m := map[string]string{}
	for _, b := range blocks {
		m[b.Name] = yamlio.Digest([]byte(strings.Join(b.Lines, "\n")))
		if items, ok := splitBlockItems(b.Name, b.Lines); ok {
			for _, it := range items {
				m[b.Name+"#"+it.id] = yamlio.Digest([]byte(it.text))
			}
		}
	}
	if data, err := yamlio.Marshal(m); err == nil {
		_ = yamlio.WriteAtomic(slotBlocksPath(e), data)
	}
}

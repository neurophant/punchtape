package doc

// The merge laws of a submission onto its standing corpus. Every
// registry part merges by its own key — the requirement by id, the
// marker by its lines' fold, the exclusion by its range, the order by
// its demand and question: a listed part replaces the standing part
// of the same key, unlisted parts stand untouched. There is no
// replace-whole list and no nil-versus-empty: an absent list and an
// empty one both carry nothing. This is the economy of submissions —
// the number of submissions does not affect the price: the listed
// bytes travel once, unchanged parts are not re-sent.

import (
	"fmt"
	"sort"
	"strings"
)

// MergeBehavior — the submitted behavior onto its standing self: the
// submitted head (name, prose, claims, covers) replaces; the rows
// merge by run.
func MergeBehavior(standing, submitted Behavior) Behavior {
	merged := submitted
	merged.Rows = MergeRows(standing.Rows, submitted.Rows)
	return merged
}

// MarkerKey — the marker's merge key: the fold of its lines — the
// parsed line numbers sorted and spelled back in the range grammar.
// Two spellings of one line set are one marker.
func MarkerKey(m Marker) string {
	nums, err := ParseRanges(m.Lines)
	if err != nil {
		// Validation has spoken before any merge runs; the fallback
		// keeps the key total for unvalidated data.
		sorted := append([]string(nil), m.Lines...)
		sort.Strings(sorted)
		return strings.Join(sorted, ",")
	}
	return SpellRanges(nums)
}

// RemoveKeyKind — the five spellings of the remove list, told apart
// mechanically: the D grammar is unique by its prefix, the line
// range by its colon.
type RemoveKeyKind int

const (
	// RemoveRequirement — the R… spelling of a standing requirement.
	RemoveRequirement RemoveKeyKind = iota
	// RemoveBehavior — the [a-z0-9-] segment spelling of a standing
	// behavior.
	RemoveBehavior
	// RemoveMarkerOrOrder — the D<n> spelling of a standing marker on
	// the line or the order riding it.
	RemoveMarkerOrOrder
	// RemoveExclusion — the path:first-last spelling of a standing
	// exclusion.
	RemoveExclusion
	// RemoveUnknown — a spelling none of the four families owns.
	RemoveUnknown
)

// ParseRemoveKey — one remove item's kind. Resolution of the
// marker-or-order spelling is contextual: first the marker standing
// on the line, then the open order riding it.
func ParseRemoveKey(s string) RemoveKeyKind {
	switch {
	case validRequirementID(s):
		return RemoveRequirement
	case validID(s):
		return RemoveBehavior
	case strings.HasPrefix(s, "D"):
		if _, err := ParseRanges([]string{s}); err == nil {
			return RemoveMarkerOrOrder
		}
		return RemoveUnknown
	case strings.Contains(s, ":"):
		if _, _, _, err := ParseLineRange(s); err == nil {
			return RemoveExclusion
		}
		return RemoveUnknown
	default:
		return RemoveUnknown
	}
}

// MergeRegistry — the submitted registry onto the standing one:
// removals retire their standing parts first, then every listed part
// merges by its own key — a repeated key replaces, a new key is
// appended, an unlisted part stands.
func MergeRegistry(standing Registry, d Document) Registry {
	retired := ParseRemovals(d.Remove)
	out := Registry{
		Requirements: make(map[string]Requirement, len(standing.Requirements)+len(d.Requirements)),
	}
	for _, id := range standing.Order {
		if retired.Requirements[id] {
			continue
		}
		out.Order = append(out.Order, id)
		out.Requirements[id] = standing.Requirements[id]
	}
	for _, req := range d.Requirements {
		if _, seen := out.Requirements[req.ID]; !seen {
			out.Order = append(out.Order, req.ID)
		}
		out.Requirements[req.ID] = req
	}
	// A marker retired by a remove line leaves whole — the D<n> key
	// names a line, the marker standing on it retires entire.
	out.Markers = mergeMarkers(dropMarkersOn(standing.Markers, retired.Lines), d.Markers)
	out.Exclusions = mergeExclusions(dropExclusions(standing.Exclusions, retired.Ranges), d.Exclusions)
	out.Open = mergeOrders(dropOrdersOn(standing.Open, retired.Lines), d.Open)
	return out
}

// Removals — the remove list resolved by kind.
type Removals struct {
	Requirements map[string]bool // requirement ids
	Lines        map[int]bool    // demand lines — the marker-or-order key
	Ranges       map[string]bool // exclusion ranges
}

// ParseRemovals — the remove list split by its five spellings. The
// D<n> spelling serves both the marker and the order: the marker
// standing on the line retires first, the order riding it when no
// marker stands.
func ParseRemovals(remove []string) Removals {
	out := Removals{
		Requirements: map[string]bool{},
		Lines:        map[int]bool{},
		Ranges:       map[string]bool{},
	}
	for _, key := range remove {
		switch ParseRemoveKey(key) {
		case RemoveRequirement:
			out.Requirements[key] = true
		case RemoveMarkerOrOrder:
			nums, _ := ParseRanges([]string{key})
			for _, n := range nums {
				out.Lines[n] = true
			}
		case RemoveExclusion:
			out.Ranges[key] = true
		}
	}
	return out
}

// dropMarkersOn — the standing markers minus those a retired line
// claims.
func dropMarkersOn(markers []Marker, lines map[int]bool) []Marker {
	out := markers[:0:0]
	for _, m := range markers {
		nums, err := ParseRanges(m.Lines)
		retired := err == nil && anyLine(nums, lines)
		if !retired {
			out = append(out, m)
		}
	}
	return out
}

// dropOrdersOn — the standing orders minus those a retired line
// rides.
func dropOrdersOn(orders []OpenOrder, lines map[int]bool) []OpenOrder {
	out := orders[:0:0]
	for _, order := range orders {
		nums, err := ParseRanges([]string{order.Demand})
		retired := err == nil && len(nums) > 0 && lines[nums[0]]
		if !retired {
			out = append(out, order)
		}
	}
	return out
}

// dropExclusions — the standing exclusions minus the retired ranges.
func dropExclusions(exclusions []Exclusion, ranges map[string]bool) []Exclusion {
	out := exclusions[:0:0]
	for _, ex := range exclusions {
		if !ranges[ex.Range] {
			out = append(out, ex)
		}
	}
	return out
}

// anyLine — whether any of the numbers is a retired line.
func anyLine(nums []int, lines map[int]bool) bool {
	for _, n := range nums {
		if lines[n] {
			return true
		}
	}
	return false
}

// mergeByKey — the registry lists' one merge law: a repeated key
// replaces the standing part whole, a new key is appended, unlisted
// parts stand.
func mergeByKey[T any](standing, submitted []T, key func(T) string) []T {
	out := make([]T, 0, len(standing)+len(submitted))
	out = append(out, standing...)
	held := make([]string, 0, len(out))
	for _, part := range out {
		held = append(held, key(part))
	}
	for _, part := range submitted {
		k := key(part)
		replaced := false
		for i, standing := range held {
			if standing == k {
				out[i] = part
				replaced = true
				break
			}
		}
		if !replaced {
			out = append(out, part)
			held = append(held, k)
		}
	}
	return out
}

// mergeMarkers — by the lines' fold; the one merge law.
func mergeMarkers(standing, submitted []Marker) []Marker {
	return mergeByKey(standing, submitted, MarkerKey)
}

// mergeExclusions — by range; the one merge law.
func mergeExclusions(standing, submitted []Exclusion) []Exclusion {
	return mergeByKey(standing, submitted, func(ex Exclusion) string { return ex.Range })
}

// mergeOrders — by the demand-and-question pair; the one merge law.
// Birth facts ride the standing copy — the caller stamps the
// submitted list first.
func mergeOrders(standing, submitted []OpenOrder) []OpenOrder {
	return mergeByKey(standing, submitted, orderKey)
}

// orderKey — the open order's merge key.
func orderKey(o OpenOrder) string {
	return fmt.Sprintf("%s\x00%s", o.Demand, o.Question)
}

// MergeRows — the standing rows with the submitted rows merged by
// run key: a submitted run that stands replaces that whole row in
// place — a partial patch does not exist; a new run is appended
// after the standing rows; rows not listed stand untouched. The
// standing order never moves: a row's place is its journal key.
func MergeRows(standing, submitted []Row) []Row {
	out := append([]Row(nil), standing...)
	var added []Row
	for _, row := range submitted {
		matched := false
		for i := range out {
			if out[i].Run.Key() == row.Run.Key() {
				out[i] = row
				matched = true
				break
			}
		}
		if !matched {
			added = append(added, row)
		}
	}
	return append(out, added...)
}

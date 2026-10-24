package doc

// The merge laws of a submission onto its standing corpus: a listed
// behavior's head replaces whole and its rows merge by their run
// key; the registry's requirements merge by id like behaviors, while
// the marker list and the artifact list replace whole when the
// submission carries them. This is the economy of submissions — the
// number of submissions does not affect the price: the listed bytes
// travel once, unchanged parts are not re-sent.

// MergeBehavior — the submitted behavior onto its standing self: the
// submitted head (name, prose, scope, covers) replaces; the rows
// merge by run.
func MergeBehavior(standing, submitted Behavior) Behavior {
	merged := submitted
	merged.Rows = MergeRows(standing.Rows, submitted.Rows)
	return merged
}

// MergeRegistry — the submitted registry onto the standing one:
// removals retire requirements first, then the listed requirements
// merge by id (a repeated id replaces its head whole — the brief
// spans); the marker list and the artifact list, when the submission
// carries one, replace the standing list entire.
func MergeRegistry(standing Registry, d Document) Registry {
	removed := make(map[string]bool, len(d.Remove))
	for _, id := range d.Remove {
		removed[id] = true
	}
	out := Registry{
		Requirements: make(map[string]Requirement, len(standing.Requirements)+len(d.Requirements)),
		Markers:      standing.Markers,
		Artifacts:    standing.Artifacts,
	}
	for _, id := range standing.Order {
		if removed[id] {
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
	if len(d.NotRequirements) > 0 {
		out.Markers = d.NotRequirements
	}
	if len(d.Artifacts) > 0 {
		out.Artifacts = d.Artifacts
	}
	return out
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

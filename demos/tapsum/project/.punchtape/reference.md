# punchtape reference — every door of the machine, always under hand.
# index: call stream document intents registry behaviors replace merge execution universe verdict faces env damage constants combinations
call — the zone is the nearest .punchtape at or above the working directory; instructions name the zone's
files from the caller's directory. No zone: the contract prints,
nothing is created, exit 0. A zone with no journal and no non-empty demands line: "no instance yet", exit 0.
The first non-empty line births the zone; parallel runs elect one winner, and every pass stands in
one line behind the zone's own lock — kernel-released, death-safe, local filesystems. Any arguments: refused, exit 1;
--version answers before any state. A zone born by another machine version: archive it, start fresh, exit 1.
stream — the human's words verbatim, one non-empty trimmed line each, numbered D1..Dn in file order. A touched
past line, a shrunken file or a removed demands.md: "zone damaged: the demand stream's past changed", exit 1 —
the chain recomputes over the bytes every run. Every new line is declared by the submission that meets it,
exactly one intent, one outcome. Accepted chains the tail; refused leaves it; no new lines sign a
no-new-demands record. A line that carries its outcome is never redeclared.
document — exactly one YAML document; anchors, aliases, tags, merge keys, duplicate keys, unknown fields and
tabs in indentation: decode refusals naming the line. A scalar holding ": " or a comma takes quotes; a flow
element holding ":" takes quotes; a multi-line pin stays one line. A done document is exactly done: true.
An accepted submission leaves the inbox; a refused one stays in place to fix.
intents — six outcomes. answer carries its text; execute carries its run; the other four carry their registry
in the same document: work and describe a requirement whose lines carry the line, mark a not-a-requirement
marker on it, ask an open entry with state awaiting on its demand. text rides an answer intent only; run
rides an execute intent only.
registry — a requirement id is R plus letters, digits and dots, unique in the submission; its lines parse in
the D grammar; refines names a standing requirement. A marker (not-a-requirement: lines + reason) never
claims a requirement's line — a line is either a demand or not. An exclusion marks a real uncovered range
with a citing reason, refs naming demand lines or laws. An order rides a standing line; its state is
awaiting or deferred; deferred carries its reason.
behaviors — an id is lowercase dash segments joined by "/", unique; name and prose non-empty. Claims are line
ranges the tree carries, path:first-last, root-relative; one range names one behavior. Covers name
requirements standing after the merge. At least one row; one row pinned: out or files with exact bytes,
"contains: " prefixing a substring; rc alone and an empty out pin nothing. One run proves one behavior; a
run is unique in its rows; expect names rc, out or files.
replace — replacing a standing behavior's head, or removing standing work, carries a mandate: the ordering
line and the target, replaced by this same submission. Removing a requirement re-opens its line and demands
the same submission re-carry every cover of it. Remove keys — R… ids, behavior ids, D<n> markers-or-orders,
"path:first-last" exclusions — every key must stand. An awaiting order closes only over a later human line,
given its own outcome, cited by a decision note; deferral is a visible mark forever.
merge — unlisted parts stand; comparison happens on canonical bytes, so reformatting changes nothing. A
behavior's head replaces whole, its rows merge by run: the same run replaces its row, a new run appends, a
row retires only with its behavior whole — remove the id under its mandate and re-add it in one document.
Remove-then-re-add in one submission starts fresh. Markers merge by lines, exclusions by range, orders by
demand and question.
execution — the build runs first on every launch; a red build stops the rows while the submission still
lands. A row's proof is its bytes, its behavior's claims digest and the environment hash; any move — pins,
PATH or HOME, the machine itself, a claimed line — restales every proof to re-earn. A behavior without
claims re-runs its rows every run. The run's own budget is 3600 s unless PUNCHTAPE_RUN_BUDGET names whole seconds — what it cannot
fit waits the next run, named by a warning. Every command's deadline is 300 s unless PUNCHTAPE_ROW_TIMEOUT names
whole seconds. A missing command carries its PATH recipe in the refusal. Execute intents run between the
submission and the rows, stdout journaled whole to 4096 bytes. Whatever a command left alive dies with its
process group.
universe — no declaration: the whole tree, minus .punchtape and version-control dirs. A range holds only
the lines the file really carries. One root-relative path a line; a root carrying no file of the tree is
refused. Files outside stand counted, never judged. An uncovered line holds the verdict; the proposal
drafts exclusions — reasons and refs are the agent's.
Temporary files go to .punchtape/tmp — cleared at each run's end, the writer mkdirs first. A declared
build's artifacts are work; anything left in the tree wants a claim or a loud exclusion.
verdict — done: true records over a non-empty corpus, every row green and pinned, the stream resolved whole,
every requirement covered, the tree equal to claims plus loud exclusions, no order awaiting, no drift
unnamed, no tree-lock standing. Blockers speak in one order: build red, tree moved, drift, unpinned, red or
stale rows, unresolved lines, uncovered requirements, holes, awaiting. The tree does not move under answer,
execute or describe — the bytes back or a work submission lifts the lock. Drift is named by a submission
landing after the witness run.
faces — state.md: the standing and the one next line. report.md: the last run's facts, refusals named field
by field. trace.md: the stream's showcase. knowledge.md: notes and machine-measured lessons. spec/: one page
per behavior, retiring with it. holes.md: the whole ledger. exclusions-proposal.yaml: the standing draft.
corpus.yaml: the machine's image. journal/: never edit.
env — one KEY=VALUE a line, the key [A-Za-z_][A-Za-z0-9_]*; a name pins once; no quotes, no absolute paths —
a pin travels with the tree. PUNCHTAPE_ROW_TIMEOUT names whole seconds: the machine's ceiling, never the
commands' own. Pins ride the commands' environment and the proofs' hash.
damage — zone damaged: the stream's past changed. A foreign version: archive and start fresh. A torn journal
tail quarantines while the intact records load. Damaged AGENTS.md markers: the machine refuses to guess. A
view write that cannot land fails loudly, never silently.
constants — deadline 300 s; the run budget 3600 s unless PUNCHTAPE_RUN_BUDGET names whole seconds — what
it cannot fit waits the next run; the document's bounds — 524288 bytes of inbox, 64 entries per registry
list, 32 claims and 16 rows per behavior, 8 expected files per row;
execution stdout 4096 bytes; mismatches cut at 80, lessons at 100; elisions at 10 (history, hole files,
verdict files); tokens 12 hex; epochs epoch-NNNNN rotated per accepted submission; a row's key is its
behavior and its place; commands see PATH, HOME and the pins, nothing else.
combinations — a red build and an accepted submission: the corpus moves, the rows wait. A tree-lock and a
drift stand independent. A work intent lifts a standing lock and arms none. An execute intent runs under
the lock. A standing verdict lives over holes; a new one demands none. A refusal is a journaled fact before
any face; a quarantined tail precedes everything.

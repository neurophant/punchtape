The model of the machine - what it is and what the verdict means:

## The verifier

punchtape is a task-agnostic verifier. An agent writes code and one spec
document; the machine stores the spec, runs its check rows in the agent's
work tree, and renders a verdict. The agnosticism law: the machine knows
nothing about tasks, stacks, domains or roles - nothing in its code,
canon or texts may encode knowledge of what is being built. It observes
bytes and exit codes, nothing else; it never reads the agent's code
files - it verifies their behavior.

## Journal ahead of files

The journal is the truth; files are renders. Every accepted submission
is recorded first (frames, digests, statuses); every artifact the agent
sees - AGENTS.md, spec.md, SUMMARY.md, the reference/ - is re-rendered
from the journal. A manual edit to a rendered file heals on the next
open; nothing the machine answers from exists outside the journal.

## Determinism

The same spec against the same tree yields the same verdict, anywhere.
Check rows are executed, not interpreted; expectations are mechanical
comparisons (exact, contains, byte-equal); nothing is sampled, scored
or judged by taste. Write deterministic rows and the machine repays
you with reproducible proof.

## The verdict is computed, not judged - and versioned

READY means: every row of every behavior green, prose present
everywhere, the tree unchanged outside the ignores. The verdict line
carries the tree digest - the signature of the exact tree the results
were measured on. It is portable proof of readiness: not an opinion,
a measurement.

A ratification is a signature, not a run: it executes nothing - no
row, no build; it signs exactly when the recorded green still stands
(every row fresh at unchanged inputs, the surface digest equal to the
last green run's), and anything stale refuses the door with the names.
Every ratification freezes an IMMUTABLE VERSIONED SNAPSHOT
(.punchtape/verdicts/vN.txt): v1 is pinned to its digest forever, v2
to its own - the model of git commits. The case lives on: the next
spec version through the same door reopens the work, and the journal
carries every verdict of every accepted state. A verdict explains
every row retired since the previous version - greening by deletion
without a trace is impossible. Row results ride versions by content:
an unchanged row in a new spec version keeps its green at unchanged
input digests (the machine's memo - a run re-executes only what went
stale).

## The report is a render, not a run

`punchtape report` prints the observed behavior of the case from the
stored runs only: every row's exact command, its demand, the measured
answer (the exit code and output digest of the last run), the run's
time and its freshness mark (fresh, stale - a named input moved,
red, never). Nothing re-runs; nothing is re-derived by hand.
`report --since vN` aims the same table at the delta after a frozen
verdict: the changed, new, re-measured and retired rows by name - the
two stored versions are the diff's ground. The report lifts the
cost of aiming a review; the duty of eyes-on verification stays with
the reader.

## Typed rows and the dry validator are deterministic functions

The mechanical row families - the refusal, the determinism pin, the
staged-bytes probe, the repeat - are HEADS of the row grammar, not
macros over it: refuse{args, says, rc} IS the row it spells (one row,
one canon, no expansion step, no bundle shorthand), witness and
repeat-run are the machine's own disciplines (the door-staged probe
bytes, the twice-run agreement; the procedural launch/tree modes ride
every head). The
judged part of a spec stays handwriting: the prose, the pinned
quotes, the witnesses' words. `spec --check <file>` is the same law
as the door's form gate run over a draft as a deterministic function - every
finding on its own line, nothing opened, nothing written; a document
that passes is taken by the door on the first try.

## The trade

The spec document is the truth of the product; the code is free
territory. You keep the spec honest; the machine keeps your claims
provable. When code and spec disagree, the spec wins: change the code,
or change the spec with a patch - never bend one to hide the other.

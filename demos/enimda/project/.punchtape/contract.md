punchtape — the contract

You ran punchtape and it printed this page: no instance here yet,
nothing was created. TO START: put the human's words verbatim, one
per line, into .punchtape/demands.md and run punchtape again — that
run births the zone.

THE CALL, always: `punchtape`, no arguments. The terminal gets one
line — the next action with the file to touch. Every other answer is
a file of the zone: reference.md (every door — read it first),
state.md (standing + next), report.md (last run's facts).

THE CYCLE: the human's words enter demands.md at any moment; the
agent appends them verbatim and the machine takes them itself at the
next submission. Submit the intent document → the machine runs the
rows → fix code or rows until green → submit `done: true` → the
machine records the verdict. state.md always says what to do now.

law:demand    The human's words, verbatim, one per line, into
              .punchtape/demands.md. Numbering D1..Dn never changes;
              the past is immutable — the journal holds the chain, a
              touched past line is zone damage. Every new line is
              declared in the submission that meets it: its outcome
              (answer, execute, describe, work, mark or ask) and the
              agent's retelling. A requirement is a demand made
              specific: the lines it occupies and their unwinding in
              prose. The tree does not move under answer, execute or
              describe: the machine digests the declared universe
              before and after, a declared build's artifacts are
              work, and a moved tree holds the verdict.

law:document  All work is one document (shape in template.yaml),
              merged by id: requirements and behaviors by their
              ids, markers by their lines, exclusions by range,
              orders by demand and question. Removal is the remove
              list; every key must stand. A done document is
              exactly `done: true`; an accepted one leaves the
              inbox, a refused one stays in place to fix. One document
              is bounded — its bytes, its lists, its behaviors' claims
              and rows, its expectations' files; the constants name them.

law:paths     Every path is relative to the working root. Absolute,
              ".." and escaping spellings are refused.

law:claims    Coverage is claimed in line ranges of files, never
              globs; one range names one behavior. The universe is
              declared in .punchtape/universe — one root-relative
              path a line, # a comment, no declaration meaning the
              whole tree; a line no claim and no loud exclusion
              names holds the verdict; the machine drafts the
              exclusions, the reason and its references are yours.

law:proof     A behavior proves by a row whose expect pins the exact
              bytes of its output; a behavior without claims re-runs
              its rows every run. One run proves one behavior. A
              moved claimed line or a changed environment — pins,
              shell, machine — makes the proof stale: the machine
              re-runs and the pin must hold again.

law:candor    What is skipped is marked, loudly. Replaced standing
              work carries its mandate: the ordering line and the
              target. An order awaiting the human holds the verdict;
              it closes only over a later human line, cited by a
              decision note; deferral is a visible mark. Notes carry
              kinds and references that resolve; exclusions carry a
              reason that cites and references that name demand
              lines or laws.

law:done      `done: true` records the verdict when: the declared build stands green; every row green
              with a pinned proof; the stream resolved whole; every requirement covered by a behavior's
              claims; the tree equal to claims plus loud exclusions; no order awaiting; and no drift left
              unnamed — a drift is named by a submission that lands after the run that first saw it.

law:your-tree Rows and the build run from the tree root with PATH
              and HOME plus the zone pins (.punchtape/env — one
              KEY=VALUE a line, no quotes, no absolute paths;
              PUNCHTAPE_ROW_TIMEOUT names whole seconds: the
              deadline of every machine-run command, never the
              commands' own). The build runs first, on every
              launch — only rows stand on their proofs. Rows run
              one at a time, in corpus order. Temporary files go
              to .punchtape/tmp — cleared at each run's end, so
              a writer creates the directory itself first;
              version-control bookkeeping stays out of the tree;
              anything left in the tree wants a claim or a loud
              exclusion.

ZONE FILES. contract.md template.yaml reference.md | demands.md inbox.yaml
universe — you write | corpus.yaml state.md report.md knowledge.md
spec/ holes.md trace.md exclusions-proposal.yaml — the machine's |
journal/ — never edit. The working root's AGENTS.md carries the
machine's marked block.

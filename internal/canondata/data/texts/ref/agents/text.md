## punchtape - what this is

The loop: `punchtape next` -> the ONE slot names the door (the INBOX
line) -> write your document into that file -> submit it -> the ack
is thin (transaction, bytes, gates, the measured one-liners); the
next slot rides your next `punchtape next` -> repeat until a VERDICT
line. You write code and spec; the machine stores, runs the rows,
renders the verdict.

## The verbs and the door

- `punchtape next ["<brief>"]` - birth the instance; the reply = state + slot.
- `punchtape submit <path>` - submit the door file (the one legal path,
  `.punchtape/inbox/current.txt`; prefilled, reset to the blank
  template after acceptance - OVERWRITE it every time: a door file
  still carrying the template, or empty, is refused unread).
- Flags: `--quiet` (no slot body; the verdict still rides at the closing)
  and `--delta` (only what changed since the last SLOT: the anchor is
  the last `next` reply, any mode - a submit's ack is not a slot and
  never moves the anchor); submit also takes `--run-after` (spec lands
  + its suite runs in the same call).
- `punchtape spec --check <file>` - the dry validator: the door's form
  rules over a draft, one finding per line; a pass = a first-try
  acceptance (write those bytes into the door).

## The document

Two kinds, one grammar: `spec` = the STATE of the case (validated,
nothing runs; the FIRST spec carries the whole table - an absent id is
retired, a visible frame records it); `run` = the same grammar, and
the machine executes the STALE rows - a row keeps its green while its
definition and the digests of the files its command names are
unchanged. Repairs ride by behavior id; a run never retires.
EVERY LATER submission is a PATCH - the change only, by row address
(BEHAVIOR-ID/N): `replace` edits a row in place, `add` appends to the
end, `remove` retires, `run` names rows to execute now, `prose`
carries the verdict texts once; the full table never rides twice (a
deliberate full rewrite spells `full-replace: true` and the ack counts
the rows). Never reorder accepted rows and never resubmit the full
table - the cache reads the same prefix.
`ratify: true` (kind: run only) SIGNS the recorded green - stale rows
or a moved tree refuse with the names; the verdict is an immutable
snapshot (`.punchtape/verdicts/vN.txt`) and the case LIVES ON - the
next spec version reopens it. Typed row heads for the mechanical
families: refuse{args,says,rc}, det{run,same}, witness{probe},
repeat-run: <command>; the procedural modes launch: nonet and
tree: unchanged compose with every head - see grammar.md.

## The cycle

1. `punchtape next "<brief>"` births the instance and prints this contract.
2. One spec document into the door; the form is validated, nothing runs.
3. Code files are yours to write freely - your territory, never read.
4. kind: run: build (if declared) + the stale rows; the table carries
   measured reds and named kept greens.
5. Repairs: a patch by row address (or a cheap partial upsert by
   behavior id).
6. A green run, then `ratify: true` (the free signature).

## Timeouts and rules

Row 60 s default (own `timeout` to 300), build 300 s. The tree may not
move during a run outside .punchtape/ and build-ignore ("tree moved").
Code files are yours - never read; the same document again is an
honest replay. Library: `.punchtape/reference/` - environment.md,
grammar.md, model.md.

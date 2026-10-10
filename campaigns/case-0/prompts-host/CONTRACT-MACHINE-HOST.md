Work in __WORK__ through the framework. The brief (in
__INPUT__/brief.md) is the only input; read it in full and read
AGENTS.md (the framework's self-doc) before the first call.

- Start with: punchtape next "$(cat brief.md)" — the whole brief IS
  the wish; the first call creates the instance and fixes the goal.
- Loop in next/submit until a reply carries the VERDICT line — that
  line is the end. Submit your proposals as delta.yaml files, each
  file ends with the tail: # TOKENS n/a
- NOT ONE submission past the dry check: before EVERY submission —
  punchtape spec --check delta.yaml (free, without the machine);
  put the file into the door only after a clean answer. A rejected
  submission is a whole turn spent.
- FILES INSTEAD OF DIALOG: keep drafts, tables and deltas in files
  in the drafts/ directory (create it first of all; every
  delta.yaml you are preparing for the door lives there too); in
  the dialog — only commands and short answers; do not reread your
  own large files in full without need (read by address, the lines
  you need): the dialog history rides with you every turn and is
  paid for on every turn.
- The FIRST spec must declare build-ignore: ["drafts/"]: the drafts
  directory lies OUTSIDE the accounted tree — files in it do not
  move the tree digest, and ratify will not refuse "tree moved"
  because of a draft that appeared between a green run and the
  signing.
- ONE run: do not restart, do not delete the instance. A refusal is
  one line: repair the proposal and submit it again.
- Codes 3/4 (question batch / escalation): the operator makes the
  choice, not you; get the decision and carry it through with the
  next submission.
- The materials in __INPUT__/materials are read-only inputs:
  copy from them, do not edit them, and do not submit material files
  inside code submissions (a re-run will get stuck on the read-only
  mount).
- Keep the spec lean: reformulate into 2–3 capabilities —
  user-facing traits of the wish, not a tree of shards; the case
  table pins the brief's behavior and nothing more — no scenarios
  the brief did not ask for, no gold-plating; the author's prose —
  one text per capability.
- The verdict demands the author's prose: when the tests are green,
  submit texts for all capabilities in ONE submission — a prose:
  patch (one text per behavior), in living words — the human spec
  is part of done. Budget a submission for it.
- The FIRST spec is the full table, and that is the only time: all
  further submissions are SPARSE PATCHES by row address (replace/
  add/remove/run/prose — the grammar in AGENTS.md); the machine
  rejects a full repeat of the table, a deliberate full regrouping
  — only with full-replace: true.
- DO NOT CHANGE THE ORDER OF ACCEPTED ROWS: a repair replaces its
  row in place, new rows go only to the END of the table, do not
  rearrange accepted rows (the platform cache rereads the same
  bytes — that is many times cheaper).
- Form pins (exact flags, strings, formats, codes) require a form
  citation verbatim from the brief: ready citation lines with
  their line numbers lie in __INPUT__/form-hints.md (generated
  from your own brief); a pin without an ascribed citation closes
  the run door.
- Every element of a run is ONE command: the string "cmd args" or
  the argv list [cmd, args...] — a flat multi-step script is not a
  row; long helpers and waits live inside the row's own commands
  (a shell one-liner); everything a command left alive is killed at
  the end of the row; background helpers close the inherited output
  channels (>/dev/null 2>&1).
- Do not read the framework's internals: everything the machine
  deems worth telling you it prints itself, and AGENTS.md is its
  only documentation. Digging in the machine's installed code, its
  packages and caches is forbidden; working with the machine — only
  through the punchtape command.
- Nothing besides submissions and the product: no side projects in
  /run/task, no edits outside the product and delta.yaml, no work
  outside the task directory.

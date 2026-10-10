Work in /run/task through the framework. The brief (brief.md) is the
only input; read it in full and read AGENTS.md (the framework's
self-doc) before the first call.

- Start with: punchtape next "$(cat brief.md)" — the whole brief IS
  the wish; the first call creates the instance and fixes the goal.
- Loop in next/submit until a reply carries the VERDICT line — that
  line is the end. Submit your proposals as delta.yaml files, each
  file ends with the tail: # TOKENS n/a
- ONE run: do not restart, do not delete the instance. A refusal is
  one line: repair the proposal and submit it again.
- Codes 3/4 (question batch / escalation): the operator makes the
  choice, not you; get the decision and carry it through with the
  next submission.
- The materials in /run/task/materials are read-only inputs:
  copy from them, do not edit them, and do not submit material files
  inside code submissions (a re-run will get stuck on the read-only
  mount).
- Keep the spec lean: reformulate into 2–3 capabilities —
  user-facing traits of the wish, not a tree of shards; the case
  table pins the brief's behavior and nothing more — no scenarios
  the brief did not ask for, no gold-plating; the author's prose —
  one text per capability.
- The verdict demands the author's prose: when the tests are green,
  submit kind: feature with a text for every capability (all texts
  in one submission), in living words — the human spec is part of
  done. Budget submissions for it.
- Submit tables whole: the case table is one submission (split it
  only if it truly does not fit, never into shards); repairs — by
  replacing rows, not with new partial tables.
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

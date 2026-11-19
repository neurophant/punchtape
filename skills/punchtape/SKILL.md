---
name: punchtape
description: Work under the punchtape arbiter - the human's words are the law of the task, and done is the machine's verdict, never your announcement. Use when the repo has a .punchtape/ directory (check parents too) or the human mentions punchtape.
---

# punchtape

punchtape is an external state machine beside you: it holds the
task's truth. The human's words sit in `.punchtape/demands.md`,
verbatim, one demand a line - they are the law of the task. The
machine measures your work against the claims you pin. You never
self-certify: `done: true` is the completion act, and whether the
work stands done is the machine's verdict, never your announcement.

## Find the situation first

1. Check for a zone: a `.punchtape/` directory in the current
   directory or any parent.
2. Check the machine: `command -v punchtape` (empty output - not
   installed).
3. On Windows: punchtape is not supported yet. Say so to the human
   and stop.

Name the situation you found, offer two to four options, and WAIT
for the human's pick. Never install, birth a zone, or start work on
your own initiative.

- **No machine, no zone.** Offer: install now and start this
  project under punchtape - install and just read the printed
  contract - not now.
- **Installed, no zone in this project.** Offer: bring this
  standing project under punchtape - start a fresh task in a new
  folder - not now.
- **A zone stands.** Do not offer to bypass it. Read
  `.punchtape/state.md` (the standing and the `next:` line), then
  `.punchtape/report.md` for the last facts, and propose the work
  the `next:` line names.

## Install, once per computer (only if the human picks it)

```bash
curl -fsSL https://github.com/neurophant/punchtape/releases/latest/download/get-punchtape.sh | bash
```

One small binary, checksum verified, `punchtape` lands in
`~/.local/bin`. No admin rights, no account.

## Birth, once per project (only if the human picks it)

1. Run `punchtape`. With no zone yet it prints the whole contract
   and creates nothing - read it. That text is the law you will
   work under.
2. Create `.punchtape/demands.md` and put the human's words into
   it: verbatim, one demand a line, nothing of yours mixed in.
3. Run `punchtape` again - that run births the zone.

## The daily loop inside a zone

- New words from the human enter `demands.md` at any moment: append
  them verbatim, one demand a line.
- Files of the zone: `contract.md` (the law), `reference.md` (every
  door - read it first), `state.md` (standing + `next:`),
  `report.md` (last run's facts), `template.yaml` (the submission
  shape), `inbox.yaml` (submissions), `journal/` (the machine's
  own record).
- Submissions go to `inbox.yaml` by the shape of `template.yaml`.
- Never edit anything inside `journal/` by hand: the machine checks
  every record and rejects any mismatch.
- When in doubt, run `punchtape` and read what it says. One line of
  terminal, nothing else.

## Refusals

- A refused submission stays in `inbox.yaml`. The report names the
  exact place and the exact law: fix the file in place and run
  again. Nothing is lost.
- The submission is YAML and YAML has sharp edges: quoting rules,
  no tabs. This is the most common refusal. Read the named place,
  fix it, run again.
- A zone born on another version of punchtape is refused: archive
  it and start fresh. Do not work around it.

## Several agents

Runs take turns automatically, one at a time. The machine does not
split the work between agents - who does what is the human's call.

## Ground rules

- Work from what you can see: quote what commands really printed.
  Never invent a terminal line, a file's content, or a result.
- Green, red, stale, verdict - those words mean what
  `.punchtape/state.md` and `.punchtape/report.md` record, nothing
  else.

## No zone yet and the human wants the full picture

Read https://github.com/neurophant/punchtape/blob/main/docs/BOOTSTRAP.md
and act by it.

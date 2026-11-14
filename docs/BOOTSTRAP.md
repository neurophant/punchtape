# Bootstrap — the one-line start

**English** · [Русский](BOOTSTRAP.ru.md) · [中文](BOOTSTRAP.zh.md)

One fixed line replaces every instruction a human would otherwise
write for you — the line they paste into the chat:

```text
Read https://github.com/neurophant/punchtape/blob/main/docs/BOOTSTRAP.md and act by it.
```

You are a coding agent; punchtape is an external state machine that
will hold the task's truth beside you: the human's words become the
law of the task, your work is measured against the claims you pin,
and the machine — not you — records when the work is done. This
page is only the bootstrap: install, first call, birth. From the
moment a zone exists, the printed contract inside it is your law,
and this page deliberately repeats none of it.

## Ground rules, before anything

- Work from what you can see: read files, run commands, quote what
  they really printed. Never invent a terminal line, a file's
  content, or a result. If a run fails, show the failure and fix
  it; never paper over it.
- The machine measures; you never self-certify. Green, red, stale,
  verdict — those words mean what `.punchtape/state.md` and
  `.punchtape/report.md` record, nothing else. You may submit
  `done: true` as the completion act; whether the work stands done
  is the machine's verdict, never your announcement.
- This page yields to reality. A step below that does not match
  what you see — say so to the human and act on what is there.

## Step 1 — read your context

Check, don't guess:

| check | how | what it tells you |
|-------|-----|-------------------|
| punchtape installed? | `command -v punchtape` | empty output — not installed |
| inside a zone? | a `.punchtape/` directory in the current directory or any parent | the project already runs on punchtape |
| the working root | look around: an empty folder, or a project with code | fresh start, or bringing a standing project in |
| system | `uname -s` and `uname -m` | Linux and macOS on amd64/arm64 are supported; Windows is not |

## Step 2 — offer the human a choice

Name the situation you found; offer two to four options; wait for
the pick.

- **No punchtape, no zone.** Offer: install now and start this
  project under punchtape · install and just read the printed
  contract, decide later · not now.
- **Installed, no zone in this project.** Offer: bring this
  standing project under punchtape — the zone starts empty, the
  code that stands gets claimed by documents as work touches it ·
  start a fresh task in a new folder · not now.
- **A zone stands.** Do not offer to bypass it. Read
  `.punchtape/state.md` — the standing and the `next:` line — then
  `.punchtape/report.md` for the last facts, and propose the work
  the `next:` line names.

## Step 3 — install, once per computer

```bash
curl -fsSL https://github.com/neurophant/punchtape/releases/latest/download/get-punchtape.sh | bash
```

Downloads one small binary for the detected OS and architecture,
verifies its sha256 checksum, puts `punchtape` into `~/.local/bin`.
No admin rights, no account, no other dependencies. From source
instead: `git clone https://github.com/neurophant/punchtape && cd
punchtape && ./install.sh` — needs Go 1.26+.

## Step 4 — the birth, once per project

In the working root the human chose:

1. Run `punchtape`. With no zone yet it prints the whole contract
   and creates nothing — read it; that text is the law you will
   work under.
2. Create `.punchtape/demands.md` and put the human's words into
   it: verbatim, one demand a line, nothing of yours mixed in.
3. Run `punchtape` again — that run births the zone.

From here this page is done with you. The zone's
`.punchtape/contract.md` is the law; `.punchtape/template.yaml`
shows the document's shape; submissions go to
`.punchtape/inbox.yaml`; `.punchtape/state.md` always says what to
do now, in the `next:` line. A refused submission stays in
`inbox.yaml` — the report names the exact place and law; fix the
file in place and run again. A run costs one line of terminal and
nothing else; when in doubt, run the machine and read what it
says.

The repository's front page — `README.md` next to this `docs/`
folder — carries the rest: features, the measured numbers, the
architecture.

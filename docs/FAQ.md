# FAQ — question by question

**English** · [Русский](FAQ.ru.md) · [中文](FAQ.zh.md)

The hard questions, asked out loud and answered with facts. The
discipline of every answer: the truth first, then the mechanism,
then the proof (a link you can open), then what is or is not
being done about it. "Yes, that is so" is a valid answer here —
honest limits are part of this machine's design, not an
embarrassment to hide.

## Contents

- [How is this different from…?](#how-is-this-different-from)
- [The science](#the-science)
- [The theory](#the-theory)
- [The mechanics](#the-mechanics)
- [Trust and security](#trust-and-security)
- [The product](#the-product)
- [The one-line objections](#the-one-line-objections)
- [About the presentation](#about-the-presentation)

## How is this different from…?

**…from CI or pytest?**
CI judges a push, after the fact, and a human reads the verdict;
it never talks to the agent that wrote the code. punchtape works
inside the task loop: it owns the task's coverage (every line of
the human's words declared, every line of the tree claimed or
crossed out openly), runs the checks itself on every change, feeds
failures to the agent, and seals the finish with a checksum. A
pytest suite is one kind of fuel a check can run — your suite
stays and becomes stronger under the machine.
Proof: [USAGE.md](USAGE.md), the [demos](../demos/) — a verdict
recorded before any push exists.

**…from a harness (Claude Code, Cursor, …)?**
A harness runs your assistant's conversation and lives inside the
agent's environment — it is one of the parties to the work.
punchtape is the external party that owns the task's truth: the
words, the proofs, the verdict. The harness stays exactly what it
is; the arbiter is added beside it.
Proof: the machine never reads your chat and never asks the agent
anything — [ARCHITECTURE.md](ARCHITECTURE.md), "The demand stream".

**…from spec-driven tools like spec-kit?**
Those tools help the agent write and hold a spec inside its own
context; the spec and its checking remain the agent's own word.
Here the human's words are taken verbatim by an external machine,
checks are pinned to lines of code, executed by the machine, and
the finish is sealed by the machine — the agent cannot sign its
own homework.
Proof: the demand-append record and the chain —
[ARCHITECTURE.md](ARCHITECTURE.md), "The demand stream".

**…from guardrails and sandboxes?**
Guardrails constrain what an agent may do; sandboxes constrain
where. Neither measures whether the finished work meets the
owner's words. punchtape does exactly that and nothing else — it
composes with any sandbox you already use.
Proof: the verdict gates — [ARCHITECTURE.md](ARCHITECTURE.md),
"The verdict".

**…from "just write tests"?**
Tests are the fuel — the machine runs them, pins their bytes,
re-runs them when their claimed lines move, and holds the finish
while any line of the task or the tree is unresolved. Writing
tests is what your agent does under the machine; keeping them
honest is what the machine does for you.
Proof: [FEATURES.md](FEATURES.md), the claims and staleness laws.

**…from make/just?**
make runs what you list, once, when you ask; it holds no words,
no coverage, no verdict, no history. punchtape is a state machine:
it knows which checks went stale, which lines of your words are
unresolved, and it will not record done while anything is open.
Proof: the run's staleness arithmetic —
[ARCHITECTURE.md](ARCHITECTURE.md), "The run".

**…from linters?**
Linters check style against fixed rules for everyone. punchtape
checks behavior against your words for this task — the checks are
written for your task, pinned to your lines, re-run on change.

**…from the agent's own self-review?**
Self-review is the agent grading its own homework — the
principal-agent problem in one act. Here the agent may write the
checks, but the machine runs them, pins their bytes, and records
the verdict; a self-declared "done" is one bare line the machine
re-checks.
Proof: the completion act — [ARCHITECTURE.md](ARCHITECTURE.md),
"The verdict".

**…from IDE watch modes?**
A watch mode re-runs a fixed command on file change. It has no
notion of your words, coverage, or a sealed finish; and it runs
inside your editor, not for your agent.

**…from an AI code reviewer (CodeRabbit, Greptile, …)?**
A reviewer reads diffs and leaves comments; the reading is still
yours, and the loop stays open. punchtape runs instead of
commenting: the checks execute in the tree, the failures go back
to the agent as work, and the finish is held until every gate
stands. A comment can be ignored; a held verdict cannot.
Proof: the run and verdict laws — [ARCHITECTURE.md](ARCHITECTURE.md).

## The science

**The paper is a preprint — where is the peer review?**
There is none; it is a preprint, self-published on three preprint
servers with DOIs. Posting there is paid hosting — it buys a slot
and a permanent DOI, never review — and the three servers are
mirrors of one text, not three papers. What stands behind it: the
methodology and the raw campaign data are open in this repository,
and every number cited here traces to a file you can open.
Proof: [RESEARCH.md](RESEARCH.md), [METRICS.md](METRICS.md),
[campaigns/METHODOLOGY.md](../campaigns/METHODOLOGY.md).

**The author evaluated his own tool — conflict of interest?**
Yes, that conflict is real and declared. The design answers it:
acceptance was measured by blind judges against the human's
original words, never by the machine; the raw per-task data is
open for independent recounting.
Proof: the declared
[conflicts of interest](RESEARCH.md#conflicts-of-interest); the
judge role — [METHODOLOGY I.7](../campaigns/METHODOLOGY.md#i7-measurement);
blind export by sealed manifest.

**Who are the "blind judges" — did you hire them?**
No one was hired. The judge is a role executed by a model under a
frozen prompt: one micro-judge per cell, the cell's full working
tree, judged against the task's own inputs — never the group
label, never the machine's own verdict. The role scored 1720
points across both groups with 0 missing.
Proof: [METHODOLOGY I.7](../campaigns/METHODOLOGY.md#i7-measurement);
the count — [METRICS.md](METRICS.md).

**n=25 tasks — where are the confidence intervals?**
This is a field evaluation, not a clinical trial. The methodology
does fix the machinery: a two-stage preregistration before the
waves, Clopper–Pearson intervals for shares, bootstrap for
differences, the Holm correction for multiple comparisons. What
is shown here is the measured layer: every task documented per
series (appendices in RESEARCH.md), raw metrics per task in
files. Treat the headline numbers as measured
field results with open data.
Proof: [METHODOLOGY I.7](../campaigns/METHODOLOGY.md#i7-measurement),
Statistics; [RESEARCH.md](RESEARCH.md),
the per-series appendices.

**25 of 25 accepted — circular? The machine decides acceptance, of course it passes.**
Acceptance was not the machine's call. The machine recorded facts;
blind judges compared the results against the original human
words. The machine cannot grade meaning — that was the judges'
work (and remains yours).
Proof: [METHODOLOGY I.7](../campaigns/METHODOLOGY.md#i7-measurement);
the verdict gates measure formal
integrity only — [ARCHITECTURE.md](ARCHITECTURE.md), "The verdict".

**The tasks were chosen by the author — cherry-picking?**
The problem ledger law forbids it: every task that entered, in
every series, stands in the ledger with its input and outcome;
nothing was dropped.
Proof: [METHODOLOGY 0.2f](../campaigns/METHODOLOGY.md#02f-the-problem-ledger-law),
the problem ledger law; the ledger lives
with the raw data.

**No comparison with CI/spec-kit/competitors?**
True — the baseline was the same work without punchtape, not a
head-to-head with other tools. That is the honest boundary of the
claim: the numbers say what changed versus no arbiter, nothing
about rivals.
Proof: [RESEARCH.md](RESEARCH.md), the group design.

**Hawthorne effect — people behave better when watched?**
The isolation law split the groups by campaign — sessions,
permissions, and materials were procedurally identical, and judges
saw anonymized exports. A residual observer effect cannot be
excluded and is not claimed away.
Proof: [METHODOLOGY 0.2](../campaigns/METHODOLOGY.md#02-the-campaign-isolation-law)
(isolation), [I.8](../campaigns/METHODOLOGY.md#i8-validity) (validity).

**Costs and tokens were counted by the author himself?**
By the per-task metrics file law: every task carries its own
metrics file — tokens, time, and money per task, open in the raw
data, recountable by anyone.
Proof: [METHODOLOGY 0.2e](../campaigns/METHODOLOGY.md#02e-the-per-task-metrics-file-law);
the raw files under campaigns/.

**The 5–19× economics rides on invented hourly rates?**
The method fixes how hours are accounted (machine's economy,
[METHODOLOGY I.11](../campaigns/METHODOLOGY.md#i11-the-machines-economy))
and reports a range across the field tasks, not
a point. The rate sensitivity is why the range is wide — that is
honesty, not padding.
Proof: [METHODOLOGY I.11](../campaigns/METHODOLOGY.md#i11-the-machines-economy);
[METRICS.md](METRICS.md).

**Was the paper written by an LLM?**
The facts are machine-checkable — films, journals, checksums,
open data; the style of a text proves nothing either way and is
not offered as proof of anything.

**The judges are LLMs — LLM-as-judge is documented to be fragile.**
True, and said in the open: the judge role is executed by a
model. The design assumes that fragility. The judge executes a
fixed table row by row, byte-strict, with no right to edit or
adapt. A doubt counts against the product. A calibration must
pass before any count is recorded, and every table is validated
twice ahead of a wave — a golden product, then one with a broken
feature.
Proof: [RESEARCH.md](RESEARCH.md), "Judges";
[METHODOLOGY I.7](../campaigns/METHODOLOGY.md#i7-measurement).

## The theory

**Goodhart: the agent will game the checks, not the meaning.**
Yes — a check is bytes, and an agent can write a hollow one
(`echo ok`). The design narrows the gap rather than denying it:
checks are pinned to claimed lines (the exact expected output is
recorded with them) and stand on the spec page — the per-behavior
page where each check is displayed beside your words — exclusions
are loud with cited reasons, and the last look at meaning stays
human.
Proof: the spec pages and the claims law —
[ARCHITECTURE.md](ARCHITECTURE.md); [FEATURES.md](FEATURES.md).

**The oracle problem just moved into YAML?**
The first statement of the words is still human — that oracle is
you, and no tool removes it. What moves out of you is the
executing of the oracle: cheap, repeatable, byte-exact, and
immutable once written.
Proof: the demand stream and its chain —
[ARCHITECTURE.md](ARCHITECTURE.md).

**Is this verification or validation?**
Verification. The machine checks formal integrity: words resolved,
rows green with pinned bytes, tree equal to claims plus loud
exclusions. Whether that matches your intent — validation — is
the human's last look, said plainly in the README.
Proof: Honest limits in the [README](../README.md).

**"Proof" and "signature" sound grand for checksums.**
They are sha256 sums over canonical bytes — records that
recompute or fail. No cryptography beyond hashing, no identity
claims, and the docs use the words to mean exactly that.
Proof: the chain and digest laws —
[ARCHITECTURE.md](ARCHITECTURE.md).

**The agent writes its own checks — principal-agent unsolved?**
The writing is the agent's; the running, the pinning, and the
verdict are the machine's, and the sight of each check beside its
claim is yours. Self-certification is split across sides; it is
not claimed solved — it is made visible and expensive to fake.
Proof: the execution and verdict laws —
[ARCHITECTURE.md](ARCHITECTURE.md).

**A sufficiently detailed spec is code — you reinvented programming.**
There is truth in that: full exactness is programming. The machine
never asks the words to be complete. Your lines stay prose; per
behavior the agent adds one small check — a command and its exact
bytes — and the machine runs it, displayed beside the words on
the spec page. Where prose cannot pin something down, the honest
form is a loud exclusion with a reason, not a pretense of
precision.
Proof: the document law — [ARCHITECTURE.md](ARCHITECTURE.md);
the shape of a real corpus — [demos/gouija](../demos/gouija/).

**The agent will meet the spec by rewriting everything — behavior nobody wrote down is lost.**
Yes: words that were never written cannot be protected — no tool
changes that. What the machine does with the gap: every line of
the tree must be claimed by a check or crossed out loudly with a
reason, claimed lines re-earn their pinned bytes after any change,
and a finish that later breaks is marked open with the run that
broke it. The hole is written down, never silent — and the last
look at meaning stays yours.
Proof: the coverage and drift laws —
[ARCHITECTURE.md](ARCHITECTURE.md).

**Agents edit the tests themselves — reward hacking is documented.**
Documented in independent evaluations, and possible here. The
structure makes it loud: a re-spelled check is a new row — the
old row still stands and still pins its bytes. Retiring any check
requires a mandate that names the human line that ordered the
change. Every row's bytes and run numbers are journaled, before
and after, so weakening a guard leaves a written trace.
Proof: the merge and mandate laws —
[ARCHITECTURE.md](ARCHITECTURE.md).

## The mechanics

**Agent-written YAML is fragile.**
It is the most common refusal in the field: quotes around values
holding ": " or a comma, tabs refused, a multi-line pin kept one
line. Every refusal names its file, line, and law; a refused
submission stays in the inbox, and no data is lost.
Proof: the document law — [ARCHITECTURE.md](ARCHITECTURE.md);
the field rehearsal numbers — [RESEARCH.md](RESEARCH.md).

**The agent will write garbage checks — everything green.**
It can write a hollow check; the machine cannot read meaning. What
the machine does: pins the check to the lines it claims and puts
it on the spec page beside the human's words — a hollow check is
visible to you at one glance, and its claimed lines are its
responsibility forever.
Proof: the spec pages — [ARCHITECTURE.md](ARCHITECTURE.md).

**This is make plus YAML — reinvented pytest.**
What pytest does not hold: your words as the law of the task,
coverage of the tree by claims or loud exclusions, staleness
arithmetic per claimed line, a sealed verdict with a checksum,
and a film that outlives every session. pytest runs inside this
machine as fuel when you want it.
Proof: the run and verdict laws —
[ARCHITECTURE.md](ARCHITECTURE.md); a live film —
[demos/enimda](../demos/enimda/).

**Arbitrary commands — rm -rf?**
A command sees only `PATH`, `HOME`, and the zone's pinned values;
it runs from the tree root through the shell, dies with its
process group, and the whole thing works offline. The owner's
words choose what runs — as they always did.
Proof: the your-tree law — [ARCHITECTURE.md](ARCHITECTURE.md),
[BOOTSTRAP.md](BOOTSTRAP.md).

**The journal got corrupted — is the zone dead?**
A torn tail quarantines itself to a sidecar and the intact records
load; damage elsewhere is refused with the file and record named,
and never absorbed silently. With the zone in git, any green state
restores with one checkout — verified in the cookbook.
Proof: the torn-tail law — [ARCHITECTURE.md](ARCHITECTURE.md);
recipes with real runs — [USAGE.md](USAGE.md).

**The protocol will change — old films rot?**
A zone born by another machine version is refused — versions
never mix silently. The film is files: archived zones stay
readable as records even when a new zone starts fresh.
Proof: the birth and version laws —
[ARCHITECTURE.md](ARCHITECTURE.md).

**NFS and locks?**
The zone lock is kernel-released flock on local filesystems; on
NFS the safe answer is: keep the zone local.
Proof: the lock law — [ARCHITECTURE.md](ARCHITECTURE.md).

**No Windows?**
Not today — Linux and macOS, amd64 and arm64. Ports are open
work, not a design refusal.
Proof: the [README](../README.md), Requirements.

**One import (yaml) and Go 1.26 — dependencies?**
One external import in the machine, stated with pride; Go 1.26+
only to build from source — the release binary needs nothing but
a unix shell.
Proof: [ARCHITECTURE.md](ARCHITECTURE.md), the package maps.

**Prompt injection through demands or reports?**
The machine never executes prose: words are lines taken verbatim,
documents are data validated by strict decode, and rows run what
their pins say. An injected sentence in a report is text you read
— the machine does not act on it.
Proof: the document and execution laws —
[ARCHITECTURE.md](ARCHITECTURE.md).

**The agent can edit .punchtape/ files directly.**
It can touch them; the machine then refuses: the chain recomputes
over the bytes, records re-compose or the zone is damaged — loud,
named, exit 1. Tampering is detectable by design, not prevented
by permissions.
Proof: the chain law — [ARCHITECTURE.md](ARCHITECTURE.md); live
refusal lines — [USAGE.md](USAGE.md).

**Vendor lock-in on the format?**
The zone is plain files — markdown, YAML, sha256-named objects —
readable without the machine, copyable with `cp`, diffable with
git. The film's readability does not depend on this binary.
Proof: the zone's file list — [ARCHITECTURE.md](ARCHITECTURE.md).

**Team workflows — zone in the repo or out of it, conflicts?**
Both shapes are walked with recipes and real refusal lines: what
to commit (the film) and what stays generated (`cache/`,
`derived/`, `run.lock`), how merges resolve (one side whole, the
journal never merges), how a damaged zone heals (restore the
bytes).
Proof: the team cookbook — [USAGE.md](USAGE.md).

**curl | bash — piping a script into a shell?**
Fair: piping unseen code into a shell is a real risk. The script
is in this repository — read it before running it. It verifies
the release checksums and installs one binary into
`~/.local/bin`, no admin rights. If you want nothing piped:
download the binary and its checksum by hand, or build from
source.
Proof: [get-punchtape.sh](../get-punchtape.sh);
[BOOTSTRAP.md](BOOTSTRAP.md).

**The agent is the machine's operator — the fox guards the henhouse.**
Yes: the operator is the agent, and an agent that controls your
whole screen can fake anything you see. The truth therefore lives
in files the machine owns: the journal recomputes its chain every
run, the views are regenerated from the records, the replays
rebuild films byte-identical, and the zone can sit in git beside
the code. What no design can give you is an agent that cannot lie
— which is why the verdict page is written for your eyes.
Proof: the chain law — [ARCHITECTURE.md](ARCHITECTURE.md); the
replays — [demos](../demos/).

**Is the checker itself an AI — will it hallucinate verdicts?**
There is no model inside the machine. Rows are commands; the
machine runs them and compares bytes; the same inputs give the
same answer on every machine. A flaky command is the task's own
problem — the deadline kills the hang, the fact is recorded red,
and the machine's own answer never varies.
Proof: [FEATURES.md](FEATURES.md), "Deterministic runs";
[ARCHITECTURE.md](ARCHITECTURE.md), "Staleness and execution".

**100% function coverage — a vanity metric?**
It is a floor, and nothing more is claimed for it. The number
means: every function of the machine executes at least once under
the build's battery — the tests plus the instrumented binary run
on the built-in example task. The semantic weight is
carried by two other gates: the example must close its verdict
through the real machine, and every enforced rule must be printed
by the contract — the same set, checked both ways.
Proof: the build-time gates — [ARCHITECTURE.md](ARCHITECTURE.md).

**Long tasks rot the context — the state cannot survive them.**
Context is the first thing to rot, true. The state lives outside
every context: the standing, the corpus, the lessons and the next
instruction are files, and the terminal line is one line. A
session or a model can be swapped mid-task — the next run reads
the zone, not the chat. The demos show work continuing on a
standing verdict across runs.
Proof: the journal and the views —
[ARCHITECTURE.md](ARCHITECTURE.md); [demos/enimda](../demos/enimda/).

## Trust and security

**A hostile agent can sabotage the zone.**
True — an agent with file access can burn the film. What it
cannot do is fake a standing verdict: the chain and the records
recompute, and sabotage shows up as loud damage, named by file.
Keep the zone in git and even sabotage is revertible.
Proof: the chain law — [ARCHITECTURE.md](ARCHITECTURE.md).

**Can the film be forged after the fact?**
Each verdict's signature is sha256 over canonical record bytes;
change anything and the sums stop matching on the next run. A
forger must recompute every hash consistently — the machine will
do that check for the reader.
Proof: the fold and signature laws —
[ARCHITECTURE.md](ARCHITECTURE.md).

**A replaced binary could lie about verdicts.**
Yes — trust the binary or build it from source; the release
carries checksums the installer verifies. The films remain
independently checkable files regardless.
Proof: the release checksums; [BOOTSTRAP.md](BOOTSTRAP.md).

**Why trust this machine at all?**
Because it asks for the smallest trust available: read the film —
plain files — and rerun the replay; everything it claims is
bytes you can check. The demos do exactly that, from scratch, on
your machine.
Proof: [demos/](../demos/) — each with a `replay.sh`.

## The product

**It's 0.5 — too early to depend on?**
Fair. The protocol is young and versions refuse to mix — that is
the honest cost of a sealed past. What backs early adopters: open
code, open data, films in plain files that outlive the binary.

**How does this make money?**
MIT license, no account, no telemetry — there is no revenue
mechanism in the machine. What comes later is an open question,
answered when there is something to answer with.

**What is the roadmap?**
Windows, team conveniences around the zone, and integrations live
or die by real users' issues — the film decides what matters.

**Bus factor of one?**
True today. Mitigation is structural: open methodology, open raw
data, plain-file films, MIT — the work survives the author.

**Why a CLI and not a service?**
Because the arbiter must be outside every conversation and
account: one binary beside your project, offline, no server to
trust, no data to leak. A service would be another party — the
thing this machine exists not to be.
Proof: the design law — [ARCHITECTURE.md](ARCHITECTURE.md).

**Integrations with my editor/CI/agent?**
The machine is agnostic by law — any agent that can run a shell
command drives it; CI can replay a zone after push. Demos show
four stacks; your harness is the fifth.
Proof: the agnosticity law — AGENTS.md of this repository; the
[demos](../demos/).

**The harnesses will build this in — why won't you be absorbed?**
Features get absorbed; that is how software goes. The property
that cannot be absorbed into a harness is the one this machine
exists for: standing outside the conversation. A built-in checker
is the agent grading its own homework again — the
self-certification this project exists to break. An external,
agnostic, offline binary works with every agent, today's and
tomorrow's; if harnesses one day ship sealed external verdicts
over open films, the users have won.
Proof: the design law — [ARCHITECTURE.md](ARCHITECTURE.md),
"What the machine is".

**Who uses this in production?**
No production users are claimed. What stands open today: three
rounds of field tests with raw per-task data, and four demo films
anyone can replay from scratch. This answer will change the day
it stops being true.
Proof: [RESEARCH.md](RESEARCH.md); [demos](../demos/).

**The demos are someone else's code outside your linters.**
Fair read of the Makefile: the demo project slices are excluded
from the lint gates, by one visible line with its reason. The
machine's own code — cmd/, internal/, the tools — runs under the
full battery. The demos carry pinned slices of real projects as
material for the films, each named with its source commit, and
each film is regenerated by its replay.
Proof: the Makefile; each demo's README with its pinned source.

## The one-line objections

**"Models got better — the problem is gone."**
The paper's control group worked with the same-class models and
produced 32 human acceptances per 25 tasks; better code does not
check itself. When it does, retire this tool with our blessing.
Proof: [METRICS.md](METRICS.md).

**"Just check it yourself."**
That is the baseline, measured: 4.4–7.3 hours per 25 tasks. The
machine's group: 0.8–1.3 hours, 25 of 25 first-submission
finishes.
Proof: [METRICS.md](METRICS.md).

**"Token overhead eats the savings."**
Counted in the field: total task cost 5–19× lower with the
machine, tokens included in the per-task metrics files.
Proof: [METHODOLOGY 0.2e](../campaigns/METHODOLOGY.md#02e-the-per-task-metrics-file-law),
[I.11](../campaigns/METHODOLOGY.md#i11-the-machines-economy);
[METRICS.md](METRICS.md).

**"This is TDD from 1999."**
TDD never held the human's words as an immutable chain, never
claimed the tree, never sealed a verdict; this is the loop closed
by an external party — the part 1999 left with the human.
Proof: the chain and verdict laws —
[ARCHITECTURE.md](ARCHITECTURE.md).

**"One more layer of bureaucracy."**
The bureaucracy you have now is you, re-reading diffs. The layer
added here hands that reading to a machine that just does it;
the words you write are the entire interface.
Proof: [METRICS.md](METRICS.md), the verifying-time rows.

**"Offline — can't do cloud?"**
Offline is the feature: the arbiter holds your task's truth
beside your code, where no account, no telemetry, and no network
call can touch it.
Proof: the your-tree law — [ARCHITECTURE.md](ARCHITECTURE.md).

**"The 'machine' style is a cult."**
The voice is chosen and kept consistent so terms mean one thing;
if it reads as ceremony, the glossary maps every word to its plain
file-and-bytes meaning.
Proof: [GLOSSARY.md](GLOSSARY.md).

**"MIT today, patents tomorrow?"**
MIT today. Nothing else is filed or planned; if that ever
changes, the license already granted cannot be taken back.

## About the presentation

**Why does the README read the way it does?**
It is built problem → fix → proof → demos → try it, with every
limit stated on the front page; the depth lives in USAGE. If a
section stops selling or starts preaching, that is a bug — say
so in an issue.

**Why the name and the tape?**
Punched tape: a hole, once punched, cannot be unpunched — the
same property the journal gives every recorded fact.
Proof: the chain law — [ARCHITECTURE.md](ARCHITECTURE.md).

**Why does the glossary read like a law book?**
Because each term names one mechanism with one law; loose
synonyms are how measured things rot. One word, one meaning, one
file.
Proof: [GLOSSARY.md](GLOSSARY.md).

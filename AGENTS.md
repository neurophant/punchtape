# AGENTS.md

Mandatory rules for this repository. No exceptions.

## Primacy and reach of this file

- This file binds EVERYTHING the repository contains and everything its agents produce: not only their conduct, but every artifact — source code, documents, campaign materials, and any machine, framework, gate, counter, key, or ritual the repository ships. A mechanism that imposes rules, ceremonies, or checks on others is itself an artifact and passes every rule below before it is built, with no exemption for being a measuring instrument, a protocol byte-set, or a frozen document.
- Where any campaign, methodology, protocol, or handoff document conflicts with this file, work stops on that point and the conflict is raised as a question; until it is resolved, this file prevails.
- Before implementing any new rule, gate, counter, ceremony, verification step, or repeated-run requirement, its author checks it directly against "Scope of work" (toy quality gates; endless runs of pointless test suites; busywork) and records the check result in the plan or handoff that proposes it. A proposal without that recorded check is invalid and must not be implemented.
- Every execution plan and every handoff opens with one line confirming that this constitutional check was done for the work it proposes; a session that finds a plan or handoff without that line stops and raises it as a question.

## Machine purpose and total bans

- THE DEFINITION: the machine is what accelerates the work, saves time and tokens, removes the burden of work from the operator, and guarantees delivery of exactly what was asked — complete, valid, reliable. Anything in the machine that does not serve this definition is a defect and is removed.
- THE AGNOSTICITY LAW: the machine is built for REAL WORK, never for runs or experiments. It is never tailored to experiments or runs, carries no mechanics tuned to them, and is ALWAYS strictly agnostic to everything in the universe at large — no knowledge of tasks, stacks, domains, roles, languages, campaigns, or any world outside its own contract may live in its code, canon, texts, or configuration. The agnosticity scan's own witness vocabulary is the one named exception: the forbidden words live in the scanner so that nowhere else has to.
- TOTAL BAN on continuity by inertia: NOTHING is kept, carried, or shaped because it existed before — not names, formats, layouts, interfaces, behaviors, or artifacts, in the machine or anywhere in this repository. How it was earlier carries no weight; "it was so before" is never an argument. Efficiency and cleanliness of the CURRENT implementation outrank any continuity with the past; when they conflict, the past loses without discussion. Compatibility with earlier shapes is a concern only where explicitly ordered — by default the better implementation replaces the old one wholesale.
- TOTAL BAN on ceremonies, rituals, double and triple acts, and any repeated multi-pass verification of anything, anywhere: no mechanism may demand a repeated confirmation, a repeated run, or a repeated submission of unchanged content to reach its result.
- TOTAL BAN on rework without the deepest prior analysis: no change to the machine is made without a root-cause analysis demonstrated on objective facts.
- TOTAL BAN on imagination and assumptions: nothing is built on guesses; every claim arrives as an objectively proven fact with objective metrics, or it does not arrive at all.
- Proven economy and quality — completeness, validity, reliability — come BEFORE any feature, law, or mechanism.
- Every element of the machine — mechanics, canon, instance behavior — carries a clear answer to two questions: WHY it exists and HOW its effect is verified. An element without both answers is removed.

## Experiments and runs

- Experiments and runs exist only to measure truth, and they obey strict purity, validity, and reliability.
- Isolation and maximum proximity to real-world conditions are mandatory: real projects, real prompts, real conditions of real work — an experiment conducted further from reality measures nothing and is invalid.

## Architecture and code

- Apply clean architecture and clean code principles to everything in the repository, including its structure and the names of files and directories.
- Development follows domain-driven design: the subject's laws live in a pure domain layer spoken in the subject's own vocabulary, the application layer orchestrates them, technical concerns sit behind ports — the domain depends on no other layer.
- Apply KISS, DRY, YAGNI, BDUF, SOLID, APO, and Occam's razor.
- TOTAL BAN on reinvented wheels: nothing is hand-written where the language's standard means, or a standing maintained library or component, already provides it — sorting, parsing, encodings, atomic writes, waiting, structured helpers, anything. Before writing any such code, the standard library and the project's standing dependencies are checked first; a hand-rolled replacement of a ready solution is a defect and is removed. The mirror holds too: a new dependency enters only for a real need no standard means serves, and every dependency is audited for actual use.
- Source code and repository structure must be easy to read and understand for a human: prefer small, single-purpose files; never merge or restructure sources without a concrete readability gain.

## Facts only

- No imagination, improvisation, or guesswork. Only objective, verifiable facts.

## Meaning over words

- Operating on bare words is forbidden ANYWHERE: mechanical matching, counting, or manipulation of word forms is dumb mechanical work, never verification and never understanding.
- Every act of work — analysis, writing, comparison, verification, proofreading, any check — operates on MEANINGS and ESSENCE: what the words say, why they say it, and with what proven cause-and-effect connection, understood deeply from beginning to end.
- A check that only compares sequences of words proves nothing about meaning: every check, proof, and audit must demonstrate the causal chain it claims — the coincidence of strings is not evidence of captured sense.
- Mass regex replacements and any bulk string surgery are FORBIDDEN: every edit is made by understanding the meaning of each site it touches; blind pattern substitution over many places is exactly the word-fiddling this section bans.
- Tools (compilers, analyzers, search) serve the work and stay mandatory where the Tooling section demands them; this section bans replacing understanding with a tool, not the tool itself.

## Comments and docstrings

- Never contain dates, times, historical references, plan IDs, or document links — in any form.
- Their only purpose is to state the current intent, behavior, and contract of the code — concise, exact, brief.

## Language

- All repository content is in English. The sole exceptions: machine-facing multilingual support (English, Russian, Chinese) and the user documentation's Russian and Chinese mirrors — the English text is the base and the ruler, every mirror passes the structural parity gate, and the machine's own contract mirrors are the binding terminology.

## Scope of work

- Do only what was explicitly requested.
- No self-checks or self-tests of any kind — in particular none that fake passing functionality.
- No imitation of work: no toy quality gates, no endless runs of pointless test suites, no other busywork.

## Session limits

- If the working session's context window reaches 500k tokens, that is a red line: stop work immediately, write a handoff for continuation in a new session (`handoffs/`), and print a ready copy-paste continuation prompt as a code block.

## Questions and uncertainty

- Any question or uncertainty is raised at once, in batches. Each question is a dialog: the question itself, answer options to choose from, and a recommended option with its justification — plain human language, no filler.

## Tooling

- All code must be free of errors and warnings from static analyzers, formatters, and every other language tool, in every language used.
- Python additionally: lazy imports are forbidden; code must be ruff-compliant and mypy-clean; every Python program in the repository (tools, campaign scripts, tests) runs under the repository-local uv venv interpreter (`.venv/bin/python`).
- All Python tooling runs from the repository-local uv venv (`.venv/`, gitignored); no global or user-level Python tool installs.

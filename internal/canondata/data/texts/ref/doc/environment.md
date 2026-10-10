The environment the machine runs check rows in:

## Direct execution

Rows and the declared build run DIRECTLY in your work dir - no
wrappers, no containers, no sandboxes: cwd is the work dir, PATH is
the system PATH; the machine adds nothing to
the environment and hides nothing from it; toolchains are whatever
the host provides.

## Timeouts - mechanism parameters

Every row runs under a ceiling of 60 s by default; a row's own
`timeout` field may raise it, up to 300 s. The build gets 300 s.
These are parameters of the mechanism, not budgets:
the machine sets no limit on how long you may work or repair.

## The digested surface

The tree digest of the verdict covers the work tree minus .punchtape/
minus the build-ignore globs - that remainder is the accounted
product. A run may not move the tree outside those two: a command
that writes elsewhere reddens the run (the red line "tree moved"
names the paths). The sibling input/ directory sits outside the work
dir and is never digested.

## Inputs and samples

The instance's input directory is the sibling input/; a seed with
`from-input:` copies one of them into the work dir, file-to-file,
for the row's duration. Sample files for `bytes:` live next to the
submission, in .punchtape/inbox/. Seeds are written before their row
and removed after it - the file, and the directories the machine
created to place it; a seed over an existing file restores the
original bytes when the row ends.

## What the machine writes

In the work dir the machine writes exactly: AGENTS.md at the root,
and .punchtape/ - state, journal, canon atoms, checkpoints, inbox,
cache, reference, spec.md (re-rendered after each accepted
submission) and SUMMARY.md at the verdict. Everything else in the
work dir is yours; the machine never reads your code files.

#!/usr/bin/env python3
"""form-hints.py — the mechanical form-hints sidcar generator of the
case-0 campaign: for a
machine hand's cell the staged input carries a sidcar of VERBATIM
brief lines a form pin can quote — a form pin wants a one-line source
for every formal pin (an exact flag, an exact output string, an exact
format), and the machine cannot pick citations itself (agnosticism:
it never sees the brief). The sidcar is PURE MECHANICS over the
cell's own brief — a deterministic grep, nothing enters from outside;
the freeze re-runs this same tool over the same brief bytes and pins
the digest, so a doctored sidcar fails the staged-freeze check.

A brief line is form-bearing when it carries a greppable formal term:
  - a flag spelling  --name  (the exact flag names the brief demands);
  - an indented usage-block line (4+ spaces: command and format
    examples the brief shows);
  - a double-quoted literal (the exact event/error/output strings the
    brief words).
Every form-bearing line rides VERBATIM with its 1-based line number.
A brief with none says so explicitly (never silence).

Usage: form-hints.py <brief.md>   (the sidcar text on stdout)
Exit 0 always (an empty hint set is a legal, printable answer).
"""
import re
import sys

FLAG = re.compile(r"--[A-Za-z][A-Za-z0-9_-]*")
QUOTED = re.compile(r'"[^"\n]+"')
EMPTY_MARK = "no form-bearing terms"


def form_hints(brief: str):
    """The form-bearing lines of the brief: (lineno, verbatim line)."""
    out = []
    for n, line in enumerate(brief.splitlines(), 1):
        if not line.strip():
            continue
        indented = line.startswith("    ")
        if indented or FLAG.search(line) or QUOTED.search(line):
            out.append((n, line.rstrip()))
    return out


def render(brief: str) -> str:
    head = [
        "# form-hints — verbatim brief lines for form citations.",
        "# Generated mechanically from brief.md:",
        "# a form pin (an exact flag/string/format/code) requires a form: citation",
        "# verbatim from the brief — take the line from here or from the brief itself.",
        "# Format: L<line number>: <the brief's verbatim line>.",
    ]
    hints = form_hints(brief)
    if not hints:
        return "\n".join(head + [f"# {EMPTY_MARK}: the brief has no lines with exact",
                                 "# flags/quotations/example blocks that need",
                                 "# a source for the form field."]) + "\n"
    body = [f"L{n}: {line}" for n, line in hints]
    return "\n".join(head + body) + "\n"


def main():
    if len(sys.argv) != 2:
        print(__doc__, file=sys.stderr)
        sys.exit(2)
    brief = open(sys.argv[1], encoding="utf-8").read()
    sys.stdout.write(render(brief))


if __name__ == "__main__":
    main()

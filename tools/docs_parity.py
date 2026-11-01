"""The user docs' mirror parity gate.

Every Russian and Chinese mirror must match its English base exactly
in structure: the same heading ladder, byte-identical code blocks and
inline code (the machine's own strings never translate), the same
table shapes with the same numbers, the same link targets, the same
count of bold entries, and its own table of contents resolving against
its own headings. Meaning is the translators' charge; structure is the
machine's — a mirror that drifts in shape has already drifted in sense.
"""

from __future__ import annotations

import re
import sys
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
BASES = [
    "README.md",
    "docs/FEATURES.md",
    "docs/ARCHITECTURE.md",
    "docs/GLOSSARY.md",
    "docs/METRICS.md",
]
LANGS = ("ru", "zh")


def split_fences(text: str) -> tuple[list[str], list[str]]:
    """The document's prose with fences cut, and the fence bodies —
    each body opening with the fence's info line."""
    prose: list[str] = []
    fences: list[str] = []
    buf: list[str] = []
    infence = False
    for line in text.split("\n"):
        if line.startswith("```"):
            if infence:
                fences.append("\n".join(buf))
                buf = []
            else:
                buf = [line[3:]]
            infence = not infence
            continue
        (buf if infence else prose).append(line)
    return prose, fences


def headings(prose: list[str]) -> list[int]:
    return [
        len(m.group(1)) for line in prose for m in [re.match(r"^(#{1,6})\s", line)] if m
    ]


def slug(text: str) -> str:
    return re.sub(r"[^\w\- ]", "", text.lower(), flags=re.UNICODE).replace(" ", "-")


def anchors(prose: list[str]) -> set[str]:
    out: set[str] = set()
    seen: dict[str, int] = {}
    for line in prose:
        m = re.match(r"^(#{1,6})\s+(.*)$", line)
        if not m:
            continue
        a = slug(m.group(2).strip())
        if a in seen:
            seen[a] += 1
            a = f"{a}-{seen[a]}"
        else:
            seen[a] = 0
        out.add(a)
    return out


def inline_code(prose: list[str]) -> Counter[str]:
    text = "\n".join(prose)
    spans = re.findall(r"`([^`]*)`", text, flags=re.DOTALL)
    spans = [re.sub(r"\s+", " ", span).strip() for span in spans]
    # Spans closing with a sentence period are user-language prompts
    # (machine lines never end with one); they may translate.
    return Counter(span for span in spans if not span.endswith((".", "。")))


def tables(prose: list[str]) -> list[tuple[int, int, Counter[str]]]:
    out: list[tuple[int, int, Counter[str]]] = []
    i = 0
    while i < len(prose):
        if (
            prose[i].startswith("|")
            and i + 1 < len(prose)
            and re.match(r"^\|[\s:-]+\|", prose[i + 1] or "")
        ):
            hdr, rows = prose[i].count("|"), 0
            nums: Counter[str] = Counter()
            j = i + 2
            while j < len(prose) and prose[j].startswith("|"):
                rows += 1
                nums.update(re.findall(r"\d+(?:\.\d+)?", prose[j]))
                j += 1
            out.append((rows, hdr, nums))
            i = j
        else:
            i += 1
    return out


def links(text: str) -> set[str]:
    return set(re.findall(r"\]\(([^)#]+)(?:#[^)]*)?\)", text))


def link_family(base: str) -> set[str]:
    stem = Path(base).name.rsplit(".", 1)[0]
    return {f"{stem}.md", f"{stem}.ru.md", f"{stem}.zh.md"}


def table_spans(prose: list[str]) -> list[tuple[int, int]]:
    spans: list[tuple[int, int]] = []
    i = 0
    while i < len(prose):
        if (
            prose[i].startswith("|")
            and i + 1 < len(prose)
            and re.match(r"^\|[\s:-]+\|", prose[i + 1] or "")
        ):
            j = i + 2
            while j < len(prose) and prose[j].startswith("|"):
                j += 1
            spans.append((i, j))
            i = j
        else:
            i += 1
    return spans


def numbers(text: str) -> Counter[str]:
    """The significant numbers of the prose: decimals and runs of
    three digits or more, tables aside — every table's numbers are
    already compared cell-multiset by their own rule."""
    prose, _ = split_fences(text)
    in_table = set().union(*(range(a, b) for a, b in table_spans(prose)))
    body = "\n".join(line for n, line in enumerate(prose) if n not in in_table)
    return Counter(re.findall(r"\d+\.\d+|\d{3,}", body))


def bold_entries(prose: list[str]) -> int:
    return "\n".join(prose).count("**")


def mermaid_skeletons(fences: list[str]) -> list[str]:
    out = []
    for body in fences:
        head, _, rest = body.partition("\n")
        if head.strip() != "mermaid":
            continue
        stripped = re.sub(r'"[^"]*"', '""', rest)
        out.append(re.sub(r"\s+", " ", stripped).strip())
    return out


def plain_code(fences: list[str]) -> list[str]:
    """Fences whose bytes never translate: code and machine output,
    mermaid aside (its labels carry human text in quotes)."""
    return [
        body
        for body in fences
        if body.partition("\n")[0].strip() not in ("mermaid", "text")
    ]


def prompt_fences(fences: list[str]) -> int:
    """Fences of the `text` family: the user's own language — prompts
    a localized reader pastes to their assistant. Their bytes may
    translate; the meaning cycle guards them."""
    return sum(1 for body in fences if body.partition("\n")[0].strip() == "text")


def check(base: str, lang: str) -> list[str]:
    issues: list[str] = []
    bpath, mpath = ROOT / base, ROOT / base.replace(".md", f".{lang}.md")
    if not mpath.is_file():
        return [f"{mpath.name}: mirror missing"]
    btext, mtext = bpath.read_text(), mpath.read_text()
    bprose, bfences = split_fences(btext)
    mprose, mfences = split_fences(mtext)
    if headings(bprose) != headings(mprose):
        issues.append(
            f"heading ladder differs: base {headings(bprose)} vs mirror {headings(mprose)}"
        )
    if plain_code(bfences) != plain_code(mfences):
        issues.append("plain code fences differ (must be byte-identical)")
    if prompt_fences(bfences) != prompt_fences(mfences):
        issues.append("prompt fence count differs")
    if mermaid_skeletons(bfences) != mermaid_skeletons(mfences):
        issues.append("mermaid skeletons differ")
    if inline_code(bprose) != inline_code(mprose):
        b, m = inline_code(bprose), inline_code(mprose)
        diff = {k for k in set(b) | set(m) if b[k] != m[k]}
        issues.append(f"inline code differs: {sorted(diff)[:5]}")
    if tables(bprose) != tables(mprose):
        issues.append("tables differ (shape or numbers)")
    fam = link_family(base)
    variant = {
        f"{t[:-3]}.{lang}.md"
        for t in links(btext)
        if t.endswith(".md") and not t.startswith("http")
    }
    allowed = links(btext) | variant | fam
    extra = links(mtext) - allowed
    if extra:
        issues.append(f"link targets differ: {sorted(extra)[:5]}")
    if numbers(btext) != numbers(mtext):
        b, m = numbers(btext), numbers(mtext)
        diff = {k for k in set(b) | set(m) if b[k] != m[k]}
        issues.append(f"numbers differ: {sorted(diff)[:8]}")
    if bold_entries(bprose) != bold_entries(mprose):
        issues.append(
            f"bold entries differ: {bold_entries(bprose)} vs {bold_entries(mprose)}"
        )
    own = anchors(mprose)
    bad = [a for a in re.findall(r"\]\(#([^)]+)\)", mtext) if a not in own]
    if bad:
        issues.append(f"mirror ToC anchors unresolved: {bad[:3]}")
    return [f"{mpath.name}: {i}" for i in issues]


def main() -> int:
    failures: list[str] = []
    for base in BASES:
        for lang in LANGS:
            failures += check(base, lang)
    for line in failures:
        print(line, file=sys.stderr)
    print(
        f"docs parity: {len(BASES) * len(LANGS) - len(failures)}/{len(BASES) * len(LANGS)} mirrors structurally exact"
    )
    return 1 if failures else 0


if __name__ == "__main__":
    raise SystemExit(main())

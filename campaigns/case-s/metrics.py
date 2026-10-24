"""Write one wave's METRICS document.

Everything the wave measured, on one page: the usage table straight
from the collector's sums, the per-brief medians that answer the
campaign's one question — is the machine cheaper than solo, and by
how much — the ownership table (each hand, its own operator, the
sum: the full cost the kinds are compared by), the meanings' etalon
pages (the yardstick every judge of a meaning read,
byte-identically), and every judge's page verbatim. The numbers are
rounded for the eye; their exact integers stand in usage.tsv, and
each of those traces back to the platform database.

Before anything is written the collection itself is VERIFIED: the
hands complete against the manifest, every hand sessioned, the sums
arithmetically consistent, every hand's judge page present, the
operators exactly one per hand when the wave ran them — a failed
check stops the metrics, nothing invalid is written. The verified
whole of the wave then lands in the wave root's METRICS.md — one
deterministic document; nothing of the run enters the repository.
"""

from __future__ import annotations

import argparse
import csv
import json
import statistics
import sys
from pathlib import Path
from typing import Any

CAMPAIGN = Path(__file__).resolve().parent


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--wave", required=True, help="the wave root")
    return parser.parse_args(argv)


def millions(tokens: int) -> str:
    return f"{tokens / 1_000_000:.2f}M"


def duration(ms: int) -> str:
    seconds = round(ms / 1000)
    return f"{seconds // 60}m{seconds % 60:02d}s" if seconds >= 60 else f"{seconds}s"


def hand_meanings(manifest: dict[str, Any]) -> dict[str, str]:
    """Every hand of the manifest with its meaning (the brief's
    sense); judge and operator rows name the same hands."""
    out: dict[str, str] = {}
    for hand in manifest["hands"]:
        out[str(hand["hand"])] = str(hand["brief"]).split("/", 1)[0]
    return out


def validate(rows: list[dict[str, str]], manifest: dict[str, Any], wave_dir: Path) -> None:
    """The collection's own correctness, checked before any write:
    completeness, sessions, arithmetic, judge pages, operators. A
    named failure stops the metrics — nothing invalid is written."""
    meanings = hand_meanings(manifest)
    hands = set(meanings)
    usage_hands = {r["hand"] for r in rows if r["kind"] in ("machine", "solo")}
    for missing in sorted(hands - usage_hands):
        print(f"hand {missing} is missing from the collection", file=sys.stderr)
    for extra in sorted(usage_hands - hands):
        print(f"usage row {extra} is no hand of the manifest", file=sys.stderr)
    if hands != usage_hands:
        raise SystemExit(1)
    for row in rows:
        if row["tail"] == "no-sessions":
            print(f"hand {row['hand']} has no sessions — run the hands first", file=sys.stderr)
            raise SystemExit(1)
        if int(row["in_out"]) != int(row["input"]) + int(row["output"]):
            print(f"row {row['hand']}: in+out != input+output", file=sys.stderr)
            raise SystemExit(1)
        if int(row["total"]) != int(row["in_out"]) + int(row["cache_read"]):
            print(f"row {row['hand']}: total != in+out+cache_read", file=sys.stderr)
            raise SystemExit(1)
    for hand in sorted(hands):
        if not (wave_dir / "judged" / f"{hand}.md").is_file():
            print(f"judge page missing: judged/{hand}.md", file=sys.stderr)
            raise SystemExit(1)
    if (wave_dir / "operators").is_dir() and list((wave_dir / "operators").glob("*.prompt")):
        expected = {f"operator/{hand}" for hand in hands}
        actual = {r["hand"] for r in rows if r["kind"] == "operator"}
        for missing in sorted(expected - actual):
            print(f"operator of {missing.split('/', 1)[1]} is missing from the collection", file=sys.stderr)
        for extra in sorted(actual - expected):
            print(f"usage row {extra} is no hand's operator of the manifest", file=sys.stderr)
        if expected != actual:
            raise SystemExit(1)


def usage_table(rows: list[dict[str, str]]) -> list[str]:
    header = (
        "| hand | kind | sessions | turns | requests | in | out | cache "
        "| in+out | total | work | span | tail |"
    )
    rule = "|---|---|---|---|---|---|---|---|---|---|---|---|---|"
    body = [
        f"| {r['hand']} | {r['kind']} | {r['sessions']} | {r['turns']} "
        f"| {r['requests']} | {millions(int(r['input']))} | {millions(int(r['output']))} "
        f"| {millions(int(r['cache_read']))} | {millions(int(r['in_out']))} "
        f"| {millions(int(r['total']))} | {duration(int(r['work_ms']))} "
        f"| {duration(int(r['span_ms']))} | {r['tail']} |"
        for r in rows
    ]
    return [header, rule, *body]


def median_table(rows: list[dict[str, str]]) -> list[str]:
    header = (
        "| brief | kind | hands | median in+out | median camp in+out "
        "| median work | median camp work |"
    )
    rule = "|---|---|---|---|---|---|---|"
    body = []
    for brief in sorted({r["hand"].rsplit("-", 2)[0] for r in rows}):
        for kind in ("machine", "solo"):
            hands = [
                r
                for r in rows
                if r["kind"] == kind and r["hand"].startswith(f"{brief}-")
            ]
            if not hands:
                continue
            in_out = statistics.median(int(r["in_out"]) for r in hands)
            work = statistics.median(int(r["work_ms"]) for r in hands)
            camps_in, camps_work = camp_sums(rows, hands)
            camp_in = statistics.median(camps_in) if camps_in else 0
            camp_work = statistics.median(camps_work) if camps_work else 0
            body.append(
                f"| {brief} | {kind} | {len(hands)} | {millions(int(in_out))} "
                f"| {millions(int(camp_in))} | {duration(int(work))} "
                f"| {duration(int(camp_work))} |"
            )
    return [header, rule, *body]


def operator_row(rows: list[dict[str, str]], hand: str) -> dict[str, str] | None:
    """The hand's own operator row, if the wave ran one."""
    for row in rows:
        if row["kind"] == "operator" and row["hand"] == f"operator/{hand}":
            return row
    return None


def camp_sums(
    rows: list[dict[str, str]], hands: list[dict[str, str]]
) -> tuple[list[int], list[int]]:
    """The camps' full costs: each hand plus its own operator, in+out
    tokens and work time — the ownership the kinds are compared by."""
    in_sums: list[int] = []
    work_sums: list[int] = []
    for hand in hands:
        camp_in = int(hand["in_out"])
        camp_work = int(hand["work_ms"])
        if op := operator_row(rows, hand["hand"]):
            camp_in += int(op["in_out"])
            camp_work += int(op["work_ms"])
        in_sums.append(camp_in)
        work_sums.append(camp_work)
    return in_sums, work_sums


def ownership_table(rows: list[dict[str, str]]) -> list[str]:
    """Every hand with its own operator and the sum — the full cost
    of ownership, side by side."""
    header = (
        "| hand | kind | in+out: hand | operator | hand+operator "
        "| work: hand | operator | hand+operator |"
    )
    rule = "|---|---|---|---|---|---|---|---|"
    body = []
    for row in [r for r in rows if r["kind"] in ("machine", "solo")]:
        op = operator_row(rows, row["hand"])
        if op is None:
            continue
        hand_in, op_in = int(row["in_out"]), int(op["in_out"])
        hand_w, op_w = int(row["work_ms"]), int(op["work_ms"])
        body.append(
            f"| {row['hand']} | {row['kind']} | {millions(hand_in)} "
            f"| {millions(op_in)} | {millions(hand_in + op_in)} "
            f"| {duration(hand_w)} | {duration(op_w)} "
            f"| {duration(hand_w + op_w)} |"
        )
    return [header, rule, *body]


def etalon_page(meaning: str) -> list[str]:
    """One meaning's etalon page, verbatim: the yardstick every judge
    of the meaning read, byte-identically. A meaning without its page
    is a stop — the instrument is written before any run, never
    improvised."""
    page = CAMPAIGN / "etalon" / f"{meaning}.md"
    if not page.is_file():
        print(f"etalon page missing: {page}", file=sys.stderr)
        raise SystemExit(1)
    return [f"### etalon/{meaning}", "", page.read_text(), ""]


def etalon_pages(manifest: dict[str, Any]) -> list[str]:
    out: list[str] = []
    for meaning in sorted({str(h["brief"]).split("/")[0] for h in manifest["hands"]}):
        out.extend(etalon_page(meaning))
    return out


def document(
    rows: list[dict[str, str]],
    manifest: dict[str, Any],
    wave_dir: Path,
    title: str,
    etalons: list[str],
) -> list[str]:
    """One metrics document over the given rows: the wave's whole, or
    one task's verified record for the repository."""
    ownership: list[str] = []
    if any(r["kind"] == "operator" for r in rows):
        ownership = [
            "## Ownership — each hand, its operator, the sum",
            "",
            *ownership_table(rows),
            "",
        ]
    lines = [
        f"# METRICS — {title}",
        "",
        (
            f"briefs: {', '.join(sorted({h['brief'] for h in manifest['hands']}))}; "
            f"model: {manifest['model']}; "
            f"hands: {len([r for r in rows if r['kind'] in ('machine', 'solo')])}; "
            f"judges: {len([r for r in rows if r['kind'] == 'judge'])}; "
            f"operators: {len([r for r in rows if r['kind'] == 'operator'])}"
        ),
        "",
        "## Usage",
        "",
        *usage_table(rows),
        "",
        "## Medians",
        "",
        *median_table(rows),
        "",
        *ownership,
        "## Etalons — the meanings' yardsticks every judge read",
        "",
        *etalons,
        "## Judge pages",
        "",
    ]
    for row in [r for r in rows if r["kind"] in ("machine", "solo")]:
        page = wave_dir / "judged" / f"{row['hand']}.md"
        lines.append(f"### {row['hand']}")
        lines.append("")
        lines.append(page.read_text() if page.is_file() else "JUDGE PAGE MISSING")
        lines.append("")
    return lines


def main(argv: list[str]) -> int:
    args = parse_args(argv)
    wave_dir = Path(args.wave)
    manifest: dict[str, Any] = json.loads(
        (wave_dir / "manifest.json").read_text()
    )
    with (wave_dir / "usage.tsv").open() as stream:
        rows = list(csv.DictReader(stream, delimiter="\t"))
    if not rows:
        print("usage.tsv is empty — collect first", file=sys.stderr)
        return 1
    validate(rows, manifest, wave_dir)

    (wave_dir / "METRICS.md").write_text(
        "\n".join(document(rows, manifest, wave_dir, wave_dir.name, etalon_pages(manifest)))
    )
    print(f"metrics: {wave_dir / 'METRICS.md'}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))

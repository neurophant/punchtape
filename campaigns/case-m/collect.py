"""Collect one wave's usage table from the platform database.

Every agent of the wave — each hand, each judge and each operator —
is found by its spawn bytes: the agent's prompt file is the
session's first user message, so a session belongs to the agent
whose full spawn text it carries; no time-window guessing, and no
visitor of a cell ever mixes into another's row. Tokens, requests,
turns and wall time come from the database's own per-request rows:
what a kill or an interruption cut off is still counted, because the
platform writes each request the moment it completes — the measured
tail cannot be lost the way a side ledger lost it. One wave, one
table: usage.tsv beside the wave's manifest, judges and operators as
their own kinds.
"""

from __future__ import annotations

import argparse
import json
import sqlite3
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Any

DEFAULT_DB = Path.home() / ".zcode" / "cli" / "db" / "db.sqlite"

COLUMNS = [
    "hand",
    "kind",
    "sessions",
    "turns",
    "requests",
    "input",
    "output",
    "cache_read",
    "in_out",
    "total",
    "work_ms",
    "span_ms",
    "tail",
]


@dataclass
class HandUsage:
    """One hand's measured usage: straight sums over its requests."""

    hand: str
    kind: str
    sessions: int
    turns: int
    requests: int
    input_tokens: int
    output_tokens: int
    cache_read: int
    work_ms: int
    span_ms: int
    tail: str

    def row(self) -> list[str]:
        return [
            self.hand,
            self.kind,
            str(self.sessions),
            str(self.turns),
            str(self.requests),
            str(self.input_tokens),
            str(self.output_tokens),
            str(self.cache_read),
            str(self.input_tokens + self.output_tokens),
            str(self.input_tokens + self.output_tokens + self.cache_read),
            str(self.work_ms),
            str(self.span_ms),
            self.tail,
        ]


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--wave", required=True, help="the wave root (holds manifest.json)")
    parser.add_argument("--db", default=str(DEFAULT_DB), help="the platform database file")
    return parser.parse_args(argv)


def first_user_text(con: sqlite3.Connection, sid: str) -> str:
    """The session's first user message as plain text, '' if none."""
    messages = con.execute(
        "select id, data from message where session_id=? order by time_created limit 10",
        (sid,),
    ).fetchall()
    spawn_id = next(
        (str(m[0]) for m in messages if json.loads(m[1]).get("role") == "user"),
        None,
    )
    if spawn_id is None:
        return ""
    parts = con.execute(
        "select data from part where message_id=? order by sequence", (spawn_id,)
    ).fetchall()
    texts = [
        str(json.loads(p[0]).get("text", ""))
        for p in parts
        if json.loads(p[0]).get("type") == "text"
    ]
    return "".join(texts)


def hand_sessions(con: sqlite3.Connection, hand: dict[str, Any]) -> list[str]:
    """The sessions spawned by exactly this hand's bytes.

    The spawn bytes — the hand's prompt file — are the session's
    first user message; tool calls, later replies and other visitors
    of the cell (judges, audits) may quote the path too and prove
    nothing. A session counts when its first user message carries the
    hand's full spawn text.
    """
    input_path = hand["input"]
    prompt_file = Path(hand["prompt"])
    if not prompt_file.is_file():
        print(f"hand prompt file missing: {prompt_file}", file=sys.stderr)
        raise SystemExit(1)
    prompt_text = prompt_file.read_text().strip()
    escaped = (
        input_path.replace("\\", "\\\\").replace("%", "\\%").replace("_", "\\_")
    )
    candidates = [
        str(r[0])
        for r in con.execute(
            "select distinct session_id from part "
            "where data like ? escape '\\' order by session_id",
            (f"%{escaped}%",),
        )
    ]
    return [sid for sid in candidates if prompt_text in first_user_text(con, sid)]


def hand_usage(
    con: sqlite3.Connection, hand: dict[str, Any], sessions: list[str]
) -> HandUsage:
    if not sessions:
        return HandUsage(hand["hand"], hand["kind"], 0, 0, 0, 0, 0, 0, 0, 0, "no-sessions")
    marks = ",".join("?" for _ in sessions)
    requests = con.execute(
        f"select count(*), coalesce(sum(input_tokens),0), coalesce(sum(output_tokens),0), "
        "coalesce(sum(cache_read_input_tokens),0), coalesce(sum(duration_ms),0), "
        "coalesce(min(started_at),0), coalesce(max(completed_at),0), "
        "sum(case when status != 'completed' then 1 else 0 end) "
        f"from model_usage where session_id in ({marks})",
        sessions,
    ).fetchone()
    turns = con.execute(
        f"select count(*), coalesce(sum(case when status != 'completed' then 1 else 0 end),0) "
        f"from turn_usage where session_id in ({marks})",
        sessions,
    ).fetchone()
    tail = "clean" if turns[1] == 0 and requests[7] == 0 else "interrupted"
    span = max(int(requests[6]) - int(requests[5]), 0)
    return HandUsage(
        hand=hand["hand"],
        kind=hand["kind"],
        sessions=len(sessions),
        turns=int(turns[0]),
        requests=int(requests[0]),
        input_tokens=int(requests[1]),
        output_tokens=int(requests[2]),
        cache_read=int(requests[3]),
        work_ms=int(requests[4]),
        span_ms=span,
        tail=tail,
    )


def judged_hands(wave_dir: Path) -> list[dict[str, Any]]:
    """The wave's judges, one per rendered prompt in judged/. The
    stem's hand name (a .killed render judges the same hand as its
    replacement) names the verdict page, whose path rides the prompt
    and prefilters the search; the full spawn text separates the
    renders."""
    out: list[dict[str, Any]] = []
    for prompt_file in sorted((wave_dir / "judged").glob("*.prompt")):
        base = prompt_file.name.split(".")[0]
        out.append(
            {
                "hand": f"judge/{prompt_file.stem}",
                "kind": "judge",
                "input": str(wave_dir / "judged" / f"{base}.md"),
                "prompt": str(prompt_file),
            }
        )
    return out


def operator_hands(wave_dir: Path) -> list[dict[str, Any]]:
    """The wave's operators, one per spawn prompt in operators/. The
    stem names the hand the operator serves through every round; the
    operator's log path rides the prompt and prefilters the search;
    the full spawn text separates the agents."""
    out: list[dict[str, Any]] = []
    for prompt_file in sorted((wave_dir / "operators").glob("*.prompt")):
        stem = prompt_file.name[: -len(".prompt")]
        out.append(
            {
                "hand": f"operator/{stem}",
                "kind": "operator",
                "input": str(wave_dir / "operators" / f"{stem}-log.md"),
                "prompt": str(prompt_file),
            }
        )
    return out


def main(argv: list[str]) -> int:
    args = parse_args(argv)
    wave_dir = Path(args.wave)
    manifest_file = wave_dir / "manifest.json"
    if not manifest_file.is_file():
        print(f"no manifest in wave root: {wave_dir}", file=sys.stderr)
        return 1
    hands: list[dict[str, Any]] = json.loads(manifest_file.read_text())["hands"]
    agents = hands + judged_hands(wave_dir) + operator_hands(wave_dir)
    con = sqlite3.connect(f"file:{args.db}?mode=ro", uri=True)

    session_sets: dict[str, set[str]] = {}
    for agent in agents:
        session_sets[agent["hand"]] = set(hand_sessions(con, agent))
    owned = [sid for ids in session_sets.values() for sid in ids]
    if len(owned) != len(set(owned)):
        print("a session carries two agents' spawn bytes — cannot attribute usage", file=sys.stderr)
        return 1

    lines = ["\t".join(COLUMNS)]
    for agent in agents:
        usage = hand_usage(con, agent, sorted(session_sets[agent["hand"]]))
        lines.append("\t".join(usage.row()))
    (wave_dir / "usage.tsv").write_text("\n".join(lines) + "\n")
    print(f"usage: {wave_dir / 'usage.tsv'}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))

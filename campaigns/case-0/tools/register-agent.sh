#!/bin/bash
# register-agent.sh — the AGENT REGISTRY of the case-0 wave (the
# 0.2.0 census discipline: every agent of the run is registered, a
# role outside the closed set refuses): appends {cell: [{uuid,
# role}]} into <run-root>/driver-agents.json (the shape
# usage-ledger.py reads). Roles: hand | operator | judge | measurer.
# A duplicate uuid, an unknown role or a second hand on a cell all
# refuse — the census is closed by construction.
#
# Usage: register-agent.sh <run-root> <cell> <role> <full-uuid>
#          [--session <platform-session-id>]
#   --session pins the agent's platform session (the micro-agents
#   spawn from the driver's sterile cwd — the directory rule cannot
#   find them; the pin rides usage-ledger.py --session later)
set -euo pipefail
RR="${1:?run root}"; CELL="${2:?cell}"; ROLE="${3:?role}"; UUID="${4:?full uuid}"; shift 4
SESS=""
while [ $# -gt 0 ]; do case "$1" in --session) SESS="$2"; shift 2;; *) echo "unknown: $1" >&2; exit 2;; esac; done
RR="$(readlink -f "$RR")"
case "$ROLE" in hand|operator|judge|measurer) ;; *) echo "role must be hand|operator|judge|measurer" >&2; exit 2 ;; esac
# the FULL agent id: the platform's agent_<uuid> form, or the bare
# 36-char uuid (a bash case glob cannot express {36} — the bare form
# is validated with a regex, never a quantified glob)
case "$UUID" in
  agent_[A-Za-z0-9_-]*) ;;
  *)
    if ! printf '%s' "$UUID" | grep -Eq '^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$'; then
      echo "uuid must be the FULL agent id" >&2; exit 2
    fi ;;
esac

python3 - "$RR" "$CELL" "$ROLE" "$UUID" "$SESS" <<'PY'
import datetime, json, os, re, sys
root, cell, role, uuid, sess = sys.argv[1:6]
# the ledger keys agents by the FIRST 8 HEX CHARS of the uuid (the
# id8 of usage-ledger.py / collect-metrics.sh) — the pin line must
# carry that same key, never uuid[:8] ("agent_xx" would never match)
m8 = re.search(r"[0-9a-fA-F]{8}", uuid)
id8 = m8.group(0).lower() if m8 else None
reg_path = f"{root}/driver-agents.json"
reg = json.load(open(reg_path, encoding="utf-8")) if os.path.isfile(reg_path) else {}
for c, lst in reg.items():
    for e in lst or []:
        if e.get("uuid") == uuid:
            print(f"REGISTER REFUSED: uuid already registered (cell {c})", file=sys.stderr)
            sys.exit(2)
if role == "hand" and any(e.get("role") == "hand" for e in reg.get(cell, [])):
    print(f"REGISTER REFUSED: cell {cell} already has its hand", file=sys.stderr)
    sys.exit(2)
entry = {"uuid": uuid, "role": role,
         "registered": datetime.datetime.now(datetime.timezone.utc)
                       .isoformat(timespec="seconds")}
if sess:
    entry["session"] = sess
    # the pin file usage-ledger.py consumes at close (--session lines)
    with open(f"{root}/session-pins.txt", "a", encoding="utf-8") as f:
        f.write(f"{id8}={sess}\n")
reg.setdefault(cell, []).append(entry)
json.dump(reg, open(reg_path, "w", encoding="utf-8"), indent=1, sort_keys=True)
rec = {"op": "register", "cell": cell, "role": role, "uuid8": id8,
       "ts": entry["registered"]}
with open(f"{root}/wave-journal.jsonl", "a", encoding="utf-8") as f:
    f.write(json.dumps(rec, ensure_ascii=False) + "\n")
print(f"registered: {cell}/{role} uuid8={id8}{' (session pinned)' if sess else ''}")
PY

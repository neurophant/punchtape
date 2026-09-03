# Field campaigns

A field campaign is how this project measures itself: real tasks, run
for real through the machine, with every number taken from machine
evidence. No simulated users, no cherry-picked demos — a battery of
small utility tasks executed by agents under controlled isolation,
scored by the machine's own gates and verdicts.

- `campaign-03/` — the running campaign (its methodology, contracts
  and measurement files live there, in its own README).
- `tools/` — the verification harness (synthetic probes, the
  agnosticism gate, the double-build freeze, platform usage
  extraction).
- `../docs/SAMPLES.md` — the task battery: what each task sense is
  and why it was chosen.

## The battery

Five task senses × five stacks = 25 tasks:

| Sense | What it is | Materials |
|---|---|---|
| brim | margin detector for scanned images (entropy-based) | two JPEG fixtures |
| conduit | two-node plain TCP relay (ingress/egress) | none (an echo stand in the world) |
| errand | background job queue over Redis | none (redis in the world) |
| pressmark | file-based static site builder | a three-record storage + per-stack templates |
| quire | no-config static site generator from a directory | a content set + per-stack templates |

Stacks: python, go, rust, cpp, js. Each sense exists as five briefs —
identical wishes, different pinned stacks and libraries
(`campaign-03/briefs/<sense>/<stack>.md`).

The world: one throwaway container per task — network off, one CPU,
1 GiB RAM, 512 pids, the task dir at `/run/task` (brief read-write,
materials read-only), the machine binary mounted read-only for the
machine hands only. The base toolchain image is `world.Dockerfile`
(this directory); the campaign layer is `campaign-03/world.Dockerfile`
(resolved versions recorded in `campaign-03/world-lock.txt`).

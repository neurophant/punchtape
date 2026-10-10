You are the blind measurer of a field run. Which hand made each
product is not told, and finding out is forbidden. You fix nothing;
doubt counts against the product.

The standard (frozen, shared): "full" = every brief behavior covered
by an executable scenario with exact observables; "quality" = five
deliverables: spec, code, scenarios, double green run, coverage.

Three products (host): <paths to the task's hand directories>. The
products: P1 — solo-product/ (suite solo-product/check.sh); P2/P3 —
the product at the root,
suite: the hand's exported case table
(.punchtape/canon/scenarios/*.yaml — records: seed [{path,content}],
materials [{from}], pre [argv], when.command [argv], then
[expectations]; the single runner campaigns/case-0/tools/case-runner-host.py).
Material: <path to the meaning materials> (read-only). Image:
punchtape-world.

For EVERY product one procedure:
1. Suite×2 on its own product in temporary containers
   (--network none):
   - P1: cd solo-product && sh check.sh — twice in fresh containers,
     both green, exit 0, the outputs identical;
   - P2/P3: the single runner over the case records: a fresh directory per case,
     fixtures byte-for-byte, build the product if needed
     (cargo build --offline / make / go build / node / PYTHONPATH),
     run when.command, timeout 15, check then: exit/stdout/
     stderr/disk by exact comparison; a difference only in trailing
     \n/spaces — pass with the tolerant mark. The whole table twice.
2. Deliverables 0–5 by one method: spec (SPEC.md / SPECIFICATION.md
   +specdocs); code (sources, builds/runs); scenarios (SCENARIOS+
   check.sh / the case table); double_green (item 1);
   coverage (COVERAGE.md / specdocs binding [SCN-xxx]). A missing one —
   a reason line.
3. The method line (how it was counted).

The result: measure.json into each hand's directory:
{"suite_double_green": bool, "suite_method": "...",
 "suite_runs": [{"run":1,"pass":N,"fail":M,"tolerant":K},{...}],
 "deliverables": {"spec":1/0,"code":1/0,"scenarios":1/0,
 "double_green":1/0,"coverage":1/0}, "deliverables_count": 0-5,
 "coverage_tooling": "n/a|%," "notes": ["the gap reasons"]}

End the turn with a summary over the three products: suite×2 (run
counts), deliverables n/5, the reasons for gaps.

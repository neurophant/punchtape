# METRICS — the case-0 battery: ALL tasks × ALL hands (five waves, one binary)

The 0.2.0-th3 protocol on the current host world; TWO TOKEN BASES (the
K3 law): `in+out` (the 0.2.0 working base) and `total`
(= in+out+cache_read) in every table; `$` is cache-aware. One series
binary f0a08dd3 (FREEZE-S, both campaigns' freezes), one set of
briefs, one accounting platform (db), TWO hands: solo and machine-top
(there is no third/flash hand). All numbers come ONLY from the five
frozen wave reports (docs/case0-*-session-report.md; each wave's
verification GREEN); not one new measurement here — reading, gluing
and summing the Σ-rows of those reports.

The battery's waves: conduit (conduit3, 20261010T142320Z) → brim
(brim2, 20261010T155105Z) → errand (20261010T174342Z) → pressmark
(20261010T202111Z) → quire (20261010T220550Z): 5 tasks × 5 stacks ×
2 hands = 50 cells. The historical September 0.2.0 runs (the source
bytes) and the aborted waves (conduit-1, conduit2, brim-1) are NOT
included — closed history. The full acceptance texts are in the wave
reports and the cells' acceptance.json; in the cells below the
verdicts and coverage method lines are shortened to the first line,
the numbers untouched.

## 1. The battery summary by waves (from the quire report's "before/after" tables)

| metric | conduit | brim | errand | pressmark | quire |
|---|---|---|---|---|---|
| cells delivered and accepted | 10/10 | 10/10 | 10/10 | 10/10 | 10/10 |
| machine hands' door refusals | 0 (of 24) | 0 (of 26) | 0 (of 24) | 0 (of 30) | 1 (of 33) |
| VERDICT READY on the first acceptance | 5/5 | 5/5 | 5/5 | 5/5 | 5/5 |
| solo-acceptance rejection cycles | 1 | 0 | 2 | 2 (rust, cycle 3) | 1 (js, accepted in cycle 2) |
| deliverables: solo hands | 5/5 ×4, 4/5 js | 5/5 ×5 | 5/5 ×5 | 5/5 ×5 | 5/5 ×5 |
| deliverables: machine hands | 5/5 ×5 | 5/5 ×4, 3/5 go | 5/5 ×5 | 5/5 ×4, 3/5 go | 5/5 ×4, 3/5 cpp |
| G1 won by the machine | 5/5 | 5/5 | 5/5 | 4/5 (go exchange) | 4/5 (cpp — solo) |
| Σ-camp total (solo/top ×) | 1.52 | 1.95 | 1.83 | 1.99 | 1.16 |
| Σ-hands total (without operators) | 1.18 | 1.56 | 1.35 | 1.45 | 0.87 |
| Σ-hands time | 1.31 | 1.41 | 1.33 | 1.30 | 0.99 |
| Σtotal top hands | 29.17M | 22.38M | 34.49M | 48.90M | 48.80M |
| Σtotal solo hands | 34.45M | 34.84M | 46.58M | 70.93M | 42.45M |
| operators top/solo, total | 11.32× | 11.82× | 26.31× | 46.01× | 10.57× |
| judge pass | 21/96 | 46/89 | 46/57 | 43/129 | 42/103 |
| verify-metrics | GREEN 307/0 | GREEN 306/0 | GREEN 308/0 | GREEN 308/0 | GREEN 307/0 |

## 2. Hands — ALL tasks × ALL stacks × both hands (50 rows)

| hand | time, s | tokens in/out | hand $ | submissions ±/− | verb-wall, s | verdict | cases | suite×2 | deliv. | prose | coverage | LOC | operator $ / decisions / wait, s |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| conduit-cpp-solo | 1258.4 | 4062962/76891 | 1.61 | — | — | ACCEPT (cycle 2) — … | own suite | green×2 | 5/5 | — | n/a | 1561 | 0.74 / 2 / 463 |
| conduit-cpp-top | 1014.7 | 3988211/68092 | 1.45 | 6/0 | 10.4 | VERDICT: READY | 29/29 | green×2 | 5/5 | 3/3 | 100% | 1010 | 0.03 / 1 / 288 |
| conduit-go-solo | 1482.6 | 4695694/88177 | 1.73 | — | — | ACCEPT — the operator himself … | own suite | green×2 | 5/5 | — | 96.2% | 2515 | 0.40 / 1 / 270 |
| conduit-go-top | 826.0 | 2130850/57702 | 0.90 | 5/0 | 6.6 | VERDICT: READY | 16/16 | green×2 | 5/5 | 3/3 | n/a | 565 | 0.04 / 1 / 66 |
| conduit-js-solo | 1067.7 | 2808340/66679 | 1.12 | — | — | ACCEPT — the operator himself … | own suite | not ×2 | 4/5 | — | 100.0% | 1362 | 0.41 / 1 / 380 |
| conduit-js-top | 872.3 | 2439774/47737 | 0.99 | 4/0 | 37.8 | VERDICT: READY | 20/20 | green×2 | 5/5 | 3/3 | n/a | 784 | 0.05 / 1 / 322 |
| conduit-python-solo | 1189.5 | 2979984/67859 | 1.16 | — | — | ACCEPT — the operator himself … | own suite | green×2 | 5/5 | — | 97.3% (146/150 statement lines; its own instrument… | 1343 | 0.33 / 1 / 347 |
| conduit-python-top | 1063.0 | 3604947/76355 | 1.39 | 5/0 | 17.4 | VERDICT: READY | 16/16 | green×2 | 5/5 | 3/3 | n/a (machine-verdict-coverage: the frozen verdict .punc… | 501 | 0.03 / 1 / 45 |
| conduit-rust-solo | 933.7 | 2772631/57532 | 1.07 | — | — | ACCEPT — the operator himself … | own suite | green×2 | 5/5 | — | n/a (measurer-eye linkage: COVERAGE.md maps all 20 brief … | 1412 | 0.43 / 1 / 590 |
| conduit-rust-top | 746.9 | 2512815/47243 | 0.94 | 4/0 | 6.6 | VERDICT: READY | 15/15 | green×2 | 5/5 | 3/3 | machine-verdict-coverage (frozen verdict v1 READY digest … | 705 | 0.04 / 1 / 77 |
| brim-cpp-solo | 1142.4 | 3699480/73811 | 1.40 | — | — | ACCEPT. Checked myself, … | own suite | green×2 | 5/5 | — | 100% (measurer-eye over COVERAGE.md: 19 brief-behavior ro… | 720 | 0.22 / 1 / 195 |
| brim-cpp-top | 877.5 | 2471793/56032 | 0.98 | 5/0 | 1.1 | VERDICT: READY | 18/18 | green×2 | 5/5 | 3/3 | 100% (machine-verdict-coverage: frozen verdict v1 READY, … | 247 | 0.03 / 1 / 50 |
| brim-go-solo | 949.0 | 3207947/67604 | 1.23 | — | — | ACCEPT. All five … | own suite | green×2 | 5/5 | — | 99.3% | 1141 | 0.42 / 1 / 344 |
| brim-go-top | 726.8 | 1805325/47173 | 0.77 | 5/0 | 0.9 | VERDICT: READY | 21/21 | not ×2 | 3/5 | 3/3 | n/a | 354 | 0.03 / 1 / 50 |
| brim-js-solo | 1228.5 | 3909145/81224 | 1.49 | — | — | ACCEPT. All five … | own suite | green×2 | 5/5 | — | 100% lines/branches/functions (node --test --experimental… | 1092 | 0.45 / 1 / 413 |
| brim-js-top | 496.4 | 1346063/33188 | 0.58 | 5/0 | 2.8 | VERDICT: READY | 16/16 | green×2 | 5/5 | 3/3 | machine-verdict-coverage: the frozen verdict v1 READY (… | 267 | 0.02 / 1 / 55 |
| brim-python-solo | 727.9 | 1529728/52577 | 0.71 | — | — | ACCEPT. All five … | own suite | green×2 | 5/5 | — | 100% — COVERAGE brim: executed 106/106 executable stateme… | 569 | 0.35 / 1 / 279 |
| brim-python-top | 614.9 | 1837775/37663 | 0.71 | 6/0 | 4.4 | VERDICT: READY | 16/16 | green×2 | 5/5 | 3/3 | n/a (machine-verdict-coverage: the frozen verdict .punc… | 148 | 0.03 / 1 / 134 |
| brim-rust-solo | 1077.1 | 5131623/67688 | 1.74 | — | — | ACCEPT. All five … | own suite | green×2 | 5/5 | — | 99,6% — runs/coverage.txt (llvm line coverage, UNION of three… | 1095 | 0.60 / 1 / 558 |
| brim-rust-top | 911.7 | 3799628/59674 | 1.34 | 5/0 | 0.5 | VERDICT: READY | 18/18 | green×2 | 5/5 | 3/3 | n/a (machine-verdict-coverage: the frozen verdict .punc… | 362 | 0.05 / 1 / 74 |
| errand-cpp-solo | 1197.2 | 3273821/87883 | 1.50 | — | — | ACCEPT — the product … | own suite | green×2 | 5/5 | — | 88.0% | 2071 | 0.57 / 1 / 942 |
| errand-cpp-top | 817.3 | 2677500/51596 | 1.01 | 5/0 | 2.8 | VERDICT: READY | 12/12 | green×2 | 5/5 | 3/3 | n/a | 712 | 0.03 / 1 / 281 |
| errand-go-solo | 1017.7 | 3029684/66591 | 1.17 | — | — | ACCEPT — the cycle 2 product … | own suite | green×2 | 5/5 | — | 94.3% | 1381 | 0.99 / 2 / 819 |
| errand-go-top | 874.7 | 3216086/46394 | 1.12 | 4/0 | 16.1 | VERDICT: READY | 15/15 | green×2 | 5/5 | 3/3 | 92.7% | 753 | 0.03 / 1 / 485 |
| errand-js-solo | 1288.5 | 4967211/81703 | 1.78 | — | — | ACCEPT. All three … | own suite | green×2 | 5/5 | — | 100% (index.js, errors.js: line/branch/funcs, node --expe… | 1249 | 1.00 / 2 / 1235 |
| errand-js-top | 732.3 | 2649681/47133 | 0.98 | 5/0 | 6.9 | VERDICT: READY | 10/10 | green×2 | 5/5 | 3/3 | n/a | 492 | 0.03 / 1 / 403 |
| errand-python-solo | 1287.6 | 4934128/80992 | 1.76 | — | — | ACCEPT. The product is accepted … | own suite | green×2 | 5/5 | — | 100% | 830 | 0.36 / 1 / 369 |
| errand-python-top | 1003.7 | 4068361/61505 | 1.43 | 4/0 | 7.8 | VERDICT: READY | 11/11 | green×2 | 5/5 | 4/4 | 100% | 493 | 0.03 / 1 / 82 |
| errand-rust-solo | 1122.0 | 7206921/62930 | 2.26 | — | — | ACCEPT. The product … | own suite | green×2 | 5/5 | — | 96.42% regions / 98.57% lines / 92.59% functions (src/lib… | 763 | 0.47 / 1 / 467 |
| errand-rust-top | 1013.6 | 4700232/62601 | 1.60 | 6/0 | 2.2 | VERDICT: READY | 9/9 | green×2 | 5/5 | 3/3 | machine-verdict-coverage | 568 | 0.03 / 1 / 154 |
| pressmark-cpp-solo | 1072.5 | 3614921/71375 | 1.36 | — | — | ACCEPTANCE NOTE … | own suite | green×2 | 5/5 | — | 100% | 1652 | 0.70 / 1 / 582 |
| pressmark-cpp-top | 1182.0 | 4935941/77687 | 1.75 | 6/0 | 18.1 | VERDICT: READY | 24/24 | green×2 | 5/5 | 3/3 | 100% | 616 | 0.03 / 1 / 150 |
| pressmark-go-solo | 989.6 | 2805573/83804 | 1.22 | — | — | ACCEPTANCE NOTE — … | own suite | green×2 | 5/5 | — | 92.4% | 2089 | 0.78 / 1 / 600 |
| pressmark-go-top | 1405.7 | 5678684/91071 | 2.01 | 5/0 | 1.8 | VERDICT: READY | 29/29 | not ×2 | 3/5 | 3/3 | n/a | 468 | 0.02 / 1 / 65 |
| pressmark-js-solo | 1029.0 | 4330271/78701 | 1.58 | — | — | ACCEPTED — pressmark … | own suite | green×2 | 5/5 | — | 100% (line coverage pressmark.mjs 360/360 = 100.0%, V8 bl… | 1147 | 0.66 / 1 / 614 |
| pressmark-js-top | 942.0 | 3475837/74750 | 1.35 | 8/0 | 3.5 | VERDICT: READY | 22/22 | green×2 | 5/5 | 3/3 | n/a | 267 | 0.03 / 1 / 85 |
| pressmark-python-solo | 1167.1 | 6192826/82283 | 2.09 | — | — | ACCEPTANCE NOTE — … | own suite | green×2 | 5/5 | — | 100% (line coverage 277/277 executable product lines = 1… | 1380 | 0.86 / 1 / 566 |
| pressmark-python-top | 1108.5 | 5607198/81906 | 1.97 | 6/0 | 15.6 | VERDICT: READY | 26/26 | green×2 | 5/5 | 3/3 | n/a | 212270 | 0.02 / 1 / 93 |
| pressmark-rust-solo | 3267.4 | 18589169/177124 | 5.88 | — | — | ACCEPTANCE NOTE — … | own suite | green×2 | 5/5 | — | 88.49% | 1733 | 2.24 / 3 / 637 |
| pressmark-rust-top | 1131.5 | 4841226/80030 | 1.74 | 5/0 | 1.4 | VERDICT: READY | 28/28 | green×2 | 5/5 | 3/3 | n/a | 500 | 0.02 / 1 / 103 |
| quire-cpp-solo | 1421.1 | 5781100/72830 | 1.94 | — | — | ACCEPT. The acceptance … | own suite | green×2 | 5/5 | — | 92.35% lines of 327, 91.71% branches of 627 (gcov -b via … | 1228 | 0.50 / 1 / 373 |
| quire-cpp-top | 1756.2 | 8381144/96240 | 2.76 | 12/0 | 17.3 | VERDICT: READY | 19/19 | not ×2 | 3/5 | 3/3 | machine-verdict-coverage (frozen verdict v1 READY digest … | 1183 | 0.04 / 1 / 115 |
| quire-go-solo | 1076.0 | 4446805/87202 | 1.66 | — | — | # Acceptance — … | own suite | green×2 | 5/5 | — | 90.7% | 2051 | 0.42 / 1 / 518 |
| quire-go-top | 1256.7 | 4731607/72027 | 1.66 | 5/0 | 1.0 | VERDICT: READY | 21/21 | green×2 | 5/5 | 3/3 | n/a | 21563 | 0.10 / 1 / 189 |
| quire-js-solo | 968.6 | 3206205/81483 | 1.31 | — | — | OPERATOR DECISION — … | own suite | green×2 | 5/5 | — | n/a | 1065 | 1.20 / 2 / 717 |
| quire-js-top | 957.3 | 3538053/75569 | 1.37 | 5/0 | 5.1 | VERDICT: READY | 24/24 | green×2 | 5/5 | 3/3 | n/a | 168 | 0.10 / 1 / 187 |
| quire-python-solo | 894.0 | 4309124/66909 | 1.51 | — | — | Acceptance note … | own suite | green×2 | 5/5 | — | 100.0% (the suite's own instrument: sys.settrace line col… | 991 | 0.47 / 1 / 484 |
| quire-python-top | 797.0 | 3253148/62572 | 1.21 | 4/0 | 5.2 | VERDICT: READY | 21/21 | green×2 | 5/5 | 3/3 | machine-verdict-coverage (frozen verdict v1 READY digest … | 154 | 0.05 / 1 / 145 |
| quire-rust-solo | 1272.2 | 3539252/75774 | 1.37 | — | — | Acceptance note … | own suite | green×2 | 5/5 | — | n/a (measurer-eye linkage: COVERAGE.md maps every brief b… | 1249 | 0.46 / 1 / 418 |
| quire-rust-top | 926.1 | 4575199/67847 | 1.60 | 6/1 | 19.1 | VERDICT: READY | 18/18 | green×2 | 5/5 | 3/3 | machine-verdict-coverage (frozen verdict v1 READY digest … | 304 | 0.07 / 1 / 186 |

Hand Σ by waves (the numbers of each wave's §6 camp Σ-rows):

- conduit: solo — 5931.9 s / 17676749 in+out / 34453965 total; top — 4522.9 s / 14973726 in+out / 29165342 total; Σ-ratios time / in+out / total = 1.31 / 1.18 / 1.18
- brim: solo — 5124.9 s / 17820827 in+out / 34843099 total; top — 3627.3 s / 11494314 in+out / 22378794 total; Σ-ratios time / in+out / total = 1.41 / 1.55 / 1.56
- errand: solo — 5913.0 s / 23791864 in+out / 46577848 total; top — 4441.6 s / 17581089 in+out / 34492513 total; Σ-ratios time / in+out / total = 1.33 / 1.35 / 1.35
- pressmark: solo — 7525.6 s / 36026047 in+out / 70925567 total; top — 5769.7 s / 24944330 in+out / 48904522 total; Σ-ratios time / in+out / total = 1.30 / 1.44 / 1.45
- quire: solo — 5631.9 s / 21666684 in+out / 42445820 total; top — 5693.3 s / 24853406 in+out / 48803230 total; Σ-ratios time / in+out / total = 0.99 / 0.87 / 0.87

## 3. Operators (separate accounting — the hard 0.2.0 rule; all waves)

| wave | cell | hand | time, s | in+out | total | decisions |
|---|---|---|---|---|---|---|
| conduit | conduit-cpp-solo | solo | 612.9 | 1836167 | 3529543 | 2 |
| conduit | conduit-cpp-top | top | 224.7 | 83888 | 162864 | 1 |
| conduit | conduit-go-solo | solo | 226.2 | 1074456 | 2077720 | 1 |
| conduit | conduit-go-top | top | 36.5 | 107388 | 208636 | 1 |
| conduit | conduit-js-solo | solo | 324.7 | 1026033 | 1981233 | 1 |
| conduit | conduit-js-top | top | 162.7 | 130620 | 253948 | 1 |
| conduit | conduit-python-solo | solo | 278.8 | 823807 | 1586943 | 1 |
| conduit | conduit-python-top | top | 22.4 | 83592 | 162824 | 1 |
| conduit | conduit-rust-solo | solo | 342.8 | 1098762 | 2123082 | 1 |
| conduit | conduit-rust-top | top | 45.2 | 108144 | 209712 | 1 |
| **conduit** | **Σ operators solo** | solo | 1785.4 | 5859225 | 11298521 | 6 |
| **conduit** | **Σ operators top (machine)** | top | 491.5 | 513632 | 997984 | 5 |
| brim | brim-cpp-solo | solo | 161.6 | 556706 | 1072546 | 1 |
| brim | brim-cpp-top | top | 30.4 | 86898 | 168818 | 1 |
| brim | brim-go-solo | solo | 277.8 | 1129601 | 2195585 | 1 |
| brim | brim-go-top | top | 25.6 | 86546 | 168274 | 1 |
| brim | brim-js-solo | solo | 356.8 | 1190087 | 2307975 | 1 |
| brim | brim-js-top | top | 20.7 | 63238 | 122630 | 1 |
| brim | brim-python-solo | solo | 240.4 | 967094 | 1881718 | 1 |
| brim | brim-python-top | top | 28.1 | 85995 | 167275 | 1 |
| brim | brim-rust-solo | solo | 445.3 | 1616098 | 3142434 | 1 |
| brim | brim-rust-top | top | 44.4 | 138846 | 269662 | 1 |
| **brim** | **Σ operators solo** | solo | 1481.9 | 5459586 | 10600258 | 5 |
| **brim** | **Σ operators top (machine)** | top | 149.2 | 461523 | 896659 | 5 |
| errand | errand-cpp-solo | solo | 319.4 | 1627333 | 3172933 | 1 |
| errand | errand-cpp-top | top | 32.0 | 66200 | 126040 | 1 |
| errand | errand-go-solo | solo | 597.4 | 2745012 | 5341044 | 2 |
| errand | errand-go-top | top | 39.8 | 66602 | 126250 | 1 |
| errand | errand-js-solo | solo | 702.0 | 2615785 | 5075433 | 2 |
| errand | errand-js-top | top | 32.7 | 65380 | 125092 | 1 |
| errand | errand-python-solo | solo | 299.0 | 911255 | 1757143 | 1 |
| errand | errand-python-top | top | 26.6 | 64823 | 124535 | 1 |
| errand | errand-rust-solo | solo | 374.6 | 1182532 | 2285252 | 1 |
| errand | errand-rust-top | top | 28.9 | 86732 | 168268 | 1 |
| **errand** | **Σ operators solo** | solo | 2292.4 | 9081917 | 17631805 | 7 |
| **errand** | **Σ operators top (machine)** | top | 160.0 | 349737 | 670185 | 5 |
| pressmark | pressmark-cpp-solo | solo | 411.2 | 1942473 | 3781833 | 1 |
| pressmark | pressmark-cpp-top | top | 20.0 | 62484 | 115284 | 1 |
| pressmark | pressmark-go-solo | solo | 446.5 | 2093080 | 4066520 | 1 |
| pressmark | pressmark-go-top | top | 17.4 | 61693 | 120445 | 1 |
| pressmark | pressmark-js-solo | solo | 428.6 | 1810489 | 3524345 | 1 |
| pressmark | pressmark-js-top | top | 21.7 | 63079 | 115623 | 1 |
| pressmark | pressmark-python-solo | solo | 464.3 | 2432013 | 4752333 | 1 |
| pressmark | pressmark-python-top | top | 18.4 | 63346 | 122802 | 1 |
| pressmark | pressmark-rust-solo | solo | 1422.2 | 5823352 | 11225720 | 3 |
| pressmark | pressmark-rust-top | top | 23.2 | 61596 | 120284 | 1 |
| **pressmark** | **Σ operators solo** | solo | 3172.8 | 14101407 | 27350751 | 7 |
| **pressmark** | **Σ operators top (machine)** | top | 100.7 | 312198 | 594438 | 5 |
| quire | quire-cpp-solo | solo | 272.2 | 1416552 | 2765928 | 1 |
| quire | quire-cpp-top | top | 38.5 | 94282 | 180298 | 1 |
| quire | quire-go-solo | solo | 410.9 | 1020733 | 1968061 | 1 |
| quire | quire-go-top | top | 89.8 | 241233 | 466321 | 1 |
| quire | quire-js-solo | solo | 960.6 | 3316237 | 6467341 | 2 |
| quire | quire-js-top | top | 80.1 | 185219 | 338307 | 1 |
| quire | quire-python-solo | solo | 361.6 | 1216897 | 2359425 | 1 |
| quire | quire-python-top | top | 45.1 | 119979 | 230187 | 1 |
| quire | quire-rust-solo | solo | 369.9 | 1237058 | 2403138 | 1 |
| quire | quire-rust-top | top | 50.1 | 154269 | 295069 | 1 |
| **quire** | **Σ operators solo** | solo | 2375.2 | 8207477 | 15963893 | 6 |
| **quire** | **Σ operators top (machine)** | top | 303.6 | 794982 | 1510182 | 5 |
| **Σ BATTERY** | **Σ hands solo** | solo | 30127.3 | 116982171 | 229246299 | — |
| **Σ BATTERY** | **Σ hands top** | top | 24054.8 | 93846865 | 183744401 | — |
| **Σ BATTERY** | **Σ operators solo** | solo | 11107.7 | 42709612 | 82845228 | — |
| **Σ BATTERY** | **Σ operators top** | top | 1205 | 2432072 | 4669448 | — |

## 4. Layers — judges and measurers (the methodology's overhead; all waves)

| wave | layer | time, s | in+out | total |
|---|---|---|---|---|
| conduit | judge/conduit-cpp | 326.5 | 902080 | 1726592 |
| conduit | judge/conduit-go | 432.0 | 1028117 | 1977365 |
| conduit | judge/conduit-js | 240.5 | 742393 | 1433657 |
| conduit | judge/conduit-python | 398.1 | 839640 | 1598232 |
| conduit | judge/conduit-rust | 330.5 | 590900 | 1129332 |
| conduit | measurer/conduit-cpp | 274.1 | 1223671 | 2384567 |
| conduit | measurer/conduit-go | 272.6 | 1366398 | 2660350 |
| conduit | measurer/conduit-js | 495.5 | 1776434 | 3468850 |
| conduit | measurer/conduit-python | 255.8 | 1019996 | 1921180 |
| conduit | measurer/conduit-rust | 429.6 | 1481739 | 2834251 |
| brim | judge/brim-cpp | 289.0 | 489499 | 930011 |
| brim | judge/brim-go | 388.5 | 1067985 | 2067921 |
| brim | judge/brim-js | 258.5 | 532454 | 1017510 |
| brim | judge/brim-python | 175.0 | 417948 | 804892 |
| brim | judge/brim-rust | 187.0 | 331915 | 630667 |
| brim | measurer/brim-cpp | 495.4 | 1926437 | 3694565 |
| brim | measurer/brim-go | 652.1 | 1918568 | 3723816 |
| brim | measurer/brim-js | 341.9 | 866269 | 1657245 |
| brim | measurer/brim-python | 324.1 | 1311613 | 2552637 |
| brim | measurer/brim-rust | 485.4 | 2015111 | 3937735 |
| errand | judge/errand-cpp | 676.2 | 2475346 | 4820882 |
| errand | judge/errand-go | 217.1 | 700851 | 1354803 |
| errand | judge/errand-js | 385.2 | 670606 | 1272782 |
| errand | judge/errand-python | 441.2 | 1218776 | 2354328 |
| errand | judge/errand-rust | 161.5 | 436322 | 832674 |
| errand | measurer/errand-cpp | 346.5 | 1462239 | 2838815 |
| errand | measurer/errand-go | 386.0 | 1585533 | 3051709 |
| errand | measurer/errand-js | 298.8 | 701878 | 1337078 |
| errand | measurer/errand-python | 247.8 | 1067201 | 2073153 |
| errand | measurer/errand-rust | 433.6 | 1913814 | 3738838 |
| pressmark | judge/pressmark-cpp | 353.3 | 1673804 | 3272524 |
| pressmark | judge/pressmark-go | 385.8 | 1424494 | 2762478 |
| pressmark | judge/pressmark-js | 249.1 | 979171 | 1896547 |
| pressmark | judge/pressmark-python | 897.0 | 2482046 | 4825790 |
| pressmark | judge/pressmark-rust | 250.5 | 705948 | 1352668 |
| pressmark | measurer/pressmark-cpp | 277.4 | 1548187 | 3016603 |
| pressmark | measurer/pressmark-go | 365.6 | 1621765 | 3156421 |
| pressmark | measurer/pressmark-js | 603.8 | 3887306 | 7643722 |
| pressmark | measurer/pressmark-python | 720.0 | 4312370 | 8471666 |
| pressmark | measurer/pressmark-rust | 520.4 | 2007710 | 3896670 |
| quire | judge/quire-cpp | 623.9 | 2279875 | 4451075 |
| quire | judge/quire-go | 536.1 | 842188 | 1601292 |
| quire | judge/quire-js | 553.7 | 1684868 | 3266436 |
| quire | judge/quire-python | 308.6 | 1139476 | 2226452 |
| quire | judge/quire-rust | 362.0 | 935069 | 1788189 |
| quire | measurer/quire-cpp | 419.8 | 1791244 | 3476748 |
| quire | measurer/quire-go | 492.4 | 1780517 | 3461093 |
| quire | measurer/quire-js | 588.2 | 4471406 | 8803374 |
| quire | measurer/quire-python | 323.3 | 2283402 | 4470666 |
| quire | measurer/quire-rust | 385.1 | 1806255 | 3532847 |

## 5. Deltas — what the machine buys (both bases, the §0.2b.3 ratios; all waves)

| wave | task | metric | solo | top | solo/top |
|---|---|---|---|---|---|
| conduit | conduit-cpp | time, s | 1258.4 | 1014.7 | 1.24 |
| conduit | conduit-cpp | in+out | 4139853 | 4056303 | 1.02 |
| conduit | conduit-cpp | total | 8018125 | 7948015 | 1.01 |
| conduit | conduit-go | time, s | 1482.6 | 826.0 | 1.79 |
| conduit | conduit-go | in+out | 4783871 | 2188552 | 2.19 |
| conduit | conduit-go | total | 9371071 | 4235144 | 2.21 |
| conduit | conduit-js | time, s | 1067.7 | 872.3 | 1.22 |
| conduit | conduit-js | in+out | 2875019 | 2487511 | 1.16 |
| conduit | conduit-js | total | 5594379 | 4795479 | 1.17 |
| conduit | conduit-python | time, s | 1189.5 | 1063.0 | 1.12 |
| conduit | conduit-python | in+out | 3047843 | 3681302 | 0.83 |
| conduit | conduit-python | total | 5948451 | 7185110 | 0.83 |
| conduit | conduit-rust | time, s | 933.7 | 746.9 | 1.25 |
| conduit | conduit-rust | in+out | 2830163 | 2560058 | 1.11 |
| conduit | conduit-rust | total | 5521939 | 5001594 | 1.10 |
| **conduit** | **Σ wave** | time, s | 5931.9 | 4522.9 | 1.31 |
| **conduit** | **Σ wave** | in+out | 17676749 | 14973726 | 1.18 |
| **conduit** | **Σ wave** | total | 34453965 | 29165342 | 1.18 |
| brim | brim-cpp | time, s | 1142.4 | 877.5 | 1.30 |
| brim | brim-cpp | in+out | 3773291 | 2527825 | 1.49 |
| brim | brim-cpp | total | 7371371 | 4916177 | 1.50 |
| brim | brim-go | time, s | 949.0 | 726.8 | 1.31 |
| brim | brim-go | in+out | 3275551 | 1852498 | 1.77 |
| brim | brim-go | total | 6395743 | 3575378 | 1.79 |
| brim | brim-js | time, s | 1228.5 | 496.4 | 2.47 |
| brim | brim-js | in+out | 3990369 | 1379251 | 2.89 |
| brim | brim-js | total | 7799649 | 2655731 | 2.94 |
| brim | brim-python | time, s | 727.9 | 614.9 | 1.18 |
| brim | brim-python | in+out | 1582305 | 1875438 | 0.84 |
| brim | brim-python | total | 3039137 | 3654830 | 0.83 |
| brim | brim-rust | time, s | 1077.1 | 911.7 | 1.18 |
| brim | brim-rust | in+out | 5199311 | 3859302 | 1.35 |
| brim | brim-rust | total | 10237199 | 7576678 | 1.35 |
| **brim** | **Σ wave** | time, s | 5124.9 | 3627.3 | 1.41 |
| **brim** | **Σ wave** | in+out | 17820827 | 11494314 | 1.55 |
| **brim** | **Σ wave** | total | 34843099 | 22378794 | 1.56 |
| errand | errand-cpp | time, s | 1197.2 | 817.3 | 1.46 |
| errand | errand-cpp | in+out | 3361704 | 2729096 | 1.23 |
| errand | errand-cpp | total | 6402216 | 5329608 | 1.20 |
| errand | errand-go | time, s | 1017.7 | 874.7 | 1.16 |
| errand | errand-go | in+out | 3096275 | 3262480 | 0.95 |
| errand | errand-go | total | 6046419 | 6409872 | 0.94 |
| errand | errand-js | time, s | 1288.5 | 732.3 | 1.76 |
| errand | errand-js | in+out | 5048914 | 2696814 | 1.87 |
| errand | errand-js | total | 9905938 | 5272430 | 1.88 |
| errand | errand-python | time, s | 1287.6 | 1003.7 | 1.28 |
| errand | errand-python | in+out | 5015120 | 4129866 | 1.21 |
| errand | errand-python | total | 9846544 | 8109322 | 1.21 |
| errand | errand-rust | time, s | 1122.0 | 1013.6 | 1.11 |
| errand | errand-rust | in+out | 7269851 | 4762833 | 1.53 |
| errand | errand-rust | total | 14376731 | 9371281 | 1.53 |
| **errand** | **Σ wave** | time, s | 5913.0 | 4441.6 | 1.33 |
| **errand** | **Σ wave** | in+out | 23791864 | 17581089 | 1.35 |
| **errand** | **Σ wave** | total | 46577848 | 34492513 | 1.35 |
| pressmark | pressmark-cpp | time, s | 1072.5 | 1182.0 | 0.91 |
| pressmark | pressmark-cpp | in+out | 3686296 | 5013628 | 0.74 |
| pressmark | pressmark-cpp | total | 7204696 | 9839804 | 0.73 |
| pressmark | pressmark-go | time, s | 989.6 | 1405.7 | 0.70 |
| pressmark | pressmark-go | in+out | 2889377 | 5769755 | 0.50 |
| pressmark | pressmark-go | total | 5592225 | 11328091 | 0.49 |
| pressmark | pressmark-js | time, s | 1029.0 | 942.0 | 1.09 |
| pressmark | pressmark-js | in+out | 4408972 | 3550587 | 1.24 |
| pressmark | pressmark-js | total | 8644300 | 6923899 | 1.25 |
| pressmark | pressmark-python | time, s | 1167.1 | 1108.5 | 1.05 |
| pressmark | pressmark-python | in+out | 6275109 | 5689104 | 1.10 |
| pressmark | pressmark-python | total | 12361061 | 11161616 | 1.11 |
| pressmark | pressmark-rust | time, s | 3267.4 | 1131.5 | 2.89 |
| pressmark | pressmark-rust | in+out | 18766293 | 4921256 | 3.81 |
| pressmark | pressmark-rust | total | 37123285 | 9651112 | 3.85 |
| **pressmark** | **Σ wave** | time, s | 7525.6 | 5769.7 | 1.30 |
| **pressmark** | **Σ wave** | in+out | 36026047 | 24944330 | 1.44 |
| **pressmark** | **Σ wave** | total | 70925567 | 48904522 | 1.45 |
| quire | quire-cpp | time, s | 1421.1 | 1756.2 | 0.81 |
| quire | quire-cpp | in+out | 5853930 | 8477384 | 0.69 |
| quire | quire-cpp | total | 11532458 | 16719176 | 0.69 |
| quire | quire-go | time, s | 1076.0 | 1256.7 | 0.86 |
| quire | quire-go | in+out | 4534007 | 4803634 | 0.94 |
| quire | quire-go | total | 8873655 | 9432178 | 0.94 |
| quire | quire-js | time, s | 968.6 | 957.3 | 1.01 |
| quire | quire-js | in+out | 3287688 | 3613622 | 0.91 |
| quire | quire-js | total | 6387336 | 7044534 | 0.91 |
| quire | quire-python | time, s | 894.0 | 797.0 | 1.12 |
| quire | quire-python | in+out | 4376033 | 3315720 | 1.32 |
| quire | quire-python | total | 8599841 | 6488392 | 1.33 |
| quire | quire-rust | time, s | 1272.2 | 926.1 | 1.37 |
| quire | quire-rust | in+out | 3615026 | 4643046 | 0.78 |
| quire | quire-rust | total | 7052530 | 9118950 | 0.77 |
| **quire** | **Σ wave** | time, s | 5631.9 | 5693.3 | 0.99 |
| **quire** | **Σ wave** | in+out | 21666684 | 24853406 | 0.87 |
| **quire** | **Σ wave** | total | 42445820 | 48803230 | 0.87 |

## 6. Camps — a hand + its operator (§0.2b.1; both K3 bases; all waves)

| wave | task | hand | hand, s | oper, s | camp, s | hand in+out | oper in+out | camp in+out | hand total | oper total | camp total |
|---|---|---|---|---|---|---|---|---|---|---|---|
| conduit | conduit-cpp | solo | 1258.4 | 612.9 | 1871.3 | 4139853 | 1836167 | 5976020 | 8018125 | 3529543 | 11547668 |
| conduit | conduit-cpp | top | 1014.7 | 224.7 | 1239.4 | 4056303 | 83888 | 4140191 | 7948015 | 162864 | 8110879 |
| conduit | conduit-cpp | solo/top × (top (G1)) | — | — | 1.51 | — | — | 1.44 | — | — | 1.42 |
| conduit | conduit-go | solo | 1482.6 | 226.2 | 1708.8 | 4783871 | 1074456 | 5858327 | 9371071 | 2077720 | 11448791 |
| conduit | conduit-go | top | 826.0 | 36.5 | 862.5 | 2188552 | 107388 | 2295940 | 4235144 | 208636 | 4443780 |
| conduit | conduit-go | solo/top × (top (G1)) | — | — | 1.98 | — | — | 2.55 | — | — | 2.58 |
| conduit | conduit-js | solo | 1067.7 | 324.7 | 1392.4 | 2875019 | 1026033 | 3901052 | 5594379 | 1981233 | 7575612 |
| conduit | conduit-js | top | 872.3 | 162.7 | 1035.0 | 2487511 | 130620 | 2618131 | 4795479 | 253948 | 5049427 |
| conduit | conduit-js | solo/top × (top (G1)) | — | — | 1.35 | — | — | 1.49 | — | — | 1.50 |
| conduit | conduit-python | solo | 1189.5 | 278.8 | 1468.3 | 3047843 | 823807 | 3871650 | 5948451 | 1586943 | 7535394 |
| conduit | conduit-python | top | 1063.0 | 22.4 | 1085.4 | 3681302 | 83592 | 3764894 | 7185110 | 162824 | 7347934 |
| conduit | conduit-python | solo/top × (top (G1)) | — | — | 1.35 | — | — | 1.03 | — | — | 1.03 |
| conduit | conduit-rust | solo | 933.7 | 342.8 | 1276.5 | 2830163 | 1098762 | 3928925 | 5521939 | 2123082 | 7645021 |
| conduit | conduit-rust | top | 746.9 | 45.2 | 792.1 | 2560058 | 108144 | 2668202 | 5001594 | 209712 | 5211306 |
| conduit | conduit-rust | solo/top × (top (G1)) | — | — | 1.61 | — | — | 1.47 | — | — | 1.47 |
| **conduit** | **Σ camps solo (hands, oper, camp)** | solo | 5931.9 | 1785.4 | 7717.3 | 17676749 | 5859225 | 23535974 | 34453965 | 11298521 | 45752486 |
| **conduit** | **Σ camps top (hands, oper, camp)** | top | 4522.9 | 491.5 | 5014.4 | 14973726 | 513632 | 15487358 | 29165342 | 997984 | 30163326 |
| **conduit** | **Σ solo/top ×** | — | 1.31 | 3.63 | 1.54 | 1.18 | 11.41 | 1.52 | 1.18 | 11.32 | 1.52 |
| brim | brim-cpp | solo | 1142.4 | 161.6 | 1304.0 | 3773291 | 556706 | 4329997 | 7371371 | 1072546 | 8443917 |
| brim | brim-cpp | top | 877.5 | 30.4 | 907.9 | 2527825 | 86898 | 2614723 | 4916177 | 168818 | 5084995 |
| brim | brim-cpp | solo/top × (top (G1)) | — | — | 1.44 | — | — | 1.66 | — | — | 1.66 |
| brim | brim-go | solo | 949.0 | 277.8 | 1226.8 | 3275551 | 1129601 | 4405152 | 6395743 | 2195585 | 8591328 |
| brim | brim-go | top | 726.8 | 25.6 | 752.4 | 1852498 | 86546 | 1939044 | 3575378 | 168274 | 3743652 |
| brim | brim-go | solo/top × (top (G1)) | — | — | 1.63 | — | — | 2.27 | — | — | 2.29 |
| brim | brim-js | solo | 1228.5 | 356.8 | 1585.3 | 3990369 | 1190087 | 5180456 | 7799649 | 2307975 | 10107624 |
| brim | brim-js | top | 496.4 | 20.7 | 517.1 | 1379251 | 63238 | 1442489 | 2655731 | 122630 | 2778361 |
| brim | brim-js | solo/top × (top (G1)) | — | — | 3.07 | — | — | 3.59 | — | — | 3.64 |
| brim | brim-python | solo | 727.9 | 240.4 | 968.3 | 1582305 | 967094 | 2549399 | 3039137 | 1881718 | 4920855 |
| brim | brim-python | top | 614.9 | 28.1 | 643.0 | 1875438 | 85995 | 1961433 | 3654830 | 167275 | 3822105 |
| brim | brim-python | solo/top × (top (G1)) | — | — | 1.51 | — | — | 1.30 | — | — | 1.29 |
| brim | brim-rust | solo | 1077.1 | 445.3 | 1522.4 | 5199311 | 1616098 | 6815409 | 10237199 | 3142434 | 13379633 |
| brim | brim-rust | top | 911.7 | 44.4 | 956.1 | 3859302 | 138846 | 3998148 | 7576678 | 269662 | 7846340 |
| brim | brim-rust | solo/top × (top (G1)) | — | — | 1.59 | — | — | 1.70 | — | — | 1.71 |
| **brim** | **Σ camps solo (hands, oper, camp)** | solo | 5124.9 | 1481.9 | 6606.8 | 17820827 | 5459586 | 23280413 | 34843099 | 10600258 | 45443357 |
| **brim** | **Σ camps top (hands, oper, camp)** | top | 3627.3 | 149.2 | 3776.5 | 11494314 | 461523 | 11955837 | 22378794 | 896659 | 23275453 |
| **brim** | **Σ solo/top ×** | — | 1.41 | 9.93 | 1.75 | 1.55 | 11.83 | 1.95 | 1.56 | 11.82 | 1.95 |
| errand | errand-cpp | solo | 1197.2 | 319.4 | 1516.6 | 3361704 | 1627333 | 4989037 | 6402216 | 3172933 | 9575149 |
| errand | errand-cpp | top | 817.3 | 32.0 | 849.3 | 2729096 | 66200 | 2795296 | 5329608 | 126040 | 5455648 |
| errand | errand-cpp | solo/top × (top (G1)) | — | — | 1.79 | — | — | 1.78 | — | — | 1.76 |
| errand | errand-go | solo | 1017.7 | 597.4 | 1615.1 | 3096275 | 2745012 | 5841287 | 6046419 | 5341044 | 11387463 |
| errand | errand-go | top | 874.7 | 39.8 | 914.5 | 3262480 | 66602 | 3329082 | 6409872 | 126250 | 6536122 |
| errand | errand-go | solo/top × (top (G1)) | — | — | 1.77 | — | — | 1.75 | — | — | 1.74 |
| errand | errand-js | solo | 1288.5 | 702.0 | 1990.5 | 5048914 | 2615785 | 7664699 | 9905938 | 5075433 | 14981371 |
| errand | errand-js | top | 732.3 | 32.7 | 765.0 | 2696814 | 65380 | 2762194 | 5272430 | 125092 | 5397522 |
| errand | errand-js | solo/top × (top (G1)) | — | — | 2.60 | — | — | 2.77 | — | — | 2.78 |
| errand | errand-python | solo | 1287.6 | 299.0 | 1586.6 | 5015120 | 911255 | 5926375 | 9846544 | 1757143 | 11603687 |
| errand | errand-python | top | 1003.7 | 26.6 | 1030.3 | 4129866 | 64823 | 4194689 | 8109322 | 124535 | 8233857 |
| errand | errand-python | solo/top × (top (G1)) | — | — | 1.54 | — | — | 1.41 | — | — | 1.41 |
| errand | errand-rust | solo | 1122.0 | 374.6 | 1496.6 | 7269851 | 1182532 | 8452383 | 14376731 | 2285252 | 16661983 |
| errand | errand-rust | top | 1013.6 | 28.9 | 1042.5 | 4762833 | 86732 | 4849565 | 9371281 | 168268 | 9539549 |
| errand | errand-rust | solo/top × (top (G1)) | — | — | 1.44 | — | — | 1.74 | — | — | 1.75 |
| **errand** | **Σ camps solo (hands, oper, camp)** | solo | 5913.0 | 2292.4 | 8205.4 | 23791864 | 9081917 | 32873781 | 46577848 | 17631805 | 64209653 |
| **errand** | **Σ camps top (hands, oper, camp)** | top | 4441.6 | 160.0 | 4601.6 | 17581089 | 349737 | 17930826 | 34492513 | 670185 | 35162698 |
| **errand** | **Σ solo/top ×** | — | 1.33 | 14.33 | 1.78 | 1.35 | 25.97 | 1.83 | 1.35 | 26.31 | 1.83 |
| pressmark | pressmark-cpp | solo | 1072.5 | 411.2 | 1483.7 | 3686296 | 1942473 | 5628769 | 7204696 | 3781833 | 10986529 |
| pressmark | pressmark-cpp | top | 1182.0 | 20.0 | 1202.0 | 5013628 | 62484 | 5076112 | 9839804 | 115284 | 9955088 |
| pressmark | pressmark-cpp | solo/top × (top (G1)) | — | — | 1.23 | — | — | 1.11 | — | — | 1.10 |
| pressmark | pressmark-go | solo | 989.6 | 446.5 | 1436.1 | 2889377 | 2093080 | 4982457 | 5592225 | 4066520 | 9658745 |
| pressmark | pressmark-go | top | 1405.7 | 17.4 | 1423.1 | 5769755 | 61693 | 5831448 | 11328091 | 120445 | 11448536 |
| pressmark | pressmark-go | solo/top × (— (exchange)) | — | — | 1.01 | — | — | 0.85 | — | — | 0.84 |
| pressmark | pressmark-js | solo | 1029.0 | 428.6 | 1457.6 | 4408972 | 1810489 | 6219461 | 8644300 | 3524345 | 12168645 |
| pressmark | pressmark-js | top | 942.0 | 21.7 | 963.7 | 3550587 | 63079 | 3613666 | 6923899 | 115623 | 7039522 |
| pressmark | pressmark-js | solo/top × (top (G1)) | — | — | 1.51 | — | — | 1.72 | — | — | 1.73 |
| pressmark | pressmark-python | solo | 1167.1 | 464.3 | 1631.4 | 6275109 | 2432013 | 8707122 | 12361061 | 4752333 | 17113394 |
| pressmark | pressmark-python | top | 1108.5 | 18.4 | 1126.9 | 5689104 | 63346 | 5752450 | 11161616 | 122802 | 11284418 |
| pressmark | pressmark-python | solo/top × (top (G1)) | — | — | 1.45 | — | — | 1.51 | — | — | 1.52 |
| pressmark | pressmark-rust | solo | 3267.4 | 1422.2 | 4689.6 | 18766293 | 5823352 | 24589645 | 37123285 | 11225720 | 48349005 |
| pressmark | pressmark-rust | top | 1131.5 | 23.2 | 1154.7 | 4921256 | 61596 | 4982852 | 9651112 | 120284 | 9771396 |
| pressmark | pressmark-rust | solo/top × (top (G1)) | — | — | 4.06 | — | — | 4.93 | — | — | 4.95 |
| **pressmark** | **Σ camps solo (hands, oper, camp)** | solo | 7525.6 | 3172.8 | 10698.4 | 36026047 | 14101407 | 50127454 | 70925567 | 27350751 | 98276318 |
| **pressmark** | **Σ camps top (hands, oper, camp)** | top | 5769.7 | 100.7 | 5870.4 | 24944330 | 312198 | 25256528 | 48904522 | 594438 | 49498960 |
| **pressmark** | **Σ solo/top ×** | — | 1.30 | 31.51 | 1.82 | 1.44 | 45.17 | 1.98 | 1.45 | 46.01 | 1.99 |
| quire | quire-cpp | solo | 1421.1 | 272.2 | 1693.3 | 5853930 | 1416552 | 7270482 | 11532458 | 2765928 | 14298386 |
| quire | quire-cpp | top | 1756.2 | 38.5 | 1794.7 | 8477384 | 94282 | 8571666 | 16719176 | 180298 | 16899474 |
| quire | quire-cpp | solo/top × (solo (G1)) | — | — | 0.94 | — | — | 0.85 | — | — | 0.85 |
| quire | quire-go | solo | 1076.0 | 410.9 | 1486.9 | 4534007 | 1020733 | 5554740 | 8873655 | 1968061 | 10841716 |
| quire | quire-go | top | 1256.7 | 89.8 | 1346.5 | 4803634 | 241233 | 5044867 | 9432178 | 466321 | 9898499 |
| quire | quire-go | solo/top × (top (G1)) | — | — | 1.10 | — | — | 1.10 | — | — | 1.10 |
| quire | quire-js | solo | 968.6 | 960.6 | 1929.2 | 3287688 | 3316237 | 6603925 | 6387336 | 6467341 | 12854677 |
| quire | quire-js | top | 957.3 | 80.1 | 1037.4 | 3613622 | 185219 | 3798841 | 7044534 | 338307 | 7382841 |
| quire | quire-js | solo/top × (top (G1)) | — | — | 1.86 | — | — | 1.74 | — | — | 1.74 |
| quire | quire-python | solo | 894.0 | 361.6 | 1255.6 | 4376033 | 1216897 | 5592930 | 8599841 | 2359425 | 10959266 |
| quire | quire-python | top | 797.0 | 45.1 | 842.1 | 3315720 | 119979 | 3435699 | 6488392 | 230187 | 6718579 |
| quire | quire-python | solo/top × (top (G1)) | — | — | 1.49 | — | — | 1.63 | — | — | 1.63 |
| quire | quire-rust | solo | 1272.2 | 369.9 | 1642.1 | 3615026 | 1237058 | 4852084 | 7052530 | 2403138 | 9455668 |
| quire | quire-rust | top | 926.1 | 50.1 | 976.2 | 4643046 | 154269 | 4797315 | 9118950 | 295069 | 9414019 |
| quire | quire-rust | solo/top × (top (G1)) | — | — | 1.68 | — | — | 1.01 | — | — | 1.00 |
| **quire** | **Σ camps solo (hands, oper, camp)** | solo | 5631.9 | 2375.2 | 8007.1 | 21666684 | 8207477 | 29874161 | 42445820 | 15963893 | 58409713 |
| **quire** | **Σ camps top (hands, oper, camp)** | top | 5693.3 | 303.6 | 5996.9 | 24853406 | 794982 | 25648388 | 48803230 | 1510182 | 50313412 |
| **quire** | **Σ solo/top ×** | — | 0.99 | 7.82 | 1.34 | 0.87 | 10.32 | 1.16 | 0.87 | 10.57 | 1.16 |
| **Σ BATTERY** | — | **Σ camps solo** | 30127.3 | 11107.7 | 41235 | 116982171 | 42709612 | 159691783 | 229246299 | 82845228 | 312091527 |
| **Σ BATTERY** | — | **Σ camps top** | 24054.8 | 1205 | 25259.8 | 93846865 | 2432072 | 96278937 | 183744401 | 4669448 | 188413849 |
| **Σ BATTERY** | — | **solo/top ×** | 1.3 | 9.2 | 1.6 | 1.2 | 17.6 | 1.7 | 1.2 | 17.7 | 1.7 |



## 7. The G1/G2 verdicts by task (§0.2b; G1 = the hand+operator camp, G2 = hands without operators)

| wave | task | G1 (camp) | camp total solo/top | G2 (hands): time × / total × |
|---|---|---|---|---|
| conduit | conduit-cpp | top (G1) | 1.42 | machine (1.24 / 1.01) |
| conduit | conduit-go | top (G1) | 2.58 | machine (1.79 / 2.21) |
| conduit | conduit-js | top (G1) | 1.50 | machine (1.22 / 1.17) |
| conduit | conduit-python | top (G1) | 1.03 | mixed (1.12 / 0.83) |
| conduit | conduit-rust | top (G1) | 1.47 | machine (1.25 / 1.10) |
| brim | brim-cpp | top (G1) | 1.66 | machine (1.30 / 1.50) |
| brim | brim-go | top (G1) | 2.29 | machine (1.31 / 1.79) |
| brim | brim-js | top (G1) | 3.64 | machine (2.47 / 2.94) |
| brim | brim-python | top (G1) | 1.29 | mixed (1.18 / 0.83) |
| brim | brim-rust | top (G1) | 1.71 | machine (1.18 / 1.35) |
| errand | errand-cpp | top (G1) | 1.76 | machine (1.46 / 1.20) |
| errand | errand-go | top (G1) | 1.74 | mixed (1.16 / 0.94) |
| errand | errand-js | top (G1) | 2.78 | machine (1.76 / 1.88) |
| errand | errand-python | top (G1) | 1.41 | machine (1.28 / 1.21) |
| errand | errand-rust | top (G1) | 1.75 | machine (1.11 / 1.53) |
| pressmark | pressmark-cpp | top (G1) | 1.10 | solo (0.91 / 0.73) |
| pressmark | pressmark-go | — (exchange) | 0.84 | solo (0.70 / 0.49) |
| pressmark | pressmark-js | top (G1) | 1.73 | machine (1.09 / 1.25) |
| pressmark | pressmark-python | top (G1) | 1.52 | machine (1.05 / 1.11) |
| pressmark | pressmark-rust | top (G1) | 4.95 | machine (2.89 / 3.85) |
| quire | quire-cpp | solo (G1) | 0.85 | solo (0.81 / 0.69) |
| quire | quire-go | top (G1) | 1.10 | solo (0.86 / 0.94) |
| quire | quire-js | top (G1) | 1.74 | mixed (1.01 / 0.91) |
| quire | quire-python | top (G1) | 1.63 | machine (1.12 / 1.33) |
| quire | quire-rust | top (G1) | 1.00 | mixed (1.37 / 0.77) |

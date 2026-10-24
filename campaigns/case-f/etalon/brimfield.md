# Etalon: brim — measuring margins on pictures

The meaning's yardstick, written into the campaign before any run:
every hand of brim, in every stack, is judged against these same
lines — the instrument is one; no product, no hand and no judge
forms it. Each line is one requirement of the brief with the run
that proves it.

1. Called as `brim detect image.jpg`; the answer is four numbers —
   margin pixels on the top, bottom, left and right — proof: run on
   the material's pictures and read the four numbers.
2. The margin is an even strip running from the edge — white, black,
   anything — measured by entropy: an even strip's disorder is near
   zero, content is alive; the boundary is where the difference in
   chaos between neighboring chunks is at its largest — proof: the
   code's measuring walk plus its answers on the material.
3. Each side is measured on its own; evenness may stretch to the end
   of the search band — then the whole band is the answer; no margin
   means zero — proof: the two material pictures between them show a
   full band, a frame and a zero.
4. `--band` — how deep from the edge to search, a fraction of the
   side, no more than half allowed, a quarter by default — proof:
   runs with and without the knob.
5. `--threshold` — the sensitivity, a fraction from 0 to 1, half by
   default; the higher, the more aggressive the hunt — proof: runs
   with and without the knob.
6. The one format is an ordinary still JPEG — proof: a non-JPEG
   picture is refused.
7. A repeat run on the same picture gives the same answer, to the
   letter — proof: run twice, compare.
8. `bordered.jpg` (250×250, a thin black even frame) gives four
   identical numbers — this material has four pixels on each side —
   proof: the run's answer is "4 4 4 4".
9. `clear.jpg` (510×350, no margin) gives zeros on all sides —
   proof: the run's answer is "0 0 0 0".
10. A search band shallower than the frame cuts down the found
    numbers; a deeper one does not: on `bordered.jpg` `--band 0.01`
    gives "2 2 2 2", `--band 0.05` leaves "4 4 4 4" — proof: both
    runs.
11. Trouble is reported briefly and clearly, in English: the file is
    missing or unreadable; this is not a picture at all or not the
    one format (JPEG); the picture is so small there is nothing to
    search for a margin in (a side shorter than two pixels) — proof:
    provoke each once and read the message.
12. A clumsy invocation — a bad knob value, an extra file argument,
    an unknown option — is an honest refusal, and each knob's
    invalid value is named along with its requirement — proof: an
    out-of-range band, a band over half, an extra path, an unknown
    flag.
13. The program only reads the picture and prints the answer: it
    writes nothing to disk, goes to no network, keeps no settings in
    files — proof: the code read plus the tree before and after the
    runs.
14. The tool carries its standing proof: a corpus of checks living
    with the tree — proof: the corpus file(s) in the tree show the
    whole law of the tool.
15. Every behavior of lines 1–13 is proven by a standing corpus row,
    green under the current fingerprints — proof: the corpus read
    against the lines, each line named by its proving row.
16. The verdict stands: the corpus's own standing verdict line closes
    the proof — proof: the verdict line over the corpus; a fresh
    re-run of the corpus reads green and the verdict holds.

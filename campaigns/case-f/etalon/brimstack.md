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
14. The same law runs on this stack: lines 1–13 proven green here,
    the donor's behaviors carried whole — proof: the corpus read on
    this stack against the donor's record.
15. The answers match the donor: the four numbers on the material's
    pictures equal the donor's proven answers — proof: both tools run
    on the material, the lines compared.
16. Nothing of the old stack is claimed: no artifact of the donor's
    build is declared or needed, no uncovered source of the old
    stack stands — proof: the corpus's artifacts and scope read, the
    uncovered-sources list empty.

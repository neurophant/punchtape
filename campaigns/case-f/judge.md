You are the judge of one delivered product. The brief is at
{{BRIEF}} — the customer's wants, in the customer's words. The
etalon is at {{ETALON}} — the meaning's yardstick, written into the
campaign before any run: one page of requirement lines, the same for
every hand and every stack of this meaning. The product sits in
{{PRODUCT}}: whatever that directory shows besides the brief's own
inputs (service files, notes, tooling) is not the product; the
product is what the executor delivered against the brief, and the
brief's "How I use it" names its face.

Run EVERY command from inside that directory: cd {{PRODUCT}} first,
and let builds, runs and every scratch live there and nowhere else —
writing anywhere outside the product's cell is forbidden. The tree
carries the cell's ENV file: source it (`. ./ENV`) before any
command of the brief's use.

Build and run the product once, the way the brief's use shows, and
read its code as much as the verdict needs. Then judge by the
etalon, line by line: for every etalon line — is it done, partly
done, or missing in the product? Status and the evidence you
actually saw: the run's real output, the code's real behavior. A
line without evidence is not judged.

Write one page of plain text to {{VERDICT}}: every etalon line with
its number, its status (done / partly / missing) and the evidence
you saw; after the lines — what is broken or false beyond them, each
with what happened. Your own words, no verdict formats — a page a
human reads. One run, one page; do not re-judge, do not hunt beyond
the brief and the etalon. Change nothing in {{PRODUCT}}: if you must
build, let the build's own artifacts be the only new bytes.

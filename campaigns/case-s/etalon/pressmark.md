# Etalon: pressmark — a small site from folders

The meaning's yardstick, written into the campaign before any run:
every hand of pressmark, in every stack, is judged against these same
lines — the instrument is one; no product, no hand and no judge
forms it. Each line is one requirement of the brief with the run
that proves it.

1. Called as `pressmark build storage --template templates --out
   site`; the build wipes the output directory and builds it anew —
   proof: build twice with a leftover file in between.
2. First-level folders in the storage are records; files in the
   storage root don't matter; an empty storage is fine — a single
   index with an empty list gets built — proof: build an empty
   storage.
3. A record folder carries a `meta.yaml` description — the creation
   moment as "year-month-day hour:minute:second", no time zone, the
   same shape in all records; a title; the full-text file name; the
   short-text file name; a list of pictures and a list of tags, both
   required, both allowed to be empty — plus full and short texts in
   markdown and pictures as just files — proof: the material's three
   records (pictures and tags, empty lists, a simpler one) all
   build.
4. A record's address: "year-month-day", a hyphen, the title brought
   to link form — lowercase letters and digits only, spaces and
   other characters become hyphens, consecutive hyphens collapse
   into one, leading and trailing hyphens are dropped, letters of
   other alphabets transliterate into Latin; 2000-08-25 with
   "Article 1" gives `2000-08-25-article-1` — proof: the material's
   titles and the built page names.
5. Each record builds a page `<out>/<address>.html`; pictures move
   to `<out>/media/<address>/<name>`; an index `<out>/index.html` is
   built — proof: the output tree of a material build.
6. The templates directory holds exactly two files: `entity.html`
   for the record page and `index.html` for the index — proof: the
   material's templates and a run with one missing.
7. Into the record template go: the moment (the same string as in
   the description), the title, the address, the ready HTML of the
   full and short texts (just markdown, no templating inside the
   text), the pictures (name and address of the form
   `media/<address>/<name>`) and the tags — proof: a template
   substituting all the fields and the built page.
8. Into the index go the same records without the full texts, from
   newest to oldest; on equal moments, by address — proof: the
   material's different moments and the index's order.
9. For every record it builds, the program says an "ok" line with
   the record's address, in build order (the same order as the
   index, only from oldest to newest); the last line is "ok index" —
   proof: the run's ok lines against the index's order.
10. A repeat build of the same storage into a new folder gives the
    same files, to the letter — proof: two builds, `cmp -r`.
11. Two records with the same address (same date and title) — an
    honest refusal, nothing is built — proof: a two-record storage
    with equal date and title.
12. Trouble is reported briefly and clearly, in English: the storage
    or the templates directory is missing; a candidate folder has no
    description; the description is crooked — a field is missing or
    the type is wrong; a text or picture file is missing; the
    templates lack `entity.html` or `index.html` — proof: provoke
    each once and read the message.
13. A clumsy invocation — extra paths, an unknown command or option
    — is an honest refusal — proof: wrong arities and an unknown
    command.
14. All files are read and written in utf-8; no network, no server;
    nothing is read outside the storage, the templates and the
    output; the storage and the templates are never changed — proof:
    the code read plus the trees before and after the runs.

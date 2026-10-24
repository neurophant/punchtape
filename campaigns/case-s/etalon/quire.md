# Etalon: quire — a page from a directory, no configuration at all

The meaning's yardstick, written into the campaign before any run:
every hand of quire, in every stack, is judged against these same
lines — the instrument is one; no product, no hand and no judge
forms it. Each line is one requirement of the brief with the run
that proves it.

1. Called as `quire source [output]`; with no output it builds right
   into the source — proof: run with one path, then with two.
2. For every created file it says an "ok" line with the path; the
   listing is stable, the root's files alphabetically — proof: run
   on the material and read the ok lines' order, then run again.
3. Markdown files in the source's root are the entries; front matter
   between `---` lines is the entry's data, the rest is the body and
   turns into HTML — proof: the material's entry with front matter
   shows its data substituted and its body rendered in the output.
4. A file without front matter is body only — proof: `about.md`
   reaches the template as a plain entry.
5. `data.yml` in the root is the data, visible to the template under
   the name `data`; a missing file means no data at all — proof: run
   with the material's data.yml, then without it.
6. The template sees all entries as a list in a stable order, by
   file name alphabetically, each with its front-matter data and its
   body's ready HTML — proof: the template output lists every entry
   alphabetically with data and rendered body.
7. One template — a root file named `index` with the template
   engine's extension — builds into `index.html`, with `data` and
   the entries available to it — proof: `index.html` exists after
   the build and shows both.
8. No template — an honest refusal: there is nothing to build —
   proof: run a source without the index file.
9. Everything that is not a text, not `data.yml` and not the
   template is copied into the output byte for byte; nested
   directories are copied plain, their contents not processed —
   proof: the material's extraneous file (the style) and its nested
   copy compare identical, `cmp`.
10. The texts and `data.yml` themselves do not land in the output —
    proof: neither exists in the output after a build.
11. A directory without a single text and without data is a
    legitimate input: the template builds with an empty list of
    entries — proof: build an empty source.
12. A repeat build of the same source into a fresh folder gives the
    same files, to the letter; the listing order is stable across
    runs — proof: two builds, `cmp -r`.
13. Trouble is reported briefly and clearly, in English, with the
    path and the reason: the source is missing or is not a
    directory; YAML does not parse; the template is missing; the
    template does not compile or breaks on substitution — proof:
    provoke each once and read the message.
14. A clumsy invocation — no paths, too many paths — is an honest
    refusal — proof: run with zero and with three paths.
15. No network, nothing read outside the source, nothing written
    into the source (except when the output is the source itself),
    no state between runs — proof: the code read plus the runs'
    behavior.

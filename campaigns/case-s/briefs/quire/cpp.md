# Brief: quire — a page from a directory, no configuration at all

I am tired of configuring site generators. I want the opposite: my
folder's contents are all the configuration there is. One run — and
the page is ready: texts become the page's lines, data become
substitutions, everything else is copied as is.

## Stack and tools — my requirement, not the executor's choice

C++17, built with g++. Libraries: md4c (markdown → HTML), inja
(templates), libyaml (data). The template is `index.inja`.

## How I use it

    quire source [output]

If I give no output, it builds right into the source. For every
created file let it say a line "ok" with the path; the listing comes
in a stable order, the root's files alphabetically.

## How my directory is laid out

- Texts are markdown in the source's root: a file may have front
  matter between `---` lines — that is the entry's data, the rest is
  the body and turns into HTML. No front matter — just the body.
- Data — one YAML file in the root: `data.yml`; the template sees it
  under the name `data`. The file may be missing — then there is no
  data at all.
- Entries — all the root's markdown files; the template sees them as
  a list in a stable order, by file name alphabetically: each has
  the front matter's data (if there was any) and the body's ready
  HTML.
- The template — one: a file named `index` with the template
  engine's extension (the exact name is in the stack requirements
  above; `index.ejs` for example), it sits in the source's root and
  builds into `index.html`; the data `data` and the list of entries
  are available to it. No template — an honest refusal: there is
  nothing to build.
- Everything that is not a text, not `data.yml`, and not the
  template is copied into the output byte for byte; nested
  directories are copied plain too, their contents are not
  processed. The texts and `data.yml` themselves do not land in the
  output: they live for the template.
- A directory without a single text and without data is a
  legitimate input: the template builds with an empty list of
  entries.

## Limits and errors

A repeat build of the same source into a fresh folder gives the same
files, to the letter; the listing order is stable across runs. Let
it report trouble briefly and clearly, in English: the source is
missing or is not a directory; YAML does not parse; the template is
missing; the template does not compile or breaks on substitution —
with the path and the reason. A clumsy invocation — no paths, too
many paths — is an honest refusal.

## What must not happen

Going to the network, reading anything outside the source, writing
into the source (except when the output is the source itself),
piling up state between runs.

## Test material

The material is simple and the same on every run: a small directory
— `data.yml` with a couple of keys, `about.md` without front matter,
an entry with front matter (a heading and a couple of body lines), an
index template that prints the data and all the entries, and one
extraneous file (a style, for example). The texts are short homemade
phrases; enough for the data substitution, the entries, the copying,
and an empty input.

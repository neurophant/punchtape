# Brief: pressmark — a small site out of folders

My content lives in files, and I like it that way: no databases. Every
entry is its own folder; with one command I want to turn these folders
into ready site pages: one page per entry and a shared index page.

## Stack and tools — my requirement, not the executor's choice

Python 3. Libraries: PyYAML 6.0.1 (descriptions), Markdown 3.6 (text →
HTML, with no extra extensions), Jinja2 3.1.4 (templates), pydantic
2.8.2 (description validation), python-slugify 8.0.4 (readable
addresses — with all of its defaults: lowercase, hyphens,
transliteration, no trimming). The product is a page builder invoked as
a command; serving the pages and pagination are not included.

## How my entries are laid out

First-level folders in the storage are entries; files in the storage's
root do not matter to me. An empty storage is fine: one index is built
with an empty list.

In an entry's folder:

- a description `meta.yaml`: the creation moment — "year-month-day
  hour:minute:second", no time zone, the same shape in all entries; a
  title; the full-text file name; the short-text file name; a list of
  pictures and a list of tags — both required, both allowed to be
  empty;
- the full and short texts are markdown;
- pictures are just files.

An entry's address: "year-month-day", a hyphen, the title run through
the slugifier — for example, an entry dated 2026-08-25 with the title
"Article 1" gets the address `2026-08-25-article-1`.

## How I build the site

    pressmark build storage --template templates-dir --out out-dir

A build wipes the output directory and builds it anew. Every entry gets
a page `<output>/<address>.html`; pictures move to
`<output>/media/<address>/<name>`; and the index `<output>/index.html`
is built.

The templates directory holds exactly two files: `entity.html` for the
entry page and `index.html` for the index. Into the entry template go:
the moment (the same string as in the description), the title, the
address, the ready HTML of the full and short texts (plain markdown, no
templating inside the text), the pictures (name and an address of the
form `media/<address>/<name>`), and the tags. Into the index go the same
entries without their full texts, newest to oldest; on equal moments,
by address. For every built entry the program says a line "ok" with the
entry's address, in build order (the same order as in the index, only
oldest to newest); the last line is "ok index".

## Test material

The material is simple and the same on every run: a storage of three
entries — one with pictures and tags, one with empty lists, one plainer
(say, the titles "first", "second", "third", different moments,
couple-line texts) — and a pair of templates that substitute every
field. The texts and pictures are homemade and short; enough for the
ordering, the substitution of every field, and a repeat build.

## Limits and errors

All files are read and written as utf-8. A repeat build of the same
storage into a fresh folder gives the same files, to the letter. Two
entries with the same address (same date and title) is an honest
refusal; nothing gets built. Let it report trouble briefly and clearly,
in English: the storage or the templates directory is missing; a
candidate folder has no description; the description is crooked — a
field is missing or has the wrong type; a text or picture file is
missing; the templates lack `entity.html` or `index.html`. A clumsy
invocation — extra paths, an unknown command or option — is an honest
refusal.

## What must not happen

Going to the network, starting a server, reading anything outside the
storage, the templates, and the output, changing my storage or
templates.

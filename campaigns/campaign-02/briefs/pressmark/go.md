# Brief: pressmark — a small site from folders

My content lives in files, and I like it that way: no databases. Each
record is its own folder; I want one command that builds finished site
pages out of these folders: one page per record and a shared index
page.

## Stack and tools — my requirement, not the executor's choice

Go. Libraries: goldmark (markdown → HTML, no extensions),
gopkg.in/yaml.v3 (descriptions); templates — the standard
html/template; a record's address follows the rule from this brief.
The product is a page builder called as a command; serving pages and
pagination are out of scope.

## How my records are arranged

First-level folders in the storage are records; files in the storage
root don't matter to me. An empty storage is fine: a single index with
an empty list gets built.

In a record folder:

- a `meta.yaml` description: the creation moment as "year-month-day
  hour:minute:second", no time zone, the same shape in all records; a
  title; the full-text file name; the short-text file name; a list of
  pictures and a list of tags — both required, both allowed to be
  empty;
- full and short text — markdown;
- pictures — just files.

A record's address: "year-month-day", a hyphen, the title brought to
link form: lowercase letters and digits only, spaces and other
characters become hyphens, consecutive hyphens collapse into one,
leading and trailing hyphens are dropped, letters of other alphabets
are transliterated into Latin — for example, a record dated 2026-08-25
with the title "Article 1" gets the address `2026-08-25-article-1`.

## How I build the site

    pressmark build storage --template templates --out site

The build wipes the output directory and builds it anew. Each record
gets a page `<out>/<address>.html`, pictures move to
`<out>/media/<address>/<name>`, and an index `<out>/index.html` is
built.

The templates directory holds exactly two files: `entity.html` for the
record page and `index.html` for the index. Into the record template
go: the moment (the same string as in the description), the title, the
address, the ready HTML of the full and short texts (just markdown, no
templating inside the text), the pictures (name and address of the
form `media/<address>/<name>`) and the tags. Into the index go the
same records without the full texts, from newest to oldest; on equal
moments, by address. For every record it builds, the program says an
"ok" line with the record's address, in build order (the same order as
the index, only from oldest to newest); the last line is "ok index".

## Test material

The material is simple and the same on every run: a storage of three
records — one with pictures and tags, one with empty lists, one
simpler (for example, the titles "first", "second", "third", different
moments, texts a couple of lines long) — and a pair of templates
substituting all the fields. The texts and pictures are homemade and
short; that is enough to check the ordering, the substitution of all
the fields, and a repeat build.

## Limits and errors

All files are read and written in utf-8. A repeat build of the same
storage into a new folder gives the same files, to the letter. Two
records with the same address (same date and title) — an honest
refusal, nothing is built. Let it report trouble briefly and clearly,
in English: the storage or the templates directory is missing; a
candidate folder has no description; the description is crooked — a
field is missing or the type is wrong; a text or picture file is
missing; the templates lack `entity.html` or `index.html`. A clumsy
invocation — extra paths, an unknown command or option — is an honest
refusal.

## What must not happen

Going to the network, starting a server, reading anything outside the
storage, the templates and the output, changing my storage and
templates.

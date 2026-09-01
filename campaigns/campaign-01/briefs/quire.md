# Brief: quire — a site from a directory, no configuration at all

I am tired of configuring static-site generators. I want the opposite:
my directory's structure is all the configuration there is. One run —
and the site is ready: texts and data become pages, everything else is
copied as is.

## Stack and tools — my requirement, not the executor's choice

Node.js (an ESM module, the system node). Libraries: marked 17.0.5
(markdown → HTML, the GitHub Flavored Markdown dialect), yeahjs 0.3.1
(compiling EJS templates; template fields are visible through the
locals object), @mourner/yeahml 1.0.0 (parsing YAML). No other
dependencies.

## How I use it

    quire source [output]

If I give no output, it builds right into the source. For every created
file let it say a line "ok" with the path; the listing comes in a
stable order (directories alphabetically, files inside a directory
alphabetically, the root first, nested ones after); the flag `--silent`
means silence, errors only.

## How my directory is laid out

- Texts are markdown: a file may have front matter between `---` lines
  — that is the entry's data; the rest is the body and turns into HTML.
  No front matter — just the body.
- Data files are YAML, whole.
- Directories become nesting: the whole tree of data is available to
  any template as "the whole site", and to a template in its own
  directory also that directory's data and the path to the site root,
  so that links work from any depth.
- Templates are EJS: an ordinary template becomes a ready file with the
  same name, and the extension can be anything — this way not only
  pages are built but also, say, a stylesheet: the style template
  builds into a style file; a special entry template is built once for
  every data file in its directory — that is how post pages come
  about; templates whose name starts with an underscore do not become
  pages themselves — other templates include them.
- Everything that is not a text, not data, and not a template is copied
  into the output byte for byte. The texts and data themselves do not
  land in the output: they live for the templates.
- An empty directory is a legitimate input: a site of nothing but
  copies.

## Test material

The material is simple and the same on every run: a small blog
directory — `data.yml` with the author's name, `about.md` without front
matter, a posts folder with two entries (front matter with a title plus
a couple of body lines), an entry template in the posts folder, an
index template in the root, an underscore-prefixed header template for
includes, an ordinary stylesheet file, and one nested data file. The
texts are short homemade phrases; enough for all the pages, the data,
the copying, and both flags.

## Two handy flags

- `--breaks`: a single line break inside a paragraph becomes a visible
  break on the site;
- `--smartypants`: typography in the texts — ellipsis, dash, and « »
  quotes instead of the machine ones.

## Limits and errors

A repeat build of the same source into a fresh folder gives the same
files, to the letter; the listing order is stable across runs. Let it
report trouble briefly and clearly, in English: the source is missing
or is not a directory; YAML does not parse; a template does not compile
or breaks on substitution — with the path and the reason. A clumsy
invocation — no paths, too many paths, an unknown flag — is an honest
refusal.

## What must not happen

Going to the network, reading anything outside the source, writing into
the source (except when the output is the source itself), piling up
state between runs.

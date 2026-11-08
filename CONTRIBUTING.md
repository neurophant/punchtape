# Contributing

Issues first, pull requests second: open an issue (typos excepted), say
what you want to change and why, and wait for a go before investing in
code.

## Build and check

Building needs Go 1.26+:

```text
make build
```

Everything the repository checks runs through one command:

```text
make venv && make check
```

`make venv` creates the repository-local Python tooling venv from
`uv.lock` (needs [uv](https://docs.astral.sh/uv/)). `make check` then
drives the whole battery: the documentation mirrors' parity gate (every
Russian and Chinese mirror structurally exact against its English
base), ruff and mypy --strict over every tracked Python file, gofmt,
go vet, staticcheck, golint, unparam, gopls and dupl over the Go code,
and the test suite with 100% coverage of non-test lines under a
120-second ceiling. A change that leaves any of this red does not go
in.

## Style

Match the surrounding code: its naming, comment density, and
structure. Repository content is in English; the README and the user
docs come in English, Russian and Chinese — when you change a base
document, change its mirrors in the same act. The parity gate enforces
the structure; the meaning is on you.

Questions and ideas: GitHub Discussions.

# Build and install punchtape.
# -buildvcs=false keeps the binary byte-identical regardless of the
# git state.

BINARY := punchtape
COV := /tmp/punchtape-check
COVERPKG := ./cmd/...,./internal/...
# Demo project slices (demos/*/project/), not machine code: their
# style is their authors', the machine never runs them.
PYFILES := $(shell git ls-files '*.py' | grep -v '^demos/')
GOPKGS := ./cmd/... ./internal/...
MYPYCACHE := /tmp/punchtape-mypy
# The battery's speed ceiling. Its ONLY job is to catch an accidental
# blow-up — a hang, a sleep, a quadratic walk — before the field
# does. It is a servant of completeness, never its judge: tests are
# never trimmed or weakened for this clock. A battery that outgrows
# the ceiling on real invariant work is a fact to raise, and only an
# explicit direct order moves the ceiling.
BATTERY_LIMIT := 120s
# The coverage verdict's mode. strict — the law: any function below
# 100% fails the battery. open — the same diagnosis, no failure: for
# platforms whose kernel keeps whole branches unreachable; the
# reference platform of the strict gate is linux.
COVERAGE_GATE ?= strict

.PHONY: build install clean vet fmt dist check docs-parity python-lint venv maps

# The repository-local tooling venv, pinned by uv.lock: ruff, mypy
# and the stubs the campaign tools import.
venv:
	uv sync --locked

# The user docs' mirror parity: every Russian and Chinese mirror
# structurally exact against its English base — the law the gate's
# own header states.
docs-parity:
	.venv/bin/python tools/docs_parity.py

# The Python law: ruff's error families (the selection adds PLC0415 —
# the lazy-import ban's own rule) and the formatter's shape, and
# mypy --strict, over every Python file the repository tracks — all
# from the repository-local venv, no global tool. mypy walks one file
# a call: the campaign case directories carry hyphens, so one call
# cannot tell their same-named modules apart.
python-lint:
	.venv/bin/ruff check --no-cache --select E4,E7,E9,F,PLC0415 $(PYFILES)
	.venv/bin/ruff format --check --no-cache $(PYFILES)
	for f in $(PYFILES); do .venv/bin/mypy --strict --cache-dir $(MYPYCACHE) "$$f" || exit 1; done

# The whole battery's gate: the analyzers, the full test suite, and
# the two standing laws — every non-test statement executed by the
# battery (the instrumented binary's own launch covers the entry
# point a unit test cannot reach), and the code's size ceiling. The
# two coverage sources merge by the block's maximum count: the same
# block seen by several test binaries counts once, as executed.
check: docs-parity python-lint
	test -z "$$(gofmt -l $(shell git ls-files '*.go' | grep -v '^demos/'))"
	go vet $(GOPKGS)
	staticcheck $(GOPKGS)
	diag=$$(golint $(GOPKGS)); test -z "$$diag" || { echo "$$diag"; exit 1; }
	diag=$$(unparam $(GOPKGS)); test -z "$$diag" || { echo "$$diag"; exit 1; }
	diag=$$(gopls check $$(git ls-files '*.go' | grep -v '^demos/')); test -z "$$diag" || { echo "$$diag"; exit 1; }
	dupl -threshold 100 $$(git ls-files '*.go' | grep -v '^demos/' | grep -v '_test\.go$$') | grep -q '^Found total 0 clone' || { echo 'duplicate code found'; exit 1; }
	start=$$(date +%s); \
	go test -count=1 $(GOPKGS) -coverpkg=$(COVERPKG) -coverprofile=$(COV).test.out || exit 1; \
	elapsed=$$(($$(date +%s) - start)); \
	echo $$elapsed > $(COV).elapsed; \
	test $$elapsed -le 120 || { echo "battery too slow: $${elapsed}s over $(BATTERY_LIMIT)"; exit 1; }
	rm -rf $(COV).bindir $(COV).rundir $(COV).bin $(COV).bin.out
	go build -cover -coverpkg=$(COVERPKG) -o $(COV).bin ./cmd/punchtape
	mkdir -p $(COV).bindir $(COV).rundir
	cd $(COV).rundir && GOCOVERDIR=$(COV).bindir $(COV).bin
	GOCOVERDIR=$(COV).bindir $(COV).bin --version
	go tool covdata textfmt -i=$(COV).bindir -o $(COV).bin.out
	awk 'FNR==1 && /^mode:/ { next } \
	     { if (!($$1 in max) || $$3+0 > max[$$1]+0) { max[$$1] = $$3+0; st[$$1] = $$2 } } \
	     END { print "mode: set"; for (k in max) print k, st[k], max[k] }' \
	    $(COV).test.out $(COV).bin.out > $(COV).merged.out
	@diag=$$(go tool cover -func=$(COV).merged.out | grep -v '100.0%' | grep -v '^total:'); \
	if [ -n "$$diag" ]; then \
		echo "$$diag"; \
		n=$$(echo "$$diag" | grep -c .); \
		if [ "$(COVERAGE_GATE)" = "strict" ]; then \
			echo "coverage: $$n function(s) below 100%"; exit 1; \
		else \
			echo "coverage: $$n function(s) below 100% — gate open, the strict reference platform is linux"; \
		fi; \
	fi
	@total=$$(git ls-files '*.go' | grep -v '_test\.go$$' | xargs wc -l | tail -1 | awk '{print $$1}'); \
	echo "check: green (coverage 100%, $$total non-test lines, battery $$(cat $(COV).elapsed) s of $(BATTERY_LIMIT))"

build:
	go build -trimpath -buildvcs=false -o $(BINARY) ./cmd/punchtape

install:
	./install.sh

vet:
	go vet $(GOPKGS)

fmt:
	gofmt -l -w $(shell git ls-files '*.go' | grep -v '^demos/')

clean:
	rm -f $(BINARY)

# The one-command install distribution (get-punchtape.sh): binaries
# for four platforms plus checksums next to them. Put the contents of
# dist/ on a static https — and `curl -fsSL <url>/get-punchtape.sh | bash`
# works without the repository and without Go.
dist:
	rm -rf dist
	GOOS=linux  GOARCH=amd64 go build -trimpath -buildvcs=false -o dist/punchtape-linux-amd64  ./cmd/punchtape
	GOOS=linux  GOARCH=arm64 go build -trimpath -buildvcs=false -o dist/punchtape-linux-arm64  ./cmd/punchtape
	GOOS=darwin GOARCH=amd64 go build -trimpath -buildvcs=false -o dist/punchtape-darwin-amd64 ./cmd/punchtape
	GOOS=darwin GOARCH=arm64 go build -trimpath -buildvcs=false -o dist/punchtape-darwin-arm64 ./cmd/punchtape
	cd dist && for f in punchtape-*; do case "$$f" in *.sha256) continue;; esac; sha256sum "$$f" > "$$f.sha256"; done
	cp get-punchtape.sh dist/
	@echo "dist/: $$(ls dist | tr '\n' ' ')"

# The code maps, regenerated straight from the code. Text carries the
# completeness — the full call list and the type sheets; two small
# diagrams carry the shape — the package graph and the protocol's life
# cycle. Nothing here fights a dense layout: every output fits a page.
CALLGRAPH := golang.org/x/tools/cmd/callgraph@v0.52.0
GODEPGRAPH := github.com/kisielk/godepgraph@v1.0.0
GOPLANTUML := github.com/jfeliu007/goplantuml/cmd/goplantuml@v1.6.3
MAPS := docs/maps

maps:
	mkdir -p $(MAPS)
	go run $(CALLGRAPH) -format=graphviz ./cmd/punchtape > $(MAPS)/calls.edges.full
	grep -E '"(\(\*)?github\.com/neurophant/punchtape[^"]*" -> "(\(\*)?github\.com/neurophant/punchtape[^"]*"' $(MAPS)/calls.edges.full \
	  | grep -v '\.init"' | LC_ALL=C sort > $(MAPS)/calls.edges
	rm $(MAPS)/calls.edges.full
	{ echo "# The call list"; echo; \
	  echo "Every internal call of the machine — caller → callee, sorted."; echo; \
	  sed -e 's/[[:space:]]*--[a-z]*-[0-9:]*-->[[:space:]]*/ → /' -e 's#github\.com/neurophant/punchtape/##g' $(MAPS)/calls.edges; } > $(MAPS)/calls.md
	awk -f tools/maps_packages.awk $(MAPS)/calls.edges > $(MAPS)/packages.dot
	rm $(MAPS)/calls.edges
	{ echo "# The type sheets"; echo; \
	  echo "The machine's types as goplantuml reads them from the code;"; \
	  echo "the source of truth is the code itself."; echo; \
	  echo '```plantuml'; \
	  go run $(GOPLANTUML) -recursive ./internal; \
	  echo '```'; } > $(MAPS)/types.md
	go run $(GODEPGRAPH) -s ./cmd/punchtape \
	  | grep -E '^(digraph|splines|nodesep|ranksep|node \[|edge \[|\})|neurophant/punchtape|\./cmd/punchtape|"github.com/goccy/go-yaml" \[' \
	  | grep -v ' -> "github.com/goccy/go-yaml/' \
	  | sed -e 's#github.com/neurophant/punchtape/##g' -e 's#"github.com/goccy/go-yaml"#"goccy/go-yaml"#g' \
	  | sed -e 's/ color="[^"]*" URL="[^"]*" target="_blank"//' \
	  | sed -e 's/node \[shape="box",style="rounded,filled"\]/node [shape=box, fontname="monospace", fontsize=11]/' \
	  | sed -e 's/edge \[arrowsize="0.5"\]/edge [fontname="monospace", fontsize=9, arrowsize=0.7]/' \
	  | dot -Tsvg -Gmargin=0.3 -o $(MAPS)/deps.svg
	dot -Tsvg $(MAPS)/packages.dot -o $(MAPS)/packages.svg
	dot -Tsvg $(MAPS)/lifecycle.dot -o $(MAPS)/lifecycle.svg
	@echo "maps/: $$(ls $(MAPS) | tr '\n' ' ')"

# Build and install punchtape.
# -buildvcs=false keeps the binary byte-identical regardless of the
# git state.

BINARY := punchtape
COV := /tmp/punchtape-check
COVERPKG := ./cmd/...,./internal/...
# The battery's speed ceiling. Its ONLY job is to catch an accidental
# blow-up — a hang, a sleep, a quadratic walk — before the field
# does. It is a servant of completeness, never its judge: tests are
# never trimmed or weakened for this clock. A battery that outgrows
# the ceiling on real invariant work is a fact to raise, and only an
# explicit direct order moves the ceiling.
BATTERY_LIMIT := 120s

.PHONY: build install clean vet fmt dist check docs-parity

# The user docs' mirror parity: every Russian and Chinese mirror
# structurally exact against its English base — the law the gate's
# own header states.
docs-parity:
	.venv/bin/python tools/docs_parity.py

# The whole battery's gate: the analyzers, the full test suite, and
# the two standing laws — every non-test statement executed by the
# battery (the instrumented binary's own launch covers the entry
# point a unit test cannot reach), and the code's size ceiling. The
# two coverage sources merge by the block's maximum count: the same
# block seen by several test binaries counts once, as executed.
check: docs-parity
	test -z "$$(gofmt -l .)"
	go vet ./...
	staticcheck ./...
	diag=$$(golint ./...); test -z "$$diag" || { echo "$$diag"; exit 1; }
	diag=$$(unparam ./...); test -z "$$diag" || { echo "$$diag"; exit 1; }
	diag=$$(gopls check $$(git ls-files '*.go')); test -z "$$diag" || { echo "$$diag"; exit 1; }
	dupl -threshold 100 $$(git ls-files '*.go' | grep -v '_test\.go$$') | grep -q '^Found total 0 clone' || { echo 'duplicate code found'; exit 1; }
	start=$$(date +%s); \
	go test -count=1 ./... -coverpkg=$(COVERPKG) -coverprofile=$(COV).test.out || exit 1; \
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
	@bad=$$(go tool cover -func=$(COV).merged.out | grep -v '100.0%' | grep -c . || true); \
	test $$bad -eq 0 || { echo "coverage: $$bad function(s) below 100%"; exit 1; }
	@total=$$(git ls-files '*.go' | grep -v '_test\.go$$' | xargs wc -l | tail -1 | awk '{print $$1}'); \
	echo "check: green (coverage 100%, $$total non-test lines, battery $$(cat $(COV).elapsed) s of $(BATTERY_LIMIT))"

build:
	go build -trimpath -buildvcs=false -o $(BINARY) ./cmd/punchtape

install:
	./install.sh

vet:
	go vet ./...

fmt:
	gofmt -l -w .

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

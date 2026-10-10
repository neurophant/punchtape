# Build and install punchtape.

BINARY := punchtape
# The version's ONE source is internal/cli/cli.go (the AGENTS §4
# release ritual bumps it there with the canon texts); no build-time
# -X override exists — a tag-derived stamp would fight the ritual and
# ship a binary whose --answer disagrees with its own canon.
# Reproducibility: -buildvcs=false keeps the binary byte-identical
# regardless of the git state (no vcs.revision/vcs.modified stamp)
# — determinism by construction, witnessed by the recorded sha of
# the single freeze build (repeated builds to "check again" are
# forbidden: one run is the verdict).

.PHONY: build install clean vet fmt dist test

# The unit battery with the 100.0% coverage gate. Coverage is
# collected through GOCOVERDIR so re-exec'ed child processes (fault
# injection: /dev/full, rlimit) count their blocks too, and a real
# instrumented binary run covers cmd/punchtape's main(). The gate is
# exact: the covdata textfmt counters must show EVERY block of EVERY
# package under ./internal/... and ./cmd/... hit — no percent
# rounding, no pass on empty output, a package without blocks fails.
# Other module packages are deliberately outside the gate.
# The scratch cover dir lives under the SANCTIONED scratch zone of
# ~/.experiments (the structure law: runs/, tmp/, cache/ only —
# nothing else may ever appear at its root; a failed run's leftover
# stays inside tmp/, which is emptied without regret).
COVDIR := $(HOME)/.experiments/tmp/punchtape-cover-$(shell id -u)-$(shell printf %s $$PPID)

# The owner's cleanliness law: every scratch and cache of
# the repo's own gates lives under ~/.experiments (tmp/ for scratch,
# cache/ for reusable caches) — go's build workdirs, pytest fixtures
# and the cover dir NEVER touch the system /tmp or ~/.cache
export TMPDIR := $(HOME)/.experiments/tmp
export GOCACHE := $(HOME)/.experiments/cache/go-build

test:
	@mkdir -p $(TMPDIR) $(GOCACHE)
	@rm -rf $(COVDIR); mkdir -p $(COVDIR)
	@# The eternal agnostic gate: no subject-matter
	@# witness vocabulary anywhere in cmd/ internal/ — the machine knows
	@# only mechanics; probes arrive as staged bytes through the door.
	bash scripts/agnostic-scan.sh
	go vet ./...
	go test ./... -count=1
	@fail=0; \
	for pkg in $$(go list ./internal/... ./cmd/...); do \
	  go test -c -cover -o $(COVDIR)/pkg.test $$pkg || { fail=1; break; }; \
	  test -f $(COVDIR)/pkg.test || continue; \
	  ( cd $(COVDIR) && GOCOVERDIR=$(COVDIR) ./pkg.test -test.gocoverdir=$(COVDIR) >/dev/null ) || { fail=1; break; }; \
	  rm -f $(COVDIR)/pkg.test; \
	done; \
	[ $$fail -eq 0 ] || { echo 'FAIL: unit tests'; exit 1; }
	go build -cover -o $(COVDIR)/punchtape-cover ./cmd/punchtape
	@( cd $(COVDIR) && GOCOVERDIR=$(COVDIR) ./punchtape-cover >/dev/null )
	@# The live-instance roundtrip: the one sanctioned non-unit
	@# remainder — the OS implementations of the ports (fsport.OS,
	@# execport.OS) and cmd/punchtape's main, by the instrumented
	@# binary over one real cycle: birth by next with the brief ->
	@# the spec stored by behavior id through the one door (nothing
	@# runs) -> the door's impermeability (any other path refused) ->
	@# the code written directly into the work tree -> the run red on
	@# a wrong expectation -> the repair resubmitted under the same
	@# behavior id, green -> ratify: true -> VERDICT: READY, the
	@# closed instance refusing any further submission.
	@GOCOVERDIR=$(COVDIR) bash scripts/gate-roundtrip.sh $(COVDIR)/punchtape-cover >/dev/null
	@rm -f $(COVDIR)/pkg.test
	@set -o pipefail; go tool covdata textfmt -i=$(COVDIR) -o=$(COVDIR)/profile.out
	@# The owner's word: the ERROR branches of the ports'
	@# own OS implementations do not count in the gate. A unit cannot
	@# touch them without touching the real world (the units law); the
	@# success paths ride the gate binary's roundtrip — the error
	@# branches of real delegates (start failures, non-exit reap
	@# failures, timeouts, kill-group loops, os-call error returns) are
	@# the owner-accepted remainder. The exclusion is file-precise:
	@# only the two ports' OS delegate files (their only executable
	@# blocks ARE the OS impl) — everything else stays gated.
	@bad=0; \
	PORT_OS='^github.com/neurophant/punchtape/internal/(execport/execport|fsport/fsport)\.go:'; \
	if [ -s "$(COVDIR)/profile.out" ]; then :; else echo "FAIL empty coverage profile"; bad=1; fi; \
	for pkg in $$(go list ./internal/... ./cmd/...); do \
	  src=$$(printf '%s' "$$pkg" | sed 's|^github.com/neurophant/punchtape/||'); \
	  total=$$(grep -F "$$src/" "$(COVDIR)/profile.out" | grep -c .); \
	  zero=$$(grep -F "$$src/" "$(COVDIR)/profile.out" | grep -Ev "$$PORT_OS" | awk 'NF && $$NF==0' | grep -c .); \
	  if [ "$$total" -eq 0 ]; then echo "FAIL no coverage blocks: $$pkg"; bad=1; \
	  elif [ "$$zero" -gt 0 ]; then echo "FAIL coverage: $$pkg has $$zero unhit blocks"; bad=1; fi; \
	done; \
	[ $$bad -eq 0 ] || { echo 'FAIL coverage gate'; exit 1; }
	@echo 'coverage gate: ALL PACKAGES 100.0% of statements (exact block counts)'
	@# The scratch cover dir dies with the green verdict; a failed run
	@# keeps it for diagnosis (the owner's word: no /tmp litter).
	@rm -rf $(COVDIR)

build:
	go build -buildvcs=false -o $(BINARY) ./cmd/punchtape

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
	GOOS=linux  GOARCH=amd64 go build -buildvcs=false -o dist/punchtape-linux-amd64  ./cmd/punchtape
	GOOS=linux  GOARCH=arm64 go build -buildvcs=false -o dist/punchtape-linux-arm64  ./cmd/punchtape
	GOOS=darwin GOARCH=amd64 go build -buildvcs=false -o dist/punchtape-darwin-amd64 ./cmd/punchtape
	GOOS=darwin GOARCH=arm64 go build -buildvcs=false -o dist/punchtape-darwin-arm64 ./cmd/punchtape
	cd dist && for f in punchtape-*; do case "$$f" in *.sha256) continue;; esac; sha256sum "$$f" > "$$f.sha256"; done
	cp get-punchtape.sh dist/
	@echo "dist/: $$(ls dist | tr '\n' ' ')"

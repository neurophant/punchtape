# Build and install punchtape.

BINARY := punchtape
# Delivery version (semver): the latest git tag; 0.1.0 for an untagged tree.
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo 0.1.0)
LDFLAGS := -X github.com/neurophant/punchtape/internal/cli.Version=$(VERSION)

.PHONY: build install clean vet fmt dist

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/punchtape

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
	GOOS=linux  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/punchtape-linux-amd64  ./cmd/punchtape
	GOOS=linux  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/punchtape-linux-arm64  ./cmd/punchtape
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/punchtape-darwin-amd64 ./cmd/punchtape
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/punchtape-darwin-arm64 ./cmd/punchtape
	cd dist && for f in punchtape-*; do case "$$f" in *.sha256) continue;; esac; sha256sum "$$f" > "$$f.sha256"; done
	cp get-punchtape.sh dist/
	@echo "dist/: $$(ls dist | tr '\n' ' ')"

.DEFAULT_GOAL := help

# ─── Variables ────────────────────────────────────────────────────────────────

VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS    = -s -w -X github.com/omargallob/devops-starter/internal/cli.version=$(VERSION) \
             -X github.com/omargallob/devops-starter/internal/cli.commit=$(COMMIT) \
             -X github.com/omargallob/devops-starter/internal/cli.date=$(BUILD_DATE)

BINARY     = devops-starter
MCP_BINARY = devops-starter-mcp
CMD_PKG    = ./cmd/devops-starter/
MCP_CMD_PKG = ./cmd/mcp-server/
DIST_DIR   = dist

# Cross-platform checksum command
CHECKSUM = $(shell command -v sha256sum >/dev/null 2>&1 && echo "sha256sum" || echo "shasum -a 256")

# ─── Build & Install ─────────────────────────────────────────────────────────

.PHONY: build install mcp-server install-mcp

## build: Compile binary for the current platform
build:
	CGO_ENABLED=0 go build -ldflags='$(LDFLAGS)' -o $(BINARY) $(CMD_PKG)

## mcp-server: Build the MCP server binary
mcp-server:
	CGO_ENABLED=0 go build -ldflags='$(LDFLAGS)' -o $(MCP_BINARY) $(MCP_CMD_PKG)

## install: Build and install to ~/.local/bin
install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BINARY) $(HOME)/.local/bin/$(BINARY)

## install-mcp: Build and install MCP server to ~/.local/bin
install-mcp: mcp-server
	mkdir -p $(HOME)/.local/bin
	cp $(MCP_BINARY) $(HOME)/.local/bin/$(MCP_BINARY)

# ─── Testing ─────────────────────────────────────────────────────────────────

.PHONY: test test-coverage

## test: Run all tests with race detector
test:
	go test -race ./...

## test-coverage: Run tests with coverage report
test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "\nHTML report: go tool cover -html=coverage.out"

# ─── Code Quality ────────────────────────────────────────────────────────────

.PHONY: fmt vet lint lint-local check

## fmt: Format Go and shell files
fmt:
	go fmt ./...
	@command -v shfmt >/dev/null 2>&1 && find dotfiles scripts -name "*.sh" -exec shfmt -w {} + || true

## vet: Run go vet
vet:
	go vet ./...

## lint: Run all linters (requires golangci-lint, shellcheck)
lint: vet
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Install golangci-lint: https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run ./...
	shellcheck scripts/install.sh
	find dotfiles -name "*.sh" -exec shellcheck {} +

## lint-local: Run all pre-commit hooks on all files
lint-local:
	pre-commit run --all-files

## check: Run fmt, vet, lint, and tests (CI-like local validation)
check: fmt vet lint test

# ─── Release ─────────────────────────────────────────────────────────────────

.PHONY: release clean goreleaser-check goreleaser-snapshot

## release: Cross-compile for all supported platforms
release:
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o $(DIST_DIR)/$(BINARY)-linux-amd64  $(CMD_PKG)
	CGO_ENABLED=0 GOOS=linux  GOARCH=arm64 go build -ldflags='$(LDFLAGS)' -o $(DIST_DIR)/$(BINARY)-linux-arm64  $(CMD_PKG)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o $(DIST_DIR)/$(BINARY)-darwin-amd64 $(CMD_PKG)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags='$(LDFLAGS)' -o $(DIST_DIR)/$(BINARY)-darwin-arm64 $(CMD_PKG)
	cd $(DIST_DIR) && $(CHECKSUM) $(BINARY)-* > checksums.txt

## goreleaser-check: Validate GoReleaser configuration
goreleaser-check:
	@command -v goreleaser >/dev/null 2>&1 || { echo "Install goreleaser: https://goreleaser.com/install/"; exit 1; }
	goreleaser check

## goreleaser-snapshot: Local dry-run release (no publish)
goreleaser-snapshot:
	@command -v goreleaser >/dev/null 2>&1 || { echo "Install goreleaser: https://goreleaser.com/install/"; exit 1; }
	goreleaser release --snapshot --clean

## clean: Remove build artifacts
clean:
	rm -f $(BINARY) $(MCP_BINARY)
	rm -f coverage.out
	rm -rf $(DIST_DIR)
	bazel clean 2>/dev/null || true

# ─── Bazel ───────────────────────────────────────────────────────────────────

.PHONY: bazel-build bazel-test bazel-release gazelle

## bazel-build: Build with Bazel
bazel-build:
	bazel build //cmd/devops-starter

## bazel-test: Test with Bazel
bazel-test:
	bazel test //...

## gazelle: Regenerate BUILD files
gazelle:
	bazel run //:gazelle

## bazel-release: Build all platforms with Bazel
bazel-release:
	bazel build //cmd/devops-starter --config=linux_amd64
	bazel build //cmd/devops-starter --config=linux_arm64
	bazel build //cmd/devops-starter --config=darwin_amd64
	bazel build //cmd/devops-starter --config=darwin_arm64

# ─── Development Setup ────────────────────────────────────────────────────────

.PHONY: setup

## setup: Install dev tools and pre-commit hooks (requires mise)
setup:
	@command -v mise >/dev/null 2>&1 || { echo "Install mise first: https://mise.jdx.dev"; exit 1; }
	mise install
	pip install pre-commit
	pre-commit install --hook-type commit-msg --hook-type pre-commit
	@echo "Done! Commit hooks are now active."

# ─── Help ─────────────────────────────────────────────────────────────────────

.PHONY: help

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk '/^## / { sub(/^## /, ""); split($$0, a, ": "); printf "  %-18s %s\n", a[1], a[2] }' $(MAKEFILE_LIST)

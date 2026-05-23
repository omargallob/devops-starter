.PHONY: build test lint clean release gazelle fmt vet install

# Default Go build
build:
	go build -o devops-starter ./cmd/devops-starter/

# Run all tests with race detector
test:
	go test -race ./...

# Run linters
lint: vet
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Install golangci-lint: https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run ./...
	shellcheck scripts/install.sh
	find dotfiles -name "*.sh" -exec shellcheck {} +

# Go vet
vet:
	go vet ./...

# Format code
fmt:
	go fmt ./...
	@command -v shfmt >/dev/null 2>&1 && find dotfiles scripts -name "*.sh" -exec shfmt -w {} + || true

# Clean build artifacts
clean:
	rm -f devops-starter
	rm -rf dist/
	bazelisk clean 2>/dev/null || true

# ─── Bazel ───────────────────────────────────────────────────────────────────

# Build with Bazel
bazel-build:
	bazelisk build //cmd/devops-starter

# Test with Bazel
bazel-test:
	bazelisk test //...

# Regenerate BUILD files
gazelle:
	bazelisk run //:gazelle

# Build all platforms with Bazel
bazel-release:
	bazelisk build //cmd/devops-starter --config=linux_amd64
	bazelisk build //cmd/devops-starter --config=linux_arm64
	bazelisk build //cmd/devops-starter --config=darwin_amd64
	bazelisk build //cmd/devops-starter --config=darwin_arm64

# ─── Release (without Bazel) ─────────────────────────────────────────────────

# Cross-compile for all platforms
release:
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/devops-starter-linux-amd64 ./cmd/devops-starter/
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/devops-starter-linux-arm64 ./cmd/devops-starter/
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/devops-starter-darwin-amd64 ./cmd/devops-starter/
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/devops-starter-darwin-arm64 ./cmd/devops-starter/
	cd dist && sha256sum devops-starter-* > checksums.txt

# Install locally
install: build
	mkdir -p $(HOME)/.local/bin
	cp devops-starter $(HOME)/.local/bin/devops-starter

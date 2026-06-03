#!/usr/bin/env bash
# Runs once after devcontainer creation, as the vscode user.
set -euo pipefail

WORKSPACE="/workspaces/devops-starter"
cd "${WORKSPACE}"

echo "==> [1/5] Installing runtimes via mise (Go 1.26.3, Python 3.12, Node 22)..."
mise install

echo "==> [2/5] Activating mise in shell profile..."
MISE_HOOK='eval "$(mise activate bash)"'
grep -qF "${MISE_HOOK}" "${HOME}/.bashrc" 2>/dev/null || echo "${MISE_HOOK}" >> "${HOME}/.bashrc"

echo "==> [3/5] Installing golangci-lint into ~/.local/bin..."
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
  | sh -s -- -b "${HOME}/.local/bin" v2.1.6

echo "==> [4/5] Installing pre-commit and git hooks..."
pip install --quiet pre-commit
mise exec -- pre-commit install --hook-type commit-msg --hook-type pre-commit

echo "==> [5/5] Pre-warming Go module cache and Bazel..."
mise exec -- go mod download
bazel version

echo ""
echo "Done. Run 'make check' to verify the environment."

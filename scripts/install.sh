#!/bin/sh
# Bootstrap installer for devops-starter.
# Downloads the correct pre-built binary for your platform from GitHub Releases.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/omargallob/devops-starter/main/scripts/install.sh | sh
#
set -e

REPO="omargallob/devops-starter"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="devops-starter"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    linux)  OS="linux" ;;
    darwin) OS="darwin" ;;
    *)      echo "Error: Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)   ARCH="amd64" ;;
    arm64|aarch64)  ARCH="arm64" ;;
    *)              echo "Error: Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Get latest version from GitHub
get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' \
        | sed -E 's/.*"v?([^"]+)".*/\1/'
}

VERSION="${VERSION:-$(get_latest_version)}"
if [ -z "$VERSION" ]; then
    echo "Error: Could not determine latest version" >&2
    exit 1
fi

# Construct download URL
URL="https://github.com/${REPO}/releases/download/v${VERSION}/${BINARY_NAME}-${OS}-${ARCH}"

echo "Installing devops-starter v${VERSION} (${OS}/${ARCH})..."
echo "  From: ${URL}"
echo "  To:   ${INSTALL_DIR}/${BINARY_NAME}"

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download binary
if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$URL" -o "${INSTALL_DIR}/${BINARY_NAME}"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$URL" -O "${INSTALL_DIR}/${BINARY_NAME}"
else
    echo "Error: curl or wget is required" >&2
    exit 1
fi

# Make executable
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

echo ""
echo "Successfully installed devops-starter v${VERSION}!"
echo ""

# Check if install dir is in PATH
case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
        echo "WARNING: ${INSTALL_DIR} is not in your PATH."
        echo "Add this to your shell profile:"
        echo ""
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        echo ""
        ;;
esac

echo "Run 'devops-starter doctor' to verify your setup."
echo "Run 'devops-starter install' to install all tools."

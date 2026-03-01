#!/usr/bin/env bash
set -euo pipefail

REPO="jansuthacheeva/ccpetline"
INSTALL_DIR="${HOME}/.local/bin"

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)
    echo "Error: unsupported OS: $OS" >&2
    exit 1
    ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)       ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Error: unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# Resolve version
if [ -z "${VERSION:-}" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')"
  if [ -z "$VERSION" ]; then
    echo "Error: could not determine latest version" >&2
    exit 1
  fi
fi

TARBALL="ccpetline-${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"

echo "Installing ccpetline ${VERSION} (${OS}/${ARCH})..."

# Download and extract
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$URL" -o "${TMPDIR}/${TARBALL}"
tar -xzf "${TMPDIR}/${TARBALL}" -C "$TMPDIR"

# Install binaries
mkdir -p "$INSTALL_DIR"
for bin in ccpetline ccpetline-hook ccpetline-config; do
  cp "${TMPDIR}/${bin}" "${INSTALL_DIR}/${bin}"
  chmod +x "${INSTALL_DIR}/${bin}"
done

# Check PATH
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "Warning: ${INSTALL_DIR} is not in your PATH."
    echo "Add it with:  export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac

echo ""
echo "ccpetline ${VERSION} installed successfully!"
echo "Run 'ccpetline-config' to configure your pet."

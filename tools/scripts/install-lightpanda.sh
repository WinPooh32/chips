#!/usr/bin/env bash
set -euo pipefail

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
if [[ "$OS" == "darwin" ]]; then
  OS="macos"
elif [[ "$OS" == "linux" ]]; then
  OS="linux"
else
  echo "Unsupported OS: $OS" >&2
  exit 1
fi

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="x86_64" ;;
  aarch64|arm64) ARCH="aarch64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

TARGET="${ARCH}-${OS}"
REPO="lightpanda-io/browser"
BIN_DIR="$(cd "$(dirname "$0")/.." && pwd)/bin"
mkdir -p "$BIN_DIR"

# Get latest release version (skip nightly)
VERSION="$(curl -s "https://api.github.com/repos/${REPO}/releases" | jq -r '.[0].tag_name')"
if [[ -z "$VERSION" || "$VERSION" == "null" ]]; then
  echo "Failed to get latest version" >&2
  exit 1
fi

FILENAME="lightpanda-${TARGET}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

echo "Downloading lightpanda ${VERSION} (${TARGET})..."
TMPFILE="$(mktemp)"

if ! curl -fsSL "$URL" -o "$TMPFILE"; then
  echo "Failed to download: $URL" >&2
  rm -f "$TMPFILE"
  exit 1
fi

mv "$TMPFILE" "$BIN_DIR/lightpanda"
chmod +x "$BIN_DIR/lightpanda"

echo "Installed lightpanda ${VERSION} to ${BIN_DIR}/lightpanda"

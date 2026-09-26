#!/usr/bin/env bash
set -euo pipefail

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
if [[ "$OS" == "darwin" ]]; then
  OS="darwin"
elif [[ "$OS" == "linux" ]]; then
  OS="linux"
else
  echo "Unsupported OS: $OS" >&2
  exit 1
fi

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

TARGET="${OS}_${ARCH}"
REPO="gastownhall/beads"
BIN_DIR="$(cd "$(dirname "$0")/.." && pwd)/bin"
mkdir -p "$BIN_DIR"

# Get latest release version
VERSION="$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | jq -r .tag_name)"
if [[ -z "$VERSION" || "$VERSION" == "null" ]]; then
  echo "Failed to get latest version" >&2
  exit 1
fi

# Strip 'v' prefix for filename
VERSION_NUM="${VERSION#v}"

# Determine archive format
if [[ "$OS" == "windows" ]]; then
  EXT="zip"
else
  EXT="tar.gz"
fi

FILENAME="beads_${VERSION_NUM}_${TARGET}.${EXT}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

echo "Downloading beads ${VERSION} (${TARGET})..."
TMPFILE="$(mktemp)"

if ! curl -fsSL "$URL" -o "$TMPFILE"; then
  echo "Failed to download: $URL" >&2
  rm -f "$TMPFILE"
  exit 1
fi

# Extract and install
if [[ "$EXT" == "zip" ]]; then
  unzip -q -o "$TMPFILE" -d "$BIN_DIR"
else
  tar -xzf "$TMPFILE" -C "$BIN_DIR"
fi

rm -f "$TMPFILE"
chmod +x "$BIN_DIR/bd"

echo "Installed beads ${VERSION} to ${BIN_DIR}/bd"

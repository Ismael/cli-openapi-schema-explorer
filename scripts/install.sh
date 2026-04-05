#!/usr/bin/env bash
set -euo pipefail

# Determine the plugin root (parent of scripts/)
PLUGIN_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BIN_DIR="${PLUGIN_DIR}/bin"
REPO="Ismael/cli-openapi-schema-explorer"
BINARY_NAME="openapi-explorer"

mkdir -p "$BIN_DIR"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  linux)   OS="linux" ;;
  darwin)  OS="darwin" ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

SUFFIX=""
if [ "$OS" = "windows" ]; then
  SUFFIX=".exe"
fi

# Get latest release tag from GitHub API
LATEST_TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST_TAG" ]; then
  echo "Error: could not determine latest release" >&2
  exit 1
fi

ASSET_NAME="${BINARY_NAME}_${OS}_${ARCH}${SUFFIX}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${ASSET_NAME}"

echo "Downloading ${BINARY_NAME} ${LATEST_TAG} for ${OS}/${ARCH}..."
curl -fsSL -o "${BIN_DIR}/${BINARY_NAME}${SUFFIX}" "$DOWNLOAD_URL"
chmod +x "${BIN_DIR}/${BINARY_NAME}${SUFFIX}"

echo "Installed ${BINARY_NAME} ${LATEST_TAG} to ${BIN_DIR}"

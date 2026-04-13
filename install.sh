#!/bin/sh
set -e

# rc CLI installer
# Usage: curl -fsSL https://raw.githubusercontent.com/andresdefi/Revenue-Cat-CLI/main/install.sh | sh

REPO="andresdefi/Revenue-Cat-CLI"
BINARY="rc"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    darwin) OS="darwin" ;;
    linux)  OS="linux" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Get latest version
VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest version"
    exit 1
fi

VERSION_NUM="${VERSION#v}"

# Build download URL
EXT="tar.gz"
if [ "$OS" = "windows" ]; then
    EXT="zip"
fi
FILENAME="${BINARY}_${VERSION_NUM}_${OS}_${ARCH}.${EXT}"
URL="https://github.com/$REPO/releases/download/$VERSION/$FILENAME"

echo "Downloading rc $VERSION for $OS/$ARCH..."

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$URL" -o "$TMPDIR/$FILENAME"

# Extract
cd "$TMPDIR"
if [ "$EXT" = "tar.gz" ]; then
    tar xzf "$FILENAME"
elif [ "$EXT" = "zip" ]; then
    unzip -q "$FILENAME"
fi

# Install
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    echo "Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv "$BINARY" "$INSTALL_DIR/$BINARY"
else
    mv "$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"

echo ""
echo "rc $VERSION installed to $INSTALL_DIR/$BINARY"
echo ""
echo "Get started:"
echo "  rc auth login"
echo "  rc projects list"

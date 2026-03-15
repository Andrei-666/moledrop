#!/bin/bash
set -e

REPO="Andrei-666/moledrop"
BINARY="mole"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case $OS in
    linux)  EXT="" ;;
    darwin) EXT="" ;;
    *)
        echo "Unsupported OS: $OS. On Windows, download the binary from:"
        echo "https://github.com/$REPO/releases/latest"
        exit 1
        ;;
esac

# Get latest release version from GitHub API
VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' \
    | sed 's/.*"tag_name": *"\(.*\)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "Could not fetch latest version. Check your internet connection."
    exit 1
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/mole-${OS}-${ARCH}${EXT}"

echo "Installing MoleDrop $VERSION for ${OS}/${ARCH}..."
curl -fsSL "$DOWNLOAD_URL" -o "/tmp/$BINARY"
chmod +x "/tmp/$BINARY"
mv "/tmp/$BINARY" "$INSTALL_DIR/$BINARY"

echo ""
echo "✅ MoleDrop installed successfully!"
echo "   Run: mole send <file>"
echo "   Run: mole receive <code>"

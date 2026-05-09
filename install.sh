#!/usr/bin/env bash
set -e

# Configuration
REPO="waiyneee/Simplesearch"
BINARY_NAME="simplesearch"
INSTALL_DIR="/usr/local/bin"

echo "========================================="
echo "🚀 Installing Simplesearch..."
echo "========================================="

# 1. Detect Operating System
OS="$(uname -s)"
case "${OS}" in
    Linux*)     PLATFORM="Linux";;
    Darwin*)    PLATFORM="Darwin";;
    *)          echo "❌ Unsupported OS: ${OS}"; exit 1;;
esac

# 2. Detect Architecture
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64*)    ARCHITECTURE="x86_64";;
    aarch64*|arm64*) ARCHITECTURE="arm64";;
    *)          echo "❌ Unsupported architecture: ${ARCH}"; exit 1;;
esac

echo "✅ Detected Platform: ${PLATFORM} (${ARCHITECTURE})"

# 3. Fetch the latest release version from GitHub API
echo "🔍 Finding latest release..."
LATEST_TAG=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    echo "❌ Could not fetch latest release. Please check your internet connection or GitHub API limits."
    exit 1
fi

echo "📦 Latest version is ${LATEST_TAG}"

# 4. Construct the download URL (Matching your GoReleaser name_template)
# e.g., simplesearch_Linux_x86_64.tar.gz
TAR_FILE="${BINARY_NAME}_${PLATFORM}_${ARCHITECTURE}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${TAR_FILE}"

echo "⬇️  Downloading from ${DOWNLOAD_URL}..."

# 5. Download and Extract
TMP_DIR=$(mktemp -d)
curl -sL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${TAR_FILE}"

echo "📂 Extracting binary..."
tar -xzf "${TMP_DIR}/${TAR_FILE}" -C "${TMP_DIR}"

# 6. Install the binary
echo "🔑 Requesting sudo permissions to move binary to ${INSTALL_DIR}..."
sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# 7. Cleanup
rm -rf "${TMP_DIR}"

echo "========================================="
echo "🎉 Simplesearch installed successfully!"
echo "========================================="
echo "You can now run the tool from anywhere by typing:"
echo "  simplesearch search -q \"your query\""
echo ""
echo "Note: Simplesearch has an embedded SQLite database and a built-in memory cache."
echo "If you want to use Redis for persistent caching, ensure Redis is installed and running,"
echo "then run: simplesearch search --cache redis -q \"...\""
echo "========================================="
#!/usr/bin/env bash
set -euo pipefail

# =========================================================
# Simplesearch Installer
# =========================================================

BINARY_NAME="simplesearch"
INSTALL_DIR="/usr/local/bin"
BUILD_DIR="$(mktemp -d)"

# =========================================================
# Colors
# =========================================================

RED='\033[1;31m'
GREEN='\033[1;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;34m'
CYAN='\033[1;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

# =========================================================
# Helpers
# =========================================================

phase() {
    echo ""
    echo -e "${BLUE}building${NC} ${CYAN}$1${NC}"
}

ok() {
    echo -e "${GREEN}✓${NC} $1"
}

warn() {
    echo -e "${YELLOW}warning:${NC} $1"
}

fail() {
    echo -e "${RED}error:${NC} $1"
    exit 1
}

cleanup() {
    rm -rf "${BUILD_DIR}"
}

trap cleanup EXIT

# =========================================================
# Header
# =========================================================

echo -e "${CYAN}${BOLD}"
echo "Simplesearch Installer"
echo -e "${NC}"

# =========================================================
# checkPhase
# =========================================================

phase "checkPhase"

# Go
if ! command -v go >/dev/null 2>&1; then
    fail "Go is not installed"
fi

ok "found Go"
echo -e "${DIM}$(go version)${NC}"

# go.mod
if [ ! -f "go.mod" ]; then
    fail "go.mod not found"
fi

ok "found go.mod"

# cmd directory
if [ ! -d "./cmd" ]; then
    fail "./cmd directory not found"
fi

ok "found cmd directory"

# Detect main package
if [ -d "./cmd/Simplesearch" ]; then
    BUILD_PATH="./cmd/Simplesearch"
elif [ -d "./cmd/simplesearch" ]; then
    BUILD_PATH="./cmd/simplesearch"
else
    fail "could not find main package under ./cmd"
fi

ok "detected main package"
echo -e "${DIM}${BUILD_PATH}${NC}"

# main.go
if ! find "${BUILD_PATH}" -name "main.go" | grep -q .; then
    fail "main.go not found in ${BUILD_PATH}"
fi

ok "found main.go"

# install dir
if [ ! -d "${INSTALL_DIR}" ]; then
    warn "${INSTALL_DIR} does not exist"
    echo -e "${DIM}creating ${INSTALL_DIR}${NC}"

    sudo mkdir -p "${INSTALL_DIR}"
fi

ok "install directory ready"

echo -e "${DIM}${INSTALL_DIR}${NC}"
# SQLite
if grep -qi "sqlite" go.mod; then
    ok "SQLite support enabled"
else
    warn "SQLite dependency not detected"
fi

# Redis
if command -v redis-server >/dev/null 2>&1; then
    ok "Redis installed"

    if command -v redis-cli >/dev/null 2>&1 && redis-cli ping >/dev/null 2>&1; then
        ok "Redis server running"
    else
        warn "Redis installed but not running"
    fi
else
    warn "Redis not installed (optional)"
fi

# =========================================================
# buildPhase
# =========================================================

phase "buildPhase"

echo -e "${DIM}building in ${BUILD_DIR}${NC}"

go build \
    -trimpath \
    -ldflags="-s -w" \
    -o "${BUILD_DIR}/${BINARY_NAME}" \
    "${BUILD_PATH}"

ok "build completed"

# =========================================================
# installPhase
# =========================================================

phase "installPhase"

sudo install -m 755 \
    "${BUILD_DIR}/${BINARY_NAME}" \
    "${INSTALL_DIR}/${BINARY_NAME}"

ok "installed binary"
echo -e "${DIM}${INSTALL_DIR}/${BINARY_NAME}${NC}"

# =========================================================
# finalPhase
# =========================================================

phase "finalPhase"

ok "Simplesearch installed successfully"

echo ""
echo -e "${BOLD}usage${NC}"
echo '  simplesearch search -q "your query"'

echo ""
echo -e "${BOLD}redis cache${NC}"
echo '  simplesearch search --cache redis -q "your query"'

echo ""
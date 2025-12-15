#!/bin/sh
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the latest release tag or use main
VERSION=${1:-main}
REPO="0xjuanma/golazo"
BINARY_NAME="golazo"

echo "${GREEN}Installing ${BINARY_NAME}...${NC}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "${RED}Unsupported architecture: ${ARCH}${NC}"; exit 1 ;;
esac

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

# Clone repository
echo "${YELLOW}Cloning repository...${NC}"
git clone --depth 1 --branch ${VERSION} https://github.com/${REPO}.git ${TMP_DIR}/golazo 2>/dev/null || \
git clone --depth 1 https://github.com/${REPO}.git ${TMP_DIR}/golazo

cd ${TMP_DIR}/golazo

# Build the binary
echo "${YELLOW}Building ${BINARY_NAME}...${NC}"
go build -o ${BINARY_NAME} .

# Determine install location
if [ -w /usr/local/bin ]; then
    INSTALL_DIR="/usr/local/bin"
elif [ -w ~/.local/bin ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p ${INSTALL_DIR}
else
    INSTALL_DIR="$HOME/bin"
    mkdir -p ${INSTALL_DIR}
fi

# Install the binary
echo "${YELLOW}Installing to ${INSTALL_DIR}...${NC}"
cp ${BINARY_NAME} ${INSTALL_DIR}/${BINARY_NAME}
chmod +x ${INSTALL_DIR}/${BINARY_NAME}

# Check if the binary is in PATH
if ! command -v ${BINARY_NAME} >/dev/null 2>&1; then
    echo "${YELLOW}Warning: ${BINARY_NAME} may not be in your PATH.${NC}"
    echo "${YELLOW}Add ${INSTALL_DIR} to your PATH if needed.${NC}"
fi

echo "${GREEN}${BINARY_NAME} installed successfully!${NC}"
echo "${GREEN}Run '${BINARY_NAME}' to start.${NC}"

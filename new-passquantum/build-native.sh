#!/bin/bash
# Simple native build for the current platform.

set -euo pipefail

APP_NAME="PassQuantum"
BUILD_DIR="build/linux"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Building PassQuantum for current platform...${NC}"
mkdir -p "$BUILD_DIR"
CGO_ENABLED=1 go build -o "$BUILD_DIR/$APP_NAME" ./ui

echo -e "${GREEN}Build complete.${NC}"
echo "Output: ./$BUILD_DIR/$APP_NAME"
echo -e "${YELLOW}Run:${NC} ./$BUILD_DIR/$APP_NAME"

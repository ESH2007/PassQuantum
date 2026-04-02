#!/bin/bash
# Simple native build without Docker (no cross-compilation)
#
# PREREQUISITES — OpenCV 4 development libraries are required because the
# biometric (GoCV) package uses CGo to link against OpenCV:
#
#   Ubuntu / Debian:  sudo apt-get install libopencv-dev pkg-config
#   Fedora / RHEL:    sudo dnf install opencv-devel pkgconf
#   macOS (Homebrew): brew install opencv pkg-config
#
# ONNX models must be present at runtime:
#   models/blazeface.onnx
#   models/face_mesh.onnx

set -e

APP_NAME="PassQuantum"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Building PassQuantum for current platform...${NC}"

# Build from project root, targeting ui directory
go build -o "$APP_NAME" ./ui/*.go

echo -e "${GREEN}✓ Build complete!${NC}"
echo ""
echo "Executable created: ./$APP_NAME"
echo ""
echo -e "${YELLOW}To run:${NC} ./$APP_NAME"

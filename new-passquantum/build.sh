#!/bin/bash
# PassQuantum Cross-Platform Build Script
# Easily build for Linux, Windows, and macOS

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

APP_NAME="PassQuantum"
BUILD_DIR="build"
VERSION="1.0.0"
CROSS_BUILD_TAGS="${CROSS_BUILD_TAGS:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}$1${NC}"
    echo -e "${GREEN}================================${NC}"
}

print_info() {
    echo -e "${YELLOW}➜${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if fyne-cross is installed
check_dependencies() {
    print_header "Checking Dependencies"
    
    if ! command -v fyne &> /dev/null; then
        print_info "Installing fyne command..."
        go install fyne.io/fyne/v2/cmd/fyne@latest
    fi
    
    if ! command -v fyne-cross &> /dev/null; then
        print_info "Installing fyne-cross..."
        go install github.com/fyne-io/fyne-cross@latest
    fi
    
    print_success "All dependencies installed"
    echo ""
}

# Check native Linux OpenCV toolchain required by gocv.
check_opencv() {
    if ! pkg-config --exists opencv4; then
        print_error "OpenCV development files not found (pkg-config: opencv4)."
        echo "Install prerequisites, for example on Ubuntu/Debian:"
        echo "  sudo apt-get install -y libopencv-dev pkg-config"
        exit 1
    fi
}

check_linux_models() {
    local blaze="models/blazeface.onnx"
    local mesh="models/face_mesh.onnx"

    for model in "$blaze" "$mesh"; do
        if [[ ! -f "$model" ]]; then
            print_error "Missing model: $model"
            exit 1
        fi
        if grep -q "git-lfs.github.com/spec/v1" "$model" 2>/dev/null; then
            print_error "Invalid model file (git-lfs pointer): $model"
            exit 1
        fi
        if head -c 64 "$model" | grep -qiE '<!doctype html|<html'; then
            print_error "Invalid model file (HTML content): $model"
            exit 1
        fi
    done
}

# Guard cross-builds for CGO/OpenCV projects.
# This repository uses gocv, which requires target OpenCV toolchains that are
# not present in default fyne-cross containers.
check_cross_cgo_requirements() {
    local target="$1"
    if [[ -n "${CROSS_BUILD_TAGS}" ]]; then
        print_info "Using cross-build tags: ${CROSS_BUILD_TAGS}"
    fi
    if [[ "$target" != "linux" ]]; then
        print_info "Non-Linux cross-builds default to '-tags=nobiometric' unless CROSS_BUILD_TAGS is explicitly set."
    fi
    return 0
}

# Build for Linux (native)
build_linux() {
    print_header "Building for Linux (AMD64)"
    mkdir -p "$BUILD_DIR/linux"

    check_opencv
    check_linux_models

    # Native build is the reliable path for gocv/OpenCV projects.
    CGO_ENABLED=1 go build -mod=vendor -o "$BUILD_DIR/linux/$APP_NAME" ./ui

    mkdir -p "$BUILD_DIR/linux/models"
    cp -f models/blazeface.onnx "$BUILD_DIR/linux/models/blazeface.onnx"
    cp -f models/face_mesh.onnx "$BUILD_DIR/linux/models/face_mesh.onnx"
    
    print_success "Linux native build complete: $BUILD_DIR/linux/$APP_NAME"
    echo ""
}

# Build for Windows
build_windows() {
    print_header "Building for Windows (AMD64)"
    mkdir -p "$BUILD_DIR/windows"
    local cross_tags="${CROSS_BUILD_TAGS:-nobiometric}"

    check_cross_cgo_requirements "windows" || return 1

    local tag_args=()
    if [[ -n "${cross_tags}" ]]; then
        tag_args=(-tags="${cross_tags}")
    fi

    if ! fyne-cross windows -arch=amd64 \
        -app-id=com.passquantum.app \
        -name="$APP_NAME" \
        -icon=Icon.png \
        "${tag_args[@]}" \
        ./ui; then
        print_error "Windows build failed"
        return 1
    fi
    
    print_success "Windows build complete"
    print_info "Note: Windows users may need updated graphics drivers for OpenGL support"
    echo ""
}

# Build for macOS
build_macos() {
    print_header "Building for macOS"
    mkdir -p "$BUILD_DIR/mac"
    local cross_tags="${CROSS_BUILD_TAGS:-nobiometric}"

    check_cross_cgo_requirements "mac" || return 1

    local mac_sdk_path="${MACOSX_SDK_PATH:-/home/lenovo/dev/PassQuantum/SDKs/MacOSX11.3.sdk}"
    if [[ ! -d "$mac_sdk_path" ]]; then
        print_error "macOS SDK path not found: $mac_sdk_path"
        print_info "Set MACOSX_SDK_PATH to a valid SDK directory and retry."
        return 1
    fi

    local tag_args=()
    if [[ -n "${cross_tags}" ]]; then
        tag_args=(-tags="${cross_tags}")
    fi
    
    if ! fyne-cross darwin -arch=* \
        -app-id=com.passquantum.app \
        -name="$APP_NAME" \
        -icon=Icon.png \
        --macosx-sdk-path="$mac_sdk_path" \
        "${tag_args[@]}" \
        ./ui; then
        print_error "macOS build failed"
        return 1
    fi
    
    print_success "macOS build complete"
    echo ""
}

# Build for all platforms
build_all() {
    check_dependencies
    build_linux

    # Optional: these can fail in default fyne-cross images when gocv/OpenCV
    # requirements are not present. Keep Linux native build as the guaranteed path.
    if ! build_windows; then
        print_error "Windows cross-build failed (likely missing OpenCV in fyne-cross environment)."
    fi
    if ! build_macos; then
        print_error "macOS cross-build failed (check SDK path and OpenCV-enabled environment)."
    fi
    
    print_header "Build Summary"
    echo "Build process completed."
    echo ""
    echo "Build artifacts:"
    echo "  • Linux native: build/linux/$APP_NAME"
    echo "  • Windows: fyne-cross/dist/windows-amd64/"
    echo "  • macOS:   fyne-cross/dist/darwin-amd64/"
    echo ""
}

# Show usage
show_usage() {
    echo "PassQuantum Build Script"
    echo ""
    echo "Usage: ./build.sh [option]"
    echo ""
    echo "Options:"
    echo "  all      - Native Linux build + best-effort cross builds"
    echo "  linux    - Build native Linux binary only (recommended)"
    echo "  windows  - Build for Windows only"
    echo "  mac      - Build for macOS only"
    echo "  deps     - Install build dependencies only"
    echo "  help     - Show this help message"
    echo ""
    echo "Environment overrides:"
    echo "  CROSS_BUILD_TAGS=...   Additional build tags for cross targets"
    echo "  MACOSX_SDK_PATH=...    Override macOS SDK path"
    echo ""
}

# Main script logic
case "${1:-all}" in
    all)
        build_all
        ;;
    linux)
        check_dependencies
        build_linux
        ;;
    windows)
        check_dependencies
        build_windows
        ;;
    mac|macos|darwin)
        check_dependencies
        build_macos
        ;;
    deps|dependencies)
        check_dependencies
        ;;
    help|--help|-h)
        show_usage
        ;;
    *)
        print_error "Unknown option: $1"
        echo ""
        show_usage
        exit 1
        ;;
esac

#!/bin/bash
# PassQuantum cross-platform build script.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

APP_NAME="PassQuantum"
BUILD_DIR="build"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_header() {
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}$1${NC}"
    echo -e "${GREEN}================================${NC}"
}

print_info() {
    echo -e "${YELLOW}=>${NC} $1"
}

print_success() {
    echo -e "${GREEN}OK${NC} $1"
}

print_error() {
    echo -e "${RED}ERR${NC} $1"
}

check_dependencies() {
    print_header "Checking Dependencies"

    if ! command -v fyne >/dev/null 2>&1; then
        print_info "Installing fyne command..."
        go install fyne.io/fyne/v2/cmd/fyne@latest
    fi

    if ! command -v fyne-cross >/dev/null 2>&1; then
        print_info "Installing fyne-cross..."
        go install github.com/fyne-io/fyne-cross@latest
    fi

    print_success "Dependencies ready"
    echo ""
}

build_linux() {
    print_header "Building for Linux (AMD64)"
    mkdir -p "$BUILD_DIR/linux"
    CGO_ENABLED=1 go build -o "$BUILD_DIR/linux/$APP_NAME" ./ui
    cp face_guard.py "$BUILD_DIR/linux/"
    print_success "Linux build complete: $BUILD_DIR/linux/$APP_NAME"
    echo ""
}

build_windows() {
    print_header "Building for Windows (AMD64)"
    mkdir -p "$BUILD_DIR/windows"

    if ! fyne-cross windows -arch=amd64 \
        -app-id=com.passquantum.app \
        -name="$APP_NAME" \
        -icon=Icon.png \
        ./ui; then
        print_error "Windows build failed"
        return 1
    fi

    cp face_guard.py "fyne-cross/dist/windows-amd64/"
    print_success "Windows build complete"
    echo ""
}

build_macos() {
    print_header "Building for macOS"
    mkdir -p "$BUILD_DIR/mac"

    local mac_sdk_path="${MACOSX_SDK_PATH:-/home/lenovo/dev/PassQuantum/SDKs/MacOSX11.3.sdk}"
    if [[ ! -d "$mac_sdk_path" ]]; then
        print_error "macOS SDK path not found: $mac_sdk_path"
        print_info "Set MACOSX_SDK_PATH to a valid SDK directory and retry."
        return 1
    fi

    if ! fyne-cross darwin -arch=* \
        -app-id=com.passquantum.app \
        -name="$APP_NAME" \
        -icon=Icon.png \
        --macosx-sdk-path="$mac_sdk_path" \
        ./ui; then
        print_error "macOS build failed"
        return 1
    fi

    cp face_guard.py "fyne-cross/dist/darwin-amd64/"
    print_success "macOS build complete"
    echo ""
}

build_all() {
    check_dependencies
    build_linux

    if ! build_windows; then
        print_error "Windows cross-build failed"
    fi

    if ! build_macos; then
        print_error "macOS cross-build failed"
    fi

    print_header "Build Summary"
    echo "Build process completed."
    echo ""
    echo "Build artifacts:"
    echo "  - Linux native: build/linux/$APP_NAME"
    echo "  - Windows: fyne-cross/dist/windows-amd64/"
    echo "  - macOS: fyne-cross/dist/darwin-amd64/"
    echo ""
}

show_usage() {
    echo "PassQuantum Build Script"
    echo ""
    echo "Usage: ./build.sh [option]"
    echo ""
    echo "Options:"
    echo "  all      - Native Linux build + best-effort cross builds"
    echo "  linux    - Build native Linux binary only"
    echo "  windows  - Build for Windows only"
    echo "  mac      - Build for macOS only"
    echo "  deps     - Install build dependencies only"
    echo "  help     - Show this help message"
    echo ""
    echo "Environment overrides:"
    echo "  MACOSX_SDK_PATH=...    Override macOS SDK path"
    echo ""
}

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

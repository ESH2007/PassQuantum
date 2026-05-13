#!/bin/bash
# PassQuantum cross-platform build script.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

APP_NAME="PassQuantum"
BUILD_DIR="build"
PYTHON_BUNDLE_OUT="ui/face_guard_bundle"   # go:embed source (built by PyInstaller)
VENV_PYTHON=".venv/bin/python"

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

# ──────────────────────────────────────────────────────────────────────────────
# Dependency checks
# ──────────────────────────────────────────────────────────────────────────────

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

# ──────────────────────────────────────────────────────────────────────────────
# Python bundle (PyInstaller --onefile)
# ──────────────────────────────────────────────────────────────────────────────

# build_python_bundle compiles all Python files into a single self-contained
# executable at $PYTHON_BUNDLE_OUT using PyInstaller.  The models/ directory
# is included via --add-data so face_landmarker.task is available at runtime.
#
# Returns 0 on success, 1 on failure.  build_linux() checks the return code
# and falls back to copying .py files if bundling fails.
build_python_bundle() {
    print_header "Building Python Bundle (PyInstaller)"

    # Require the project venv
    if [[ ! -x "$VENV_PYTHON" ]]; then
        print_error "Python venv not found at .venv — run:"
        print_info  "  python3 -m venv .venv && .venv/bin/pip install -r requirements.txt"
        return 1
    fi

    # Install PyInstaller into the venv if not already present
    if ! "$VENV_PYTHON" -c "import PyInstaller" 2>/dev/null; then
        print_info "Installing PyInstaller into .venv..."
        .venv/bin/pip install --quiet pyinstaller
    fi

    # Wipe previous artefacts so a stale bundle cannot be embedded
    rm -f "$PYTHON_BUNDLE_OUT"

    print_info "Running PyInstaller..."
    "$VENV_PYTHON" -m PyInstaller \
        --onefile \
        --name face_guard_bundle \
        --distpath "$(dirname "$PYTHON_BUNDLE_OUT")" \
        --workpath /tmp/passquantum_pyinstaller_work \
        --specpath /tmp/passquantum_pyinstaller_spec \
        --add-data "$SCRIPT_DIR/models:models" \
        --collect-all mediapipe \
        --hidden-import cv2 \
        --hidden-import numpy \
        --noconfirm \
        --clean \
        face_guard.py

    if [[ ! -f "$PYTHON_BUNDLE_OUT" ]]; then
        print_error "PyInstaller did not produce $PYTHON_BUNDLE_OUT"
        return 1
    fi

    local size
    size=$(du -sh "$PYTHON_BUNDLE_OUT" | cut -f1)
    print_success "Python bundle ready: $PYTHON_BUNDLE_OUT ($size)"
    echo ""
}

# ──────────────────────────────────────────────────────────────────────────────
# Platform builds
# ──────────────────────────────────────────────────────────────────────────────

build_linux() {
    print_header "Building for Linux (AMD64)"
    mkdir -p "$BUILD_DIR/linux"

    local go_tags=""

    # Attempt to embed the Python bundle into the Go binary
    if build_python_bundle; then
        go_tags="-tags with_face_bundle"
        print_info "Building Go binary with embedded Python bundle..."
    else
        print_info "Python bundle unavailable — copying .py files instead..."
    fi

    # shellcheck disable=SC2086
    CGO_ENABLED=1 go build $go_tags -o "$BUILD_DIR/linux/$APP_NAME" ./ui

    if [[ -z "$go_tags" ]]; then
        # Fallback: ship .py files and the models directory alongside the binary
        cp face_guard.py geometric_encoder.py liveness_detector.py \
           face_authenticator.py "$BUILD_DIR/linux/"
        cp -r models "$BUILD_DIR/linux/"
        print_info "Copied Python source files to $BUILD_DIR/linux/"
    else
        # Bundle is embedded; clean up the temporary artefact from ui/
        rm -f "$PYTHON_BUNDLE_OUT"
        print_info "Python bundle embedded — no separate .py files needed."
    fi

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

    # PyInstaller cannot cross-compile; ship .py files for Windows
    cp face_guard.py geometric_encoder.py liveness_detector.py \
       face_authenticator.py "fyne-cross/dist/windows-amd64/"
    cp -r models "fyne-cross/dist/windows-amd64/"
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

    # PyInstaller cannot cross-compile; ship .py files for macOS
    cp face_guard.py geometric_encoder.py liveness_detector.py \
       face_authenticator.py "fyne-cross/dist/darwin-amd64/"
    cp -r models "fyne-cross/dist/darwin-amd64/"
    print_success "macOS build complete"
    echo ""
}

# ──────────────────────────────────────────────────────────────────────────────
# Aggregate targets
# ──────────────────────────────────────────────────────────────────────────────

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
    echo "  - Linux (self-contained): build/linux/$APP_NAME"
    echo "  - Windows:                fyne-cross/dist/windows-amd64/"
    echo "  - macOS:                  fyne-cross/dist/darwin-amd64/"
    echo ""
    echo "Note: The Linux binary embeds the Python runtime via PyInstaller."
    echo "      Windows and macOS distributions still require Python + .py files."
    echo ""
}

# ──────────────────────────────────────────────────────────────────────────────
# Entry point
# ──────────────────────────────────────────────────────────────────────────────

show_usage() {
    echo "PassQuantum Build Script"
    echo ""
    echo "Usage: ./build.sh [option]"
    echo ""
    echo "Options:"
    echo "  all      - Native Linux build (with embedded Python) + best-effort cross builds"
    echo "  linux    - Build native Linux binary only (with embedded Python)"
    echo "  windows  - Build for Windows only"
    echo "  mac      - Build for macOS only"
    echo "  bundle   - Build Python bundle only (output: $PYTHON_BUNDLE_OUT)"
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
    bundle)
        build_python_bundle
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


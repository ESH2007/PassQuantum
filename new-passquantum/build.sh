#!/bin/bash
# PassQuantum cross-platform build script.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Ensure Go-installed binaries (fyne, fyne-cross, …) are reachable.
export PATH="$(go env GOPATH)/bin:$PATH"

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
# Python bundle for Windows (PyInstaller inside a Docker + Wine container)
# ──────────────────────────────────────────────────────────────────────────────

# build_python_bundle_windows produces a Windows PE executable
# (face_guard_bundle.exe) using PyInstaller running inside the tobix/pywine
# Docker image — Python 3 pre-installed inside Wine, no host Wine required.
#
# Prerequisites on the build machine:
#   • Docker (already required by fyne-cross)
#
# The Docker image used can be overridden:
#   PYWINE_IMAGE="tobix/pywine:3.11"   (default)
#
# At runtime on Windows the Go binary looks for the bundle at:
#   %APPDATA%\PassQuantum\face_guard_bundle.exe
# Ship the produced .exe via an installer (NSIS/Inno Setup) targeting that path.
#
# Returns 0 on success, non-zero on failure.
build_python_bundle_windows() {
    print_header "Building Windows Python Bundle (PyInstaller via Docker + Wine)"

    local win_dist="fyne-cross/dist/windows-amd64"
    mkdir -p "$win_dist"

    local pywine_image="${PYWINE_IMAGE:-tobix/pywine:3.11}"

    # ── sanity check ───────────────────────────────────────────────────────────
    if ! command -v docker >/dev/null 2>&1; then
        print_error "docker not found — cannot build Windows Python bundle."
        print_info  "Docker is already required by fyne-cross; install it and retry."
        return 1
    fi

    # Pull the image (no-op if already cached)
    print_info "Pulling Docker image: $pywine_image"
    if ! docker pull "$pywine_image" 2>&1; then
        print_error "Could not pull $pywine_image"
        return 1
    fi

    rm -f "$win_dist/face_guard_bundle.exe"

    # ── run PyInstaller inside the container ───────────────────────────────────
    # The container's Wine Python lives at /wine/bin/python.
    # /src  → project root (read-only source)
    # /dist → output directory (writable, mapped to win_dist on the host)
    #
    # --add-data uses ';' as the separator (Windows convention).
    # Inside the container, /src/models maps to Z:\src\models in Wine.
    print_info "Running PyInstaller inside $pywine_image..."
    docker run --rm \
        -v "$SCRIPT_DIR:/src:ro" \
        -v "$(realpath "$win_dist"):/dist" \
        -w /src \
        "$pywine_image" \
        bash -c '
            set -euo pipefail
            wine python -m pip install --quiet pyinstaller mediapipe opencv-python-headless numpy
            wine python -m PyInstaller \
                --onefile \
                --name face_guard_bundle \
                --distpath /dist \
                --workpath /tmp/pq_win_work \
                --specpath /tmp/pq_win_spec \
                --add-data "Z:\\src\\models;models" \
                --collect-all mediapipe \
                --hidden-import cv2 \
                --hidden-import numpy \
                --noconfirm \
                --clean \
                face_guard.py
        '

    if [[ ! -f "$win_dist/face_guard_bundle.exe" ]]; then
        print_error "PyInstaller did not produce face_guard_bundle.exe"
        return 1
    fi

    local size
    size=$(du -sh "$win_dist/face_guard_bundle.exe" | cut -f1)
    print_success "Windows Python bundle: $win_dist/face_guard_bundle.exe ($size)"
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

    local win_dist="fyne-cross/dist/windows-amd64"

    # Attempt to build a self-contained Windows Python bundle via PyInstaller/Wine.
    # On success, face_guard_bundle.exe lands in $win_dist alongside the Go binary.
    # At install time it must be placed in %APPDATA%\PassQuantum\ on the target.
    if build_python_bundle_windows; then
        print_info "Bundle built. At install time place face_guard_bundle.exe in:"
        print_info "  %APPDATA%\\PassQuantum\\face_guard_bundle.exe"
    else
        # Fallback: ship Python source files in a dedicated AppData subfolder.
        # The contents of PassQuantum_AppData/ must be copied to
        # %APPDATA%\PassQuantum\ on the target machine (e.g. by an NSIS installer).
        print_info "Wine/PyInstaller unavailable — copying .py source files as fallback."
        local appdata_dir="$win_dist/PassQuantum_AppData"
        mkdir -p "$appdata_dir"
        cp face_guard.py geometric_encoder.py liveness_detector.py \
           face_authenticator.py "$appdata_dir/"
        cp -r models "$appdata_dir/"
        print_info "Copied Python source files to $win_dist/PassQuantum_AppData/"
        print_info "At install time place the contents of PassQuantum_AppData/ in:"
        print_info "  %APPDATA%\\PassQuantum\\"
    fi

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
    echo "      Windows uses PyInstaller via Wine (face_guard_bundle.exe) when Wine is"
    echo "      available, otherwise falls back to .py source files."
    echo "      Either way, the Python files must be placed in %APPDATA%\\PassQuantum\\"
    echo "      on the target Windows machine (e.g. via an NSIS/Inno Setup installer)."
    echo "      macOS distributions still require Python + .py files."
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
    echo "  windows  - Build for Windows only (PyInstaller via Wine, falls back to .py files)"
    echo "  mac      - Build for macOS only"
    echo "  bundle   - Build Python bundle only (output: $PYTHON_BUNDLE_OUT)"
    echo "  deps     - Install build dependencies only"
    echo "  help     - Show this help message"
    echo ""
    echo "Environment overrides:"
    echo "  MACOSX_SDK_PATH=...    Override macOS SDK path"
    echo "  PYWINE_IMAGE=...       Docker image used to build the Windows Python bundle"
    echo "                         (default: 'tobix/pywine:3.11')"
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


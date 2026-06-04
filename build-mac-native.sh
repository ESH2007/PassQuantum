#!/bin/bash
# PassQuantum native macOS build script (Apple Silicon arm64 / Intel amd64).
#
# Runs ON a macOS machine — it does NOT cross-compile. Mirrors the two-phase
# flow of Build-FaceBundle.ps1 (Windows):
#
#   Phase 1  PyInstaller --onefile face_guard bundle (native arch) into ui/
#   Phase 2  fyne package -os darwin → PassQuantum.app with embedded bundle
#   Phase 3  Inject NSCameraUsageDescription into Info.plist
#   Phase 4  Ad-hoc code signing of the .app
#   Phase 5  DMG packaging + ad-hoc signing
#   Phase 6  Summary
#
# The Python bundle is embedded into the Go binary via go:embed + the
# "with_face_bundle" build tag (see ui/python_bundle.go, which is compiled for
# every non-Windows OS — including darwin). No Python install is required on
# the target Mac. If PyInstaller fails, the script falls back to shipping the
# .py source files + models/ inside the .app bundle.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Ensure Go-installed binaries (fyne, …) are reachable.
export PATH="$(go env GOPATH 2>/dev/null)/bin:$PATH"

APP_NAME="PassQuantum"
APP_ID="com.passquantum.app"
BUILD_DIR="build/mac"
APP_BUNDLE="$BUILD_DIR/$APP_NAME.app"
DMG_OUT="$BUILD_DIR/$APP_NAME.dmg"
FACEGUARD_VENV=".venv-faceguard"             # isolated from .venv to avoid torch/tf contamination
# PyInstaller --onedir output (NOT --onefile): the helper must run in place from
# inside the signed .app for the macOS camera/TCC permission to work, so onefile's
# /tmp self-extraction is unusable here.
FACEGUARD_DIST="build/mac-faceguard"         # temp PyInstaller dist dir
FACEGUARD_ONEDIR="$FACEGUARD_DIST/face_guard_bundle"        # dir PyInstaller produces
FACEGUARD_EXE="$FACEGUARD_ONEDIR/face_guard_bundle"         # helper executable in it
HELPER_DEST_REL="Contents/Resources/faceguard"             # location inside the .app
MODEL_FILE="models/face_landmarker.task"
CAMERA_USAGE="PassQuantum uses the camera for face biometric authentication."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_header() {
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}$1${NC}"
    echo -e "${GREEN}================================${NC}"
}

print_info()    { echo -e "${YELLOW}=>${NC} $1"; }
print_success() { echo -e "${GREEN}OK${NC} $1"; }
print_error()   { echo -e "${RED}ERR${NC} $1"; }

# ──────────────────────────────────────────────────────────────────────────────
# Arguments
# ──────────────────────────────────────────────────────────────────────────────

SKIP_BUNDLE=0
SKIP_SIGN=0
SKIP_DMG=0

show_usage() {
    echo "PassQuantum native macOS build script"
    echo ""
    echo "Usage: ./build-mac-native.sh [options]"
    echo ""
    echo "Options:"
    echo "  --skip-bundle   Skip PyInstaller; ship .py source files inside the .app instead"
    echo "  --skip-sign     Skip ad-hoc code signing of the .app and .dmg"
    echo "  --skip-dmg      Skip DMG creation (only produce the .app)"
    echo "  --help, -h      Show this help message"
    echo ""
    echo "Outputs:"
    echo "  $APP_BUNDLE"
    echo "  $DMG_OUT"
    echo ""
}

for arg in "$@"; do
    case "$arg" in
        --skip-bundle) SKIP_BUNDLE=1 ;;
        --skip-sign)   SKIP_SIGN=1 ;;
        --skip-dmg)    SKIP_DMG=1 ;;
        --help|-h)     show_usage; exit 0 ;;
        *)
            print_error "Unknown option: $arg"
            echo ""
            show_usage
            exit 1
            ;;
    esac
done

# ──────────────────────────────────────────────────────────────────────────────
# Preflight checks
# ──────────────────────────────────────────────────────────────────────────────

print_header "PassQuantum macOS Native Build"

if [[ "$(uname -s)" != "Darwin" ]]; then
    print_error "This script must run on macOS (uname -s reported: $(uname -s))."
    print_info  "On Linux, use ./build.sh mac to cross-compile via fyne-cross."
    exit 1
fi

# Detect native architecture.
case "$(uname -m)" in
    arm64)  GOARCH="arm64";  ARCH_LABEL="Apple Silicon (arm64)" ;;
    x86_64) GOARCH="amd64";  ARCH_LABEL="Intel (amd64)" ;;
    *)
        print_error "Unsupported architecture: $(uname -m)"
        exit 1
        ;;
esac
export GOARCH
print_info "Target architecture: $ARCH_LABEL → GOARCH=$GOARCH"

# Prerequisite tools.
missing=0
for tool in go python3 codesign; do
    if ! command -v "$tool" >/dev/null 2>&1; then
        print_error "Required tool not found in PATH: $tool"
        missing=1
    fi
done
if ! xcode-select -p >/dev/null 2>&1; then
    print_error "Xcode Command Line Tools not found."
    print_info  "Install them with: xcode-select --install"
    missing=1
fi
if [[ "$missing" -ne 0 ]]; then
    print_error "Missing prerequisites — aborting."
    exit 1
fi

# Model asset (required by the MediaPipe face workflow; bundled via --add-data).
if [[ ! -f "$MODEL_FILE" ]]; then
    print_error "Required model file not found: $MODEL_FILE"
    print_info  "Download face_landmarker.task and place it in models/."
    print_info  "See models/README.md for the source URL and instructions."
    exit 1
fi

print_success "Preflight checks passed"
echo ""

# ──────────────────────────────────────────────────────────────────────────────
# Phase 1 — Python bundle (PyInstaller, native arch)
# ──────────────────────────────────────────────────────────────────────────────

# build_python_bundle compiles the Python face guard into a self-contained
# PyInstaller --onedir directory at $FACEGUARD_ONEDIR. --onedir (not --onefile)
# is required so the helper runs in place from inside the signed .app, which the
# macOS camera/TCC permission system needs. Returns 0 on success, 1 on failure.
build_python_bundle() {
    print_header "Phase 1 — Python Bundle (PyInstaller --onedir, $GOARCH)"

    local venv_python="$FACEGUARD_VENV/bin/python"

    # Create or reuse the isolated face-guard venv (kept separate from .venv).
    if [[ ! -x "$venv_python" ]]; then
        print_info "Creating isolated venv at $FACEGUARD_VENV ..."
        python3 -m venv "$FACEGUARD_VENV"
    else
        print_info "Reusing existing venv at $FACEGUARD_VENV"
    fi

    print_info "Installing PyInstaller + face-guard dependencies (headless OpenCV)..."
    "$venv_python" -m pip install --quiet --upgrade pip
    "$venv_python" -m pip install --quiet \
        pyinstaller \
        mediapipe \
        opencv-python-headless \
        "numpy<2"

    # Wipe any previous output so a stale bundle cannot be shipped.
    rm -rf "$FACEGUARD_DIST"
    mkdir -p "$FACEGUARD_DIST"

    print_info "Running PyInstaller (--onedir)..."
    "$venv_python" -m PyInstaller \
        --onedir \
        --name face_guard_bundle \
        --distpath "$FACEGUARD_DIST" \
        --workpath /tmp/passquantum_pyinstaller_work \
        --specpath /tmp/passquantum_pyinstaller_spec \
        --paths "$SCRIPT_DIR/python" \
        --add-data "$SCRIPT_DIR/models:models" \
        --collect-all mediapipe \
        --hidden-import cv2 \
        --hidden-import numpy \
        --hidden-import geometric_encoder \
        --hidden-import liveness_detector \
        --hidden-import face_authenticator \
        --noconfirm \
        --clean \
        python/face_guard.py

    if [[ ! -x "$FACEGUARD_EXE" ]]; then
        print_error "PyInstaller did not produce $FACEGUARD_EXE"
        return 1
    fi

    local size
    size=$(du -sh "$FACEGUARD_ONEDIR" | cut -f1)
    print_success "Python bundle ready: $FACEGUARD_ONEDIR ($size)"
    echo ""
}

HELPER_READY=0   # 1 → PyInstaller onedir built, will be shipped inside the .app

if [[ "$SKIP_BUNDLE" -eq 1 ]]; then
    print_info "--skip-bundle set — skipping PyInstaller, will ship .py source files."
    rm -rf "$FACEGUARD_DIST"
elif build_python_bundle; then
    HELPER_READY=1
else
    print_error "PyInstaller failed — falling back to shipping .py source files."
    rm -rf "$FACEGUARD_DIST"
fi

# ──────────────────────────────────────────────────────────────────────────────
# Phase 2 — Go build + Fyne package (.app bundle)
# ──────────────────────────────────────────────────────────────────────────────

print_header "Phase 2 — Fyne Package (.app)"

if ! command -v fyne >/dev/null 2>&1; then
    print_info "Installing fyne command..."
    go install fyne.io/fyne/v2/cmd/fyne@latest
fi

mkdir -p "$BUILD_DIR"
# Remove any stale bundle so the build is idempotent.
rm -rf "$APP_BUNDLE"

# The Go binary itself does NOT embed the helper on macOS — the helper ships as
# a real file inside the bundle (signed in place). So no -tags here; the darwin
# locator (ui/python_bundle_darwin.go) finds it at Contents/Resources/faceguard.
# fyne package writes <name>.app into the current directory; build there, move it.
print_info "Packaging $APP_NAME.app (CGO_ENABLED=1, GOARCH=$GOARCH)..."

# fyne resolves -icon relative to -sourceDir, so pass an absolute path to the
# root-level Icon.png (otherwise it looks for ui/Icon.png and fails).
CGO_ENABLED=1 GOARCH="$GOARCH" fyne package \
    -os darwin \
    -icon "$SCRIPT_DIR/Icon.png" \
    -name "$APP_NAME" \
    -appID "$APP_ID" \
    -sourceDir ./ui

# fyne emits the bundle next to -sourceDir (./ui) or in the cwd depending on
# version; locate it and move it into build/mac/.
produced_app=""
for candidate in "./$APP_NAME.app" "ui/$APP_NAME.app"; do
    if [[ -d "$candidate" ]]; then
        produced_app="$candidate"
        break
    fi
done

if [[ -z "$produced_app" ]]; then
    print_error "fyne package did not produce $APP_NAME.app"
    exit 1
fi

mv "$produced_app" "$APP_BUNDLE"
print_success "App bundle: $APP_BUNDLE"

# Ship the face guard helper INSIDE the bundle so it is signed in place and the
# camera/TCC permission attributes to PassQuantum. Falls back to .py sources when
# PyInstaller was skipped or failed.
if [[ "$HELPER_READY" -eq 1 ]]; then
    helper_dest="$APP_BUNDLE/$HELPER_DEST_REL"
    print_info "Copying face guard helper into $helper_dest ..."
    rm -rf "$helper_dest"
    mkdir -p "$helper_dest"
    # Copy the onedir CONTENTS (executable + _internal/) into faceguard/.
    cp -R "$FACEGUARD_ONEDIR"/. "$helper_dest"/
    chmod +x "$helper_dest/face_guard_bundle"
    print_success "Embedded in-bundle helper at $HELPER_DEST_REL/face_guard_bundle"
else
    res_dir="$APP_BUNDLE/Contents/Resources"
    print_info "No PyInstaller bundle — copying Python sources + models into the .app..."
    mkdir -p "$res_dir/python"
    cp python/face_guard.py python/geometric_encoder.py python/liveness_detector.py \
       python/face_authenticator.py "$res_dir/python/"
    cp -R models "$res_dir/"
    print_info "Copied .py sources to $res_dir/python/ (requires Python at runtime)."
fi
echo ""

# ──────────────────────────────────────────────────────────────────────────────
# Phase 3 — Info.plist camera entitlement
# ──────────────────────────────────────────────────────────────────────────────

print_header "Phase 3 — Camera Permission (Info.plist)"

PLIST="$APP_BUNDLE/Contents/Info.plist"
if [[ ! -f "$PLIST" ]]; then
    print_error "Info.plist not found at $PLIST"
    exit 1
fi

# Idempotent: only add the key if it isn't already present.
if /usr/libexec/PlistBuddy -c "Print :NSCameraUsageDescription" "$PLIST" >/dev/null 2>&1; then
    print_info "NSCameraUsageDescription already present — leaving as-is."
else
    /usr/libexec/PlistBuddy -c \
        "Add :NSCameraUsageDescription string \"$CAMERA_USAGE\"" "$PLIST"
    print_success "Added NSCameraUsageDescription to Info.plist"
fi
echo ""

# ──────────────────────────────────────────────────────────────────────────────
# Phase 4 — Ad-hoc code signing
# ──────────────────────────────────────────────────────────────────────────────

if [[ "$SKIP_SIGN" -eq 1 ]]; then
    print_info "--skip-sign set — skipping code signing."
else
    print_header "Phase 4 — Ad-hoc Code Signing"

    print_info "Signing $APP_BUNDLE (ad-hoc) ..."
    codesign --deep --force --sign - "$APP_BUNDLE"

    if codesign --verify --deep --strict "$APP_BUNDLE" 2>/dev/null; then
        print_success "Signature verified."
    else
        print_error "Signature verification failed (ad-hoc signing is best-effort) — continuing."
    fi
    echo ""
fi

# ──────────────────────────────────────────────────────────────────────────────
# Phase 5 — DMG packaging
# ──────────────────────────────────────────────────────────────────────────────

if [[ "$SKIP_DMG" -eq 1 ]]; then
    print_info "--skip-dmg set — skipping DMG creation."
else
    print_header "Phase 5 — DMG Packaging"

    rm -f "$DMG_OUT"
    print_info "Creating $DMG_OUT ..."
    hdiutil create \
        -volname "$APP_NAME" \
        -srcfolder "$APP_BUNDLE" \
        -ov -format UDZO \
        "$DMG_OUT"

    if [[ "$SKIP_SIGN" -eq 0 ]]; then
        print_info "Signing DMG (ad-hoc) ..."
        codesign --force --sign - "$DMG_OUT"
    fi
    print_success "DMG ready: $DMG_OUT"
    echo ""
fi

# ──────────────────────────────────────────────────────────────────────────────
# Cleanup — remove the temporary PyInstaller dist (now copied into the .app)
# ──────────────────────────────────────────────────────────────────────────────

rm -rf "$FACEGUARD_DIST"

# ──────────────────────────────────────────────────────────────────────────────
# Phase 6 — Summary
# ──────────────────────────────────────────────────────────────────────────────

print_header "Build Complete"

if [[ -d "$APP_BUNDLE" ]]; then
    app_size=$(du -sh "$APP_BUNDLE" | cut -f1)
    print_success "$APP_NAME.app  ($app_size)  → $APP_BUNDLE"
fi
if [[ "$SKIP_DMG" -eq 0 && -f "$DMG_OUT" ]]; then
    dmg_size=$(du -sh "$DMG_OUT" | cut -f1)
    print_success "$APP_NAME.dmg  ($dmg_size)  → $DMG_OUT"
fi
echo ""
echo "Architecture:  $ARCH_LABEL"
if [[ "$HELPER_READY" -eq 1 ]]; then
    echo "The .app is fully self-contained:"
    echo "  - Go UI + password manager"
    echo "  - $HELPER_DEST_REL/face_guard_bundle (mediapipe + opencv + all Python deps)"
    echo "  - models/face_landmarker.task (bundled inside the helper)"
    echo ""
    echo "The face guard helper runs in place from inside the signed bundle so the"
    echo "camera permission attributes to PassQuantum (NSCameraUsageDescription)."
else
    echo "The .app ships Python .py sources in Contents/Resources/python/"
    echo "and requires a Python 3 interpreter (with mediapipe, opencv, numpy) at runtime."
fi
if [[ "$SKIP_SIGN" -eq 0 ]]; then
    echo ""
    echo "Signing: ad-hoc. On first launch macOS may warn the app is unidentified —"
    echo "right-click → Open, or run: xattr -rd com.apple.quarantine '$APP_BUNDLE'"
fi
echo ""

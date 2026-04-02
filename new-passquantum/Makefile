# PassQuantum Build Makefile
# Easy cross-platform compilation for Linux, Windows, and macOS
#
# ─────────────────────────────────────────────────────────────────────────────
# BIOMETRIC FEATURE PREREQUISITES
# ─────────────────────────────────────────────────────────────────────────────
# PassQuantum includes 3D Face Mesh recognition via GoCV (OpenCV DNN + ONNX).
# Building requires OpenCV 4.x development libraries on the HOST machine:
#
#   Ubuntu / Debian:  sudo apt-get install libopencv-dev pkg-config
#   Fedora / RHEL:    sudo dnf install opencv-devel pkgconf
#   macOS (Homebrew): brew install opencv pkg-config
#   Windows:          Install OpenCV 4 and set PKG_CONFIG_PATH appropriately.
#
# ONNX model files must be placed in the models/ directory before runtime:
#   models/blazeface.onnx   — BlazeFace short-range face detector
#   models/face_mesh.onnx   — MediaPipe Face Mesh (468 3D landmarks)
#
# Cross-compilation with fyne-cross:
#   fyne-cross containers do NOT include OpenCV. For cross-compiled biometric
#   builds you must supply a custom Docker image with the target platform's
#   OpenCV 4 shared libraries. Native builds (build-simple / build-native.sh)
#   work as long as OpenCV is installed on the local machine.
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: all clean build-linux build-windows build-mac build-all install-deps help

# Application details
APP_NAME=PassQuantum
MAIN_FILE=ui/main.go
BUILD_DIR=build
VERSION?=1.0.0

# Detect current OS
UNAME_S := $(shell uname -s)

all: build

help:
	@echo "PassQuantum Build System"
	@echo "========================"
	@echo ""
	@echo "Usage:"
	@echo "  make build           - Build for current platform"
	@echo "  make build-linux     - Build for Linux (AMD64)"
	@echo "  make build-windows   - Build for Windows (AMD64)"
	@echo "  make build-mac       - Build for macOS (AMD64)"
	@echo "  make build-all       - Build for all platforms"
	@echo "  make install-deps    - Install build dependencies"
	@echo "  make clean           - Remove build artifacts"
	@echo ""

# Install fyne-cross for cross-compilation
install-deps:
	@echo "Installing fyne-cross for cross-compilation..."
	go install fyne.io/fyne/v2/cmd/fyne@latest
	go install github.com/fyne-io/fyne-cross@latest
	@echo "Dependencies installed successfully!"

# Build for current platform
build:
	@echo "Building for current platform..."
	@mkdir -p $(BUILD_DIR)
	fyne package -os $(shell echo $(UNAME_S) | tr '[:upper:]' '[:lower:]' | sed 's/darwin/darwin/' | sed 's/linux/linux/' | sed 's/mingw.*/windows/') -name $(APP_NAME)
	@echo "Build complete! Check the current directory for the package."

# Build for Linux
build-linux:
	@echo "Building for Linux (AMD64)..."
	@mkdir -p $(BUILD_DIR)/linux
	fyne-cross linux -arch=amd64 -app-id=com.passquantum.app -name $(APP_NAME) -icon=Icon.png ./ui
	@echo "Linux build complete! Check fyne-cross/dist/linux-amd64/"

# Build for Windows
build-windows:
	@echo "Building for Windows (AMD64)..."
	@mkdir -p $(BUILD_DIR)/windows
	fyne-cross windows -arch=amd64 -app-id=com.passquantum.app -name $(APP_NAME) -icon=Icon.png ./ui
	@echo "Windows build complete! Check fyne-cross/dist/windows-amd64/"
	@echo ""
	@echo "NOTE: If users get OpenGL errors on Windows, they should:"
	@echo "  1. Update graphics drivers"
	@echo "  2. Or download Mesa3D software renderer from https://fdossena.com/?p=mesa/index.frag"

# Build for macOS
build-mac:
	@echo "Building for macOS (AMD64)..."
	@mkdir -p $(BUILD_DIR)/mac
	fyne-cross darwin -arch=amd64 -app-id=com.passquantum.app -name $(APP_NAME) -icon=Icon.png ./ui
	@echo "macOS build complete! Check fyne-cross/dist/darwin-amd64/"

# Build for all platforms
build-all: build-linux build-windows build-mac
	@echo ""
	@echo "All builds complete!"
	@echo "  Linux:   fyne-cross/dist/linux-amd64/"
	@echo "  Windows: fyne-cross/dist/windows-amd64/"
	@echo "  macOS:   fyne-cross/dist/darwin-amd64/"

# Simple Go build (no packaging, for quick testing)
build-simple:
	@echo "Building simple executable for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "Simple build complete! Output: $(BUILD_DIR)/$(APP_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf fyne-cross
	rm -f $(APP_NAME)
	rm -f *.exe *.app
	rm -rf *.tar.xz
	@echo "Clean complete!"

# Run the application
run:
	@echo "Running PassQuantum..."
	go run $(MAIN_FILE)

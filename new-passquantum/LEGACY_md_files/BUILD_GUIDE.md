# PassQuantum Build Guide

This guide explains how to easily build PassQuantum for Linux, Windows, and macOS.

## Quick Start

### Option 1: Using the Build Script (Recommended)

```bash
# Make the script executable
chmod +x build.sh

# Build for all platforms
./build.sh all

# Or build for specific platforms
./build.sh linux
./build.sh windows
./build.sh mac
```

### Option 2: Using Make

```bash
# Build for all platforms
make build-all

# Or build for specific platforms
make build-linux
make build-windows
make build-mac

# Install dependencies first (if needed)
make install-deps
```

## Prerequisites

### Required Tools

1. **Go 1.22 or later**
   ```bash
   go version
   ```

2. **Fyne command-line tools** (auto-installed by scripts)
   ```bash
   go install fyne.io/fyne/v2/cmd/fyne@latest
   go install github.com/fyne-io/fyne-cross@latest
   ```

3. **Docker** (required for fyne-cross to build for other platforms)
   - Linux: `sudo apt install docker.io` or follow [Docker installation guide](https://docs.docker.com/engine/install/)
   - macOS: Install [Docker Desktop](https://www.docker.com/products/docker-desktop)
   - Windows: Install [Docker Desktop](https://www.docker.com/products/docker-desktop)

### Setting up Docker (Linux)

```bash
# Install Docker
sudo apt update
sudo apt install docker.io

# Add your user to the docker group (to run without sudo)
sudo usermod -aG docker $USER

# Log out and log back in, or run:
newgrp docker

# Verify Docker is working
docker --version
```

## Building for Different Platforms

### Linux Build

```bash
./build.sh linux
# or
make build-linux
```

Output: `fyne-cross/dist/linux-amd64/PassQuantum`

### Windows Build

```bash
./build.sh windows
# or
make build-windows
```

Output: `fyne-cross/dist/windows-amd64/PassQuantum.exe`

**Important for Windows users:** See "Windows OpenGL Issue" section below.

### macOS Build

```bash
./build.sh mac
# or
make build-mac
```

Output: `fyne-cross/dist/darwin-amd64/PassQuantum.app`

### Build All Platforms at Once

```bash
./build.sh all
# or
make build-all
```

## Windows OpenGL Issue - IMPORTANT!

The error you encountered:
```
Fyne error: window creation error
Cause: APIUnavailable: WGL: The driver does not appear to support OpenGL
```

This happens when Windows doesn't have proper OpenGL drivers installed.

### Solutions for End Users

Provide these solutions to users who encounter this error:

#### Solution 1: Update Graphics Drivers (Recommended)

1. **NVIDIA Graphics:**
   - Visit [NVIDIA Driver Downloads](https://www.nvidia.com/Download/index.aspx)
   - Download and install the latest driver

2. **AMD Graphics:**
   - Visit [AMD Driver Downloads](https://www.amd.com/en/support)
   - Download and install the latest driver

3. **Intel Graphics:**
   - Visit [Intel Driver Downloads](https://www.intel.com/content/www/us/en/download-center/home.html)
   - Download and install the latest graphics driver

#### Solution 2: Software OpenGL Renderer (If hardware drivers can't be updated)

If updating drivers doesn't work or isn't possible (e.g., old hardware, virtual machines):

1. Download Mesa3D software renderer:
   - Visit: https://fdossena.com/?p=mesa/index.frag
   - Download the latest Mesa3D release

2. Extract `opengl32.dll` from the archive

3. Place `opengl32.dll` in the same folder as `PassQuantum.exe`

4. Run `PassQuantum.exe`

**Note:** Software rendering will be slower than hardware rendering, but the application will work.

#### Solution 3: Build with Software Rendering Flag

You can also build a version that uses software rendering by default:

```bash
# Set environment variable before building
export LIBGL_ALWAYS_SOFTWARE=1
./build.sh windows
```

## Advanced Building Options

### Manual Build (without fyne-cross)

For quick testing on your current platform:

```bash
# Simple build
go build -o PassQuantum ui/main.go

# Or use fyne package for better integration
fyne package -os linux -name PassQuantum
```

### Custom Build with Icons

If you have a custom icon (icon.png):

```bash
fyne-cross windows -arch=amd64 \
    -app-id=com.passquantum.app \
    -name=PassQuantum \
    -icon=icon.png
```

### Building for ARM platforms

```bash
# Linux ARM64
fyne-cross linux -arch=arm64

# macOS ARM (Apple Silicon)
fyne-cross darwin -arch=arm64
```

## Troubleshooting

### "fyne-cross: command not found"

Install fyne-cross:
```bash
go install github.com/fyne-io/fyne-cross@latest
```

Make sure your `$GOPATH/bin` is in your `$PATH`:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Docker permission denied

On Linux:
```bash
sudo usermod -aG docker $USER
newgrp docker
```

### Build fails with "cannot find package"

Update dependencies:
```bash
go mod download
go mod tidy
```

## Distribution

After building, you'll find your executables in:
- `fyne-cross/dist/linux-amd64/` - Linux executable
- `fyne-cross/dist/windows-amd64/` - Windows .exe
- `fyne-cross/dist/darwin-amd64/` - macOS .app

### Creating Release Packages

For Linux:
```bash
cd fyne-cross/dist/linux-amd64/
tar -czf PassQuantum-linux-amd64.tar.gz PassQuantum
```

For Windows:
```bash
cd fyne-cross/dist/windows-amd64/
zip PassQuantum-windows-amd64.zip PassQuantum.exe
# Optionally include opengl32.dll for users with graphics issues
```

For macOS:
```bash
cd fyne-cross/dist/darwin-amd64/
zip -r PassQuantum-macos-amd64.zip PassQuantum.app
```

## Tips for Users

Include these in your README or release notes:

### System Requirements

**Windows:**
- Windows 10 or later
- OpenGL 2.0 compatible graphics card (or Mesa3D software renderer)
- Updated graphics drivers recommended

**Linux:**
- Modern Linux distribution (Ubuntu 20.04+, Fedora 35+, etc.)
- OpenGL support
- Libraries: `libgl1`, `libx11-6`, `libxcursor1`, `libxrandr2`, `libxinerama1`, `libxi6`

**macOS:**
- macOS 10.13 (High Sierra) or later
- For Apple Silicon Macs, use the ARM64 build

## Clean Build Artifacts

```bash
make clean
# or
rm -rf fyne-cross build
```

## Need Help?

- Fyne documentation: https://developer.fyne.io/
- Fyne-cross documentation: https://github.com/fyne-io/fyne-cross

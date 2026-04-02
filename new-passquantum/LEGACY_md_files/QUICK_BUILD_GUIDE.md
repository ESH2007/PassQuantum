# PassQuantum - Quick Build Guide

## TL;DR - How to Build

### For Linux (Current Platform) - No Docker Needed ✅
```bash
./build-native.sh
```
This creates `PassQuantum` executable immediately!

### For Windows/macOS Cross-Compilation - Requires Docker 🐳
```bash
# First, start Docker Desktop on Windows
# Then enable WSL2 integration in Docker Desktop settings

# Build for specific platform
./build.sh windows   # Creates .exe
./build.sh mac       # Creates .app
./build.sh linux     # Creates Linux binary

# Or build all at once
./build.sh all
```

---

## Build Options Explained

### Option 1: Native Build (Fastest, No Docker) ✅

**Best for:** Testing, Linux development, quick builds

```bash
./build-native.sh
```

**Pros:**
- ✅ No Docker required
- ✅ Fast compilation
- ✅ Works immediately on Linux/WSL2

**Cons:**
- ❌ Only builds for your current platform (Linux)
- ❌ Cannot create Windows .exe or macOS .app

**Output:** `./PassQuantum` (Linux binary)

---

### Option 2: Cross-Platform Build (Docker Required) 🐳

**Best for:** Creating releases for Windows/macOS/Linux

```bash
./build.sh windows   # Windows .exe
./build.sh mac       # macOS .app
./build.sh linux     # Linux binary
./build.sh all       # All platforms
```

**Pros:**
- ✅ Build for Windows, macOS, and Linux
- ✅ Professional packaging
- ✅ Automated build process

**Cons:**
- ❌ Requires Docker Desktop running
- ❌ First build downloads Docker images (slow initially)
- ❌ Needs WSL2 integration enabled

**Output:** `fyne-cross/dist/{platform}-amd64/`

**Docker Setup Required:**
1. Install [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop)
2. Open Docker Desktop
3. Go to: Settings → Resources → WSL Integration
4. Enable integration with your Linux distro
5. Click "Apply & Restart"

---

## Windows OpenGL Error - IMPORTANT! ⚠️

When Windows users run your `.exe`, they might see:
```
Fyne error: window creation error
Cause: APIUnavailable: WGL: The driver does not appear to support OpenGL
```

### Solutions for Windows Users:

**Solution 1: Update Graphics Drivers** (Recommended)
- NVIDIA: https://www.nvidia.com/Download/index.aspx
- AMD: https://www.amd.com/en/support
- Intel: https://www.intel.com/content/www/us/en/download-center/home.html

**Solution 2: Software Renderer** (If drivers can't be updated)
1. Download Mesa3D: https://fdossena.com/?p=mesa/index.frag
2. Extract `opengl32.dll`
3. Place it next to `PassQuantum.exe`
4. Run the application

See [WINDOWS_OPENGL_FIX.md](WINDOWS_OPENGL_FIX.md) for detailed instructions to share with Windows users.

---

## Quick Reference Commands

```bash
# Native build (Linux only, no Docker)
./build-native.sh

# Cross-platform builds (requires Docker)
./build.sh linux      # Linux
./build.sh windows    # Windows
./build.sh mac        # macOS
./build.sh all        # All platforms

# Using Make
make build-linux
make build-windows
make build-mac
make build-all

# Clean build artifacts
make clean
# or
rm -rf fyne-cross build PassQuantum

# Run the app
./PassQuantum
```

---

## Where Are My Built Files?

### Native Build
```
./PassQuantum  (Linux executable in current directory)
```

### Cross-Platform Build
```
fyne-cross/dist/
├── linux-amd64/
│   └── PassQuantum  (Linux executable)
├── windows-amd64/
│   └── PassQuantum.exe  (Windows executable)
└── darwin-amd64/
    └── PassQuantum.app  (macOS application bundle)
```

---

## Creating Release Packages

### For Linux
```bash
cd fyne-cross/dist/linux-amd64/
tar -czf PassQuantum-linux-amd64.tar.gz PassQuantum
```

### For Windows
```bash
cd fyne-cross/dist/windows-amd64/
zip PassQuantum-windows-amd64.zip PassQuantum.exe

# Optionally include Mesa3D for users with graphics issues
# Download opengl32.dll from Mesa3D and add to zip
```

### For macOS
```bash
cd fyne-cross/dist/darwin-amd64/
zip -r PassQuantum-macos-amd64.zip PassQuantum.app
```

---

## Distribution Checklist

When distributing your application:

- [ ] Include `WINDOWS_OPENGL_FIX.md` with Windows builds
- [ ] Test on actual Windows/macOS machines (not just WSL)
- [ ] Include LICENSE file
- [ ] Include USER_GUIDE.md
- [ ] Create a README with system requirements
- [ ] Consider code signing (for macOS/Windows)

---

## Troubleshooting

### "Docker not running"
- Start Docker Desktop on Windows
- Enable WSL2 integration in Docker Desktop settings

### "command not found: fyne-cross"
```bash
go install github.com/fyne-io/fyne-cross@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

### "package passquantum/core/crypto not found"
You're building from the wrong directory. Always build from:
```bash
cd /home/lenovo/dev/PassQuantum/new-passquantum
./build-native.sh
```

### Build is very slow
- First cross-platform build downloads Docker images (10-15 min)
- Subsequent builds are much faster (2-3 min per platform)
- Use `./build-native.sh` for quick testing

---

## Summary

**For development and testing:**
```bash
./build-native.sh  # Quick Linux build
```

**For releases (after setting up Docker):**
```bash
./build.sh all  # Creates Windows, macOS, and Linux builds
```

**For detailed documentation:**
- See [BUILD_GUIDE.md](BUILD_GUIDE.md) for comprehensive instructions
- See [WINDOWS_OPENGL_FIX.md](WINDOWS_OPENGL_FIX.md) for Windows troubleshooting

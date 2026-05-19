<#
.SYNOPSIS
    Build PassQuantum.exe — a Windows binary with the face-recognition Python
    runtime embedded directly inside the Go binary.

.DESCRIPTION
    Two-phase build:

    Phase 1 — PyInstaller
        Packs the following Python modules and all their dependencies into a
        single self-contained face_guard_bundle.exe:
            face_guard.py          (entry point / Go IPC layer)
            geometric_encoder.py   (MediaPipe landmark encoder)
            liveness_detector.py   (EAR blink anti-spoofing)
            face_authenticator.py  (high-level enroll / verify API)
            models\face_landmarker.task  (MediaPipe model asset)

    Phase 2 — Go build  (CGO_ENABLED=1, -tags with_face_bundle)
        Embeds face_guard_bundle.exe into the Go binary via go:embed
        (ui\python_bundle_windows.go).  The resulting PassQuantum.exe
        contains everything — no separate Python installation needed on
        the target machine.

    Output:
        <script dir>\build\windows\PassQuantum.exe

    IMPORTANT — isolated virtual environment
    ----------------------------------------
    PyInstaller runs inside a dedicated venv (.venv-faceguard) with only
    mediapipe, opencv-python, numpy, and pyinstaller installed.  This keeps
    torch / tensorflow / keras out of the analysis graph, avoiding:
      • multi-GB bundles
      • multi-hour analysis runs
      • NumPy 1.x / 2.x ABI crashes in PyInstaller hooks

    Pass -RebuildVenv to force a clean venv rebuild.

.PARAMETER GoBuildOutput
    Full path to the produced Go binary.
    Defaults to  <script dir>\build\windows\PassQuantum.exe

.PARAMETER WorkDir
    Scratch directory used by PyInstaller during the build.
    Defaults to  %TEMP%\passquantum_pyinstaller_work

.PARAMETER SpecDir
    Directory where PyInstaller writes the generated .spec file.
    Defaults to  %TEMP%\passquantum_pyinstaller_spec

.PARAMETER Python
    Base Python interpreter used to create the venv.  Defaults to 'python'.
    Override with a full path, e.g.  -Python "C:\Python311\python.exe"

.PARAMETER RebuildVenv
    Delete and recreate the isolated venv before building.

.EXAMPLE
    # Basic build — PassQuantum.exe lands in .\build\windows\
    .\Build-FaceBundle.ps1

.EXAMPLE
    # Force a clean venv rebuild
    .\Build-FaceBundle.ps1 -RebuildVenv

.EXAMPLE
    # Custom output path + specific Python interpreter
    .\Build-FaceBundle.ps1 -GoBuildOutput "C:\Publish\PassQuantum.exe" -Python "C:\Python311\python.exe"
#>

[CmdletBinding()]
param(
    [string] $GoBuildOutput = (Join-Path $PSScriptRoot "build\windows\PassQuantum.exe"),
    [string] $WorkDir       = (Join-Path $env:TEMP "passquantum_pyinstaller_work"),
    [string] $SpecDir       = (Join-Path $env:TEMP "passquantum_pyinstaller_spec"),
    [string] $Python        = "python",
    [switch] $RebuildVenv
)

# PyInstaller always outputs face_guard_bundle.exe into ui\ (next to the Go
# source files) so the go:embed directive in python_bundle_windows.go can find it.
$PyInstallerOutputDir = Join-Path $PSScriptRoot "ui"

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# ─────────────────────────────────────────────────────────────────────────────
# Helpers
# ─────────────────────────────────────────────────────────────────────────────

function Write-Header([string]$msg) {
    Write-Host ""
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host "  $msg" -ForegroundColor Cyan
    Write-Host "================================" -ForegroundColor Cyan
}

function Write-Step([string]$msg)  { Write-Host "=> $msg" -ForegroundColor Yellow }
function Write-OK([string]$msg)    { Write-Host "OK  $msg" -ForegroundColor Green  }
function Write-Fail([string]$msg)  { Write-Host "ERR $msg" -ForegroundColor Red    }

# ─────────────────────────────────────────────────────────────────────────────
# 1. Verify base Python
# ─────────────────────────────────────────────────────────────────────────────

Write-Header "Checking Python"

try {
    $pyVer = & $Python --version 2>&1
    Write-OK "Found: $pyVer  ($Python)"
} catch {
    Write-Fail "Python not found at '$Python'."
    Write-Step "Install Python 3.11+ from https://www.python.org/downloads/ and retry,"
    Write-Step "or pass -Python 'C:\path\to\python.exe'."
    exit 1
}

# ─────────────────────────────────────────────────────────────────────────────
# 2. Create / reuse isolated virtual environment
#
#    Why: the system Python may have torch, tensorflow, keras, scipy, etc.
#    PyInstaller's hook runner imports every reachable package, so those
#    multi-GB frameworks end up in the bundle and their NumPy 1.x ABI
#    conflicts with numpy 2.x cause analysis crashes.
#    A clean venv with only mediapipe + opencv-python + numpy avoids all of
#    that and produces a ~200 MB binary in minutes instead of hours.
# ─────────────────────────────────────────────────────────────────────────────

$VenvDir    = Join-Path $PSScriptRoot ".venv-faceguard"
$VenvPython = Join-Path $VenvDir "Scripts\python.exe"

Write-Header "Isolated Virtual Environment"

if ($RebuildVenv -and (Test-Path $VenvDir)) {
    Write-Step "RebuildVenv requested — removing $VenvDir"
    Remove-Item -Recurse -Force $VenvDir
}

if (-not (Test-Path $VenvPython)) {
    Write-Step "Creating venv at $VenvDir ..."
    & $Python -m venv $VenvDir
    if ($LASTEXITCODE -ne 0) {
        Write-Fail "Failed to create virtual environment."
        exit 1
    }
    Write-OK "Venv created."
} else {
    Write-OK "Reusing existing venv: $VenvDir"
}

# ─────────────────────────────────────────────────────────────────────────────
# 3. Install dependencies into the venv
#    Only the three packages the app actually uses + PyInstaller.
#    numpy<2 is pinned because mediapipe 0.10.x ships C extensions compiled
#    against the NumPy 1.x ABI.
# ─────────────────────────────────────────────────────────────────────────────

Write-Header "Installing Dependencies (isolated)"

$deps = @("pyinstaller", "mediapipe", "opencv-python", "numpy<2")
Write-Step "pip install $($deps -join ' ')"

& $VenvPython -m pip install --quiet --upgrade pip
& $VenvPython -m pip install --quiet @deps
if ($LASTEXITCODE -ne 0) {
    Write-Fail "pip install failed (exit $LASTEXITCODE). Try -RebuildVenv and retry."
    exit 1
}
Write-OK "Dependencies installed in isolated venv."

# ─────────────────────────────────────────────────────────────────────────────
# 4. Resolve absolute paths
#    Using $PSScriptRoot ensures PyInstaller finds assets regardless of where
#    the spec file is written (the same root cause fixed for the Linux CI job).
# ─────────────────────────────────────────────────────────────────────────────

$ProjectRoot  = $PSScriptRoot
$EntryPoint   = Join-Path $ProjectRoot "face_guard.py"
$ModelsDir    = Join-Path $ProjectRoot "models"
$BundleName   = "face_guard_bundle"
$BundleExe    = Join-Path $PyInstallerOutputDir "$BundleName.exe"

$RequiredFiles = @(
    $EntryPoint,
    (Join-Path $ProjectRoot "geometric_encoder.py"),
    (Join-Path $ProjectRoot "liveness_detector.py"),
    (Join-Path $ProjectRoot "face_authenticator.py"),
    (Join-Path $ModelsDir   "face_landmarker.task")
)

Write-Header "Validating Source Files"
foreach ($f in $RequiredFiles) {
    if (-not (Test-Path $f)) {
        Write-Fail "Required file not found: $f"
        exit 1
    }
    Write-OK (Split-Path $f -Leaf)
}

# ─────────────────────────────────────────────────────────────────────────────
# 5. Clean previous artefacts
# ─────────────────────────────────────────────────────────────────────────────

Write-Header "Cleaning Previous Build Artefacts"

foreach ($path in @($BundleExe, $WorkDir, $SpecDir)) {
    if (Test-Path $path) {
        Remove-Item -Recurse -Force $path
        Write-Step "Removed: $path"
    }
}

New-Item -ItemType Directory -Force -Path $PyInstallerOutputDir | Out-Null

# ─────────────────────────────────────────────────────────────────────────────
# 6. Run PyInstaller (using the venv Python)
#
#    --add-data  uses ';' as the path separator on Windows.
#    --paths     adds $ProjectRoot to sys.path so PyInstaller resolves the
#                local modules the same way Python does at runtime.
# ─────────────────────────────────────────────────────────────────────────────

Write-Header "Running PyInstaller"

$PyInstallerArgs = @(
    "-m", "PyInstaller",
    "--onefile",
    "--name",          $BundleName,
    "--distpath",      $PyInstallerOutputDir,
    "--workpath",      $WorkDir,
    "--specpath",      $SpecDir,
    "--paths",         $ProjectRoot,
    "--add-data",      "${ModelsDir};models",
    "--collect-all",   "mediapipe",
    "--hidden-import", "cv2",
    "--hidden-import", "numpy",
    "--hidden-import", "geometric_encoder",
    "--hidden-import", "liveness_detector",
    "--hidden-import", "face_authenticator",
    "--noconfirm",
    "--clean",
    $EntryPoint
)

Write-Step "Interpreter: $VenvPython"
Write-Host ""

& $VenvPython @PyInstallerArgs
if ($LASTEXITCODE -ne 0) {
    Write-Fail "PyInstaller exited with code $LASTEXITCODE."
    exit 1
}

# ─────────────────────────────────────────────────────────────────────────────
# 7. Verify PyInstaller output
# ─────────────────────────────────────────────────────────────────────────────

Write-Header "Python Bundle Ready"

if (-not (Test-Path $BundleExe)) {
    Write-Fail "PyInstaller did not produce: $BundleExe"
    exit 1
}

$pySize = (Get-Item $BundleExe).Length / 1MB
Write-OK ("face_guard_bundle.exe  ({0:F1} MB)  → {1}" -f $pySize, $BundleExe)

# ─────────────────────────────────────────────────────────────────────────────
# 8. Embed Windows application manifest (Windows 10 / 11 compatibility)
#
#    Without a manifest the OS reports the binary as "not compatible with
#    your version of Windows".  rsrc (github.com/akavel/rsrc) converts
#    ui\app.manifest into a ui\rsrc.syso resource object that the Go linker
#    automatically picks up — no extra linker flags needed.
# ─────────────────────────────────────────────────────────────────────────────

Write-Header "Embedding Windows Application Manifest"

$ManifestSrc = Join-Path $PSScriptRoot "ui\app.manifest"
$ManifestSyso = Join-Path $PSScriptRoot "ui\rsrc.syso"

if (-not (Test-Path $ManifestSrc)) {
    Write-Fail "app.manifest not found at $ManifestSrc"
    exit 1
}

Write-Step "go install github.com/akavel/rsrc  →  ui\rsrc.syso"
& go install github.com/akavel/rsrc@v0.10.2
if ($LASTEXITCODE -ne 0) {
    Write-Fail "Failed to install rsrc tool (exit $LASTEXITCODE)."
    exit 1
}
& rsrc -arch amd64 -manifest $ManifestSrc -o $ManifestSyso
if ($LASTEXITCODE -ne 0) {
    Write-Fail "rsrc failed (exit $LASTEXITCODE)."
    exit 1
}
Write-OK "rsrc.syso generated — manifest will be linked into PassQuantum.exe"

# ─────────────────────────────────────────────────────────────────────────────
# 9. Build the Go binary with the Python bundle embedded
#
#    python_bundle_windows.go (//go:build with_face_bundle && windows) picks up
#    ui\face_guard_bundle.exe via //go:embed and extracts it at runtime into
#    %TEMP%\passquantum-face-guard\face_guard.exe.
#
#    Requirements on the build machine:
#      • Go 1.22+  (go.exe in PATH)
#      • A C compiler for CGO (TDM-GCC, MSYS2/MinGW-w64, or LLVM/clang)
#        Fyne requires CGO; without it the build will fail.
# ─────────────────────────────────────────────────────────────────────────────

Write-Header "Building Go Binary (with embedded Python bundle)"

# Verify Go is available
try {
    $goVer = & go version 2>&1
    Write-OK "Found: $goVer"
} catch {
    Write-Fail "go not found in PATH."
    Write-Step "Install Go 1.22+ from https://go.dev/dl/ and ensure it is in PATH."
    exit 1
}

$GoBuildDir = Split-Path $GoBuildOutput -Parent
New-Item -ItemType Directory -Force -Path $GoBuildDir | Out-Null

$env:CGO_ENABLED = "1"

# Prefer MSYS2 MinGW-w64 over TDM-GCC — TDM-GCC produces CGO binaries that
# Windows rejects at load time ("not a valid application for this OS platform").
$Msys2Gcc = "C:\msys64\mingw64\bin\gcc.exe"
if (Test-Path $Msys2Gcc) {
    $env:CC   = $Msys2Gcc
    $env:PATH = "C:\msys64\mingw64\bin;$($env:PATH)"
    Write-OK "Using MSYS2 MinGW-w64 GCC: $Msys2Gcc"
} else {
    Write-Step "MSYS2 not found at C:\msys64 — falling back to GCC in PATH."
    Write-Step "If the build fails, install MSYS2 from https://www.msys2.org/ and run: pacman -S mingw-w64-x86_64-gcc"
}

Write-Step "go build -tags with_face_bundle -o $GoBuildOutput .\ui"
& go build -tags with_face_bundle -o $GoBuildOutput .\ui
if ($LASTEXITCODE -ne 0) {
    # Clean up the generated syso so a partial build doesn't linger
    if (Test-Path $ManifestSyso) { Remove-Item -Force $ManifestSyso }
    Write-Fail "Go build failed (exit $LASTEXITCODE)."
    Write-Step "Make sure MSYS2 MinGW-w64 is installed (https://www.msys2.org/) and run: pacman -S mingw-w64-x86_64-gcc"
    exit 1
}

# rsrc.syso is a build-time artefact — remove it so it is not accidentally
# committed and so that a plain 'go build' (without the manifest step) still
# works on non-Windows machines.
if (Test-Path $ManifestSyso) {
    Remove-Item -Force $ManifestSyso
    Write-Step "Cleaned up ui\rsrc.syso (build artefact)"
}

# ─────────────────────────────────────────────────────────────────────────────
# 10. Final summary
# ─────────────────────────────────────────────────────────────────────────────

Write-Header "Build Complete"

$goSize = (Get-Item $GoBuildOutput).Length / 1MB
Write-OK ("PassQuantum.exe  ({0:F1} MB)  → {1}" -f $goSize, $GoBuildOutput)
Write-Host ""
Write-Host "The binary is fully self-contained:" -ForegroundColor Cyan
Write-Host "  - Go UI + password manager" -ForegroundColor White
Write-Host "  - face_guard_bundle.exe (mediapipe + opencv + all Python deps)" -ForegroundColor White
Write-Host "  - models\face_landmarker.task" -ForegroundColor White
Write-Host ""
Write-Host "On first launch the Python bundle is extracted to:" -ForegroundColor Cyan
Write-Host "  %TEMP%\passquantum-face-guard\face_guard.exe" -ForegroundColor White

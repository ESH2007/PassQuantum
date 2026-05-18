<#
.SYNOPSIS
    Build face_guard_bundle.exe — a self-contained Windows executable that
    bundles all face-recognition Python modules and their dependencies.

.DESCRIPTION
    Uses PyInstaller to pack:
        face_guard.py          (entry point / Go IPC layer)
        geometric_encoder.py   (MediaPipe landmark encoder)
        liveness_detector.py   (EAR blink anti-spoofing)
        face_authenticator.py  (high-level enroll / verify API)
        models\                (face_landmarker.task runtime asset)
    plus all transitive libraries (mediapipe, opencv-python, numpy).

    IMPORTANT — isolated virtual environment
    ----------------------------------------
    PyInstaller is run inside a dedicated venv that contains ONLY the three
    packages the app needs.  This prevents the global Python environment
    (which may have torch, tensorflow, scipy, pandas, etc.) from being
    scanned and bundled, which would:
      • bloat the binary to several GB
      • take hours to analyse
      • cause NumPy 1.x / 2.x ABI crashes during hook analysis

    The venv is created at  <script dir>\.venv-faceguard  and reused on
    subsequent runs.  Pass -RebuildVenv to force a clean rebuild.

    The produced binary is written to:
        <repo>\new-passquantum\ui\face_guard_bundle.exe

    It is intended to be placed at  %APPDATA%\PassQuantum\face_guard_bundle.exe
    on the target Windows machine.

.PARAMETER OutputDir
    Directory that receives face_guard_bundle.exe.
    Defaults to  <script dir>\ui

.PARAMETER WorkDir
    Scratch directory used by PyInstaller during the build.
    Defaults to  %TEMP%\passquantum_pyinstaller_work

.PARAMETER SpecDir
    Directory where PyInstaller writes the generated .spec file.
    Defaults to  %TEMP%\passquantum_pyinstaller_spec

.PARAMETER Python
    Base Python interpreter used to create the venv.  Defaults to 'python'.
    Override with a full path when multiple interpreters are installed,
    e.g.  -Python "C:\Python311\python.exe"

.PARAMETER RebuildVenv
    Delete and recreate the isolated venv before building.

.EXAMPLE
    # Basic — output lands in .\ui\
    .\Build-FaceBundle.ps1

.EXAMPLE
    # Force a clean venv rebuild
    .\Build-FaceBundle.ps1 -RebuildVenv

.EXAMPLE
    # Custom output directory + specific interpreter
    .\Build-FaceBundle.ps1 -OutputDir "C:\Publish\PassQuantum" -Python "C:\Python311\python.exe"
#>

[CmdletBinding()]
param(
    [string] $OutputDir   = (Join-Path $PSScriptRoot "ui"),
    [string] $WorkDir     = (Join-Path $env:TEMP "passquantum_pyinstaller_work"),
    [string] $SpecDir     = (Join-Path $env:TEMP "passquantum_pyinstaller_spec"),
    [string] $Python      = "python",
    [switch] $RebuildVenv
)

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

$ProjectRoot = $PSScriptRoot
$EntryPoint  = Join-Path $ProjectRoot "face_guard.py"
$ModelsDir   = Join-Path $ProjectRoot "models"
$BundleName  = "face_guard_bundle"
$BundleExe   = Join-Path $OutputDir "$BundleName.exe"

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

New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

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
    "--distpath",      $OutputDir,
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
# 7. Verify output
# ─────────────────────────────────────────────────────────────────────────────

Write-Header "Build Result"

if (-not (Test-Path $BundleExe)) {
    Write-Fail "PyInstaller did not produce: $BundleExe"
    exit 1
}

$size = (Get-Item $BundleExe).Length / 1MB
Write-OK ("face_guard_bundle.exe  ({0:F1} MB)" -f $size)
Write-OK "Output: $BundleExe"
Write-Host ""
Write-Host "Deploy to the target machine at:" -ForegroundColor Cyan
Write-Host "  %APPDATA%\PassQuantum\face_guard_bundle.exe" -ForegroundColor White

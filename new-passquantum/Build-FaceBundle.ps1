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
    plus all transitive libraries (mediapipe, opencv-python, numpy, …).

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
    Python interpreter to use.  Defaults to 'python'.
    Override with a full path when multiple interpreters are installed,
    e.g.  -Python "C:\Python311\python.exe"

.EXAMPLE
    # Basic — output lands in .\ui\
    .\Build-FaceBundle.ps1

.EXAMPLE
    # Custom output directory
    .\Build-FaceBundle.ps1 -OutputDir "C:\Publish\PassQuantum"

.EXAMPLE
    # Use a specific Python interpreter
    .\Build-FaceBundle.ps1 -Python "C:\Python311\python.exe"
#>

[CmdletBinding()]
param(
    [string] $OutputDir = (Join-Path $PSScriptRoot "ui"),
    [string] $WorkDir   = (Join-Path $env:TEMP "passquantum_pyinstaller_work"),
    [string] $SpecDir   = (Join-Path $env:TEMP "passquantum_pyinstaller_spec"),
    [string] $Python    = "python"
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

function Write-Step([string]$msg)    { Write-Host "=> $msg" -ForegroundColor Yellow }
function Write-OK([string]$msg)      { Write-Host "OK  $msg" -ForegroundColor Green  }
function Write-Fail([string]$msg)    { Write-Host "ERR $msg" -ForegroundColor Red    }

# ─────────────────────────────────────────────────────────────────────────────
# 1. Verify Python
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
# 2. Install / verify PyInstaller and runtime dependencies
# ─────────────────────────────────────────────────────────────────────────────

Write-Header "Installing Dependencies"

$deps = @("pyinstaller", "mediapipe", "opencv-python", "numpy")
Write-Step "pip install $($deps -join ' ')"
& $Python -m pip install --quiet @deps
if ($LASTEXITCODE -ne 0) {
    Write-Fail "pip install failed (exit $LASTEXITCODE)."
    exit 1
}
Write-OK "Dependencies ready."

# ─────────────────────────────────────────────────────────────────────────────
# 3. Resolve absolute paths
#    Using $PSScriptRoot ensures PyInstaller finds assets regardless of where
#    the spec file is written (the same root cause fixed for the Linux CI job).
# ─────────────────────────────────────────────────────────────────────────────

$ProjectRoot = $PSScriptRoot
$EntryPoint  = Join-Path $ProjectRoot "face_guard.py"
$ModelsDir   = Join-Path $ProjectRoot "models"
$BundleName  = "face_guard_bundle"
$BundleExe   = Join-Path $OutputDir "$BundleName.exe"

# Validate required source files exist before invoking PyInstaller
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
# 4. Clean previous artefacts
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
# 5. Run PyInstaller
#
#    --add-data uses ';' as the separator on Windows.
#    --paths    adds $ProjectRoot to sys.path so PyInstaller resolves the local
#               modules (geometric_encoder, liveness_detector, face_authenticator)
#               the same way Python does at runtime.
# ─────────────────────────────────────────────────────────────────────────────

Write-Header "Running PyInstaller"

$PyInstallerArgs = @(
    "-m", "PyInstaller",
    "--onefile",
    "--name",         $BundleName,
    "--distpath",     $OutputDir,
    "--workpath",     $WorkDir,
    "--specpath",     $SpecDir,
    "--paths",        $ProjectRoot,
    "--add-data",     "${ModelsDir};models",
    "--collect-all",  "mediapipe",
    "--hidden-import","cv2",
    "--hidden-import","numpy",
    "--hidden-import","geometric_encoder",
    "--hidden-import","liveness_detector",
    "--hidden-import","face_authenticator",
    "--noconfirm",
    "--clean",
    $EntryPoint
)

Write-Step "Command: $Python $($PyInstallerArgs -join ' ')"
Write-Host ""

& $Python @PyInstallerArgs
if ($LASTEXITCODE -ne 0) {
    Write-Fail "PyInstaller exited with code $LASTEXITCODE."
    exit 1
}

# ─────────────────────────────────────────────────────────────────────────────
# 6. Verify output
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

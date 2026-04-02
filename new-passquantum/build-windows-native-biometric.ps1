#requires -Version 5.1
<#
Native Windows biometric build bootstrap for PassQuantum.
- Installs missing dependencies (winget + MSYS2 packages)
- Verifies OpenCV/CGO toolchain
- Verifies ONNX model files are present and not HTML placeholders
- Builds PassQuantum.exe with biometric support enabled (no nobiometric tag)

Run in PowerShell (preferably as Administrator for installs):
  powershell -ExecutionPolicy Bypass -File .\build-windows-native-biometric.ps1

Optional:
  powershell -ExecutionPolicy Bypass -File .\build-windows-native-biometric.ps1 -ProjectRoot C:\dev\PassQuantum\new-passquantum
#>

param(
    [string]$ProjectRoot = "C:\dev\PassQuantum\new-passquantum",
    [switch]$SkipInstalls
)

$ErrorActionPreference = "Stop"

function Write-Step($msg) { Write-Host "`n=== $msg ===" -ForegroundColor Cyan }
function Write-Info($msg) { Write-Host "[INFO] $msg" -ForegroundColor Yellow }
function Write-Ok($msg)   { Write-Host "[OK]   $msg" -ForegroundColor Green }
function Write-Fail($msg) { Write-Host "[FAIL] $msg" -ForegroundColor Red }

function Require-Command([string]$Name, [string]$Hint) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Missing command '$Name'. $Hint"
    }
}

function Ensure-WingetPackage([string]$Id, [string]$DisplayName) {
    $installed = winget list --id $Id --accept-source-agreements 2>$null
    if ($LASTEXITCODE -eq 0 -and $installed) {
        Write-Ok "$DisplayName already installed"
        return
    }

    if ($SkipInstalls) {
        throw "$DisplayName is not installed and -SkipInstalls was passed"
    }

    Write-Info "Installing $DisplayName via winget..."
    winget install --id $Id --exact --accept-package-agreements --accept-source-agreements
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to install $DisplayName via winget"
    }
    Write-Ok "$DisplayName installed"
}

function Ensure-MSYS2Packages([string[]]$Packages) {
    $bash = "C:\msys64\usr\bin\bash.exe"
    if (-not (Test-Path $bash)) {
        throw "MSYS2 bash not found at $bash"
    }

    Write-Info "Updating MSYS2 package database..."
    & $bash -lc "pacman -Sy --noconfirm"

    $pkgLine = ($Packages -join " ")
    Write-Info "Installing MSYS2 packages: $pkgLine"
    & $bash -lc "pacman -S --needed --noconfirm $pkgLine"
    if ($LASTEXITCODE -ne 0) {
        throw "Failed installing MSYS2 packages"
    }

    Write-Ok "MSYS2 packages installed"
}

function Ensure-ProjectRoot {
    if (-not (Test-Path $ProjectRoot)) {
        throw "Project root does not exist: $ProjectRoot"
    }
    $resolved = (Resolve-Path $ProjectRoot).Path
    Set-Location $resolved
    Write-Ok "Using project root: $resolved"
}

function Ensure-Models {
    $blaze = Join-Path $ProjectRoot "models\blazeface.onnx"
    $mesh  = Join-Path $ProjectRoot "models\face_mesh.onnx"

    foreach ($model in @($blaze, $mesh)) {
        if (-not (Test-Path $model)) {
            throw "Missing model file: $model"
        }

        $bytes = [System.IO.File]::ReadAllBytes($model)
        if ($bytes.Length -lt 1024) {
            throw "Model file too small to be valid ONNX: $model ($($bytes.Length) bytes)"
        }

        $headLen = [Math]::Min($bytes.Length, 256)
        $headRaw = [System.Text.Encoding]::UTF8.GetString($bytes, 0, $headLen)
        $head = $headRaw.Trim().ToLowerInvariant()

        if ($head.StartsWith("<!doctype html") -or $head.StartsWith("<html")) {
            throw "Model file appears to be HTML placeholder, not ONNX binary: $model"
        }
        if ($head.StartsWith("version https://git-lfs.github.com/spec/v1")) {
            throw "Model file is a git-lfs pointer, not ONNX binary: $model"
        }
    }

    Write-Ok "Model files look valid"
}

function Configure-ToolchainEnv {
    $mingwBin = "C:\msys64\mingw64\bin"
    $pkgCfgDir = "C:\msys64\mingw64\lib\pkgconfig"

    if (-not (Test-Path (Join-Path $mingwBin "gcc.exe"))) {
        throw "gcc not found in $mingwBin"
    }
    if (-not (Test-Path (Join-Path $mingwBin "pkg-config.exe"))) {
        throw "pkg-config not found in $mingwBin"
    }

    $env:CC = Join-Path $mingwBin "gcc.exe"
    $env:CXX = Join-Path $mingwBin "g++.exe"
    $env:CGO_ENABLED = "1"
    $env:PKG_CONFIG = Join-Path $mingwBin "pkg-config.exe"
    $env:PKG_CONFIG_PATH = $pkgCfgDir
    $env:Path = "$mingwBin;$env:Path"

    Write-Ok "Configured CGO/OpenCV toolchain environment"
}

function Verify-Toolchain {
    Write-Info "Go version:"
    go version

    Write-Info "GCC version:"
    & $env:CC --version | Select-Object -First 1

    Write-Info "pkg-config version:"
    & $env:PKG_CONFIG --version

    Write-Info "OpenCV (pkg-config opencv4):"
    & $env:PKG_CONFIG --modversion opencv4
    if ($LASTEXITCODE -ne 0) {
        throw "OpenCV pkg-config metadata (opencv4) not found"
    }

    Write-Info "go env (CGO + compiler):"
    go env CGO_ENABLED CC CXX

    Write-Ok "Toolchain verification succeeded"
}

function Build-App {
    $outDir = Join-Path $ProjectRoot "build\windows"
    $outExe = Join-Path $outDir "PassQuantum.exe"

    if (-not (Test-Path $outDir)) {
        New-Item -ItemType Directory -Path $outDir | Out-Null
    }

    if (-not (Test-Path (Join-Path $ProjectRoot "vendor"))) {
        Write-Info "vendor/ not found, running go mod vendor..."
        go mod vendor
    }

    Write-Info "Building native Windows EXE with biometric enabled..."
    go build -mod=vendor -o $outExe .\ui
    if ($LASTEXITCODE -ne 0) {
        throw "go build failed"
    }

    Write-Ok "Build completed: $outExe"
}

function Write-RunLauncher {
    $launcher = Join-Path $ProjectRoot "build\windows\run-passquantum.cmd"
    $content = @"
@echo off
setlocal
set PATH=C:\msys64\mingw64\bin;%PATH%
cd /d %~dp0
PassQuantum.exe
"@
    Set-Content -Path $launcher -Value $content -Encoding ASCII
    Write-Ok "Created launcher: $launcher"
}

try {
    Write-Step "Checking base tools"
    Require-Command "winget" "Install App Installer from Microsoft Store first."

    Write-Step "Installing required applications"
    Ensure-WingetPackage -Id "GoLang.Go" -DisplayName "Go"
    Ensure-WingetPackage -Id "MSYS2.MSYS2" -DisplayName "MSYS2"

    Write-Step "Installing MSYS2 toolchain and OpenCV"
    Ensure-MSYS2Packages -Packages @(
        "mingw-w64-x86_64-gcc",
        "mingw-w64-x86_64-pkgconf",
        "mingw-w64-x86_64-opencv",
        "mingw-w64-x86_64-toolchain",
        "make"
    )

    Write-Step "Preparing project"
    Ensure-ProjectRoot
    Ensure-Models

    Write-Step "Configuring build environment"
    Configure-ToolchainEnv
    Verify-Toolchain

    Write-Step "Compiling PassQuantum"
    Build-App
    Write-RunLauncher

    Write-Step "Done"
    Write-Ok "Native Windows biometric build is ready."
    Write-Host "Run this on Windows from your project folder:" -ForegroundColor Cyan
    Write-Host "  .\build\windows\run-passquantum.cmd" -ForegroundColor White
    Write-Host "or" -ForegroundColor Cyan
    Write-Host "  .\build\windows\PassQuantum.exe" -ForegroundColor White
}
catch {
    Write-Fail $_.Exception.Message
    exit 1
}

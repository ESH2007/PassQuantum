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
    [string]$ProjectRoot = "C:\Users\Lenovo\OneDrive\Documents\Projectos-VSCode2\PassQuantum\new-passquantum",
    [switch]$SkipInstalls
)

$BlazeFaceModelURL = "https://github.com/PINTO0309/PINTO_model_zoo/raw/main/030_BlazeFace/01_float32/blazeface.onnx"
$FaceMeshModelURL  = "https://github.com/PINTO0309/PINTO_model_zoo/raw/main/032_FaceMesh/01_float32/face_mesh.onnx"

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

function Test-ModelFile([string]$modelPath) {
    if (-not (Test-Path $modelPath)) {
        return "missing"
    }

    $bytes = [System.IO.File]::ReadAllBytes($modelPath)
    if ($bytes.Length -lt 1024) {
        return "too-small"
    }

    $headLen = [Math]::Min($bytes.Length, 256)
    $headRaw = [System.Text.Encoding]::UTF8.GetString($bytes, 0, $headLen)
    $head = $headRaw.Trim().ToLowerInvariant()

    if ($head.StartsWith("<!doctype html") -or $head.StartsWith("<html")) {
        return "html"
    }
    if ($head.StartsWith("version https://git-lfs.github.com/spec/v1")) {
        return "git-lfs"
    }

    return "ok"
}

function Download-Model([string]$url, [string]$destination) {
    if ($SkipInstalls) {
        throw "Model download needed but -SkipInstalls was passed. Download manually: $url"
    }

    $destDir = Split-Path -Parent $destination
    if (-not (Test-Path $destDir)) {
        New-Item -Path $destDir -ItemType Directory -Force | Out-Null
    }

    Write-Info "Downloading model: $url"
    $downloaded = $false

    $curlCmd = Get-Command "curl.exe" -ErrorAction SilentlyContinue
    if ($curlCmd) {
        & $curlCmd.Path -L --fail --silent --show-error -o $destination $url
        if ($LASTEXITCODE -eq 0) {
            $downloaded = $true
        }
    }

    if (-not $downloaded) {
        try {
            Invoke-WebRequest -Uri $url -OutFile $destination -UseBasicParsing
            $downloaded = $true
        }
        catch {
            throw "Failed to download model from ${url}: $($_.Exception.Message)"
        }
    }

    Write-Ok "Downloaded model: $destination"
}

function Ensure-Models {
    $blaze = Join-Path $ProjectRoot "models\blazeface.onnx"
    $mesh  = Join-Path $ProjectRoot "models\face_mesh.onnx"

    $sources = @(
        @{ Name = "BlazeFace"; Path = $blaze; URL = $BlazeFaceModelURL },
        @{ Name = "FaceMesh";  Path = $mesh;  URL = $FaceMeshModelURL }
    )

    foreach ($source in $sources) {
        $status = Test-ModelFile $source.Path
        if ($status -ne "ok") {
            Write-Info "$($source.Name) model at $($source.Path) is '$status'. Refreshing from PINTO source."
            Download-Model -url $source.URL -destination $source.Path
            $status = Test-ModelFile $source.Path
            if ($status -ne "ok") {
                throw "$($source.Name) model is still invalid after download ($status): $($source.Path). Source URL: $($source.URL)"
            }
        } else {
            Write-Ok "$($source.Name) model present: $($source.Path)"
        }
    }

    Write-Ok "Model files look valid (PINTO OpenCV-compatible sources configured)"
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

    # Use pkg-config to provide exact OpenCV include/library flags for MSYS2 layout.
    $env:CGO_CFLAGS = (& $env:PKG_CONFIG --cflags opencv4).Trim()
    $env:CGO_LDFLAGS = (& $env:PKG_CONFIG --libs opencv4).Trim()
    
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
        throw "OpenCV pkg-config metadata (opencv4) not found. Ensure mingw-w64-x86_64-opencv is installed."
    }

    # Verify the canonical OpenCV header path used by gocv.
    $opencvHeader = "C:\msys64\mingw64\include\opencv4\opencv2\opencv.hpp"
    if (-not (Test-Path $opencvHeader)) {
        throw "OpenCV header not found at $opencvHeader. Reinstall mingw-w64-x86_64-opencv."
    }

    Write-Info "go env (CGO + compiler):"
    go env CGO_ENABLED CC CXX CGO_CFLAGS CGO_LDFLAGS

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
    go build -tags customenv -mod=vendor -o $outExe .\ui
    if ($LASTEXITCODE -ne 0) {
        throw "go build failed"
    }

    Write-Ok "Build completed: $outExe"
}

function Copy-RuntimeDependencies {
    $mingwBin = "C:\msys64\mingw64\bin"
    $outDir = Join-Path $ProjectRoot "build\windows"

    if (-not (Test-Path $mingwBin)) {
        throw "MSYS2 MinGW bin directory not found: $mingwBin"
    }

    # Bundle MinGW runtime DLLs so the EXE can run outside shells that already have MSYS2 in PATH.
    $files = Get-ChildItem -Path $mingwBin -Filter "*.dll" -ErrorAction SilentlyContinue
    $copied = 0
    foreach ($file in $files) {
        Copy-Item -Path $file.FullName -Destination (Join-Path $outDir $file.Name) -Force
        $copied++
    }

    if ($copied -eq 0) {
        throw "No runtime DLLs were copied from $mingwBin"
    }

    Write-Ok "Copied runtime DLLs: $copied files"
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
    Copy-RuntimeDependencies
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

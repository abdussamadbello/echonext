# EchoNext CLI Installer for Windows
# Run: iwr -useb https://raw.githubusercontent.com/abdussamadbello/echonext/master/install.ps1 | iex

$ErrorActionPreference = "Stop"

$InstallDir = "$env:LOCALAPPDATA\bin"
$BinaryName = "echonext.exe"

Write-Host "EchoNext CLI Installer" -ForegroundColor Cyan
Write-Host "======================" -ForegroundColor Cyan
Write-Host ""

# Check Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Go is required but not installed." -ForegroundColor Red
    Write-Host "Please install Go from https://go.dev/dl/" -ForegroundColor Yellow
    exit 1
}

$GoVersion = (go version) -replace "go version ", ""
Write-Host "Found $GoVersion"

# Create install directory
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Write-Host "Created directory: $InstallDir"
}

# Install using go install
Write-Host "Installing EchoNext CLI..."
go install github.com/abdussamadbello/echonext/cmd/echonext-cli@latest

if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: go install failed" -ForegroundColor Red
    exit 1
}

# Get GOBIN path
$GoBin = (go env GOBIN)
if ([string]::IsNullOrEmpty($GoBin)) {
    $GoBin = (go env GOPATH) + "\bin"
}

$SourceBinary = Join-Path $GoBin "echonext-cli.exe"
$DestBinary = Join-Path $InstallDir $BinaryName

# Copy to install directory with renamed binary
if (Test-Path $SourceBinary) {
    Copy-Item $SourceBinary $DestBinary -Force
    Write-Host "Installed $BinaryName to $InstallDir"
} else {
    Write-Host "Error: Binary not found at $SourceBinary" -ForegroundColor Red
    exit 1
}

# Check if install directory is in PATH
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host ""
    Write-Host "Adding $InstallDir to your PATH..." -ForegroundColor Yellow

    # Add to user PATH
    $NewPath = if ([string]::IsNullOrEmpty($UserPath)) {
        $InstallDir
    } else {
        "$UserPath;$InstallDir"
    }
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")

    Write-Host "PATH updated successfully." -ForegroundColor Green
    Write-Host ""
    Write-Host "NOTE: Restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
} else {
    Write-Host ""
    Write-Host "You can now use 'echonext' from anywhere!" -ForegroundColor Green
}

Write-Host ""
Write-Host "Verify installation:" -ForegroundColor Cyan
Write-Host "  echonext version"
Write-Host ""
Write-Host "Get started:" -ForegroundColor Cyan
Write-Host "  echonext init myproject"

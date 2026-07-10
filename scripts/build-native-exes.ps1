$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

New-Item -ItemType Directory -Force -Path "dist" | Out-Null

Write-Host "Building Windows x64 executable..."
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -trimpath -ldflags "-s -w -H=windowsgui" -o "dist/ProxyEnvSwitch_x64.exe" "native/proxyenv_switch_windows.go"

Write-Host "Building Windows ARM64 executable..."
$env:GOARCH = "arm64"
go build -trimpath -ldflags "-s -w -H=windowsgui" -o "dist/ProxyEnvSwitch_ARM64.exe" "native/proxyenv_switch_windows.go"

Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue

Write-Host "Build completed. Files are in $repoRoot\dist"

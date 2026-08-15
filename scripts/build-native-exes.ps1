$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go was not found. Install Go 1.23 or later and reopen PowerShell."
}

New-Item -ItemType Directory -Force -Path "dist" | Out-Null

$oldGoos = $env:GOOS
$oldGoarch = $env:GOARCH
try {
    Write-Host "Formatting Go source..."
    gofmt -w native/*.go
    if ($LASTEXITCODE -ne 0) { throw "gofmt failed." }

    Write-Host "Building Windows x64 executable..."
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -trimpath -ldflags "-s -w -H=windowsgui" -o "dist/ProxyEnvSwitch_x64.exe" ./native
    if ($LASTEXITCODE -ne 0) { throw "The x64 Go build failed." }

    Write-Host "Building Windows ARM64 executable..."
    $env:GOARCH = "arm64"
    go build -trimpath -ldflags "-s -w -H=windowsgui" -o "dist/ProxyEnvSwitch_ARM64.exe" ./native
    if ($LASTEXITCODE -ne 0) { throw "The ARM64 Go build failed." }
}
finally {
    $env:GOOS = $oldGoos
    $env:GOARCH = $oldGoarch
}

$checksumLines = @()
foreach ($file in @("dist/ProxyEnvSwitch_x64.exe", "dist/ProxyEnvSwitch_ARM64.exe")) {
    $hash = (Get-FileHash -Algorithm SHA256 $file).Hash.ToLower()
    $checksumLines += "$hash  $([System.IO.Path]::GetFileName($file))"
}
$checksumLines | Set-Content -Encoding ascii "dist/SHA256SUMS.txt"

Write-Host "Build completed. Files are in $repoRoot\dist"

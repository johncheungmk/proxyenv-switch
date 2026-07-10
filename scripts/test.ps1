$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

if (-not (Get-Command py -ErrorAction SilentlyContinue)) {
    throw "Python Launcher (py.exe) was not found."
}

Write-Host "Compiling Python files..."
py -m py_compile src/proxyenv_core.py src/proxyenv_switch.pyw
if ($LASTEXITCODE -ne 0) { throw "Python compilation failed." }

Write-Host "Running Python unit tests..."
py -m unittest discover -s tests -v
if ($LASTEXITCODE -ne 0) { throw "Python tests failed." }

if (Get-Command go -ErrorAction SilentlyContinue) {
    Write-Host "Checking Go formatting and cross-compilation..."
    $unformatted = gofmt -l native/proxyenv_switch_windows.go
    if ($unformatted) { throw "Run gofmt on: $unformatted" }

    $oldGoos = $env:GOOS
    $oldGoarch = $env:GOARCH
    try {
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        go build -o "$env:TEMP/ProxyEnvSwitch-test.exe" ./native
        if ($LASTEXITCODE -ne 0) { throw "Go cross-compilation failed." }
        Remove-Item "$env:TEMP/ProxyEnvSwitch-test.exe" -ErrorAction SilentlyContinue
    }
    finally {
        $env:GOOS = $oldGoos
        $env:GOARCH = $oldGoarch
    }
} else {
    Write-Warning "Go is not installed; native build validation was skipped."
}

Write-Host "All available checks passed."

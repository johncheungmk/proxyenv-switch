# ProxyEnv Switch

![Version](https://img.shields.io/badge/version-1.2.1-blue)
![Platform](https://img.shields.io/badge/platform-Windows%2011-0078D4)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A lightweight Windows 11 utility for adding, updating, or removing the current user's `HTTP_PROXY` and `HTTPS_PROXY` environment variables.

<p align="left">
  <img src="assets/proxyenv-switch.png" alt="ProxyEnv Switch icon" width="128">
</p>

## Download

### Windows 11 on Intel or AMD (x64)

**[Download ProxyEnvSwitch_x64.exe](https://github.com/johncheungmk/proxyenv-switch/raw/refs/heads/main/dist/ProxyEnvSwitch_x64.exe)**

### Windows 11 on ARM64

**[Download ProxyEnvSwitch_ARM64.exe](https://github.com/johncheungmk/proxyenv-switch/raw/refs/heads/main/dist/ProxyEnvSwitch_ARM64.exe)**

You can also open the **[GitHub Releases page](https://github.com/johncheungmk/proxyenv-switch/releases/latest)** for release bundles and checksums.

> The direct links above download the executables stored in the repository's `dist` folder. This makes the download available even before GitHub Actions or release assets are configured.

The executables are not commercially code-signed. Microsoft Defender SmartScreen may show an “unrecognized app” warning. Verify [`dist/SHA256SUMS.txt`](dist/SHA256SUMS.txt) or build the application from source before running it.

## What it does

Enter a local proxy port such as `60505`, then choose an action:

- **Add / Update Proxy** sets both variables to the same value:

  ```text
  HTTP_PROXY=http://127.0.0.1:60505
  HTTPS_PROXY=http://127.0.0.1:60505
  ```

- **Remove Proxy** deletes both variables.

The utility changes only these current-user registry values:

```text
HKEY_CURRENT_USER\Environment\HTTP_PROXY
HKEY_CURRENT_USER\Environment\HTTPS_PROXY
```

Administrator rights are not required. ProxyEnv Switch does not modify machine-wide variables, Windows Settings proxy, WinHTTP proxy, browser settings, or VPN configuration.

> Applications must support the `HTTP_PROXY` and `HTTPS_PROXY` environment-variable convention. Already-running applications normally keep the values inherited when they started, so reopen them after a change.

## Version 1.2.1 improvements

- Added prominent direct-download links to the README.
- Replaced fragile Build and dynamic Release badges with stable static badges.
- Added a GitHub web-upload guide for workflows and release assets.
- Updated the included x64 and ARM64 executables to version 1.2.1.
- Retained the version 1.2 registry, rollback, verification, DPI-awareness, and test improvements.

## Repository structure

```text
proxyenv-switch/
├── .github/                 Workflows, issue template, Dependabot
├── assets/                  Application icon
├── dist/                    Ready-to-download Windows executables
├── docs/                    Repository and release guidance
├── native/                  Native Win32 implementation in Go
├── scripts/                 Build and test scripts
├── src/                     Python/Tkinter implementation and core helpers
├── tests/                   Platform-independent Python unit tests
├── CHANGELOG.md
├── LICENSE
├── README.md
├── SECURITY.md
├── VERSION
├── go.mod
└── requirements-dev.txt
```

## Build and test

### Python application

Requirements: Windows 11 and Python 3.10 or later.

```bat
scripts\build-python-exe.bat
```

Output:

```text
dist\ProxyEnvSwitch.exe
```

### Native x64 and ARM64 applications

Requirement: Go 1.23 or later.

```powershell
powershell -ExecutionPolicy Bypass -File scripts\build-native-exes.ps1
```

Outputs:

```text
dist\ProxyEnvSwitch_x64.exe
dist\ProxyEnvSwitch_ARM64.exe
```

### Tests

```powershell
powershell -ExecutionPolicy Bypass -File scripts\test.ps1
```

## GitHub Actions and releases

The Build badge was deliberately removed from the README because it shows a broken image until `.github/workflows/build.yml` is actually committed to the repository.

When using the GitHub website rather than Git, create or upload these exact paths:

```text
.github/workflows/build.yml
.github/workflows/release.yml
```

See [`docs/GITHUB_WEB_UPLOAD_FIX.md`](docs/GITHUB_WEB_UPLOAD_FIX.md) for step-by-step instructions.

For an automated release, commit the complete repository and push a matching version tag:

```powershell
git tag v1.2.1
git push origin v1.2.1
```

The release workflow validates the tag against `VERSION`, builds the executables, generates checksums, creates a portable ZIP, and publishes the assets.

## Security and privacy

ProxyEnv Switch:

- Writes or removes only `HTTP_PROXY` and `HTTPS_PROXY` for the current user
- Does not require administrator rights
- Does not connect to the Internet
- Does not transmit or collect data
- Does not install a service or run in the background

Review [`SECURITY.md`](SECURITY.md) for vulnerability reporting.

## License

MIT License. See [`LICENSE`](LICENSE).

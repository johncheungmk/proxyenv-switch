# ProxyEnv Switch

[![Build](https://github.com/johncheungmk/proxyenv-switch/actions/workflows/build.yml/badge.svg)](https://github.com/johncheungmk/proxyenv-switch/actions/workflows/build.yml)
[![Latest release](https://img.shields.io/github/v/release/johncheungmk/proxyenv-switch)](https://github.com/johncheungmk/proxyenv-switch/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A lightweight Windows 11 utility for adding, updating, or removing the current user's `HTTP_PROXY` and `HTTPS_PROXY` environment variables.

<img src="assets/proxyenv-switch.png" alt="ProxyEnv Switch icon" width="96">

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

## Download

Open the repository's **Releases** page and download:

- `ProxyEnvSwitch_x64.exe` for most Intel and AMD Windows 11 computers
- `ProxyEnvSwitch_ARM64.exe` for Windows 11 on ARM
- `ProxyEnvSwitch.exe` for the Python/Tkinter build produced on a Windows GitHub runner
- `ProxyEnv-Switch-Windows-vX.Y.Z.zip` for the complete portable release bundle

The executables are not commercially code-signed. Microsoft Defender SmartScreen may therefore show an “unrecognized app” warning. Verify the published SHA-256 checksum or build the application from source before running it.

## Version 1.2 improvements

- Explicit Windows DPI-awareness handling in both implementations
- Resizable interface with wrapped text and larger minimum dimensions
- Complete display of the current persistent proxy values
- Direct Windows Registry API access in the native implementation
- Verified two-variable updates with rollback after a partial failure
- Correct pointer-sized Win32 environment-change notification handling
- Refresh button and `F5` refresh in the Python interface
- Unit tests, build validation, release packaging, and checksums

## Repository structure

```text
proxyenv-switch/
├── .github/                 Workflows, issue template, Dependabot
├── assets/                  Application icon
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

## Publish a GitHub release

After committing the complete repository, including the hidden `.github` directory:

```powershell
git tag v1.2.0
git push origin v1.2.0
```

The release workflow validates the tag against `VERSION`, builds all executables, generates checksums, creates a portable ZIP, and publishes the assets.

See [`docs/REPOSITORY_SETUP.md`](docs/REPOSITORY_SETUP.md) before updating the existing repository. This is important because a normal Windows drag-and-drop upload can omit hidden files such as `.github` and `.gitignore`.

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

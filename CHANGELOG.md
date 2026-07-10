# Changelog

All notable changes to this project are documented here.

## [1.2.0] - 2026-07-10

### Added

- Platform-independent Python core helpers and unit tests.
- Refresh button and `F5` refresh in the Python interface.
- Version display in both interfaces.
- Release ZIP generation and tag-to-version validation in GitHub Actions.
- Repository setup guidance for preserving hidden `.github` files.
- Dependabot configuration and a structured bug-report template.

### Changed

- Native implementation now uses the Windows Registry API directly instead of parsing `reg.exe` output.
- Both implementations verify that `HTTP_PROXY` and `HTTPS_PROXY` were changed together.
- Failed partial changes are rolled back to the previous values when possible.
- Native interface dimensions and fonts now scale with Windows DPI.
- README now explains that environment variables do not change the Windows system proxy.
- Build scripts use pinned development requirements and generate SHA-256 checksums.

### Fixed

- Python `SendMessageTimeoutW` now uses an explicit function signature and pointer-sized result storage.
- Missing GitHub workflow/release guidance that could leave the public repository without active Actions.
- README download guidance when no GitHub Release had yet been published.
- Potential native registry-read failures on non-English Windows installations.

## [1.1.0] - 2026-07-10

### Added

- Project renamed to **ProxyEnv Switch**.
- Current persistent `HTTP_PROXY` and `HTTPS_PROXY` values are displayed.
- GitHub Actions workflows for native and Python/PyInstaller builds.
- x64 and ARM64 native Windows build targets.
- Application icon, security policy, license, and release documentation.

### Changed

- Window is larger and resizable.
- Text wraps dynamically instead of being held in fixed-height controls.
- Native app is DPI-aware for Windows display scaling.
- Buttons expand cleanly and status messages have more space.

### Fixed

- Subtitle and status text could be clipped or partially hidden at higher Windows display scaling.

## [1.0.0] - 2026-07-10

- Initial add/update/remove proxy utility.

# Changelog

All notable changes to this project are documented here.

## [1.1.0] - 2026-07-10

### Added

- Project renamed to **ProxyEnv Switch**.
- Current persistent `HTTP_PROXY` and `HTTPS_PROXY` values are displayed.
- GitHub Actions workflows for native and Python/PyInstaller builds.
- x64 and ARM64 native Windows build targets.
- Application icon, security policy, license, and release documentation.

### Changed

- Window is now larger and resizable.
- Text wraps dynamically instead of being held in fixed-height controls.
- Native app is DPI-aware for Windows display scaling.
- Buttons expand cleanly and status messages have more space.

### Fixed

- Subtitle and status text could be clipped or partially hidden, particularly at 125%, 150%, or higher Windows display scaling.

## [1.0.0] - 2026-07-10

- Initial add/update/remove proxy utility.

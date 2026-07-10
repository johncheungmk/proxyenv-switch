# Changelog

All notable changes to this project are documented here.

## [1.2.1] - 2026-07-10

### Added

- Prominent direct download links for Windows x64 and ARM64 in the README.
- A step-by-step GitHub website upload and release guide.

### Changed

- Replaced the broken workflow badge and fragile dynamic release badge with stable version and platform badges.
- Updated bundled executables and checksums to version 1.2.1.

### Fixed

- README users could not directly download the application.
- Badges displayed `no releases or repo not found` or a broken image when workflows were not committed.

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

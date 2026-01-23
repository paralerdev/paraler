# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Detailed error messages when service fails to start (shows command and directory)
- Directory existence check before starting process

### Fixed
- Project detection for custom-named subdirectories (e.g., `myproject-api`, `myproject-web`)

## [0.2.0] - 2025-01-23

### Added
- **Copy mode** — press `y` to enter, select lines with `v`, copy with `y`/`Enter`
- **Fullscreen logs** — press `f` to hide sidebar for easier mouse selection
- **Status indicator in logs** — shows `[running]`/`[stopped]`/`[failed]` in log panel title
- **System messages** — logs show `▶ Service started`, `■ Service stopped`, `✖ Service failed`
- Move service between projects (`m` key)
- Rename project (`Ctrl+R` key)

### Changed
- Health indicator only shows when health check is configured (no more `?`)
- Focus returns to sidebar when exiting fullscreen mode

### Fixed
- Layout issues with ANSI escape codes from process output

## [0.1.0] - 2025-01-22

### Added
- Initial release
- TUI with sidebar (services) and main panel (logs)
- Start/stop/restart services with single keypress
- Bulk operations (start all, stop all)
- Log viewing with filtering and export
- Health checks (HTTP and port monitoring)
- Auto-restart on crash
- Service dependencies (start order)
- Hot config reload (Ctrl+R)
- Auto-discovery for common frameworks (NestJS, React, Vue, Go, etc.)
- Multi-select mode for bulk operations
- Custom service colors
- Environment variables display
- Port conflict detection

[Unreleased]: https://github.com/paralerdev/paraler/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/paralerdev/paraler/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/paralerdev/paraler/releases/tag/v0.1.0

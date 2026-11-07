# Changelog

All notable changes are documented here. The project follows Semantic Versioning once a stable release is published.

## 3.0.0-preview.1 — First public preview

### Added

- Per-user Windows installer and one-click signed updates from GitHub Releases. Installed copies automatically update on startup by default; portable installation requires an explicit click.
- Application updates tab with release notes, saved opt-out, distinct tray navigation and Windows single-instance activation.

- Go/Fyne desktop application for Windows and macOS without a WebView.
- Redesigned quota window, separate settings navigation and light/dark design tokens.
- Two-row 184×88 floating widget and dynamic tray indicators.
- Per-user autostart using Windows Run registry values and macOS LaunchAgents.
- Portable Windows and macOS archives with SHA-256 checksums and CI artifacts for Intel and Apple Silicon Macs.
- Public-repository documentation, issue templates and release guidance.

### Changed

- Settings are opened from the main window and appearance changes apply immediately.
- Missing or failed quota responses clear stale values instead of showing them as current.
- Settings are validated and saved through a temporary-file replacement.
- UI tests cover both themes, three display scales, settings pages and widget states.

### Known limitations

- No Windows Authenticode or macOS notarization; Ed25519 authentication covers in-app installer updates only.
- macOS installation and updating remain manual; real-hardware acceptance is pending. Popup quota notifications and quiet hours are not implemented.

## 2.0.0 — Superseded preview

- Earlier standalone Windows preview with configurable indicators and update UI scaffolding.
- Superseded by the Go/Fyne migration; no legacy WinForms source is included in this repository.

## 1.0.0

- Numeric tray indicator, app-server integration and variable quota windows.

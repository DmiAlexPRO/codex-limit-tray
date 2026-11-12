# Changelog

This file records user-facing changes. Release versions follow Semantic Versioning.

## 3.4.0 — Earned Codex limit resets

- See your available resets and their expiration dates when Codex provides them. Lists longer than five entries scroll in place.
- Use a reset from the dashboard after confirming. The app then refreshes the limits through the same local Codex connection.

See the [3.4.0 release notes](docs/releases/v3.4.0.md).

## 3.3.0 — Tray icon visibility

- Added tray icon visibility instructions in Appearance and both READMEs.
- On Windows, Appearance can open the system taskbar settings directly. On macOS, it explains where to find menu bar controls.

See the [3.3.0 release notes](docs/releases/v3.3.0.md).

## 3.2.0 — 27 interface languages

- Added Belarusian, Kazakh, Ukrainian, Polish, Italian, Turkish, Kyrgyz and Armenian, bringing the app to 27 interface languages.
- Gave dashboard actions room for longer translated labels.

See the [3.2.0 release notes](docs/releases/v3.2.0.md).

## 3.1.0 — 19 interface languages

- Localized the dashboard, settings, tray menu, quota labels, connection status and update screens.
- Added a saved language preference under **Appearance → Language**. The change takes effect after restarting.
- The Windows installer initializes the app language from the system on a fresh install. Portable copies and macOS detect it on first launch. English is the fallback; reinstalling or upgrading preserves an existing choice.

See the [3.1.0 release notes](docs/releases/v3.1.0.md).

## 3.0.1 — First release after the repository history cleanup

- Rebuilt with an updated Go toolchain and updated security-sensitive dependencies.
- The version increases from 3.0.0 so installed Windows copies can receive the signed update and keep their settings.

See the [3.0.1 release notes](docs/releases/v3.0.1.md).

## 3.0.0 — Stable for Windows x64

- 3.0.0 is the first GitHub release. The Windows edition follows stable releases, not prereleases. The primary download points to the latest stable installer.
- Both Windows editions retain their update behavior: the installed app can apply signed updates at startup; the portable ZIP installs nothing without an explicit action. Existing preferences are reused.
- The stable designation applies to Windows only. macOS remains experimental, with manual installation and updates. No Authenticode publisher signature or Apple notarization is provided.

See the [3.0.0 release notes](docs/releases/v3.0.0.md).

## Before 3.0.0 — Unpublished development history

The entries below describe development builds, not published GitHub releases.

### 3.0.0-preview.1

#### Added

- Per-user Windows installer and one-click signed updates from GitHub Releases. Installed copies automatically update on startup by default; portable installation requires an explicit click.
- Application updates tab with release notes, saved opt-out, distinct tray navigation and Windows single-instance activation.

- Go/Fyne desktop application for Windows and macOS without a WebView.
- Redesigned quota window, separate settings navigation and light/dark design tokens.
- Two-row 184×88 floating widget and dynamic tray indicators.
- Per-user autostart using Windows Run registry values and macOS LaunchAgents.
- Portable Windows and macOS archives with SHA-256 checksums and CI artifacts for Intel and Apple Silicon Macs.
- Public-repository documentation, issue templates and release guidance.

#### Changed

- Settings are opened from the main window and appearance changes apply immediately.
- Missing or failed quota responses clear stale values instead of showing them as current.
- Settings are validated and saved through a temporary-file replacement.
- UI tests cover both themes, three display scales, settings pages and widget states.

#### Known limitations

- No Windows Authenticode or macOS notarization; Ed25519 authentication covers in-app installer updates only.
- macOS installation and updating remain manual; real-hardware acceptance is pending. Popup quota notifications and quiet hours are not implemented.

### 2.0.0 — Superseded preview

- Earlier standalone Windows preview with configurable indicators and update UI scaffolding.
- Superseded by the Go/Fyne migration; no legacy WinForms source is included in this repository.

### 1.0.0

- Numeric tray indicator, app-server integration and variable quota windows.

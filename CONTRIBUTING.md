# Contributing

Bug reports and focused pull requests are welcome. By contributing, you agree that your changes are licensed under the project's MIT license.

## Development setup

Use Go 1.27.4+, Fyne 2.8.2 and CGO. Windows requires GCC/MinGW; macOS requires Xcode Command Line Tools. Do not commit `.tools`, `.cache`, `dist`, editor state or local design assets.

Before submitting:

```sh
gofmt -w cmd internal
go test ./...
go vet ./...
```

On Windows, `build.ps1 -Preview` runs the full build and generates UI screenshots. On macOS, run `sh scripts/build-macos.sh`. A Windows result does not validate native macOS behavior.

## Guidelines

- Add tests for parsing, retry, settings, autostart and UI behavior that you change.
- Use synthetic fixtures. Never include credentials, personal paths, raw server logs or real account responses.
- Treat missing data as unknown, not zero or full.
- Keep protocol, policy, rendering and native platform responsibilities separate.
- Check both themes, 100/150/200% scaling and keyboard access for UI changes.
- Preserve third-party attribution and review license compatibility before adding dependencies or assets.
- Use Conventional Commits, for example `fix: recover after network reconnect`.

## Platform changes

Treat the Windows portable ZIP and installed Setup edition as separate modes. Verify both when changing the UI or behavior. The portable edition checks for newer versions and opens the release on GitHub; only the installed edition can download and apply updates in the app.

Autostart must remain per-user and require no administrator privileges. Test enable, disable, application relocation and a path containing spaces. macOS pull requests should state which behaviors were checked on physical hardware.

## Localization

English is the source language. Add or update app strings in every catalog in `internal/i18n/locales/`; `go test ./...` checks coverage and placeholders. Keep both README files up to date when changing public instructions.

For GitHub Pages, update all 11 translations in `scripts/pages/locales.json`, then run `python scripts/build-pages.py` and commit the generated HTML and sitemap. Use `python scripts/build-pages.py --check` to verify they are current.

## Reporting security issues

Do not open a public issue for an exploitable vulnerability. Follow [SECURITY.md](SECURITY.md).

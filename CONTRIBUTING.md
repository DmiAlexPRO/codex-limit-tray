# Contributing

Bug reports and focused pull requests are welcome. By contributing, you agree that your changes are licensed under the project's MIT license.

## Development setup

Use Go 1.26+, Fyne 2.8.2 and CGO. Windows requires GCC/MinGW; macOS requires Xcode Command Line Tools. Do not commit `.tools`, `.cache`, `dist`, editor state or local design assets.

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
- Keep Russian UI wording consistent and English documentation clear.
- Preserve third-party attribution and review license compatibility before adding dependencies or assets.
- Use Conventional Commits, for example `fix: recover after network reconnect`.

## Platform changes

Autostart must remain per-user and require no administrator privileges. Test enable, disable, application relocation and a path containing spaces. macOS pull requests should state which behaviors were checked on physical hardware.

## Reporting security issues

Do not open a public issue for an exploitable vulnerability. Follow [SECURITY.md](SECURITY.md).

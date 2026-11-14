# Release checklist

### One-time setup

1. Enable GitHub Actions and create a protected environment named `release` in [the release repository](https://github.com/DmiAlexPRO/codex-limit-tray). Restrict release tags to trusted maintainers and require review of the tagged code and signing job. Configure a private vulnerability reporting channel. Never give PR builds release secrets.
2. **For this project the first key pair has already been generated.** The public verification key is in `packaging/update-public-key.txt`. The maintainer's local copy of the private seed is encrypted with Windows DPAPI (CurrentUser) in the ignored `.local/update-signing-seed.dpapi`. Do not put that file in Git. From the original Windows account, after logging into GitHub CLI as a maintainer, run the following once from the repository root to send the secret directly to the protected environment without printing it:

   ```powershell
   Add-Type -AssemblyName System.Security
   $blob = [IO.File]::ReadAllBytes((Resolve-Path '.local/update-signing-seed.dpapi').Path)
   $seed = [Security.Cryptography.ProtectedData]::Unprotect($blob, $null, [Security.Cryptography.DataProtectionScope]::CurrentUser)
   try {
       [Convert]::ToBase64String($seed) | gh secret set UPDATE_SIGNING_SEED --env release --repo DmiAlexPRO/codex-limit-tray
       if ($LASTEXITCODE -ne 0) { throw 'Secret was not configured.' }
   } finally { [Array]::Clear($seed, 0, $seed.Length) }
   ```

   Protect your account and keep a separate encrypted, recoverable offline backup: this DPAPI file is not a portable backup. GitHub will not let you read the environment secret back. Do not regenerate the key for each release.

3. **For a fresh independent fork only:** generate a *different* Ed25519 key on a trusted machine, in a private folder outside the checkout:

   ```sh
   go run ./cmd/sign-update -generate-key /private/path/codex-limit-tray.seed
   ```

   The command refuses overwrites; stdout is the **public key only**. Replace the fork's pinned public key, set the matching seed as the environment secret and change the local build defaults. Never paste a seed into chat, logs or commits. Changing the key in CI alone will not update the trust anchor in existing clients.
4. CI repository identity comes from `github.repository`; no PAT belongs in the desktop app. CI uses the checked-in public key unless overridden by `UPDATE_PUBLIC_KEY`. The Windows build script defaults to the official repo and pinned key. For a custom build: `./build.ps1 -Installer -Version <semver> -Repository <owner/repo> -PublicKey <public-base64>`.

### Each release

1. Review exactly what goes into Git. Exclude `.local`, `.tools`, `.cache`, `dist`, private keys, personal paths and unlicensed assets. Update the changelog, third-party notices and `docs/releases/v<version>.md`. The workflow uses that file as the GitHub Release description when present. Run `build.ps1 -Installer -Preview` and the platform checks below.
2. Tag the reviewed commit with a strictly increasing version, e.g. `v3.0.0-preview.1`, then `v3.0.0-preview.2`; publish `v3.0.0` when ready for stable. Do not reuse tags or replace old assets. Stable users never receive previews.
3. Tag CI stages the Windows installer, portable ZIP and both Mac archives. The protected job checks the signing key pair, signs the exact version and Setup bytes, creates combined checksums and publishes a GitHub Release with notes. A preview tag creates a prerelease. The signed asset name is always `CodexLimitTray-windows-x64-Setup.exe`, with detached `.sig` beside it.
4. Review generated notes; edit the **description of the GitHub Release** to provide short user-facing changes. Edit text only, not already published binaries. Clients display this body as plain text. If adding OS signing, sign the app before packaging, then sign the installer **before** computing its update signature and checksums.

## Manual acceptance

- Language: on a clean Windows install, confirm the system language (or English fallback) is selected. Change it in Appearance and confirm the dashboard, settings, tray menu and dialogs switch before restarting. Confirm the choice survives a restart and an upgrade. Spot-check Chinese, an Arabic-script locale, and an Indic-script locale.
- Fresh user: first launch, real login, one/two quota windows, logout/account switch and missing-data behavior.
- Network loss/reconnect, timeout, retry backoff and suspend/resume.
- Tray at 100/150/200% DPI; hide/show, minimize/close, widget drag/hide and taskbar/Alt+Tab behavior.
- Four indicator styles, 0/100%, both themes, system-theme changes, threshold colors and sound.
- Autostart enable/disable, log out/in, paths containing spaces and moving the portable application. Confirm that disabling removes the Windows Run value or macOS LaunchAgent.
- Windows: Explorer restart and behavior after replacing the executable.
- On a disposable Windows account, install release A without elevation, change a setting, then publish a signed B in a separate test repo/key. Verify the default startup update and tray restart, that B shows its notes and does not install B again, manual-only mode after opting out, and clean uninstall of the Run entry. A portable A must show a GitHub link for B without downloading or installing anything. Test offline, missing/corrupt signature and locked files.
- macOS: CI checks a macOS 13.0 deployment target for both Intel and Apple Silicon. On both architectures, check the menu bar, accessory mode, always-on-top behavior and LaunchAgent on real hardware.
- Signatures, notarization and SmartScreen/Gatekeeper presentation. Never describe unsigned builds as signed or trusted.

Automated tests use synthetic data. They do not establish live API compatibility, clean-machine acceptance or platform signing reputation.

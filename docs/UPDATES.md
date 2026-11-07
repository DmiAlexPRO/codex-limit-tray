# Application updates

**Installed Windows x64:** checks once on each launch (including launch at login). Auto-install is on by default and can be turned off in **Настройки → Обновления**. When disabled, there is no startup check. Manual checks show the version and the plain-text body of the GitHub Release. One click downloads and applies it without asking the user to exit or find an installer. A successful update restarts in the tray.

**Portable Windows x64:** does not check or install automatically. The explicit install button converts the latest published version, including the exact current one, into a per-user installation. It keeps settings and leaves the original portable EXE alone. **macOS:** manual checks and notes only. **Demo:** no update network access, download or installation.

The installed copy is identified by a marker shipped only by Setup. The updater never overwrites its running binary: it stages and verifies a complete installer, then hands off to Inno Setup. The installer waits for the caller to exit. Updates keep the existing settings and autostart preference. On download, verification or handoff failure the running copy is not replaced. If a completed download's installation fails after exit, relaunch from Start, or reinstall a known-good release manually; there is no automatic post-install rollback.

## Channel and trust

- The repository, version and public Ed25519 key are embedded at build time. Stable builds query the latest public non-preview release. Preview builds compare the 30 most recent releases (stable and preview); drafts, malformed tags and downgrades are ignored. Equal versions show notes but never trigger automatic reinstallation.
- Release notes are untrusted plain text (up to 4,096 characters); no remote images, executable markup or account tokens. GitHub failures or rate limits do not interfere with quota monitoring.
- The client only requests predetermined HTTPS URLs for the configured repository and expected filenames. Release asset redirects are restricted to GitHub's asset hosts. Metadata, signature, payload and request duration are bounded.
- An update is verified using Ed25519 over a domain-separated message containing its canonical version, fixed installer name and SHA-256 digest. An invalid or missing signature is never bypassed. A public key fetched alongside its own release would not be a trust anchor.
- This protects subsequent updates, assuming the user trusts the original app and the maintainer keeps the private key secure. It is not Authenticode, notarization or a guarantee that a signed update is bug-free. If the signing key is compromised, stop publishing and distribute a new trusted build through an independently verified channel.

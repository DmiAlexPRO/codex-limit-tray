# Security

Authentication is delegated to installed Codex. The app does not read token files or collect analytics. Users still need to trust Codex, their operating system and the application binary. A manually selected Codex executable runs with the current user's permissions: select only a trusted installation.

Never publish tokens, cookies, authentication files or raw server logs. Redact the username in diagnostic paths. Use GitHub private vulnerability reporting when enabled; otherwise contact a maintainer privately before posting exploit details.

The Windows in-app updater trusts only installers whose version and content match a signature verified with its embedded Ed25519 key. A hash from the release page alone is not publisher authentication. The first installation still requires trusting its source; OS code signing is not configured. Release descriptions are shown as plain text. Updates use unauthenticated public GitHub endpoints and never send Codex credentials. See [update policy](docs/UPDATES.md). Only the latest published version receives fixes; no security SLA is provided.

#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
if [ "$(uname -s)" != Darwin ]; then echo 'Run this script on macOS with Xcode Command Line Tools installed.' >&2; exit 1; fi
MACOSX_DEPLOYMENT_TARGET=13.0
export MACOSX_DEPLOYMENT_TARGET
go test ./...
go vet ./...
version=${APP_VERSION:-3.0.0-preview}
repository=${UPDATE_REPOSITORY:-DmiAlexPRO/codex-limit-tray}
key=${UPDATE_PUBLIC_KEY:-$(tr -d '\r\n' < packaging/update-public-key.txt)}
case "$version" in *[!A-Za-z0-9.-]*) echo 'Invalid version' >&2; exit 1;; esac
printf '%s\n' "$version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.-]+)?$' || { echo 'Invalid version' >&2; exit 1; }
case "$repository$key" in *[!A-Za-z0-9./_+=-]*) echo 'Invalid update metadata' >&2; exit 1;; esac
printf '%s\n' "$key" | grep -Eq '^[A-Za-z0-9+/]{43}=$' || { echo 'Invalid update public key' >&2; exit 1; }
bundle=dist/CodexLimitTray.app
rm -rf "$bundle"
mkdir -p "$bundle/Contents/MacOS"
mkdir -p "$bundle/Contents/Resources"
cp LICENSE THIRD_PARTY_NOTICES.txt "$bundle/Contents/Resources/"
CGO_ENABLED=1 go build -trimpath -ldflags="-s -w -X codexlimittray/internal/update.Version=$version -X codexlimittray/internal/update.Repository=$repository -X codexlimittray/internal/update.PublicKey=$key" -o "$bundle/Contents/MacOS/codex-limit-tray" ./cmd/codex-limit-tray
cat > "$bundle/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleExecutable</key><string>codex-limit-tray</string>
<key>CFBundleIdentifier</key><string>io.codexlimittray.preview</string>
<key>CFBundleName</key><string>Codex Limit Tray</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleShortVersionString</key><string>${version%%-*}</string>
<key>CFBundleVersion</key><string>${version%%-*}</string>
<key>LSMinimumSystemVersion</key><string>${MACOSX_DEPLOYMENT_TARGET}</string>
<key>LSUIElement</key><true/>
<key>NSHighResolutionCapable</key><true/>
</dict></plist>
PLIST
architecture=$(go env GOARCH)
case "$architecture" in amd64) architecture=x86_64 ;; arm64) ;; *) echo 'Unsupported macOS architecture' >&2; exit 1 ;; esac
archive="CodexLimitTray-macOS-${architecture}.zip"
rm -f "dist/$archive"
ditto -c -k --keepParent "$bundle" "dist/$archive"
(
    cd dist
    shasum -a 256 "$archive" > SHA256SUMS
)
echo "Built unsigned preview: $bundle ($architecture)"
echo "Portable archive: dist/$archive"

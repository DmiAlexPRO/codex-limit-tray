param(
    [switch]$Preview,
    [switch]$Installer,
    [ValidatePattern('^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$')][string]$Version='3.0.0-preview',
    [string]$Repository='DmiAlexPRO/codex-limit-tray',
    [string]$PublicKey='',
    [string]$IsccPath=''
)
$ErrorActionPreference='Stop'
if (-not $PublicKey) { $PublicKey=(Get-Content -LiteralPath "$PSScriptRoot\packaging\update-public-key.txt" -Raw).Trim() }
if ((($Version -split '-')[0].Split('.') | Where-Object { [double]$_ -gt 65535 })) { throw 'Windows file version components cannot exceed 65535.' }
if ($Repository -and $Repository -notmatch '^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$') { throw 'Invalid repository (expected owner/repo).' }
if ($PublicKey -cnotmatch '^[A-Za-z0-9+/]{43}=$' -or [Convert]::FromBase64String($PublicKey).Length -ne 32) { throw 'Invalid Ed25519 public key.' }
$root=$PSScriptRoot
Set-Location $root
$localGo=Join-Path $root '.tools\go\bin'
if (Test-Path -LiteralPath (Join-Path $localGo 'go.exe')) { $env:PATH="$localGo;$env:PATH" }
$env:GOMODCACHE=Join-Path $root '.cache\modules'
$env:GOCACHE=Join-Path $root '.cache\build'
$compiler=Join-Path $root '.tools\w64devkit\bin\gcc.exe'
if (Test-Path -LiteralPath $compiler) {
    $env:CC='gcc'
    $env:PATH="$(Split-Path $compiler);$env:PATH"
}
$env:CGO_ENABLED='1'
if ($Preview) { $env:CODEX_LIMIT_PREVIEW='1' }
New-Item -ItemType Directory -Force (Join-Path $root 'dist') | Out-Null
go test ./...
if ($LASTEXITCODE -ne 0) { throw 'Tests failed.' }
go vet ./...
if ($LASTEXITCODE -ne 0) { throw 'Vet failed.' }
$flags="-H=windowsgui -s -w -X codexlimittray/internal/update.Version=$Version -X codexlimittray/internal/update.Repository=$Repository -X codexlimittray/internal/update.PublicKey=$PublicKey"
$fileVersion=($Version -split '-')[0]
$resource=Join-Path $root 'cmd\codex-limit-tray\resources_windows_amd64.syso'
if (Test-Path -LiteralPath $resource) { throw "Resource file already present (concurrent/interrupted build): $resource" }
try {
    $rc=(Get-Content -LiteralPath 'packaging/windows.rc.in' -Raw).Replace('@VERSION@', $Version).Replace('@NUMERIC_VERSION@', ($fileVersion.Replace('.', ',')+',0'))
    Set-Content -LiteralPath dist/windows.rc -Value $rc -Encoding ascii
    windres -I packaging -i dist/windows.rc -O coff -o $resource
    if ($LASTEXITCODE -ne 0) { throw 'Failed to compile Windows icon and version resources.' }
    go build -trimpath -ldflags $flags -o dist/CodexLimitTray.exe ./cmd/codex-limit-tray
    if ($LASTEXITCODE -ne 0) { throw 'Build failed.' }
} finally {
    if (Test-Path -LiteralPath $resource) { Remove-Item -LiteralPath $resource }
}
Copy-Item LICENSE,THIRD_PARTY_NOTICES.txt -Destination dist -Force
$archive=Join-Path $root 'dist\CodexLimitTray-windows-x64.zip'
Compress-Archive -LiteralPath (Join-Path $root 'dist\CodexLimitTray.exe'),(Join-Path $root 'dist\LICENSE'),(Join-Path $root 'dist\THIRD_PARTY_NOTICES.txt') -DestinationPath $archive -Force
$artifacts=@($archive)
if ($Installer) {
    if (-not $IsccPath) {
        $IsccPath=@("$root\.tools\inno\ISCC.exe", "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe", "$env:ProgramFiles\Inno Setup 7\ISCC.exe", (Get-Command ISCC.exe -ErrorAction SilentlyContinue).Source) | Where-Object { $_ -and (Test-Path -LiteralPath $_) } | Select-Object -First 1
    }
    if (-not $IsccPath) { throw 'Inno Setup 6.3+ required: specify -IsccPath or run scripts/get-inno.ps1.' }
    & $IsccPath "/DAppSemVer=$Version" "/DAppFileVersion=$fileVersion.0" (Join-Path $root 'packaging\windows.iss')
    if ($LASTEXITCODE -ne 0) { throw 'Installer build failed.' }
    $artifacts+=Join-Path $root 'dist\CodexLimitTray-windows-x64-Setup.exe'
}
$checksums=foreach ($file in $artifacts) {
    $hash=(Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash.ToLowerInvariant()
    "$hash  $([IO.Path]::GetFileName($file))"
}
Set-Content -LiteralPath (Join-Path $root 'dist\SHA256SUMS') -Value $checksums -Encoding ascii
Write-Output $checksums

# Destructive installation/upgrade/uninstallation smoke test for a disposable GitHub VM only.
$ErrorActionPreference='Stop'
if ($env:GITHUB_ACTIONS -ne 'true' -or $env:RUNNER_ENVIRONMENT -ne 'github-hosted') { throw 'Disposable GitHub-hosted runner required.' }
# Do not let MinGW/MSYS DLLs on the build VM's PATH mask missing runtime dependencies in the installer.
$env:Path="$env:SystemRoot\System32;$env:SystemRoot;$env:SystemRoot\System32\WindowsPowerShell\v1.0"
$setup=(Resolve-Path "$PSScriptRoot\..\dist\CodexLimitTray-windows-x64-Setup.exe").Path
$folder=Join-Path $env:LOCALAPPDATA 'Programs\Codex Limit Tray'
$app=Join-Path $folder 'CodexLimitTray.exe'
$runKey='HKCU:\Software\Microsoft\Windows\CurrentVersion\Run'
if (Test-Path -LiteralPath $folder) { throw 'Expected a clean installation.' }
$configDir=Join-Path $env:APPDATA 'CodexLimitTray'
if (Test-Path -LiteralPath $configDir) { throw 'Expected clean user settings.' }
New-Item -ItemType Directory -Path $configDir | Out-Null
$config=Join-Path $configDir 'fyne-preview.json'
Set-Content -LiteralPath $config -Value '{"autoUpdate":false}' -Encoding ascii
$configHash=(Get-FileHash -LiteralPath $config).Hash
function Invoke-Installer([string]$Path, [string[]]$Options) {
    $p=Start-Process -FilePath $Path -ArgumentList $Options -WindowStyle Hidden -PassThru
    if (-not $p.WaitForExit(120000) -or $p.ExitCode -ne 0) { throw "Installer failed: $Path, exit $($p.ExitCode)" }
}
$quiet=@('/VERYSILENT','/SUPPRESSMSGBOXES','/NORESTART')
Invoke-Installer $setup $quiet
foreach ($name in @('CodexLimitTray.exe','installed.flag','unins000.exe','LICENSE','THIRD_PARTY_NOTICES.txt')) {
    if (-not (Test-Path -LiteralPath (Join-Path $folder $name))) { throw "Missing installed file: $name" }
}
# A running, short-lived predecessor exercises the update wait; /UPDATE must relaunch the app.
$old=Start-Process powershell.exe -ArgumentList '-NoProfile','-Command','Start-Sleep 2' -WindowStyle Hidden -PassThru
Invoke-Installer $setup ($quiet + '/UPDATE' + "/WAITPID=$($old.Id)")
$deadline=[DateTime]::UtcNow.AddSeconds(15)
do {
    $launched=Get-Process CodexLimitTray -ErrorAction SilentlyContinue | Where-Object Path -eq $app
    if ($launched) { break }
    Start-Sleep -Milliseconds 200
} while ([DateTime]::UtcNow -lt $deadline)
if (-not $launched) { throw 'Update failed to relaunch the installed app.' }
$launched | Stop-Process
$launched | Wait-Process -ErrorAction SilentlyContinue
if (-not (Test-Path $runKey)) { New-Item -Path $runKey | Out-Null }
New-ItemProperty -LiteralPath $runKey -Name CodexLimitTray -Value ('"' + $app + '" --minimized') -PropertyType String -Force | Out-Null
Invoke-Installer (Join-Path $folder 'unins000.exe') $quiet
$deadline=[DateTime]::UtcNow.AddSeconds(10)
while ((Test-Path -LiteralPath $app) -and [DateTime]::UtcNow -lt $deadline) { Start-Sleep -Milliseconds 200 }
if (Test-Path -LiteralPath $app) { throw 'Uninstall left the executable behind.' }
if (Get-ItemPropertyValue -LiteralPath $runKey -Name CodexLimitTray -ErrorAction SilentlyContinue) { throw 'Uninstall left a stale autostart command.' }
if ((Get-FileHash -LiteralPath $config).Hash -ne $configHash) { throw 'Installer lifecycle changed the user settings.' }

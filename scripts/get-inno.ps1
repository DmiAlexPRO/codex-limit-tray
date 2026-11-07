# Extract the official compiler in portable mode; no machine installation or associations.
$ErrorActionPreference='Stop'
$directory=Join-Path $PSScriptRoot '..\.tools'
New-Item -ItemType Directory -Force $directory | Out-Null
$directory=(Resolve-Path $directory).Path
$bootstrap=Join-Path $directory 'innosetup-7.1.1-x64.exe'
$target=Join-Path $directory 'inno'
$digest='8a5a6711c3766a6e5fb74f280ba0e3551e719ce94538d901fdbd1f8f9920b3f0'
Invoke-WebRequest 'https://github.com/jrsoftware/issrc/releases/download/is-7_1_1/innosetup-7.1.1-x64.exe' -UseBasicParsing -OutFile $bootstrap
if ((Get-FileHash $bootstrap -Algorithm SHA256).Hash -ne $digest -or (Get-AuthenticodeSignature $bootstrap).Status -ne 'Valid') {
    throw 'Inno Setup compiler bootstrap authentication failed.'
}
$compiler=Start-Process -FilePath $bootstrap -ArgumentList '/CURRENTUSER','/PORTABLE=1','/VERYSILENT','/NORESTART','/SUPPRESSMSGBOXES',("/DIR=`"$target`"") -WindowStyle Hidden -Wait -PassThru
if ($compiler.ExitCode -ne 0 -or -not (Test-Path -LiteralPath "$target\ISCC.exe")) { throw 'Could not extract portable compiler.' }

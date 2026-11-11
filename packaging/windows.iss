#ifndef AppSemVer
  #error AppSemVer is required
#endif
#ifndef AppFileVersion
  #error AppFileVersion is required
#endif
#ifndef ArtifactDir
  #define ArtifactDir "..\dist"
#endif
#define AppName "Codex Limit Tray"
[Setup]
AppId={{F4E71354-C880-45C1-8137-94D7D588A9F5}
AppName={#AppName}
AppVersion={#AppSemVer}
AppPublisher=Codex Limit Tray contributors
VersionInfoVersion={#AppFileVersion}
VersionInfoProductVersion={#AppFileVersion}
VersionInfoProductTextVersion={#AppSemVer}
DefaultDirName={localappdata}\Programs\Codex Limit Tray
DefaultGroupName={#AppName}
PrivilegesRequired=lowest
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
DisableDirPage=yes
DisableProgramGroupPage=yes
ShowLanguageDialog=no
LanguageDetectionMethod=uilanguage
UninstallDisplayIcon={app}\CodexLimitTray.exe
OutputDir={#ArtifactDir}
OutputBaseFilename=CodexLimitTray-windows-x64-Setup
SetupIconFile=app.ico
Compression=lzma2
SolidCompression=yes
RestartApplications=no
WizardStyle=modern

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "russian"; MessagesFile: "compiler:Languages\Russian.isl"
Name: "arabic"; MessagesFile: "compiler:Languages\Arabic.isl"
Name: "chinese"; MessagesFile: "compiler:Languages\ChineseSimplified.isl"
Name: "french"; MessagesFile: "compiler:Languages\French.isl"
Name: "german"; MessagesFile: "compiler:Languages\German.isl"
Name: "japanese"; MessagesFile: "compiler:Languages\Japanese.isl"
Name: "portuguese"; MessagesFile: "compiler:Languages\BrazilianPortuguese.isl"
Name: "portugal"; MessagesFile: "compiler:Languages\Portuguese.isl"
Name: "spanish"; MessagesFile: "compiler:Languages\Spanish.isl"

[Files]
Source: "{#ArtifactDir}\CodexLimitTray.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\LICENSE"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\THIRD_PARTY_NOTICES.txt"; DestDir: "{app}"; Flags: ignoreversion
Source: "installed.flag"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#AppName}"; Filename: "{app}\CodexLimitTray.exe"

[Run]
Filename: "{app}\CodexLimitTray.exe"; Parameters: "--initialize-language"; Flags: runhidden waituntilterminated; Check: not IsUpdate
Filename: "{app}\CodexLimitTray.exe"; Description: "{cm:LaunchProgram,{#AppName}}"; Flags: postinstall nowait skipifsilent; Check: not IsUpdate
Filename: "{app}\CodexLimitTray.exe"; Parameters: "--minimized"; Flags: nowait; Check: IsUpdate

[Code]
function OpenProcess(DesiredAccess: LongWord; InheritHandle: Boolean; ProcessId: LongWord): THandle;
  external 'OpenProcess@kernel32.dll stdcall';
function WaitForSingleObject(Handle: THandle; Milliseconds: LongWord): LongWord;
  external 'WaitForSingleObject@kernel32.dll stdcall';
function CloseHandle(Handle: THandle): Boolean;
  external 'CloseHandle@kernel32.dll stdcall';

function IsUpdate: Boolean;
var I: Integer;
begin
  Result := False;
  for I := 1 to ParamCount do
    if CompareText(ParamStr(I), '/UPDATE') = 0 then begin
      Result := True;
      Exit;
    end;
end;

function PrepareToInstall(var NeedsRestart: Boolean): String;
var Pid: Integer; Process: THandle;
begin
  Result := '';
  Pid := StrToIntDef(ExpandConstant('{param:WAITPID|0}'), 0);
  if Pid <= 0 then Exit;
  Process := OpenProcess($00100000, False, Pid); { SYNCHRONIZE only }
  if Process = 0 then Exit; { Already exited; Restart Manager handles any remaining file locks. }
  try
    if WaitForSingleObject(Process, 30000) <> 0 then
      Result := 'Codex Limit Tray is still running. Exit the app and retry.';
  finally
    CloseHandle(Process);
  end;
end;

procedure CurUninstallStepChanged(Step: TUninstallStep);
var Command, Exe, LegacyExe: String;
begin
  if Step <> usUninstall then Exit;
  { Only clean this user's native x64 Run entry, never one belonging to a portable copy. }
  if not RegQueryStringValue(HKCU64, 'Software\Microsoft\Windows\CurrentVersion\Run', 'CodexLimitTray', Command) then Exit;
  Exe := ExpandConstant('{app}\CodexLimitTray.exe');
  if CompareText(Copy(Exe, 1, 8), '\\?\UNC\') = 0 then
    Exe := '\\' + Copy(Exe, 9, Length(Exe))
  else if Copy(Exe, 1, 4) = '\\?\' then
    Delete(Exe, 1, 4);
  { Compare the old Go-quoted spelling separately: do not collapse UNC paths. }
  LegacyExe := Exe;
  StringChangeEx(LegacyExe, '\', '\\', True);
  if (CompareText(Command, '"' + Exe + '" --minimized') <> 0) and
     (CompareText(Command, Exe + ' --minimized') <> 0) and
     (CompareText(Command, '"' + LegacyExe + '" --minimized') <> 0) then begin
    Log('Keeping autostart: the Run command belongs to a different installation.');
    Exit;
  end;
  if RegDeleteValue(HKCU64, 'Software\Microsoft\Windows\CurrentVersion\Run', 'CodexLimitTray') then
    Log('Removed installed application autostart.')
  else
    Log('Could not remove installed application autostart.');
end;

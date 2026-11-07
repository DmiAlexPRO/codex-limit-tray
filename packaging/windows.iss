#ifndef AppSemVer
  #error AppSemVer is required
#endif
#ifndef AppFileVersion
  #error AppFileVersion is required
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
UninstallDisplayIcon={app}\CodexLimitTray.exe
OutputDir=..\dist
OutputBaseFilename=CodexLimitTray-windows-x64-Setup
SetupIconFile=app.ico
Compression=lzma2
SolidCompression=yes
RestartApplications=no
WizardStyle=modern

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "russian"; MessagesFile: "compiler:Languages\Russian.isl"

[Files]
Source: "..\dist\CodexLimitTray.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\LICENSE"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\THIRD_PARTY_NOTICES.txt"; DestDir: "{app}"; Flags: ignoreversion
Source: "installed.flag"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#AppName}"; Filename: "{app}\CodexLimitTray.exe"

[Run]
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
var Command, Exe: String;
begin
  if Step <> usPostUninstall then Exit;
  if not RegQueryStringValue(HKEY_CURRENT_USER, 'Software\Microsoft\Windows\CurrentVersion\Run', 'CodexLimitTray', Command) then Exit;
  Exe := ExpandConstant('{app}\CodexLimitTray.exe');
  { Older builds used Go string quoting and doubled the backslashes. }
  StringChangeEx(Command, '\\', '\', True);
  if (CompareText(Command, '"' + Exe + '" --minimized') = 0) or
     (CompareText(Command, Exe + ' --minimized') = 0) then
    RegDeleteValue(HKEY_CURRENT_USER, 'Software\Microsoft\Windows\CurrentVersion\Run', 'CodexLimitTray');
end;

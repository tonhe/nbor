; nbor Windows Installer
; Built with Inno Setup 6.x
; https://jrsoftware.org/isinfo.php

#define MyAppName "nbor"
#define MyAppPublisher "nbor"
#define MyAppURL "https://github.com/tmattke/nbor"
#define MyAppExeName "nbor.exe"

; Version is passed from CI: iscc /DMyAppVersion=1.0.0 nbor.iss
#ifndef MyAppVersion
  #define MyAppVersion "0.0.0-dev"
#endif

; Npcap download URL - update version as needed
#define NpcapVersion "1.80"
#define NpcapURL "https://npcap.com/dist/npcap-" + NpcapVersion + ".exe"

[Setup]
; NOTE: Generate a new GUID for AppId if you fork this project
AppId={{B8F5E3A1-7D2C-4E9F-A1B3-8C6D5E4F7A2B}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}/issues
AppUpdatesURL={#MyAppURL}/releases
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
LicenseFile=LICENSE.txt
OutputDir=..\..

OutputBaseFilename=nbor-setup-windows-amd64
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=admin
PrivilegesRequiredOverridesAllowed=dialog
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
SetupIconFile=nbor.ico
UninstallDisplayIcon={app}\nbor.ico

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "addtopath"; Description: "Add nbor to system PATH (recommended for command-line use)"; GroupDescription: "Additional options:"; Flags: unchecked

[Files]
Source: "..\..\nbor-windows-amd64.exe"; DestDir: "{app}"; DestName: "{#MyAppExeName}"; Flags: ignoreversion
Source: "nbor.ico"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\nbor.ico"; Comment: "Network neighbor discovery tool"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"

[Code]
var
  NpcapMissing: Boolean;
  NpcapPage: TWizardPage;
  NpcapCheckbox: TNewCheckBox;
  NpcapLabel: TNewStaticText;
  DownloadPage: TDownloadWizardPage;

function IsNpcapInstalled: Boolean;
begin
  Result := RegKeyExists(HKEY_LOCAL_MACHINE, 'SOFTWARE\Npcap') or
            RegKeyExists(HKEY_LOCAL_MACHINE, 'SOFTWARE\WOW6432Node\Npcap');
end;

procedure InitializeWizard;
begin
  NpcapMissing := not IsNpcapInstalled;

  if NpcapMissing then
  begin
    // Create custom page for Npcap installation
    NpcapPage := CreateCustomPage(wpSelectTasks,
      'Npcap Required',
      'nbor requires Npcap for packet capture functionality.');

    NpcapLabel := TNewStaticText.Create(NpcapPage);
    NpcapLabel.Parent := NpcapPage.Surface;
    NpcapLabel.Caption :=
      'Npcap was not detected on your system.' + #13#10 + #13#10 +
      'Npcap is required for nbor to capture network packets. Without it, ' +
      'nbor will not be able to discover network neighbors.' + #13#10 + #13#10 +
      'If you choose to install Npcap, it will be downloaded from npcap.com ' +
      'and you will need to accept the Npcap license agreement.' + #13#10 + #13#10 +
      'The installer will enable "WinPcap API-compatible Mode" automatically.';
    NpcapLabel.AutoSize := True;
    NpcapLabel.WordWrap := True;
    NpcapLabel.Width := NpcapPage.SurfaceWidth;
    NpcapLabel.Top := 0;

    NpcapCheckbox := TNewCheckBox.Create(NpcapPage);
    NpcapCheckbox.Parent := NpcapPage.Surface;
    NpcapCheckbox.Caption := 'Download and install Npcap {#NpcapVersion} (recommended)';
    NpcapCheckbox.Checked := True;
    NpcapCheckbox.Top := NpcapLabel.Top + NpcapLabel.Height + 20;
    NpcapCheckbox.Width := NpcapPage.SurfaceWidth;

    // Create download page
    DownloadPage := CreateDownloadPage(SetupMessage(msgWizardPreparing), SetupMessage(msgPreparingDesc), nil);
  end;
end;

function NextButtonClick(CurPageID: Integer): Boolean;
var
  NpcapInstaller: String;
  ResultCode: Integer;
begin
  Result := True;

  // Handle Npcap download and installation
  if NpcapMissing and (CurPageID = NpcapPage.ID) and NpcapCheckbox.Checked then
  begin
    NpcapInstaller := ExpandConstant('{tmp}\npcap-{#NpcapVersion}.exe');

    // Download Npcap
    DownloadPage.Clear;
    DownloadPage.Add('{#NpcapURL}', 'npcap-{#NpcapVersion}.exe', '');
    DownloadPage.Show;
    try
      try
        DownloadPage.Download;

        // Run Npcap installer with WinPcap compatibility mode
        // /S = silent, /winpcap_mode=yes = enable WinPcap API compatibility
        DownloadPage.SetText('Installing Npcap...', 'Please complete the Npcap installation wizard.');
        if not Exec(NpcapInstaller, '/winpcap_mode=yes', '', SW_SHOW, ewWaitUntilTerminated, ResultCode) then
        begin
          MsgBox('Failed to run Npcap installer. You can install it manually from npcap.com', mbError, MB_OK);
        end
        else if ResultCode <> 0 then
        begin
          MsgBox('Npcap installation may not have completed successfully. ' +
                 'If nbor does not work, please install Npcap manually from npcap.com', mbInformation, MB_OK);
        end;
      except
        MsgBox('Failed to download Npcap. You can install it manually from npcap.com', mbError, MB_OK);
      end;
    finally
      DownloadPage.Hide;
    end;
  end;
end;

function ShouldSkipPage(PageID: Integer): Boolean;
begin
  Result := False;
  // Skip Npcap page if already installed
  if NpcapMissing and (NpcapPage <> nil) and (PageID = NpcapPage.ID) then
    Result := IsNpcapInstalled;
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  Path: string;
  NewPath: string;
begin
  if CurStep = ssPostInstall then
  begin
    // Add to PATH if selected
    if WizardIsTaskSelected('addtopath') then
    begin
      RegQueryStringValue(HKEY_LOCAL_MACHINE,
        'SYSTEM\CurrentControlSet\Control\Session Manager\Environment',
        'Path', Path);

      if Pos(LowerCase(ExpandConstant('{app}')), LowerCase(Path)) = 0 then
      begin
        NewPath := Path + ';' + ExpandConstant('{app}');
        RegWriteStringValue(HKEY_LOCAL_MACHINE,
          'SYSTEM\CurrentControlSet\Control\Session Manager\Environment',
          'Path', NewPath);
        // Notify the system of environment change
        // Note: New cmd windows will pick up the change; existing ones need to be restarted
      end;
    end;
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  Path: string;
  AppPath: string;
  P: Integer;
begin
  if CurUninstallStep = usPostUninstall then
  begin
    // Remove from PATH if present
    RegQueryStringValue(HKEY_LOCAL_MACHINE,
      'SYSTEM\CurrentControlSet\Control\Session Manager\Environment',
      'Path', Path);

    AppPath := ExpandConstant('{app}');

    // Try to remove ';{app}' first
    P := Pos(';' + LowerCase(AppPath), LowerCase(Path));
    if P > 0 then
    begin
      Delete(Path, P, Length(';' + AppPath));
      RegWriteStringValue(HKEY_LOCAL_MACHINE,
        'SYSTEM\CurrentControlSet\Control\Session Manager\Environment',
        'Path', Path);
    end
    else
    begin
      // Try to remove '{app};'
      P := Pos(LowerCase(AppPath) + ';', LowerCase(Path));
      if P > 0 then
      begin
        Delete(Path, P, Length(AppPath + ';'));
        RegWriteStringValue(HKEY_LOCAL_MACHINE,
          'SYSTEM\CurrentControlSet\Control\Session Manager\Environment',
          'Path', Path);
      end
      else
      begin
        // Try to remove '{app}' alone (if it's the only entry)
        P := Pos(LowerCase(AppPath), LowerCase(Path));
        if P > 0 then
        begin
          Delete(Path, P, Length(AppPath));
          RegWriteStringValue(HKEY_LOCAL_MACHINE,
            'SYSTEM\CurrentControlSet\Control\Session Manager\Environment',
            'Path', Path);
        end;
      end;
    end;
  end;
end;

function UpdateReadyMemo(Space, NewLine, MemoUserInfoInfo, MemoDirInfo, MemoTypeInfo,
  MemoComponentsInfo, MemoGroupInfo, MemoTasksInfo: String): String;
begin
  Result := MemoDirInfo + NewLine + NewLine;

  if MemoTasksInfo <> '' then
    Result := Result + MemoTasksInfo + NewLine + NewLine;

  if NpcapMissing and not IsNpcapInstalled then
  begin
    if NpcapCheckbox.Checked then
      Result := Result + 'Npcap: Will be downloaded and installed' + NewLine
    else
      Result := Result + 'WARNING: Npcap is not installed. nbor requires Npcap to function.' + NewLine +
                         'Please install Npcap from npcap.com after setup completes.' + NewLine;
  end
  else
    Result := Result + 'Npcap: Already installed' + NewLine;
end;

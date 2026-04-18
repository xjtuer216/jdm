; ============================================================================
; JDM (JDK Version Manager) - Windows Installer
; ============================================================================
; Build: iscc.exe jdm.iss
; Requires: Inno Setup 6.x
; ============================================================================

#define MyAppName "JDM"
#ifndef MyAppVersion
  #define MyAppVersion "1.0.0"
#endif
#define MyAppPublisher "xjtuer216"
#define MyAppURL "https://github.com/xjtuer216/jdm"
#define MyAppExeName "jdm.exe"
#define MyAppId "B4E84F72-3C5A-4D1B-9F2E-8A7C6D5E4F3A"
#define ProjectRoot "."

[Setup]
; Application identity
AppId={{{#MyAppId}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}

; Install location
DefaultDirName={localappdata}\{#MyAppName}
DisableDirPage=no
DirExistsWarning=no

; Start Menu
DefaultGroupName={#MyAppName}
AllowNoIcons=yes

; Language detection - auto-detect system language
LanguageDetectionMethod=uilanguage
ShowLanguageDialog=yes

; Environment variables - Inno Setup auto-notifies Explorer
ChangesEnvironment=yes
ChangesAssociations=no

; Privileges - lowest so no UAC prompt needed
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog

; Architecture - 64-bit only
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64

; Output
OutputDir={#ProjectRoot}\dist
OutputBaseFilename=jdm-setup-{#MyAppVersion}
UninstallDisplayIcon={app}\{#MyAppExeName}
UninstallDisplayName={#MyAppName} (JDK Version Manager)

; Compression
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
WizardSizePercent=100,100

; Misc
DisableWelcomePage=no
DisableProgramGroupPage=yes
UsePreviousAppDir=yes
UsePreviousGroup=yes
UsePreviousSetupType=yes
UsePreviousTasks=yes
UsePreviousLanguage=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "chinesesimplified"; MessagesFile: ".\Languages\ChineseSimplified.isl"

[CustomMessages]
; English custom messages
english.JDKStorageTitle=JDK Storage Location
english.JDKStorageDesc=All downloaded JDK versions will be stored in this directory.
english.JDKStorageSubCaption=Select the folder where JDK versions should be stored.
english.AddToPath=Add JDM to system PATH environment variable
english.CreateDesktopIcon=Create a &desktop shortcut
english.ConfigGenerated=Configuration file generated based on your selections.
english.EnvVarSet=JDM_HOME environment variable has been set.
english.PathUpdated=System PATH has been updated to include JDM.
english.InstallComplete=JDM has been successfully installed on your computer.
english.OpenNewTerminal=Please open a new Command Prompt or PowerShell window to use JDM.
english.UninstallWarning=You are about to uninstall JDM (JDK Version Manager).%n%nThis will:%n  - Remove the jdm command%n  - Delete all installed JDK versions%n  - Remove all JDM configuration files%n  - Clean up environment variables%n%nAre you sure you want to continue?
english.UninstallComplete=JDM has been completely removed from your computer.
english.PathExists=The path "%1" already exists in PATH.
english.PathAdded=Added to PATH: %1
english.PathRemoved=Removed from PATH: %1
english.ViewReadme=View README file
english.AdditionalIcons=Additional icons:

; Chinese custom messages
chinesesimplified.JDKStorageTitle=JDK 存储位置
chinesesimplified.JDKStorageDesc=所有下载的 JDK 版本将存储在此目录中。
chinesesimplified.JDKStorageSubCaption=选择 JDK 版本的存储文件夹。
chinesesimplified.AddToPath=将 JDM 添加到系统 PATH 环境变量
chinesesimplified.CreateDesktopIcon=创建桌面快捷方式(&D)
chinesesimplified.ConfigGenerated=已根据您的选择自动生成配置文件。
chinesesimplified.EnvVarSet=已设置 JDM_HOME 环境变量。
chinesesimplified.PathUpdated=已更新系统 PATH 环境变量。
chinesesimplified.InstallComplete=JDM 已成功安装到您的电脑。
chinesesimplified.OpenNewTerminal=请打开新的命令提示符或 PowerShell 窗口以使用 JDM。
chinesesimplified.UninstallWarning=您即将卸载 JDM（JDK 版本管理器）。%n%n此操作将：%n  - 删除 jdm 命令%n  - 删除所有已安装的 JDK 版本%n  - 删除所有 JDM 配置文件%n  - 清理环境变量%n%n确定要继续吗？
chinesesimplified.UninstallComplete=JDM 已从您的电脑上完全删除。
chinesesimplified.PathExists=路径 "%1" 已存在于 PATH 中。
chinesesimplified.PathAdded=已添加到 PATH: %1
chinesesimplified.PathRemoved=已从 PATH 删除: %1
chinesesimplified.ViewReadme=查看自述文件
chinesesimplified.AdditionalIcons=其他图标：

[Tasks]
Name: "addtopath"; Description: "{cm:AddToPath}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
; Main executable
Source: "{#ProjectRoot}\assets\{#MyAppExeName}"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#ProjectRoot}\assets\license.txt"; DestDir: "{app}"; Flags: ignoreversion isreadme
Source: "{#ProjectRoot}\assets\README.txt"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
; Start Menu
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\Uninstall {#MyAppName}"; Filename: "{uninstallexe}"

; Desktop
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
; Open README after installation
Filename: "{app}\README.txt"; Description: "{cm:ViewReadme}"; Flags: postinstall shellexec skipifsilent

[UninstallDelete]
; Delete JDM installation directory
Type: filesandordirs; Name: "{app}"

; Delete JDK versions directory (from config)
Type: filesandordirs; Name: "{userappdata}\{#MyAppName}"

[Code]
{ ============================================================================ }
{ JDM Installer - Pascal Script                                                }
{ ============================================================================ }

const
  EnvironmentKey = 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment';
  JDMHomeEnvName = 'JDM_HOME';

var
  JDKStoragePage: TInputDirWizardPage;

{ ============================================================================ }
{ PATH Management Functions                                                    }
{ ============================================================================ }

function IsPathInPath(RootKey: Integer; Path: string): Boolean;
var
  Paths: string;
begin
  Result := False;
  if not RegQueryStringValue(RootKey, EnvironmentKey, 'Path', Paths) then
    Exit;
  if Pos(';' + Lowercase(Path) + ';', ';' + Lowercase(Paths) + ';') > 0 then
    Result := True;
end;

function IsAdminInstall: Boolean;
begin
  Result := IsAdminInstallMode;
end;

procedure AddPathToRegistry(RootKey: Integer; RegKey: string; Path: string);
var
  Paths: string;
begin
  if not RegQueryStringValue(RootKey, RegKey, 'Path', Paths) then
    Paths := '';
  if Pos(';' + Lowercase(Path) + ';', ';' + Lowercase(Paths) + ';') > 0 then
  begin
    Log('[JDM] Path already in PATH: ' + Path);
    Exit;
  end;
  if Paths <> '' then
    Paths := Paths + ';' + Path
  else
    Paths := Path;
  if RegWriteExpandStringValue(RootKey, RegKey, 'Path', Paths) then
    Log('[JDM] Added to PATH: ' + Path)
  else
    Log('[JDM] Failed to add to PATH: ' + Path);
end;

procedure AddToSystemPath(Path: string);
begin
  if IsAdminInstall then
    AddPathToRegistry(HKEY_LOCAL_MACHINE, EnvironmentKey, Path)
  else
    AddPathToRegistry(HKEY_CURRENT_USER, 'Environment', Path);
end;

procedure RemovePathFromRegistry(RootKey: Integer; RegKey: string; Path: string);
var
  Paths: string;
  P: Integer;
  SearchPath: string;
begin
  if not RegQueryStringValue(RootKey, RegKey, 'Path', Paths) then
  begin
    Log('[JDM] PATH not found in registry');
    Exit;
  end;
  SearchPath := ';' + Lowercase(Path) + ';';
  P := Pos(SearchPath, ';' + Lowercase(Paths) + ';');
  if P = 0 then
  begin
    Log('[JDM] Path not found in PATH: ' + Path);
    Exit;
  end;
  if P > 1 then
    P := P - 1;
  Delete(Paths, P, Length(Path) + 1);
  while Pos(';;', Paths) > 0 do
    StringChangeEx(Paths, ';;', ';', True);
  if Length(Paths) > 0 then
  begin
    if Paths[1] = ';' then
      Delete(Paths, 1, 1);
    if Paths[Length(Paths)] = ';' then
      Delete(Paths, Length(Paths), 1);
  end;
  if RegWriteExpandStringValue(RootKey, RegKey, 'Path', Paths) then
    Log('[JDM] Removed from PATH: ' + Path)
  else
    Log('[JDM] Failed to remove from PATH: ' + Path);
end;

procedure RemoveFromSystemPath(Path: string);
begin
  if IsAdminInstall then
  begin
    RemovePathFromRegistry(HKEY_LOCAL_MACHINE, EnvironmentKey, Path);
  end;
  RemovePathFromRegistry(HKEY_CURRENT_USER, 'Environment', Path);
end;

procedure AddToUserPath(Path: string);
var
  Paths: string;
begin
  if not RegQueryStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', Paths) then
    Paths := '';
  if Pos(';' + Lowercase(Path) + ';', ';' + Lowercase(Paths) + ';') > 0 then
    Exit;
  if Paths <> '' then
    Paths := Paths + ';' + Path
  else
    Paths := Path;
  RegWriteExpandStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', Paths);
  Log('[JDM] Added to user PATH: ' + Path);
end;

procedure RemoveFromUserPath(Path: string);
var
  Paths: string;
  P: Integer;
  SearchPath: string;
begin
  if not RegQueryStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', Paths) then
    Exit;
  SearchPath := ';' + Lowercase(Path) + ';';
  P := Pos(SearchPath, ';' + Lowercase(Paths) + ';');
  if P = 0 then Exit;
  if P > 1 then P := P - 1;
  Delete(Paths, P, Length(Path) + 1);
  while Pos(';;', Paths) > 0 do
    StringChangeEx(Paths, ';;', ';', True);
  if Length(Paths) > 0 then
  begin
    if Paths[1] = ';' then Delete(Paths, 1, 1);
    if Paths[Length(Paths)] = ';' then Delete(Paths, Length(Paths), 1);
  end;
  RegWriteExpandStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', Paths);
  Log('[JDM] Removed from user PATH: ' + Path);
end;

{ ============================================================================ }
{ Environment Variable Management                                              }
{ ============================================================================ }

procedure SetJDMHome(Path: string);
begin
  if IsAdminInstall then
    RegWriteExpandStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, JDMHomeEnvName, Path);
  RegWriteExpandStringValue(HKEY_CURRENT_USER, 'Environment', JDMHomeEnvName, Path);
  Log('[JDM] Set JDM_HOME = ' + Path);
end;

procedure RemoveJDMHome;
begin
  RegDeleteValue(HKEY_LOCAL_MACHINE, EnvironmentKey, JDMHomeEnvName);
  RegDeleteValue(HKEY_CURRENT_USER, 'Environment', JDMHomeEnvName);
  Log('[JDM] Removed JDM_HOME environment variable');
end;

{ ============================================================================ }
{ Wizard Page Initialization                                                   }
{ ============================================================================ }

procedure InitializeWizard;
var
  DefaultJDKPath: string;
begin
  DefaultJDKPath := ExpandConstant('{userappdata}\{#MyAppName}\versions');
  JDKStoragePage := CreateInputDirPage(wpSelectDir,
    ExpandConstant('{cm:JDKStorageTitle}'),
    ExpandConstant('{cm:JDKStorageDesc}'),
    ExpandConstant('{cm:JDKStorageSubCaption}'),
    False, '');
  JDKStoragePage.Add('');
  JDKStoragePage.Values[0] := DefaultJDKPath;
end;

{ ============================================================================ }
{ Configuration File Generation                                                }
{ ============================================================================ }

procedure GenerateConfig;
var
  ConfigPath: string;
  ConfigContent: string;
  JDMHomePath: string;
  JDKHomePath: string;
begin
  JDMHomePath := WizardForm.DirEdit.Text;
  JDKHomePath := JDKStoragePage.Values[0];
  if Length(JDKHomePath) > 0 then
  begin
    if JDKHomePath[Length(JDKHomePath)] = '\' then
      JDKHomePath := Copy(JDKHomePath, 1, Length(JDKHomePath) - 1);
  end;
  ConfigContent := '{' + #13#10;
  ConfigContent := ConfigContent + '  "jdm_home": "' + JDMHomePath + '",' + #13#10;
  ConfigContent := ConfigContent + '  "jdk_home": "' + JDKHomePath + '",' + #13#10;
  ConfigContent := ConfigContent + '  "mirror": "https://api.adoptium.net",' + #13#10;
  ConfigContent := ConfigContent + '  "download_mirror": "https://ghproxy.net",' + #13#10;
  ConfigContent := ConfigContent + '  "default": "",' + #13#10;
  ConfigContent := ConfigContent + '  "aliases": {}' + #13#10;
  ConfigContent := ConfigContent + '}' + #13#10;
  ConfigPath := JDMHomePath + '\config.json';
  ForceDirectories(JDMHomePath);
  SaveStringToFile(ConfigPath, ConfigContent, False);
  Log('[JDM] Generated config at: ' + ConfigPath);
end;

{ ============================================================================ }
{ Installation Events                                                          }
{ ============================================================================ }

procedure CurStepChanged(CurStep: TSetupStep);
var
  AppPath: string;
  JDKHomePath: string;
  CurrentBinPath: string;
begin
  if CurStep = ssPostInstall then
  begin
    AppPath := ExpandConstant('{app}');
    JDKHomePath := JDKStoragePage.Values[0];
    CurrentBinPath := JDKHomePath + '\current\bin';
    Log('[JDM] Post-install: AppPath=' + AppPath + ', JDKHome=' + JDKHomePath);
    GenerateConfig;
    SetJDMHome(AppPath);
    AddToSystemPath(AppPath);
    AddToSystemPath(CurrentBinPath);
    Log('[JDM] Installation completed successfully');
  end;
end;

{ ============================================================================ }
{ Uninstall Events                                                             }
{ ============================================================================ }

function InitializeUninstall(): Boolean;
begin
  Result := SuppressibleMsgBox(
    ExpandConstant('{cm:UninstallWarning}'),
    mbConfirmation,
    MB_YESNO,
    IDYES
  ) = IDYES;
  if not Result then
    Log('[JDM] Uninstall cancelled by user');
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  AppPath: string;
  JDKHomePath: string;
  CurrentBinPath: string;
begin
  case CurUninstallStep of
    usUninstall:
      begin
        Log('[JDM] Starting uninstall cleanup...');
        AppPath := ExpandConstant('{app}');
        JDKHomePath := ExpandConstant('{userappdata}\{#MyAppName}\versions');
        CurrentBinPath := JDKHomePath + '\current\bin';
        RemoveJDMHome;
        RemoveFromSystemPath(AppPath);
        RemoveFromSystemPath(CurrentBinPath);
        RemoveFromUserPath(AppPath);
        RemoveFromUserPath(CurrentBinPath);
        Log('[JDM] Environment cleanup completed');
      end;
    usPostUninstall:
      begin
        Log('[JDM] Uninstall finished');
      end;
  end;
end;

; ============================================================================
; JDM (JDK Version Manager) - Windows Installer
; ============================================================================
; Build: iscc.exe jdm.iss
; Requires: Inno Setup 6.x
; ============================================================================

#define MyAppName "JDM"
#ifndef MyAppVersion
  #define MyAppVersion "0.1.1"
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
{ Path Helpers                                                                 }
{ ============================================================================ }

{ 追加子目录：基础目录已以反斜杠结尾时不再重复添加分隔符 }
function AppendSubDir(const BaseDir, SubDir: string): string;
begin
  Result := BaseDir;
  if (Length(Result) > 0) and (Result[Length(Result)] = '\') then
    Result := Result + SubDir
  else
    Result := Result + '\' + SubDir;
end;

{ JDK 版本目录固定为 JDM 主目录下的 versions 子目录，
  与 Go 端 config.NewConfig 的默认值保持一致 }
function GetJDKHomePath(const JDMHomePath: string): string;
begin
  Result := AppendSubDir(JDMHomePath, 'versions');
end;

{ ============================================================================ }
{ Configuration File Initialization                                           }
{ ============================================================================ }

{ 配置默认值只由 JDM 自身维护，安装器不再重复生成 JSON：
  config init --install 会在配置不存在时写入完整模板，
  已存在时只补充缺失的键，保留用户自定义值 }
procedure InitializeInstallConfig;
var
  AppPath: string;
  ConfigPath: string;
  ResultCode: Integer;
begin
  AppPath := ExpandConstant('{app}');
  ConfigPath := AppPath + '\config.json';
  if not Exec(AppPath + '\{#MyAppExeName}', 'config init --install', AppPath,
      SW_HIDE, ewWaitUntilTerminated, ResultCode) then
  begin
    Log('[JDM] Failed to launch config init: ' + AppPath);
    Exit;
  end;
  if ResultCode <> 0 then
  begin
    Log('[JDM] config init failed with exit code ' + IntToStr(ResultCode));
    Exit;
  end;
  Log('[JDM] Initialized config at: ' + ConfigPath);
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
    JDKHomePath := GetJDKHomePath(AppPath);
    { Go 端把 current 链接创建在 JDM 主目录下，而不是 jdk_home 下 }
    CurrentBinPath := AppendSubDir(AppPath, 'current') + '\bin';
    Log('[JDM] Post-install: AppPath=' + AppPath + ', JDKHome=' + JDKHomePath);
    InitializeInstallConfig;
    SetJDMHome(AppPath);
    AddToSystemPath(AppPath);
    AddToSystemPath(CurrentBinPath);
    Log('[JDM] Installation completed successfully');
  end;
end;

{ ============================================================================ }
{ Uninstall Events                                                             }
{ ============================================================================ }

{ 安装时写入的 JDM_HOME 才是实际主目录；旧版本缺失时回退到安装目录 }
function GetInstalledJDMHome: string;
var
  Value: string;
begin
  Result := '';
  if RegQueryStringValue(HKEY_CURRENT_USER, 'Environment', JDMHomeEnvName, Value) then
    Result := Value
  else if RegQueryStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, JDMHomeEnvName, Value) then
    Result := Value;
  if Result = '' then
    Result := ExpandConstant('{app}');
end;

{ 删除 current 链接本身。RemoveDir 直接调用 RemoveDirectory，
  对目录符号链接/联接点只移除链接，不会删除链接指向的 JDK 目录 }
procedure RemoveCurrentLink(const JDMHomePath: string);
var
  CurrentPath: string;
begin
  CurrentPath := AppendSubDir(JDMHomePath, 'current');
  if DirExists(CurrentPath) then
  begin
    if RemoveDir(CurrentPath) then
      Log('[JDM] Removed current link: ' + CurrentPath)
    else
      Log('[JDM] Failed to remove current link: ' + CurrentPath);
  end
  else if FileExists(CurrentPath) then
  begin
    if DeleteFile(CurrentPath) then
      Log('[JDM] Removed current file link: ' + CurrentPath)
    else
      Log('[JDM] Failed to remove current file link: ' + CurrentPath);
  end;
end;

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
  JDMHomePath: string;
  JDKHomePath: string;
  CurrentBinPath: string;
  LegacyCurrentBinPath: string;
  LegacyDefaultBinPath: string;
begin
  case CurUninstallStep of
    usUninstall:
      begin
        Log('[JDM] Starting uninstall cleanup...');
        AppPath := ExpandConstant('{app}');
        JDMHomePath := GetInstalledJDMHome;
        JDKHomePath := GetJDKHomePath(JDMHomePath);
        CurrentBinPath := AppendSubDir(JDMHomePath, 'current') + '\bin';
        { 旧版安装器把 PATH 指向所选 jdk_home\current\bin，默认 jdk_home 在 userappdata 下 }
        LegacyCurrentBinPath := AppendSubDir(JDKHomePath, 'current') + '\bin';
        LegacyDefaultBinPath :=
            ExpandConstant('{userappdata}\{#MyAppName}\versions\current\bin');
        Log('[JDM] Uninstall: AppPath=' + AppPath + ', JDMHome=' + JDMHomePath +
            ', JDKHome=' + JDKHomePath);
        RemoveCurrentLink(JDMHomePath);
        RemoveCurrentLink(JDKHomePath);
        RemoveJDMHome;
        RemoveFromSystemPath(AppPath);
        RemoveFromSystemPath(CurrentBinPath);
        RemoveFromSystemPath(LegacyCurrentBinPath);
        RemoveFromSystemPath(LegacyDefaultBinPath);
        RemoveFromUserPath(AppPath);
        RemoveFromUserPath(CurrentBinPath);
        RemoveFromUserPath(LegacyCurrentBinPath);
        RemoveFromUserPath(LegacyDefaultBinPath);
        Log('[JDM] Environment cleanup completed');
      end;
    usPostUninstall:
      begin
        Log('[JDM] Uninstall finished');
      end;
  end;
end;

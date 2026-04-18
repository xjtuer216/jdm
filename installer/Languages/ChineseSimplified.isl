; *** Inno Setup version 6.2.2+ Chinese Simplified messages ***
; Maintainer: JDM Project
; Encoding: UTF-8

[LangOptions]
LanguageName=简体中文
LanguageID=$0804
LanguageCodePage=936

[Messages]
; *** Application titles
SetupAppTitle=安装
SetupWindowTitle=安装 - %1
UninstallAppTitle=卸载
UninstallAppFullTitle=%1 卸载

; *** Misc. common
ButtonBack=< 上一步(&B)
ButtonNext=下一步(&N) >
ButtonInstall=安装(&I)
ButtonOK=确定
ButtonCancel=取消
ButtonYes=是(&Y)
ButtonYesToAll=全是(&A)
ButtonNo=否(&N)
ButtonNoToAll=全否(&L)
ButtonFinish=完成(&F)
ButtonBrowse=浏览(&B)
ButtonWizardBrowse=浏览(&R)
ButtonNewFolder=新建文件夹(&M)

; *** "Select Language" dialog messages
SelectLanguageTitle=选择安装语言
SelectLanguageLabel=选择安装时使用的语言。

; *** "Welcome" wizard page
WelcomeLabel1=欢迎安装 [name]
WelcomeLabel2=现在将在您的电脑上安装 [name/ver]。%n%n建议在继续之前关闭所有其他应用程序。%n%n单击"下一步"继续，或单击"取消"退出安装。

; *** "License Agreement" wizard page
WizardLicense=许可协议
LicenseLabel=安装前请阅读以下许可协议。
LicenseLabel3=请仔细阅读以下许可协议。您必须接受这些条款才能继续安装。
LicenseAccepted=我接受协议(&A)
LicenseNotAccepted=我不接受协议(&D)

; *** "Select Destination Directory" wizard page
WizardSelectDir=选择目标位置
SelectDirDesc=您想将 [name] 安装在哪里？
SelectDirLabel3=安装程序将安装 [name] 到以下文件夹。%n%n要安装到不同的文件夹，请单击"浏览"并选择其他文件夹。%n%n单击"下一步"继续。
SelectDirBrowseLabel=要安装到不同的文件夹，请单击"浏览"。
DiskSpaceGBLabel=至少需要 [gb] GB 的可用磁盘空间。
DiskSpaceMBLabel=至少需要 [mb] MB 的可用磁盘空间。
CannotInstallToUncPath=安装程序无法安装到 UNC 路径。请从本地驱动器或映射的网络驱动器安装。

; *** "Select Start Menu Folder" wizard page
WizardSelectProgramGroup=选择开始菜单文件夹
SelectStartMenuLabel3=安装程序应在哪个开始菜单文件夹中创建程序的快捷方式？
SelectStartMenuBrowseLabel=要安装到不同的文件夹，请单击"浏览"。

; *** "Preparing to Install" wizard page
WizardPreparing=准备安装
PreparingDesc=安装程序现在准备在您的电脑上安装 [name]。
ClickNext=单击"下一步"继续，或单击"上一步"查看之前的设置。
PreparingError=安装程序无法继续。%n%n%1
PreparingYes=是，我想退出安装程序
PreparingNo=否，我想继续安装

; *** "Installing" wizard page
WizardInstalling=正在安装
InstallingLabel=正在安装。请等待。

; *** "Setup Completed" wizard page
WizardFinished=安装完成
FinishedHeadingLabel=[name] 安装完成
FinishedLabelNoIcons=[name] 已成功安装到您的电脑上。
FinishedLabel=单击"完成"退出安装程序。
FinishedRestartLabel=要完成 [name] 的安装，安装程序必须重新启动您的电脑。是否立即重新启动？
FinishedRestartMessage=要完成 [name] 的安装，安装程序必须重新启动您的电脑。%n%n是否立即重新启动？
FinishedLabelNoIcons2=单击"完成"退出安装程序。

; *** "Uninstall" messages
UninstallNotFound=无法卸载，因为卸载程序不存在。
UninstallOpenError=无法卸载，因为无法打开卸载程序。
UninstallRunError=卸载程序无法运行。
ConfirmUninstall=您确定要完全删除 %1 及其所有组件吗？
OnlyAdminCanUninstall=只有管理员才能卸载此应用程序。
UninstalledAll=%1 已从您的电脑上删除。
UninstalledMost=%1 卸载完成。%n%n某些元素无法删除。可以手动删除它们。
UninstalledAndNeedsRestart=要完成 %1 的卸载，电脑必须重新启动。%n%n是否立即重新启动？
UninstallDataCorrupted=卸载数据已损坏，卸载无法继续。
UninstallUnsupported=无法卸载 %1。此程序与当前系统不兼容。
UninstallStatusLabel=正在从您的电脑上删除 %1。请等待。

; *** Exit dialog messages
ExitDialogTitle=退出？
ExitDialogMessage=安装未完成。%n%n如果现在退出，程序将无法安装。%n%n您可以稍后再次运行安装程序来完成安装。%n%n现在退出安装程序？

; *** Error messages
ErrorCreatingDir=安装程序无法创建文件夹"%1"。
ErrorTooManyFilesInDir=无法在文件夹"%1"中创建文件，因为其中包含太多文件。

; *** Misc. common
ErrorTitle=安装程序错误
ErrorAborting=安装程序中止。
ErrorAbortingNoCancel=安装程序中止。
ErrorRestartingComputer=安装程序无法重新启动电脑。请手动重新启动。
ErrorOpeningReadme=打开自述文件时出错。
ErrorReadingExistingDir=安装程序无法读取现有目录的信息。
ExistingFileReadOnly=现有文件标记为只读。%n%n%1%n%n请删除只读属性后单击"重试"，或单击"取消"中止安装。
FileNotRegisted=文件未注册。
FileExists=文件已存在。%n%n是否覆盖？
FileExists2=文件已存在。%n%n安装程序无法写入文件：%n%n%1%n%n请确保文件未被使用，然后单击"重试"，或单击"取消"中止安装。
FileAbortRetry=安装程序无法写入文件：%n%n%1%n%n请确保文件未被使用，然后单击"重试"，或单击"取消"中止安装。
FileAbortRetrySkip=安装程序无法写入文件：%n%n%1%n%n请确保文件未被使用，然后单击"重试"，或单击"跳过"跳过此文件（不推荐），或单击"取消"中止安装。

; *** Setup startup errors
SetupAlreadyRunning=安装程序已在运行中。
SetupAppRunning=以下应用程序正在运行。%n%n请关闭它们，然后单击"重试"继续，或单击"取消"退出。%n%n%1
SetupUnableToCreateTempDir=安装程序无法创建临时目录。%n%n%1
SetupMissingSetupExe=安装程序缺少必要的文件。
SetupMissingSetupIni=安装程序缺少必要的文件。
SetupWrongPlatform=此版本的 [name] 无法安装在此版本的 Windows 上。%n%n请获取适合此电脑的版本。
SetupWrongPlatform64=此版本的 [name] 无法安装在此版本的 Windows 上。%n%n请获取 64 位版本。

; *** Misc. common
SetupFileMissing=安装程序缺少文件 %1。请修复此问题或获取新的安装程序副本。
SetupFileCorrupt=安装文件已损坏。请获取新的安装程序副本。

; *** Tasks
WizardSelectTasks=选择附加任务
SelectTasksDesc=要执行哪些附加任务？
SelectTasksLabel2=选择要执行的附加任务，然后单击"下一步"。

; *** Custom messages for JDM
JDMJDKStorageTitle=JDK 存储位置
JDMJDKStorageDesc=所有下载的 JDK 版本将存储在此目录中。
JDMJDKStorageSubCaption=选择 JDK 版本的存储文件夹。
JDMPathAddLabel=将 JDM 添加到系统 PATH 环境变量
JDMDesktopIconLabel=创建桌面快捷方式
JDMStartMenuIconLabel=创建开始菜单快捷方式
JDMConfigGenerated=已根据您的选择自动生成配置文件。
JDMEnvVarSet=已设置 JDM_HOME 环境变量。
JDMPathUpdated=已更新系统 PATH 环境变量。
JDMInstallComplete=JDM 已成功安装到您的电脑。
JDMOpenNewTerminal=请打开新的命令提示符或 PowerShell 窗口以使用 JDM。

; *** Uninstaller custom messages
JDMUninstallWarning=您即将卸载 JDM（JDK 版本管理器）。%n%n%n此操作将：%n%n  • 删除 jdm 命令%n%n  • 删除所有已安装的 JDK 版本%n%n  • 删除所有 JDM 配置文件%n%n  • 清理环境变量%n%n%n确定要继续吗？
JDMUninstallComplete=JDM 已从您的电脑上完全删除。

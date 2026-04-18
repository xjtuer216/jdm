@echo off
setlocal enabledelayedexpansion

:: ============================================================================
:: JDM Windows Installer - Build Script
:: ============================================================================
:: Usage: build.bat [version]
:: Example: build.bat 1.0.0
::
:: Prerequisites:
::   - Go 1.21+ installed and in PATH
::   - Inno Setup 6.x installed (iscc.exe in PATH or default location)
:: ============================================================================

echo ========================================
echo JDM Installer Build Script
echo ========================================
echo.

:: Get version from argument or default
if "%~1"=="" (
    set "VERSION=1.0.0"
) else (
    set "VERSION=%~1"
)

echo Version: %VERSION%
echo.

:: Get script directory
set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%.."
set "CORE_DIR=%PROJECT_ROOT%\core"
set "ASSETS_DIR=%SCRIPT_DIR%assets"
set "DIST_DIR=%SCRIPT_DIR%dist"

:: Check Go installation
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed or not in PATH.
    echo Please install Go from https://go.dev/dl/
    pause
    exit /b 1
)

:: Check Inno Setup installation
set "ISCC="
where iscc >nul 2>&1
if %errorlevel% equ 0 (
    set "ISCC=iscc"
) else (
    :: Try default installation paths
    if exist "D:\develop\Inno Setup 6\ISCC.exe" (
        set "ISCC=D:\develop\Inno Setup 6\ISCC.exe"
    ) else if exist "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" (
        set "ISCC=C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
    ) else if exist "C:\Program Files\Inno Setup 6\ISCC.exe" (
        set "ISCC=C:\Program Files\Inno Setup 6\ISCC.exe"
    ) else (
        :: Try to find via Registry
        for /f "tokens=2*" %%A in ('reg query "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Inno Setup 6_is1" /v InstallLocation 2^>nul') do set "ISCC_REG=%%B"
        if defined ISCC_REG (
            if exist "!ISCC_REG!ISCC.exe" set "ISCC=!ISCC_REG!ISCC.exe"
        )
    )
)

if "%ISCC%"=="" (
    echo [ERROR] Inno Setup 6 is not installed.
    echo Please install from https://jrsoftware.org/isdl.php
    pause
    exit /b 1
)

echo [1/4] Building jdm.exe...
cd /d "%CORE_DIR%"
if %errorlevel% neq 0 (
    echo [ERROR] Failed to change to core directory: %CORE_DIR%
    pause
    exit /b 1
)

go build -ldflags "-X github.com/xjtuer216/jdm/internal/jdk.Version=%VERSION%" -o "%ASSETS_DIR%\jdm.exe" .
if %errorlevel% neq 0 (
    echo [ERROR] Failed to build jdm.exe
    pause
    exit /b 1
)
echo [OK] jdm.exe built successfully
echo.

echo [2/4] Verifying assets...
if not exist "%ASSETS_DIR%\jdm.exe" (
    echo [ERROR] jdm.exe not found in assets directory
    pause
    exit /b 1
)
if not exist "%ASSETS_DIR%\license.txt" (
    echo [WARNING] license.txt not found in assets directory
)
if not exist "%ASSETS_DIR%\README.txt" (
    echo [WARNING] README.txt not found in assets directory
)
echo [OK] Assets verified
echo.

echo [3/4] Creating output directory...
if not exist "%DIST_DIR%" mkdir "%DIST_DIR%"
echo [OK] Output directory ready: %DIST_DIR%
echo.

echo [4/4] Building installer...
cd /d "%SCRIPT_DIR%"
"%ISCC%" /dMyAppVersion=%VERSION% "jdm.iss"
if %errorlevel% neq 0 (
    echo [ERROR] Failed to build installer
    pause
    exit /b 1
)
echo.

echo ========================================
echo Build Complete!
echo ========================================
echo.
echo Installer: %DIST_DIR%\jdm-setup-%VERSION%.exe
echo.
dir "%DIST_DIR%\jdm-setup-%VERSION%.exe" 2>nul
echo.

pause

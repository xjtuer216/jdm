# JDM Windows Installer - Design Document

> **Historical design note (2026-04)**: kept for rationale. Current behavior lives in `installer/jdm.iss` and `docs/RELEASE_OPERATIONS.md`.

## Overview

Create a Windows installer for JDM (JDK Version Manager) using Inno Setup, with bilingual support (English + Chinese Simplified), automatic PATH configuration, environment variable management, and complete uninstaller cleanup.

## Architecture

```
installer/
├── jdm.iss                          # Main Inno Setup script
├── build.bat                        # Build automation
├── Languages/
│   └── ChineseSimplified.isl        # Chinese language file
├── assets/
│   ├── jdm.exe                      # Built binary (copied by build.bat)
│   ├── jdm.ico                      # Application icon
│   ├── license.txt                  # MIT License
│   └── README.txt                   # Installation notes
└── dist/                            # Build output (gitignored)
    └── jdm-setup-{version}.exe
```

## Installation Flow

1. **Language Detection**: Auto-detect system language (EN/ZH), show selection dialog
2. **Welcome Page**: Standard Inno Setup welcome
3. **License Page**: MIT License agreement
4. **Install Path Page**: User selects JDM installation directory (default: `%LOCALAPPDATA%\jdm`)
5. **JDK Storage Path Page**: User selects JDK versions storage directory (default: `%USERPROFILE%\.jdm\versions`)
6. **Tasks Page**: Checkboxes for PATH configuration, desktop shortcut
7. **Ready to Install**: Summary
8. **Installation**: Copy files, set environment variables, update PATH
9. **Finish**: Launch option

## Environment Variables

| Variable | Scope | Value | Purpose |
|----------|-------|-------|---------|
| `JDM_HOME` | System + User | `{app}` | JDM data directory |
| `PATH` | System | `{app}` | Access jdm.exe |
| `PATH` | System | `{JDM_HOME}\current\bin` | Access java/javac |

## Uninstaller Cleanup

1. Remove `{app}` directory (jdm.exe + config)
2. Remove `{JDM_HOME}\versions` directory (all JDK versions)
3. Remove `JDM_HOME` from registry (system + user)
4. Remove `{app}` from PATH (system)
5. Remove `{JDM_HOME}\current\bin` from PATH (system)
6. Refresh environment

## Technical Details

### Inno Setup Configuration
- **App GUID**: `B4E84F72-3C5A-4D1B-9F2E-8A7C6D5E4F3A`
- **PrivilegesRequired**: `lowest` (no admin needed for most operations)
- **ChangesEnvironment**: `yes`
- **LanguageDetectionMethod**: `uilanguage`
- **ShowLanguageDialog**: `yes`
- **ArchitecturesInstallIn64BitMode**: `x64`

### PATH Manipulation (Pascal Script)
- Use `RegQueryStringValue` / `RegWriteExpandStringValue` on `HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
- Idempotent: check for existing path before adding
- Case-insensitive matching
- Clean up double semicolons

### Build Process
```
build.bat [version]
  1. Build jdm.exe with version injection
  2. Copy to installer/assets/
  3. Run iscc.exe jdm.iss
  4. Output to installer/dist/jdm-setup-{version}.exe
```

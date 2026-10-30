JDM - JDK Version Manager for Windows
======================================

System Requirements:
- Windows 10/11 (64-bit)
- No administrator privileges required (uses lowest privilege mode)

Installation Notes:
- The installer will add JDM to your system PATH environment variable
- JDM_HOME environment variable will be set to the installation directory
- JDK versions are stored in the "versions" subdirectory of the installation directory
- Only the installation directory can be customized during setup; the JDK storage path follows it

After Installation:
1. Open a new command prompt or PowerShell window
2. Run 'jdm --help' to see available commands
3. Run 'jdm install 17' to install your first JDK version
4. Run 'jdm use 17' to switch to the installed version

Uninstallation:
- Use Windows Settings > Apps > JDM to uninstall
- Or run the uninstaller from the Start Menu
- Uninstallation will remove all installed JDK versions and configuration

For more information, visit: https://github.com/xjtuer216/jdm

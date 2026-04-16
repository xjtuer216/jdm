package jdk

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/whimsy/jdm/internal/config"
	"github.com/whimsy/jdm/internal/file"
	"github.com/whimsy/jdm/internal/web"
)

type InstalledVersion struct {
	Version   string `json:"version"`
	Path      string `json:"path"`
	IsCurrent bool   `json:"is_current"`
	IsDefault bool   `json:"is_default"`
}

type VersionManager struct {
	Config *config.Config
	Client *web.AdoptiumClient
}

func NewVersionManager(cfg *config.Config) *VersionManager {
	return &VersionManager{
		Config: cfg,
		Client: web.NewAdoptiumClient(cfg.Mirror),
	}
}

// ListLocal lists all locally installed JDK versions
func (vm *VersionManager) ListLocal() ([]InstalledVersion, error) {
	entries, err := os.ReadDir(vm.Config.JDKHome)
	if err != nil {
		return nil, err
	}

	var versions []InstalledVersion
	currentVersion, _ := vm.GetCurrent()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		version := entry.Name()
		// Extract version from directory name (e.g., "jdk-17.0.2+2" -> "17.0.2+2")
		version = strings.TrimPrefix(version, "jdk-")

		path := filepath.Join(vm.Config.JDKHome, entry.Name())

		versions = append(versions, InstalledVersion{
			Version:   version,
			Path:      path,
			IsCurrent: currentVersion != nil && currentVersion.Version == version,
			IsDefault: vm.Config.Default == version,
		})
	}

	return versions, nil
}

// ListRemote lists all available remote JDK versions for a major version
func (vm *VersionManager) ListRemote(version string) ([]web.RemoteVersion, error) {
	majorVersion := vm.parseMajorVersion(version)
	return vm.Client.FetchVersions(majorVersion)
}

// Install installs a JDK version
func (vm *VersionManager) Install(version string) error {
	// Resolve version
	resolvedVersion, err := vm.Client.ResolveVersion(version)
	if err != nil {
		return fmt.Errorf("failed to resolve version: %w", err)
	}

	// Check if already installed
	versions, err := vm.ListLocal()
	if err == nil {
		for _, v := range versions {
			if v.Version == resolvedVersion {
				return fmt.Errorf("version %s is already installed", resolvedVersion)
			}
		}
	}

	// Get download URL
	remoteVersions, err := vm.Client.FetchVersions(vm.parseMajorVersion(version))
	if err != nil {
		return fmt.Errorf("failed to fetch versions: %w", err)
	}

	var downloadURL string
	var fileSize int64

	for _, v := range remoteVersions {
		if v.Version == resolvedVersion {
			downloadURL = v.DownloadURL
			fileSize = v.FileSize
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("download URL not found for version %s", resolvedVersion)
	}

	// Download
	fmt.Printf("Downloading JDK %s (%s)...\n", resolvedVersion, formatSize(fileSize))

	tmpFile, err := vm.downloadFile(downloadURL, fileSize)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer os.Remove(tmpFile)

	// Extract
	installPath := filepath.Join(vm.Config.JDKHome, fmt.Sprintf("jdk-%s", resolvedVersion))
	fmt.Printf("Extracting to %s...\n", installPath)

	if err := file.Extract(tmpFile, installPath); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	fmt.Printf("JDK %s installed successfully!\n", resolvedVersion)
	return nil
}

// Uninstall uninstalls a JDK version
func (vm *VersionManager) Uninstall(version string) error {
	versions, err := vm.ListLocal()
	if err != nil {
		return err
	}

	var targetPath string
	for _, v := range versions {
		if v.Version == version {
			targetPath = v.Path
			break
		}
	}

	if targetPath == "" {
		return fmt.Errorf("version %s is not installed", version)
	}

	// Check if it's the current version
	current, _ := vm.GetCurrent()
	if current != nil && current.Version == version {
		return fmt.Errorf("cannot uninstall current version. Use 'jdm use <other>' first")
	}

	// Remove directory
	if err := os.RemoveAll(targetPath); err != nil {
		return fmt.Errorf("failed to remove version: %w", err)
	}

	// Update config if it was default
	if vm.Config.Default == version {
		vm.Config.Default = ""
		vm.Config.Save()
	}

	fmt.Printf("JDK %s uninstalled successfully!\n", version)
	return nil
}

// GetCurrent gets the currently active JDK version
func (vm *VersionManager) GetCurrent() (*InstalledVersion, error) {
	currentPath := filepath.Join(vm.Config.JDMHome, "current")

	info, err := os.Lstat(currentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	if (info.Mode() & os.ModeSymlink) == 0 {
		return nil, fmt.Errorf("current is not a symlink")
	}

	target, err := os.Readlink(currentPath)
	if err != nil {
		return nil, err
	}

	// Extract version from path
	version := filepath.Base(target)
	version = strings.TrimPrefix(version, "jdk-")

	return &InstalledVersion{
		Version:   version,
		Path:      target,
		IsCurrent: true,
	}, nil
}

// Use switches to a different JDK version temporarily
func (vm *VersionManager) Use(version string) error {
	versions, err := vm.ListLocal()
	if err != nil {
		return err
	}

	var targetPath string
	for _, v := range versions {
		if v.Version == version || vm.Config.Aliases[version] == v.Version {
			targetPath = v.Path
			break
		}
	}

	if targetPath == "" {
		return fmt.Errorf("version %s is not installed", version)
	}

	currentLink := filepath.Join(vm.Config.JDMHome, "current")

	if err := file.EnsureCurrentSymlink(targetPath, currentLink); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Update current_version file
	currentVersionFile := filepath.Join(vm.Config.JDMHome, "current_version")
	os.WriteFile(currentVersionFile, []byte(version), 0644)

	fmt.Printf("Now using JDK %s\n", version)
	fmt.Printf("JAVA_HOME: %s\n", targetPath)
	return nil
}

// SetDefault sets the default JDK version
func (vm *VersionManager) SetDefault(version string) error {
	versions, err := vm.ListLocal()
	if err != nil {
		return err
	}

	var targetPath string
	for _, v := range versions {
		if v.Version == version {
			targetPath = v.Path
			break
		}
	}

	if targetPath == "" {
		return fmt.Errorf("version %s is not installed", version)
	}

	vm.Config.Default = version
	if err := vm.Config.Save(); err != nil {
		return err
	}

	// Also create symlink
	currentLink := filepath.Join(vm.Config.JDMHome, "current")
	if err := file.EnsureCurrentSymlink(targetPath, currentLink); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	fmt.Printf("Default JDK version set to %s\n", version)
	return nil
}

func (vm *VersionManager) parseMajorVersion(version string) int {
	version = strings.TrimPrefix(version, "jdk-")
	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		var major int
		fmt.Sscanf(parts[0], "%d", &major)
		return major
	}
	return 0
}

func (vm *VersionManager) downloadFile(url string, expectedSize int64) (string, error) {
	resp, err := vm.Client.Client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "jdm-*.zip")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Init initializes the version manager (creates directories)
func (vm *VersionManager) Init() error {
	return vm.Config.Load()
}
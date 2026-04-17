package jdk

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xjtuer216/jdm/internal/config"
	"github.com/xjtuer216/jdm/internal/file"
	"github.com/xjtuer216/jdm/internal/progress"
	"github.com/xjtuer216/jdm/internal/web"
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
	if version == "" {
		// Get all available major versions
		releases, ltsReleases, _, _, err := vm.Client.ListAvailableReleases()
		if err != nil {
			return nil, fmt.Errorf("failed to list available releases: %w", err)
		}

		// Create a set of LTS major versions for quick lookup
		ltsSet := make(map[int]bool)
		for _, v := range ltsReleases {
			ltsSet[v] = true
		}

		var allVersions []web.RemoteVersion
		for _, major := range releases {
			latest, err := vm.Client.GetLatestPerMajor(major)
			if err != nil {
				return nil, fmt.Errorf("failed to get latest for major %d: %w", major, err)
			}
			if latest != nil {
				latest.IsLTS = ltsSet[major]
				allVersions = append(allVersions, *latest)
			}
		}

		// Sort by version descending (newest first)
		sort.Slice(allVersions, func(i, j int) bool {
			return web.CompareVersions(allVersions[i].Version, allVersions[j].Version) > 0
		})

		return allVersions, nil
	}

	majorVersion := vm.parseMajorVersion(version)
	return vm.Client.FetchVersions(majorVersion)
}

// Install installs a JDK version
func (vm *VersionManager) Install(version string) error {
	// Resolve version
	fmt.Println("Resolving version...")
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
	tmpFile, err := vm.downloadFile(downloadURL, fileSize)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer os.Remove(tmpFile)

	// Extract
	installPath := filepath.Join(vm.Config.JDKHome, fmt.Sprintf("jdk-%s", resolvedVersion))
	if err := vm.extractWithProgress(tmpFile, installPath); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	fmt.Printf("JDK %s installed successfully!\n", resolvedVersion)
	return nil
}

// extractWithProgress extracts an archive with a progress bar.
func (vm *VersionManager) extractWithProgress(archive, dest string) error {
	// Count total files first
	totalFiles, err := countArchiveFiles(archive)
	if err != nil {
		return err
	}

	bar := progress.NewFiles("Extracting", int64(totalFiles))
	defer bar.Done()

	return file.ExtractWithProgress(archive, dest, func(current, total int64) {
		bar.Update(current)
	})
}

// countArchiveFiles counts the number of files in a zip archive.
func countArchiveFiles(archive string) (int, error) {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	count := 0
	for _, f := range r.File {
		if !f.FileInfo().IsDir() {
			count++
		}
	}
	return count, nil
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
	// Use download mirror if configured
	// Mirror format: prefix the original URL, e.g. "https://ghproxy.net/https://github.com/..."
	downloadURL := url
	if vm.Config.DownloadMirror != "" && strings.Contains(url, "github.com") {
		downloadURL = vm.Config.DownloadMirror + "/" + url
	}

	// Use a separate client with longer timeout for large file downloads
	downloadClient := &http.Client{
		Timeout: 10 * time.Minute,
	}

	resp, err := downloadClient.Get(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d (URL: %s)", resp.StatusCode, downloadURL)
	}

	tmpFile, err := os.CreateTemp("", "jdm-*.zip")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// Create progress bar
	bar := progress.New("Downloading", expectedSize)
	defer bar.Done()

	// Wrap response body with progress tracking
	reader := &progressReader{
		reader: resp.Body,
		bar:    bar,
	}

	if _, err := io.Copy(tmpFile, reader); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

// progressReader wraps an io.Reader and updates a progress bar on each Read.
type progressReader struct {
	reader io.Reader
	bar    *progress.ProgressBar
	total  int64
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	if n > 0 {
		pr.total += int64(n)
		pr.bar.Update(pr.total)
	}
	return
}

// Init initializes the version manager (creates directories)
func (vm *VersionManager) Init() error {
	return vm.Config.Load()
}

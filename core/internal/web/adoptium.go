package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Adoptium API types
type AdoptiumAsset struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	BinaryLink   string `json:"binary_link"`
	Package      struct {
		Link string `json:"link"`
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"package"`
}

type AdoptiumVersion struct {
	Version     string          `json:"version"`
	ReleaseName string          `json:"release_name"`
	Vendor      string          `json:"vendor"`
	Source      string          `json:"source"`
	Assets      []AdoptiumAsset `json:"assets"`
}

type FeatureVersion struct {
	Version  int               `json:"version"`
	Releases []AdoptiumVersion `json:"releases"`
}

type RemoteVersion struct {
	Version     string `json:"version"`
	ReleaseName string `json:"release_name"`
	Vendor      string `json:"vendor"`
	DownloadURL string `json:"download_url"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	IsLTS       bool   `json:"is_lts"`
}

// AdoptiumRelease represents a release from the feature_releases API
type AdoptiumRelease struct {
	ReleaseName string           `json:"release_name"`
	Vendor      string           `json:"vendor"`
	VersionData VersionDataField `json:"version_data"`
	Binaries    []AdoptiumBinary `json:"binaries"`
}

// VersionDataField represents version information from the API
type VersionDataField struct {
	Semver string `json:"semver"`
}

// AdoptiumBinary represents a binary in a release
type AdoptiumBinary struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Package      struct {
		Link string `json:"link"`
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"package"`
}

type AdoptiumClient struct {
	Mirror string
	Client *http.Client
}

func NewAdoptiumClient(mirror string) *AdoptiumClient {
	return &AdoptiumClient{
		Mirror: mirror,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *AdoptiumClient) FetchVersions(javaVersion int) ([]RemoteVersion, error) {
	url := fmt.Sprintf("%s/v3/assets/feature_releases/%d/ga?image_type=jdk&os=windows&architecture=x64", c.Mirror, javaVersion)

	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch versions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var releases []AdoptiumRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if this major version is LTS
	_, ltsReleases, _, _, err := c.ListAvailableReleases()
	isLTSVersion := false
	if err == nil {
		for _, lts := range ltsReleases {
			if lts == javaVersion {
				isLTSVersion = true
				break
			}
		}
	}

	var versions []RemoteVersion
	for _, release := range releases {
		for _, binary := range release.Binaries {
			versions = append(versions, RemoteVersion{
				Version:     release.VersionData.Semver,
				ReleaseName: release.ReleaseName,
				Vendor:      release.Vendor,
				DownloadURL: binary.Package.Link,
				FileName:    binary.Package.Name,
				FileSize:    binary.Package.Size,
				IsLTS:       isLTSVersion,
			})
		}
	}

	// Sort by version (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return CompareVersions(versions[i].Version, versions[j].Version) > 0
	})

	return versions, nil
}

// FetchLatestLTS fetches the latest LTS version for a given major version
func (c *AdoptiumClient) FetchLatestLTS(javaVersion int) (*RemoteVersion, error) {
	versions, err := c.FetchVersions(javaVersion)
	if err != nil {
		return nil, err
	}

	for _, v := range versions {
		if isLTS(v.ReleaseName) {
			return &v, nil
		}
	}

	if len(versions) > 0 {
		return &versions[0], nil
	}

	return nil, fmt.Errorf("no versions available for Java %d", javaVersion)
}

// ListAllVersions lists all available versions across major versions
func (c *AdoptiumClient) ListAllVersions() (map[int][]RemoteVersion, error) {
	result := make(map[int][]RemoteVersion)
	versionsToCheck := []int{8, 11, 17, 21, 25}

	for _, v := range versionsToCheck {
		versions, err := c.FetchVersions(v)
		if err != nil {
			continue
		}
		if len(versions) > 0 {
			result[v] = versions
		}
	}

	return result, nil
}

// AvailableReleases represents the available releases information from Adoptium API
type AvailableReleases struct {
	AvailableReleases        []int `json:"available_releases"`
	AvailableLTSReleases     []int `json:"available_lts_releases"`
	MostRecentLTS            int   `json:"most_recent_lts"`
	MostRecentFeatureVersion int   `json:"most_recent_feature_version"`
}

// ListAvailableReleases fetches available releases information from Adoptium API
func (c *AdoptiumClient) ListAvailableReleases() ([]int, []int, int, int, error) {
	url := fmt.Sprintf("%s/v3/info/available_releases", c.Mirror)

	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("failed to fetch available releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, 0, 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("failed to read response: %w", err)
	}

	var releases AvailableReleases
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, nil, 0, 0, fmt.Errorf("failed to parse response: %w", err)
	}

	return releases.AvailableReleases, releases.AvailableLTSReleases, releases.MostRecentLTS, releases.MostRecentFeatureVersion, nil
}

// GetLatestPerMajor returns the latest version for a given major version
func (c *AdoptiumClient) GetLatestPerMajor(majorVersion int) (*RemoteVersion, error) {
	versions, err := c.FetchVersions(majorVersion)
	if err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions available for Java %d", majorVersion)
	}

	return &versions[0], nil
}

// ResolveVersion resolves a version string (e.g., "17" -> "17.0.2+2")
func (c *AdoptiumClient) ResolveVersion(version string) (string, error) {
	majorVersion := parseVersionNumber(version)

	versions, err := c.FetchVersions(majorVersion)
	if err != nil {
		return "", err
	}

	// Exact match
	for _, v := range versions {
		if v.Version == version {
			return v.Version, nil
		}
	}

	// Short version (e.g., "17") -> latest
	if version == strconv.Itoa(majorVersion) {
		latest, err := c.FetchLatestLTS(majorVersion)
		if err != nil {
			return "", err
		}
		return latest.Version, nil
	}

	return "", fmt.Errorf("version %s not found", version)
}

func parseVersionNumber(version string) int {
	version = strings.TrimPrefix(version, "jdk-")
	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		major, _ := strconv.Atoi(parts[0])
		return major
	}
	major, _ := strconv.Atoi(version)
	return major
}

func CompareVersions(v1, v2 string) int {
	parts1 := strings.Split(strings.TrimPrefix(v1, "jdk-"), ".")
	parts2 := strings.Split(strings.TrimPrefix(v2, "jdk-"), ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		p1, _ := strconv.Atoi(parts1[i])
		p2, _ := strconv.Atoi(parts2[i])
		if p1 != p2 {
			return p1 - p2
		}
	}
	return len(parts1) - len(parts2)
}

func isLTS(releaseName string) bool {
	return strings.Contains(strings.ToLower(releaseName), "lts")
}

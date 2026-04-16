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
	url := fmt.Sprintf("%s/v3/assets/feature_versions/%d", c.Mirror, javaVersion)

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

	var featureVersion FeatureVersion
	if err := json.Unmarshal(body, &featureVersion); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var versions []RemoteVersion
	for _, release := range featureVersion.Releases {
		for _, asset := range release.Assets {
			// Filter for Windows x64
			if asset.OS == "windows" && asset.Architecture == "x64" {
				versions = append(versions, RemoteVersion{
					Version:     release.Version,
					ReleaseName: release.ReleaseName,
					Vendor:      release.Vendor,
					DownloadURL: asset.Package.Link,
					FileName:    asset.Package.Name,
					FileSize:    asset.Package.Size,
				})
			}
		}
	}

	// Sort by version (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i].Version, versions[j].Version) > 0
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

func compareVersions(v1, v2 string) int {
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

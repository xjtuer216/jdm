package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var versionRegex = regexp.MustCompile(`^(\d+)\.(\d+)\.?(\d*)[+]?(\S*)$`)

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
	Build string // e.g., "+2" or "-beta"
}

// Parse parses a version string into a Version struct
func Parse(v string) (*Version, error) {
	// Remove "jdk-" prefix if present
	v = strings.TrimPrefix(v, "jdk-")
	v = strings.TrimPrefix(v, "jdk")

	// Match patterns like "17.0.2+2", "17.0.2", "17.0.2-beta"
	matches := versionRegex.FindStringSubmatch(v)

	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid version format: %s", v)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])

	patch := 0
	if len(matches) >= 4 && matches[3] != "" {
		patch, _ = strconv.Atoi(matches[3])
	}

	build := ""
	if len(matches) >= 5 {
		build = matches[4]
	}

	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Build: build,
	}, nil
}

// ParseMajor extracts major version from a version string
func ParseMajor(v string) (int, error) {
	ver, err := Parse(v)
	if err != nil {
		return 0, err
	}
	return ver.Major, nil
}

// String returns the version string
func (v *Version) String() string {
	if v.Build != "" {
		return fmt.Sprintf("%d.%d.%d+%s", v.Major, v.Minor, v.Patch, v.Build)
	}
	if v.Patch > 0 {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	}
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// Compare compares two versions
// Returns: -1 if v < other, 0 if v == other, 1 if v > other
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

// IsCompatible checks if this version is compatible with another (same major version)
func (v *Version) IsCompatible(other *Version) bool {
	return v.Major == other.Major
}

// Matches checks if version matches a pattern (e.g., "17", "17.0", "17.0.2")
func (v *Version) Matches(pattern string) bool {
	p, err := Parse(pattern)
	if err != nil {
		return false
	}

	if p.Major != v.Major {
		return false
	}
	if p.Minor > 0 && p.Minor != v.Minor {
		return false
	}
	if p.Patch > 0 && p.Patch != v.Patch {
		return false
	}
	return true
}
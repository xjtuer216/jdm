package arch

import (
	"os"
	"runtime"
	"strings"
)

// Supported architectures
const (
	X64   = "x64"
	X86   = "x86"
	ARM64 = "arm64"
)

// GetSystemArch returns the system architecture
func GetSystemArch() string {
	arch := runtime.GOARCH
	if arch == "amd64" {
		return X64
	}
	if arch == "386" {
		return X86
	}
	if arch == "arm64" {
		return ARM64
	}
	return arch
}

// NormalizeArch normalizes architecture string
func NormalizeArch(arch string) string {
	arch = strings.ToLower(arch)
	if arch == "x64" || arch == "amd64" || arch == "64" {
		return X64
	}
	if arch == "x86" || arch == "386" || arch == "32" {
		return X86
	}
	if arch == "arm64" || arch == "aarch64" {
		return ARM64
	}
	return arch
}

// IsValidArch checks if the architecture is supported
func IsValidArch(arch string) bool {
	arch = NormalizeArch(arch)
	return arch == X64 || arch == X86 || arch == ARM64
}

// GetJavaArch returns the architecture string used in JDK download filenames
func GetJavaArch() string {
	arch := GetSystemArch()
	if arch == X64 {
		return "x64"
	}
	if arch == ARM64 {
		return "aarch64"
	}
	return "i586" // x86 in OpenJDK naming
}

// GetEnvArch returns architecture from environment or detects from executable
func GetEnvArch() string {
	// Check PROCESSOR_ARCHITECTURE environment variable
	if arch := os.Getenv("PROCESSOR_ARCHITECTURE"); arch != "" {
		return NormalizeArch(arch)
	}
	// Fall back to system detection
	return GetSystemArch()
}
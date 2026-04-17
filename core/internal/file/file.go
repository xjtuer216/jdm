package file

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// CreateSymlink creates a symbolic link on Windows.
// Falls back to directory junction if symlink fails due to privilege issues.
func CreateSymlink(target, link string) error {
	// Remove existing link if exists
	if _, err := os.Lstat(link); err == nil {
		if err := os.Remove(link); err != nil {
			return err
		}
	}

	// Try symlink first
	err := os.Symlink(target, link)
	if err == nil {
		return nil
	}

	// On Windows, fall back to directory junction if privilege error
	if runtime.GOOS == "windows" {
		return createJunction(target, link)
	}

	return err
}

// createJunction creates a directory junction on Windows using mklink /J.
// Junctions don't require administrator privileges.
func createJunction(target, link string) error {
	// Ensure target is an absolute path
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("failed to resolve target path: %w", err)
	}

	cmd := exec.Command("cmd", "/c", "mklink", "/J", link, absTarget)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create junction (%s): %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// RemoveSymlink removes a symbolic link
func RemoveSymlink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return os.Remove(path)
	}

	return os.Remove(path)
}

// ReadSymlink reads the target of a symbolic link
func ReadSymlink(path string) (string, error) {
	return os.Readlink(path)
}

// EnsureCurrentSymlink ensures the current symlink points to the target
func EnsureCurrentSymlink(target, currentPath string) error {
	currentDir := filepath.Dir(currentPath)

	// Ensure parent directory exists
	if err := os.MkdirAll(currentDir, 0755); err != nil {
		return err
	}

	return CreateSymlink(target, currentPath)
}

// Exists checks if a file or directory exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// IsFile checks if a path is a regular file
func IsFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// EnsureDir ensures a directory exists, creating it if necessary
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// Remove removes a file or directory
func Remove(path string) error {
	return os.RemoveAll(path)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// Extract extracts an archive (zip or tar.gz) to the destination directory
func Extract(archive, dest string) error {
	return ExtractWithProgress(archive, dest, nil)
}

// ExtractWithProgress extracts an archive with progress callback.
// The progress callback receives (current, total) file counts.
func ExtractWithProgress(archive, dest string, progress func(current, total int64)) error {
	// Ensure destination exists
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// Determine archive type
	lower := strings.ToLower(archive)
	if strings.HasSuffix(lower, ".zip") {
		return extractZipWithProgress(archive, dest, progress)
	}
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return extractTarGzWithProgress(archive, dest, progress)
	}

	return fmt.Errorf("unsupported archive format: %s", archive)
}

func extractZipWithProgress(archive, dest string, progress func(current, total int64)) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer r.Close()

	// Count total files for progress tracking
	totalFiles := int64(0)
	for _, f := range r.File {
		if !f.FileInfo().IsDir() {
			totalFiles++
		}
	}

	var extracted int64
	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)

		// Prevent zip slip vulnerability
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		inFile, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, copyErr := io.Copy(outFile, inFile)
		outFile.Close()
		inFile.Close()
		if copyErr != nil {
			return copyErr
		}

		extracted++
		if progress != nil {
			progress(extracted, totalFiles)
		}
	}

	return nil
}

func extractTarGzWithProgress(archive, dest string, progress func(current, total int64)) error {
	file, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer file.Close()

	tr := tar.NewReader(file)

	// First pass: count total files
	totalFiles := int64(0)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header.Typeflag == tar.TypeReg {
			totalFiles++
		}
	}

	// Reset reader for second pass
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	// Re-open to reset tar reader
	file.Close()
	file, err = os.Open(archive)
	if err != nil {
		return err
	}
	defer file.Close()
	tr = tar.NewReader(file)

	var extracted int64
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(dest, header.Name)

		// Prevent tar slip
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}

			_, copyErr := io.Copy(outFile, tr)
			outFile.Close()
			if copyErr != nil {
				return copyErr
			}

			extracted++
			if progress != nil {
				progress(extracted, totalFiles)
			}
		}
	}

	return nil
}

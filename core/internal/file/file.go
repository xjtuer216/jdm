package file

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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

// Extract extracts an archive (zip or tar.gz) to the destination directory.
// An archive whose entries share one top-level directory has that directory
// stripped, so dest becomes the archive's content root.
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

// normalizeArchiveEntry converts an archive entry name to a clean relative
// path with forward slashes, so names from different writers compare equally.
func normalizeArchiveEntry(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimPrefix(name, "./")
	return strings.TrimLeft(name, "/")
}

// rootDirDetector tracks the top-level directory shared by archive entries.
type rootDirDetector struct {
	root       string
	hasSubpath bool
	multiple   bool
}

// Add records an archive entry name.
func (d *rootDirDetector) Add(name string) {
	name = normalizeArchiveEntry(name)
	if name == "" {
		return
	}

	first, rest, found := strings.Cut(name, "/")
	if found && rest != "" {
		d.hasSubpath = true
	}
	if d.root == "" {
		d.root = first
	} else if first != d.root {
		d.multiple = true
	}
}

// Root returns the single top-level directory that wraps every entry, or ""
// when entries are flat or live under multiple roots. A root is only reported
// when at least one entry has a subpath, so a lone top-level file is kept.
func (d *rootDirDetector) Root() string {
	if d.multiple || !d.hasSubpath || d.root == "" || d.root == "." || d.root == ".." {
		return ""
	}
	return d.root
}

// stripArchiveRoot removes root from an already-normalized entry name.
// ok is false when the entry is the root directory itself, a no-op ".",
// or empty, meaning there is no content to extract.
func stripArchiveRoot(name, root string) (string, bool) {
	if name == "" || name == "." {
		return "", false
	}
	if root == "" {
		return name, true
	}
	if name == root {
		return "", false
	}
	rest, ok := strings.CutPrefix(name, root+"/")
	if !ok {
		return name, true
	}
	if rest == "" {
		return "", false
	}
	return rest, true
}

func extractZipWithProgress(archive, dest string, progress func(current, total int64)) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer r.Close()

	// Count files for progress tracking and detect the archive root at once.
	totalFiles := int64(0)
	var root rootDirDetector
	for _, f := range r.File {
		root.Add(f.Name)
		if !f.FileInfo().IsDir() {
			totalFiles++
		}
	}

	// JDK archives wrap their contents in one directory (e.g. jdk-17.0.1+12/).
	// Keep it only when the archive has no single, unambiguous root.
	rootDir := root.Root()

	var extracted int64
	for _, f := range r.File {
		name, ok := stripArchiveRoot(normalizeArchiveEntry(f.Name), rootDir)
		if !ok {
			// Only the root directory entry has no content; a file named
			// like the root would silently disappear, so reject it.
			if f.FileInfo().IsDir() {
				continue
			}
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		path := filepath.Join(dest, name)

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

// walkTarGz opens a gzip-compressed tarball and visits every entry. The archive
// is opened per pass because gzip and tar streams cannot be rewound.
func walkTarGz(archive string, visit func(header *tar.Header, content io.Reader) error) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		// PAX headers are archive metadata, not content. Global headers are
		// returned by Reader.Next and must not take part in root detection.
		if header.Typeflag == tar.TypeXGlobalHeader || header.Typeflag == tar.TypeXHeader {
			continue
		}
		if err := visit(header, tr); err != nil {
			return err
		}
	}
}

func extractTarGzWithProgress(archive, dest string, progress func(current, total int64)) error {
	// First pass: count files for progress tracking and detect the archive root.
	totalFiles := int64(0)
	var root rootDirDetector
	if err := walkTarGz(archive, func(header *tar.Header, _ io.Reader) error {
		root.Add(header.Name)
		if header.Typeflag == tar.TypeReg {
			totalFiles++
		}
		return nil
	}); err != nil {
		return err
	}

	rootDir := root.Root()

	// Second pass: extract entries, stripping the shared root directory.
	var extracted int64
	return walkTarGz(archive, func(header *tar.Header, content io.Reader) error {
		name, ok := stripArchiveRoot(normalizeArchiveEntry(header.Name), rootDir)
		if !ok {
			if header.Typeflag == tar.TypeDir {
				return nil
			}
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		path := filepath.Join(dest, name)

		// Prevent tar slip
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			return os.MkdirAll(path, os.FileMode(header.Mode))
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}

			_, copyErr := io.Copy(outFile, content)
			outFile.Close()
			if copyErr != nil {
				return copyErr
			}

			extracted++
			if progress != nil {
				progress(extracted, totalFiles)
			}
		}
		return nil
	})
}

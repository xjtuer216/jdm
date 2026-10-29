package file_test

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xjtuer216/jdm/internal/file"
)

type archiveEntry struct {
	name     string
	content  string
	typeflag byte
}

func TestExtractZipStripsSingleRootDirectory(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "jdk.zip")
	writeZip(t, archive, []archiveEntry{
		{name: "jdk-17.0.1+12/"},
		{name: "jdk-17.0.1+12/bin/java.exe", content: "java"},
		{name: "jdk-17.0.1+12/lib/modules", content: "modules"},
	})

	dest := filepath.Join(t.TempDir(), "jdk-17.0.1+12")
	if err := file.Extract(archive, dest); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "bin", "java.exe"), "java")
	assertFileContent(t, filepath.Join(dest, "lib", "modules"), "modules")
	assertNotExist(t, filepath.Join(dest, "jdk-17.0.1+12"))
}

func TestExtractZipStripsRootWithoutDirectoryEntry(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "jdk.zip")
	writeZip(t, archive, []archiveEntry{
		{name: "jdk-17.0.1+12/bin/java.exe", content: "java"},
		{name: "jdk-17.0.1+12/lib/modules", content: "modules"},
	})

	dest := filepath.Join(t.TempDir(), "jdk-17.0.1+12")
	if err := file.Extract(archive, dest); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "bin", "java.exe"), "java")
	assertFileContent(t, filepath.Join(dest, "lib", "modules"), "modules")
	assertNotExist(t, filepath.Join(dest, "jdk-17.0.1+12"))
}

func TestExtractZipKeepsFlatArchive(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "jdk.zip")
	writeZip(t, archive, []archiveEntry{
		{name: "bin/java.exe", content: "java"},
		{name: "lib/modules", content: "modules"},
	})

	dest := filepath.Join(t.TempDir(), "jdk")
	if err := file.Extract(archive, dest); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "bin", "java.exe"), "java")
	assertFileContent(t, filepath.Join(dest, "lib", "modules"), "modules")
}

func TestExtractZipKeepsMultipleTopLevelDirectories(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "jdk.zip")
	writeZip(t, archive, []archiveEntry{
		{name: "jdk-a/bin/java.exe", content: "a"},
		{name: "jdk-b/bin/java.exe", content: "b"},
	})

	dest := filepath.Join(t.TempDir(), "jdk")
	if err := file.Extract(archive, dest); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "jdk-a", "bin", "java.exe"), "a")
	assertFileContent(t, filepath.Join(dest, "jdk-b", "bin", "java.exe"), "b")
}

func TestExtractTarGzStripsSingleRootDirectory(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "jdk.tar.gz")
	writeTarGz(t, archive, []archiveEntry{
		{name: "jdk-17.0.1+12/"},
		{name: "jdk-17.0.1+12/bin/java", content: "java"},
		{name: "jdk-17.0.1+12/lib/modules", content: "modules"},
	})

	dest := filepath.Join(t.TempDir(), "jdk-17.0.1+12")
	if err := file.Extract(archive, dest); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "bin", "java"), "java")
	assertFileContent(t, filepath.Join(dest, "lib", "modules"), "modules")
	assertNotExist(t, filepath.Join(dest, "jdk-17.0.1+12"))
}

func TestExtractTarGzKeepsFlatArchive(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "jdk.tar.gz")
	writeTarGz(t, archive, []archiveEntry{
		{name: "bin/java", content: "java"},
		{name: "lib/modules", content: "modules"},
	})

	dest := filepath.Join(t.TempDir(), "jdk")
	if err := file.Extract(archive, dest); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "bin", "java"), "java")
	assertFileContent(t, filepath.Join(dest, "lib", "modules"), "modules")
}

func TestExtractTarGzKeepsMultipleTopLevelDirectories(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "jdk.tar.gz")
	writeTarGz(t, archive, []archiveEntry{
		{name: "jdk-a/bin/java", content: "a"},
		{name: "jdk-b/bin/java", content: "b"},
	})

	dest := filepath.Join(t.TempDir(), "jdk")
	if err := file.Extract(archive, dest); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "jdk-a", "bin", "java"), "a")
	assertFileContent(t, filepath.Join(dest, "jdk-b", "bin", "java"), "b")
}

func TestExtractTarGzStripsRootWithPaxGlobalHeader(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "jdk.tar.gz")
	writeTarGz(t, archive, []archiveEntry{
		{name: "pax_global_header", typeflag: tar.TypeXGlobalHeader},
		{name: "jdk-17.0.1+12/bin/java", content: "java"},
	})

	dest := filepath.Join(t.TempDir(), "jdk-17.0.1+12")
	if err := file.Extract(archive, dest); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "bin", "java"), "java")
	assertNotExist(t, filepath.Join(dest, "jdk-17.0.1+12"))
}

func TestExtractTarGzRejectsPathTraversal(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "jdk.tar.gz")
	writeTarGz(t, archive, []archiveEntry{
		{name: "jdk-17/../../evil.txt", content: "evil"},
	})

	dest := filepath.Join(t.TempDir(), "jdk")
	err := file.Extract(archive, dest)
	if err == nil {
		t.Fatal("Extract() error = nil, want path traversal error")
	}
	assertNotExist(t, filepath.Join(filepath.Dir(filepath.Dir(dest)), "evil.txt"))
}

func TestExtractZipKeepsSingleTopLevelFile(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "jdk.zip")
	writeZip(t, archive, []archiveEntry{
		{name: "readme.txt", content: "readme"},
	})

	dest := filepath.Join(t.TempDir(), "jdk")
	if err := file.Extract(archive, dest); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "readme.txt"), "readme")
}

func TestExtractZipRejectsPathTraversal(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "jdk.zip")
	writeZip(t, archive, []archiveEntry{
		{name: "jdk-17/../../evil.txt", content: "evil"},
	})

	dest := filepath.Join(t.TempDir(), "jdk")
	err := file.Extract(archive, dest)
	if err == nil {
		t.Fatal("Extract() error = nil, want path traversal error")
	}
	assertNotExist(t, filepath.Join(filepath.Dir(filepath.Dir(dest)), "evil.txt"))
}

func TestExtractZipRejectsFileNamedLikeRoot(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "jdk.zip")
	writeZip(t, archive, []archiveEntry{
		{name: "jdk-17", content: "file"},
		{name: "jdk-17/bin/java.exe", content: "java"},
	})

	err := file.Extract(archive, filepath.Join(t.TempDir(), "jdk"))
	if err == nil {
		t.Fatal("Extract() error = nil, want invalid file path error")
	}
}

func writeZip(t *testing.T, path string, entries []archiveEntry) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}

	w := zip.NewWriter(f)
	for _, entry := range entries {
		fw, err := w.Create(entry.name)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", entry.name, err)
		}
		if _, err := fw.Write([]byte(entry.content)); err != nil {
			t.Fatalf("write zip entry %s: %v", entry.name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}
}

func writeTarGz(t *testing.T, path string, entries []archiveEntry) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create tarball: %v", err)
	}

	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for _, entry := range entries {
		header := &tar.Header{
			Name:     entry.name,
			Mode:     0644,
			Size:     int64(len(entry.content)),
			Typeflag: tar.TypeReg,
		}
		switch {
		case entry.typeflag == tar.TypeXGlobalHeader:
			// The writer only accepts PAX records for global headers.
			header = &tar.Header{
				Name:       entry.name,
				Typeflag:   tar.TypeXGlobalHeader,
				PAXRecords: map[string]string{},
			}
		case strings.HasSuffix(entry.name, "/"):
			header.Typeflag = tar.TypeDir
			header.Mode = 0755
			header.Size = 0
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("write tar header %s: %v", entry.name, err)
		}
		if _, err := tw.Write([]byte(entry.content)); err != nil {
			t.Fatalf("write tar entry %s: %v", entry.name, err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close tarball: %v", err)
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != want {
		t.Fatalf("content of %s = %q, want %q", path, got, want)
	}
}

func assertNotExist(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s not to exist, stat error = %v", path, err)
	}
}

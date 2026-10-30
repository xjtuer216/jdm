package jdk

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestBuildDownloadSources(t *testing.T) {
	const (
		mirror   = "https://mirror.example.com/Adoptium"
		proxy    = "https://ghproxy.net"
		fileName = "OpenJDK17U-jdk_x64_windows_hotspot_17.0.12_7.zip"
		github   = "https://github.com/adoptium/temurin17-binaries/releases/download/jdk-17.0.12_7/" + fileName
	)

	tests := []struct {
		name        string
		mirror      string
		major       int
		fileName    string
		proxy       string
		originalURL string
		want        []DownloadSource
	}{
		{
			name:        "all sources in priority order",
			mirror:      mirror,
			major:       17,
			fileName:    fileName,
			proxy:       proxy,
			originalURL: github,
			want: []DownloadSource{
				{Name: "jdk_mirror", URL: "https://mirror.example.com/Adoptium/17/jdk/x64/windows/" + fileName},
				{Name: "proxy", URL: proxy + "/" + github},
				{Name: "github", URL: github},
			},
		},
		{
			name:        "empty mirror skips mirror",
			mirror:      "",
			major:       17,
			fileName:    fileName,
			proxy:       proxy,
			originalURL: github,
			want: []DownloadSource{
				{Name: "proxy", URL: proxy + "/" + github},
				{Name: "github", URL: github},
			},
		},
		{
			name:        "empty mirror and proxy fall back to github",
			mirror:      "",
			major:       17,
			fileName:    fileName,
			proxy:       "",
			originalURL: github,
			want:        []DownloadSource{{Name: "github", URL: github}},
		},
		{
			name:        "proxy only applies to github urls",
			mirror:      mirror,
			major:       21,
			fileName:    fileName,
			proxy:       proxy,
			originalURL: "https://download.example.com/" + fileName,
			want: []DownloadSource{
				{Name: "jdk_mirror", URL: "https://mirror.example.com/Adoptium/21/jdk/x64/windows/" + fileName},
				{Name: "github", URL: "https://download.example.com/" + fileName},
			},
		},
		{
			name:        "empty file name skips mirror layout",
			mirror:      mirror,
			major:       17,
			fileName:    "",
			proxy:       "",
			originalURL: github,
			want:        []DownloadSource{{Name: "github", URL: github}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildDownloadSources(tt.mirror, tt.major, tt.fileName, tt.proxy, tt.originalURL)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d sources, got %d: %+v", len(tt.want), len(got), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("source %d = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestDownloader_FallsBackAndVerifies(t *testing.T) {
	payload := []byte("PK\x03\x04jdm-test-jdk-archive-payload")
	corrupt := append([]byte(nil), payload...)
	corrupt[len(corrupt)-1] ^= 0xff
	goodChecksum := fmt.Sprintf("%x", sha256.Sum256(payload))

	writePayload := func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(payload)
	}
	writeHTML := func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<html><body>not found</body></html>"))
	}
	writeStatus := func(code int) func(http.ResponseWriter) {
		return func(w http.ResponseWriter) {
			w.WriteHeader(code)
		}
	}

	tests := []struct {
		name       string
		mirror     func(http.ResponseWriter)
		checksum   string
		wantSource string
	}{
		{
			name:       "mirror succeeds",
			mirror:     writePayload,
			checksum:   goodChecksum,
			wantSource: "jdk_mirror",
		},
		{
			name:       "mirror 404 falls back to proxy",
			mirror:     writeStatus(http.StatusNotFound),
			checksum:   goodChecksum,
			wantSource: "proxy",
		},
		{
			name:       "mirror 500 falls back to proxy",
			mirror:     writeStatus(http.StatusInternalServerError),
			checksum:   goodChecksum,
			wantSource: "proxy",
		},
		{
			name:       "html error page falls back to proxy",
			mirror:     writeHTML,
			checksum:   goodChecksum,
			wantSource: "proxy",
		},
		{
			name: "truncated response falls back to proxy",
			mirror: func(w http.ResponseWriter) {
				w.Header().Set("Content-Type", "application/zip")
				w.Write(payload[:8])
			},
			checksum:   goodChecksum,
			wantSource: "proxy",
		},
		{
			name: "checksum mismatch falls back to proxy",
			mirror: func(w http.ResponseWriter) {
				w.Header().Set("Content-Type", "application/zip")
				w.Write(corrupt)
			},
			checksum:   goodChecksum,
			wantSource: "proxy",
		},
		{
			name: "missing checksum still verifies size",
			mirror: func(w http.ResponseWriter) {
				w.Header().Set("Content-Type", "application/zip")
				w.Write(payload[:8])
			},
			checksum:   "",
			wantSource: "proxy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var userAgents []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userAgents = append(userAgents, r.Header.Get("User-Agent"))
				if r.URL.Path == "/mirror" {
					tt.mirror(w)
					return
				}
				writePayload(w)
			}))
			defer srv.Close()

			sources := []DownloadSource{
				{Name: "jdk_mirror", URL: srv.URL + "/mirror"},
				{Name: "proxy", URL: srv.URL + "/proxy"},
				{Name: "github", URL: srv.URL + "/github"},
			}
			d := newDownloader(srv.Client(), "JDM/test")

			path, used, err := d.Download(sources, int64(len(payload)), tt.checksum)
			if err != nil {
				t.Fatalf("Download failed: %v", err)
			}
			defer os.Remove(path)

			if used.Name != tt.wantSource {
				t.Errorf("expected source %q, got %q (%s)", tt.wantSource, used.Name, used.URL)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read downloaded file: %v", err)
			}
			if string(data) != string(payload) {
				t.Errorf("downloaded content does not match payload")
			}

			if len(userAgents) == 0 {
				t.Fatal("expected the downloader to send requests")
			}
			for _, ua := range userAgents {
				if ua != "JDM/test" {
					t.Errorf("expected User-Agent %q, got %q", "JDM/test", ua)
				}
			}
		})
	}
}

func TestDownloader_AggregatesAllSourceFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	sources := []DownloadSource{
		{Name: "jdk_mirror", URL: srv.URL + "/mirror"},
		{Name: "proxy", URL: srv.URL + "/proxy"},
		{Name: "github", URL: srv.URL + "/github"},
	}
	d := newDownloader(srv.Client(), "JDM/test")

	_, _, err := d.Download(sources, 10, "")
	if err == nil {
		t.Fatal("expected all download sources to fail")
	}
	for _, want := range []string{"jdk_mirror", "proxy", "github", "404"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("expected error to mention %q, got %v", want, err)
		}
	}
}

func TestDownloader_CleansUpTempFilesAfterVerificationFailure(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("TMP", tmpDir)
	t.Setenv("TEMP", tmpDir)
	t.Setenv("TMPDIR", tmpDir)

	payload := []byte("PK\x03\x04partial-download")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(payload)
	}))
	defer srv.Close()

	sources := []DownloadSource{{Name: "jdk_mirror", URL: srv.URL + "/mirror"}}
	d := newDownloader(srv.Client(), "JDM/test")

	_, _, err := d.Download(sources, int64(len(payload))+100, "")
	if err == nil {
		t.Fatal("expected the size mismatch to fail the download")
	}

	entries, readErr := os.ReadDir(tmpDir)
	if readErr != nil {
		t.Fatalf("failed to inspect temp dir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Errorf("expected failed downloads to clean up temp files, found %d", len(entries))
	}
}

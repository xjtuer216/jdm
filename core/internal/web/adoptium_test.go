package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseVersionNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"17", 17},
		{"17.0.2", 17},
		{"21", 21},
		{"8", 8},
	}

	for _, tt := range tests {
		result := parseVersionNumber(tt.input)
		if result != tt.expected {
			t.Errorf("parseVersionNumber(%s) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}

func TestAdoptiumClient_ParsesChecksumAndSendsUserAgent(t *testing.T) {
	const userAgent = "JDM/test"
	var userAgents []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgents = append(userAgents, r.Header.Get("User-Agent"))
		switch {
		case strings.HasSuffix(r.URL.Path, "/available_releases"):
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"available_releases":[17],"available_lts_releases":[17],"most_recent_lts":17,"most_recent_feature_version":21}`))
		case strings.HasSuffix(r.URL.Path, "/ga"):
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{
				"release_name": "jdk-17.0.12+7",
				"vendor": "eclipse",
				"version_data": {"semver": "17.0.12+7"},
				"binaries": [{
					"architecture": "x64",
					"os": "windows",
					"package": {
						"link": "https://github.com/adoptium/temurin17-binaries/releases/download/jdk-17.0.12%2B7/OpenJDK17U-jdk_x64_windows_hotspot_17.0.12_7.zip",
						"name": "OpenJDK17U-jdk_x64_windows_hotspot_17.0.12_7.zip",
						"size": 190000000,
						"checksum": "6f1c6d1f1f5f2f2d9d6f7c1a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e"
					}
				}]
			}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewAdoptiumClient(srv.URL, userAgent)
	versions, err := client.FetchVersions(17)
	if err != nil {
		t.Fatalf("FetchVersions failed: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}

	got := versions[0]
	const wantChecksum = "6f1c6d1f1f5f2f2d9d6f7c1a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e"
	if got.Checksum != wantChecksum {
		t.Errorf("expected checksum %q, got %q", wantChecksum, got.Checksum)
	}
	if got.FileSize != 190000000 {
		t.Errorf("expected file size 190000000, got %d", got.FileSize)
	}
	if got.FileName != "OpenJDK17U-jdk_x64_windows_hotspot_17.0.12_7.zip" {
		t.Errorf("unexpected file name %q", got.FileName)
	}

	if len(userAgents) == 0 {
		t.Fatal("expected the client to send requests")
	}
	for _, ua := range userAgents {
		if ua != userAgent {
			t.Errorf("expected User-Agent %q, got %q", userAgent, ua)
		}
	}
}

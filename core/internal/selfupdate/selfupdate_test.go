package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xjtuer216/jdm/internal/jdk"
)

// setVersion 临时替换本地版本号，测试结束后恢复
func setVersion(t *testing.T, v string) {
	t.Helper()
	old := jdk.Version
	jdk.Version = v
	t.Cleanup(func() { jdk.Version = old })
}

// fakeRelease 构造 release 元数据 JSON 响应体
func fakeRelease(t *testing.T, tag string, assets map[string]string, exeSize int64) string {
	t.Helper()
	type asset struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	}
	list := make([]asset, 0, len(assets))
	for name, url := range assets {
		list = append(list, asset{Name: name, BrowserDownloadURL: url, Size: exeSize})
	}
	b, err := json.Marshal(map[string]any{"tag_name": tag, "assets": list})
	if err != nil {
		t.Fatalf("marshal release: %v", err)
	}
	return string(b)
}

func TestCheck_ParsesReleaseAndComparesVersions(t *testing.T) {
	setVersion(t, "0.1.0")

	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		fmt.Fprint(w, fakeRelease(t, "v0.2.0", nil, 0))
	}))
	defer srv.Close()

	giteeReleaseURL = srv.URL
	githubReleaseURL = srv.URL + "/unused"
	t.Cleanup(restoreEndpoints)

	res, err := Check()
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if res.LatestVersion != "0.2.0" {
		t.Errorf("LatestVersion = %q, want %q", res.LatestVersion, "0.2.0")
	}
	if !res.HasUpdate {
		t.Error("HasUpdate = false, want true for 0.1.0 -> 0.2.0")
	}
	if gotUA != jdk.UserAgent() {
		t.Errorf("User-Agent = %q, want %q", gotUA, jdk.UserAgent())
	}
}

func TestCheck_NoUpdateWhenLatestEqualsCurrent(t *testing.T) {
	setVersion(t, "0.2.0")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakeRelease(t, "v0.2.0", nil, 0))
	}))
	defer srv.Close()

	giteeReleaseURL = srv.URL
	githubReleaseURL = srv.URL + "/unused"
	t.Cleanup(restoreEndpoints)

	res, err := Check()
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if res.HasUpdate {
		t.Error("HasUpdate = true, want false for 0.2.0 -> 0.2.0")
	}
}

func TestCheck_FallsBackToSecondSource(t *testing.T) {
	setVersion(t, "0.1.0")

	gitee := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer gitee.Close()

	var githubHit bool
	github := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		githubHit = true
		fmt.Fprint(w, fakeRelease(t, "v0.2.0", nil, 0))
	}))
	defer github.Close()

	giteeReleaseURL = gitee.URL
	githubReleaseURL = github.URL
	t.Cleanup(restoreEndpoints)

	res, err := Check()
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !githubHit {
		t.Error("fallback source was not queried")
	}
	if res.LatestVersion != "0.2.0" || !res.HasUpdate {
		t.Errorf("Check() = %+v, want latest 0.2.0 with update", res)
	}
}

func TestCheck_ErrorsWhenAllSourcesFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	giteeReleaseURL = srv.URL
	githubReleaseURL = srv.URL
	t.Cleanup(restoreEndpoints)

	_, err := Check()
	if err == nil {
		t.Fatal("Check() error = nil, want aggregated failure")
	}
	if !strings.Contains(err.Error(), "Gitee") || !strings.Contains(err.Error(), "GitHub") {
		t.Errorf("error should aggregate per-source attempts, got: %v", err)
	}
}

func TestApply_ReplacesBinary(t *testing.T) {
	setVersion(t, "0.1.0")

	exePath := tempExe(t, "old-bytes")
	exePathFn = func() (string, error) { return exePath, nil }
	t.Cleanup(restoreEndpoints)

	exeContent := []byte("new-exe-bytes-0123456789")
	sum := sha256.Sum256(exeContent)
	hexSum := hex.EncodeToString(sum[:])

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			fmt.Fprint(w, fakeRelease(t, "v0.2.0", map[string]string{
				assetExeName:       srv.URL + "/jdm.exe",
				assetChecksumsName: srv.URL + "/checksums.txt",
			}, int64(len(exeContent))))
		case "/checksums.txt":
			// sha256sum 格式：多行、双空格分隔、容忍 CRLF 行尾
			fmt.Fprintf(w, "deadbeef  some-other-file\r\n%s  %s\r\n", hexSum, assetExeName)
		case "/jdm.exe":
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(exeContent)
		default:
			t.Errorf("unexpected request path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	giteeReleaseURL = srv.URL + "/releases/latest"
	githubReleaseURL = srv.URL + "/unused"
	t.Cleanup(restoreEndpoints)

	res, err := Apply()
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if res.UpToDate {
		t.Error("UpToDate = true, want false")
	}
	if res.Version != "0.2.0" {
		t.Errorf("Version = %q, want %q", res.Version, "0.2.0")
	}
	got, err := os.ReadFile(exePath)
	if err != nil || string(got) != string(exeContent) {
		t.Errorf("new exe content = %q (err %v), want %q", got, err, exeContent)
	}
	backup, err := os.ReadFile(exePath + ".old")
	if err != nil || string(backup) != "old-bytes" {
		t.Errorf("backup content = %q (err %v), want old-bytes", backup, err)
	}
}

// tempExe 在临时目录写一个假的当前可执行文件，返回其路径
func tempExe(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "jdm.exe")
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write temp exe: %v", err)
	}
	return path
}

func TestApply_ChecksumMismatchKeepsOriginal(t *testing.T) {
	setVersion(t, "0.1.0")

	exePath := tempExe(t, "old-bytes")
	exePathFn = func() (string, error) { return exePath, nil }
	t.Cleanup(restoreEndpoints)

	exeContent := []byte("tampered-or-corrupted-payload")

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			fmt.Fprint(w, fakeRelease(t, "v0.2.0", map[string]string{
				assetExeName:       srv.URL + "/jdm.exe",
				assetChecksumsName: srv.URL + "/checksums.txt",
			}, int64(len(exeContent))))
		case "/checksums.txt":
			// 清单声明的是正确摘要，但服务器返回的载荷与之不符
			fmt.Fprintf(w, "%s  %s\n", strings.Repeat("0", 64), assetExeName)
		case "/jdm.exe":
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(exeContent)
		case "/unused":
			w.WriteHeader(http.StatusNotFound)
		default:
			t.Errorf("unexpected request path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	giteeReleaseURL = srv.URL + "/releases/latest"
	githubReleaseURL = srv.URL + "/unused"
	t.Cleanup(restoreEndpoints)

	if _, err := Apply(); err == nil {
		t.Fatal("Apply() error = nil, want checksum mismatch rejection")
	}
	got, err := os.ReadFile(exePath)
	if err != nil || string(got) != "old-bytes" {
		t.Errorf("original exe content = %q (err %v), want unchanged old-bytes", got, err)
	}
	if _, err := os.Stat(exePath + ".old"); !os.IsNotExist(err) {
		t.Errorf("backup should not exist, stat err = %v", err)
	}
}

func TestApply_MissingChecksumsVerifiesSizeOnly(t *testing.T) {
	setVersion(t, "0.1.0")

	exePath := tempExe(t, "old-bytes")
	exePathFn = func() (string, error) { return exePath, nil }
	t.Cleanup(restoreEndpoints)

	exeContent := []byte("new-exe-bytes-without-checksums")
	declaredSize := int64(len(exeContent))

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			// release 只有 exe 资产，无 checksums.txt
			fmt.Fprint(w, fakeRelease(t, "v0.2.0", map[string]string{
				assetExeName: srv.URL + "/jdm.exe",
			}, declaredSize))
		case "/jdm.exe":
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(exeContent)
		case "/unused":
			w.WriteHeader(http.StatusNotFound)
		default:
			t.Errorf("unexpected request path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	giteeReleaseURL = srv.URL + "/releases/latest"
	githubReleaseURL = srv.URL + "/unused"
	t.Cleanup(restoreEndpoints)

	res, err := Apply()
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if res.Version != "0.2.0" || res.UpToDate {
		t.Errorf("Apply() = %+v, want replaced 0.2.0", res)
	}
	if got, err := os.ReadFile(exePath); err != nil || string(got) != string(exeContent) {
		t.Errorf("exe content = %q (err %v), want %q", got, err, exeContent)
	}
}

func TestApply_SizeMismatchKeepsOriginal(t *testing.T) {
	setVersion(t, "0.1.0")

	exePath := tempExe(t, "old-bytes")
	exePathFn = func() (string, error) { return exePath, nil }
	t.Cleanup(restoreEndpoints)

	exeContent := []byte("truncated-payload")
	declaredSize := int64(len(exeContent) + 100)

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			fmt.Fprint(w, fakeRelease(t, "v0.2.0", map[string]string{
				assetExeName: srv.URL + "/jdm.exe",
			}, declaredSize))
		case "/jdm.exe":
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(exeContent)
		case "/unused":
			w.WriteHeader(http.StatusNotFound)
		default:
			t.Errorf("unexpected request path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	giteeReleaseURL = srv.URL + "/releases/latest"
	githubReleaseURL = srv.URL + "/unused"
	t.Cleanup(restoreEndpoints)

	if _, err := Apply(); err == nil {
		t.Fatal("Apply() error = nil, want size mismatch rejection")
	}
	if got, err := os.ReadFile(exePath); err != nil || string(got) != "old-bytes" {
		t.Errorf("original exe content = %q (err %v), want unchanged old-bytes", got, err)
	}
}

func TestApply_UpToDateSkipsDownload(t *testing.T) {
	setVersion(t, "0.2.0")

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			fmt.Fprint(w, fakeRelease(t, "v0.2.0", map[string]string{
				assetExeName:       srv.URL + "/jdm.exe",
				assetChecksumsName: srv.URL + "/checksums.txt",
			}, 0))
		case "/unused":
			w.WriteHeader(http.StatusNotFound)
		default:
			t.Errorf("up-to-date Apply must not download, got request %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	giteeReleaseURL = srv.URL + "/releases/latest"
	githubReleaseURL = srv.URL + "/unused"
	t.Cleanup(restoreEndpoints)

	res, err := Apply()
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !res.UpToDate || res.Version != "0.2.0" {
		t.Errorf("Apply() = %+v, want up-to-date 0.2.0", res)
	}
}

func TestApply_FallsBackWhenDownloadFails(t *testing.T) {
	setVersion(t, "0.1.0")

	exePath := tempExe(t, "old-bytes")
	exePathFn = func() (string, error) { return exePath, nil }
	t.Cleanup(restoreEndpoints)

	exeContent := []byte("exe-from-github-fallback")
	sum := sha256.Sum256(exeContent)
	hexSum := hex.EncodeToString(sum[:])

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		// Gitee 元数据可达，但其 exe 资产下载失败
		case "/gitee/releases/latest":
			fmt.Fprint(w, fakeRelease(t, "v0.2.0", map[string]string{
				assetExeName:       srv.URL + "/gitee/jdm.exe",
				assetChecksumsName: srv.URL + "/gitee/checksums.txt",
			}, int64(len(exeContent))))
		case "/gitee/checksums.txt":
			fmt.Fprintf(w, "%s  %s\n", hexSum, assetExeName)
		case "/gitee/jdm.exe":
			w.WriteHeader(http.StatusInternalServerError)
		// GitHub 元数据与资产均可用
		case "/github/releases/latest":
			fmt.Fprint(w, fakeRelease(t, "v0.2.0", map[string]string{
				assetExeName:       srv.URL + "/github/jdm.exe",
				assetChecksumsName: srv.URL + "/github/checksums.txt",
			}, int64(len(exeContent))))
		case "/github/checksums.txt":
			fmt.Fprintf(w, "%s  %s\n", hexSum, assetExeName)
		case "/github/jdm.exe":
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(exeContent)
		default:
			t.Errorf("unexpected request path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	giteeReleaseURL = srv.URL + "/gitee/releases/latest"
	githubReleaseURL = srv.URL + "/github/releases/latest"
	t.Cleanup(restoreEndpoints)

	res, err := Apply()
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if res.Version != "0.2.0" || res.UpToDate {
		t.Errorf("Apply() = %+v, want replaced 0.2.0 via fallback", res)
	}
	if got, err := os.ReadFile(exePath); err != nil || string(got) != string(exeContent) {
		t.Errorf("exe content = %q (err %v), want %q", got, err, exeContent)
	}
}

func TestCleanupBackup_RemovesStaleBackup(t *testing.T) {
	exePath := tempExe(t, "current")
	exePathFn = func() (string, error) { return exePath, nil }
	t.Cleanup(restoreEndpoints)
	if err := os.WriteFile(exePath+".old", []byte("stale"), 0o755); err != nil {
		t.Fatalf("write backup: %v", err)
	}

	CleanupBackup()

	if _, err := os.Stat(exePath + ".old"); !os.IsNotExist(err) {
		t.Errorf("backup should be removed, stat err = %v", err)
	}
	if got, err := os.ReadFile(exePath); err != nil || string(got) != "current" {
		t.Errorf("current exe content = %q (err %v), want current", got, err)
	}
}

func TestCleanupBackup_NoOpWithoutBackup(t *testing.T) {
	exePath := tempExe(t, "current")
	exePathFn = func() (string, error) { return exePath, nil }
	t.Cleanup(restoreEndpoints)

	CleanupBackup()

	if got, err := os.ReadFile(exePath); err != nil || string(got) != "current" {
		t.Errorf("current exe content = %q (err %v), want current", got, err)
	}
}

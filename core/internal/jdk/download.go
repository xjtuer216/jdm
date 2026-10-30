package jdk

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/xjtuer216/jdm/internal/progress"
)

// 下载源镜像布局中固定的产物类型、架构与操作系统
const (
	mirrorImageType = "jdk"
	mirrorArch      = "x64"
	mirrorOS        = "windows"
)

// DownloadSource 是一个可尝试的下载来源，Name 用于成功提示与失败汇总
type DownloadSource struct {
	Name string
	URL  string
}

// BuildDownloadSources 按优先级构造下载来源：下载源 → 代理 → GitHub 直连
// 入参是配置层给出的生效值（none 与空值已归一化为空字符串）
// 下载源按 {base}/{major}/jdk/x64/windows/{fileName} 拼出地址，fileName 逐字取自元数据
func BuildDownloadSources(mirrorBase string, major int, fileName, proxy, originalURL string) []DownloadSource {
	sources := make([]DownloadSource, 0, 3)

	if mirrorBase != "" && fileName != "" {
		sources = append(sources, DownloadSource{
			Name: "jdk_mirror",
			URL:  fmt.Sprintf("%s/%d/%s/%s/%s/%s", mirrorBase, major, mirrorImageType, mirrorArch, mirrorOS, fileName),
		})
	}
	if proxy != "" && originalURL != "" && strings.Contains(originalURL, "github.com") {
		sources = append(sources, DownloadSource{
			Name: "proxy",
			URL:  proxy + "/" + originalURL,
		})
	}
	if originalURL != "" {
		sources = append(sources, DownloadSource{Name: "github", URL: originalURL})
	}

	return sources
}

// downloader 负责按候选链下载并用元数据校验完整性
type downloader struct {
	client    *http.Client
	userAgent string
}

func newDownloader(client *http.Client, userAgent string) *downloader {
	return &downloader{client: client, userAgent: userAgent}
}

// downloadAttempt 记录单个来源的失败原因
type downloadAttempt struct {
	Source DownloadSource
	Err    error
}

// DownloadError 汇总所有下载来源的失败原因
type DownloadError struct {
	Attempts []downloadAttempt
}

func (e *DownloadError) Error() string {
	var b strings.Builder
	b.WriteString("all download sources failed")
	for _, attempt := range e.Attempts {
		fmt.Fprintf(&b, "\n  %s (%s): %v", attempt.Source.Name, attempt.Source.URL, attempt.Err)
	}
	return b.String()
}

// Download 依次尝试候选来源，返回临时文件路径与实际使用的来源
// 每个来源只尝试一次；失败或校验不符时删除临时文件
func (d *downloader) Download(sources []DownloadSource, expectedSize int64, checksum string) (string, DownloadSource, error) {
	if len(sources) == 0 {
		return "", DownloadSource{}, errors.New("no download sources available")
	}
	if checksum == "" {
		logrus.Warn("SHA-256 checksum unavailable; skipping checksum verification")
	}

	attempts := make([]downloadAttempt, 0, len(sources))
	for _, source := range sources {
		path, err := d.fetch(source, expectedSize, checksum)
		if err == nil {
			return path, source, nil
		}
		attempts = append(attempts, downloadAttempt{Source: source, Err: err})
	}

	return "", DownloadSource{}, &DownloadError{Attempts: attempts}
}

// fetch 从单个来源下载到临时文件，并校验响应类型、字节数与 SHA-256
func (d *downloader) fetch(source DownloadSource, expectedSize int64, checksum string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, source.URL, nil)
	if err != nil {
		return "", err
	}
	if d.userAgent != "" {
		req.Header.Set("User-Agent", d.userAgent)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %s", resp.Status)
	}
	if strings.HasPrefix(strings.ToLower(resp.Header.Get("Content-Type")), "text/html") {
		return "", errors.New("response is HTML, not a JDK archive")
	}

	return d.save(resp.Body, expectedSize, checksum)
}

// save 把响应体写入临时文件并校验字节数与 SHA-256，失败时删除临时文件
func (d *downloader) save(body io.Reader, expectedSize int64, checksum string) (string, error) {
	tmpFile, err := os.CreateTemp("", "jdm-*.zip")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()
	success := false
	defer func() {
		if !success {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	bar := progress.New("Downloading", expectedSize)
	defer bar.Done()

	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(tmpFile, hasher), &progressReader{reader: body, bar: bar})
	if err != nil {
		return "", err
	}
	if expectedSize > 0 && written != expectedSize {
		return "", fmt.Errorf("size mismatch: got %d bytes, expected %d", written, expectedSize)
	}
	if checksum != "" {
		actual := hex.EncodeToString(hasher.Sum(nil))
		if !strings.EqualFold(actual, checksum) {
			return "", fmt.Errorf("SHA-256 mismatch: got %s, expected %s", actual, checksum)
		}
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to finalize download: %w", err)
	}

	success = true
	return tmpPath, nil
}

// progressReader wraps an io.Reader and updates a progress bar on each Read.
type progressReader struct {
	reader io.Reader
	bar    *progress.ProgressBar
	total  int64
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	if n > 0 {
		pr.total += int64(n)
		pr.bar.Update(pr.total)
	}
	return
}

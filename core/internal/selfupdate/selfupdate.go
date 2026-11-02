// Package selfupdate 实现 JDM 自升级：Check() 供版本探针与 jdm update
// 共用，Apply() 执行下载、校验与替换的完整升级动作。
//
// 更新来源候选链：Gitee Releases API 为主、GitHub 镜像仓库回退，每个来源
// 只尝试一次（与 ADR 0001 的候选链精神一致）。更新来源是独立通道，与
// 元数据源、下载源、代理三个配置概念无关；仓库地址为编译期常量，不进配置层。
//
// 发布资产命名契约：release 必须附带名为 jdm-windows-amd64.exe 的 Windows
// x64 可执行文件与名为 checksums.txt 的摘要清单（sha256sum 格式，每行
// "<sha256>  <文件名>"）。Release API 不返回资产摘要，摘要一律以
// checksums.txt 为准；缺失时告警并跳过哈希校验、仍校验字节数（ADR 0001 先例）。
package selfupdate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xjtuer216/jdm/internal/jdk"
	"github.com/xjtuer216/jdm/internal/web"
)

// 发布资产的固定命名契约
const (
	assetExeName       = "jdm-windows-amd64.exe"
	assetChecksumsName = "checksums.txt"
)

// 更新来源的编译期常量地址（测试通过同名的包级变量覆写注入）
const (
	defaultGiteeReleaseURL  = "https://gitee.com/api/v5/repos/xjtuer216/jdm/releases/latest"
	defaultGitHubReleaseURL = "https://api.github.com/repos/xjtuer216/jdm/releases/latest"
)

var (
	giteeReleaseURL  = defaultGiteeReleaseURL
	githubReleaseURL = defaultGitHubReleaseURL
	exePathFn        = defaultExePathFn
)

func defaultExePathFn() (string, error) { return os.Executable() }

// restoreEndpoints 供测试恢复默认注入值
func restoreEndpoints() {
	giteeReleaseURL = defaultGiteeReleaseURL
	githubReleaseURL = defaultGitHubReleaseURL
	exePathFn = defaultExePathFn
}

var metadataClient = &http.Client{Timeout: 30 * time.Second}

// updateSource 是更新候选链上的一个来源
type updateSource struct {
	Name       string
	ReleaseURL string
}

// updateChain 按优先级返回候选来源，每次调用读取当前端点值以便测试注入
func updateChain() []updateSource {
	return []updateSource{
		{Name: "Gitee", ReleaseURL: giteeReleaseURL},
		{Name: "GitHub", ReleaseURL: githubReleaseURL},
	}
}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type releaseInfo struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

// CheckResult 是一次版本探测的结果
type CheckResult struct {
	LatestVersion string
	HasUpdate     bool
}

// Check 沿候选链获取最新 release 并与本地版本比较。
// 本地版本来自 -ldflags 注入；比较语义与 web.CompareVersions 一致。
func Check() (CheckResult, error) {
	rel, _, err := fetchLatestRelease()
	if err != nil {
		return CheckResult{}, err
	}
	latest := versionFromTag(rel.TagName)
	return CheckResult{
		LatestVersion: latest,
		HasUpdate:     web.CompareVersions(latest, jdk.Version) > 0,
	}, nil
}

// fetchLatestRelease 依次尝试候选来源的 release 元数据，每个来源只尝试一次
func fetchLatestRelease() (releaseInfo, updateSource, error) {
	var attempts []string
	for _, src := range updateChain() {
		rel, err := fetchRelease(src)
		if err == nil {
			return rel, src, nil
		}
		attempts = append(attempts, fmt.Sprintf("  %s (%s): %v", src.Name, src.ReleaseURL, err))
	}
	return releaseInfo{}, updateSource{}, fmt.Errorf("all update sources failed\n%s", strings.Join(attempts, "\n"))
}

func fetchRelease(src updateSource) (releaseInfo, error) {
	body, err := httpGetBody(src.ReleaseURL)
	if err != nil {
		return releaseInfo{}, err
	}
	var rel releaseInfo
	if err := json.Unmarshal(body, &rel); err != nil {
		return releaseInfo{}, fmt.Errorf("invalid release metadata: %w", err)
	}
	if rel.TagName == "" {
		return releaseInfo{}, fmt.Errorf("release metadata has no tag_name")
	}
	return rel, nil
}

// httpGetBody 发起带 User-Agent 的 GET 请求并读取响应体（上限 1 MiB），
// 非 2xx 视为失败
func httpGetBody(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", jdk.UserAgent())

	resp, err := metadataClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %s", resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 1<<20))
}

func versionFromTag(tag string) string {
	return strings.TrimPrefix(tag, "v")
}

func findAsset(rel releaseInfo, name string) (releaseAsset, bool) {
	for _, asset := range rel.Assets {
		if asset.Name == name {
			return asset, true
		}
	}
	return releaseAsset{}, false
}

// ApplyResult 是一次升级动作的结果
type ApplyResult struct {
	// Version 是候选链返回的最新版本号
	Version string
	// UpToDate 为 true 时表示本地已是最新版本，未发生任何下载
	UpToDate bool
}

// applyAttempt 记录单个更新来源的失败原因
type applyAttempt struct {
	Source string
	Err    error
}

// ApplyError 汇总所有更新来源的失败原因
type ApplyError struct {
	Attempts []applyAttempt
}

func (e *ApplyError) Error() string {
	var b strings.Builder
	b.WriteString("self-update failed on all sources")
	for _, attempt := range e.Attempts {
		fmt.Fprintf(&b, "\n  %s: %v", attempt.Source, attempt.Err)
	}
	return b.String()
}

// Apply 沿候选链执行完整的升级动作：获取 release 元数据 → 解析
// checksums.txt → 下载并校验新 exe → 替换运行中的可执行文件。
// 本地已是最新版本时不做下载，返回 UpToDate=true。
// 替换成功后旧程序保留在 <exe>.old，下次启动由 CleanupBackup 清理。
func Apply() (ApplyResult, error) {
	var attempts []applyAttempt
	for _, src := range updateChain() {
		res, err := applyFromSource(src)
		if err == nil {
			return res, nil
		}
		attempts = append(attempts, applyAttempt{Source: src.Name, Err: err})
	}
	return ApplyResult{}, &ApplyError{Attempts: attempts}
}

// applyFromSource 对单个来源完成升级，任何一步失败即交给候选链下一个来源
func applyFromSource(src updateSource) (ApplyResult, error) {
	rel, err := fetchRelease(src)
	if err != nil {
		return ApplyResult{}, err
	}
	latest := versionFromTag(rel.TagName)
	if web.CompareVersions(latest, jdk.Version) <= 0 {
		return ApplyResult{Version: latest, UpToDate: true}, nil
	}

	exeAsset, ok := findAsset(rel, assetExeName)
	if !ok {
		return ApplyResult{}, fmt.Errorf("release %s has no %s asset", rel.TagName, assetExeName)
	}

	checksum, err := checksumFor(rel, assetExeName)
	if err != nil {
		logrus.Warnf("self-update: %v", err)
	}

	tmpPath, err := jdk.DownloadBinary(
		[]jdk.DownloadSource{{Name: src.Name, URL: exeAsset.BrowserDownloadURL}},
		exeAsset.Size, checksum)
	if err != nil {
		return ApplyResult{}, err
	}
	if err := replaceBinary(tmpPath); err != nil {
		os.Remove(tmpPath)
		return ApplyResult{}, err
	}
	return ApplyResult{Version: latest}, nil
}

// checksumFor 解析 release 附带的 checksums.txt，返回目标资产的摘要。
// 摘要清单或条目缺失时返回错误，由调用方告警后跳过哈希校验（仍校验字节数）。
func checksumFor(rel releaseInfo, assetName string) (string, error) {
	sumAsset, ok := findAsset(rel, assetChecksumsName)
	if !ok {
		return "", fmt.Errorf("release %s has no %s asset; skipping SHA-256 verification (byte count still verified)", rel.TagName, assetChecksumsName)
	}
	checksums, err := fetchChecksums(sumAsset.BrowserDownloadURL)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", assetChecksumsName, err)
	}
	checksum, ok := checksums[assetName]
	if !ok {
		return "", fmt.Errorf("%s has no entry for %s; skipping SHA-256 verification (byte count still verified)", assetChecksumsName, assetName)
	}
	return checksum, nil
}

// fetchChecksums 以 sha256sum 格式（每行 "<sha256>  <文件名>"）解析摘要清单
func fetchChecksums(url string) (map[string]string, error) {
	body, err := httpGetBody(url)
	if err != nil {
		return nil, err
	}

	checksums := make(map[string]string)
	for _, line := range strings.Split(string(body), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		checksums[fields[1]] = fields[0]
	}
	if len(checksums) == 0 {
		return nil, fmt.Errorf("no sha256sum entries found")
	}
	return checksums, nil
}

// replaceBinary 将运行中的可执行文件重命名为 .old 后在原路径写入新文件。
// 运行中的 exe 不可覆盖但可重命名；新文件写入失败时回滚恢复原文件。
func replaceBinary(tmpPath string) error {
	exePath, err := exePathFn()
	if err != nil {
		return fmt.Errorf("resolve current executable: %w", err)
	}
	backupPath := exePath + ".old"

	mode := os.FileMode(0o755)
	if info, err := os.Stat(backupPath); err == nil {
		mode = info.Mode().Perm()
	}
	if err := os.Rename(exePath, backupPath); err != nil {
		return fmt.Errorf("back up current executable: %w", err)
	}
	if err := installFile(tmpPath, exePath, mode); err != nil {
		if rollbackErr := os.Rename(backupPath, exePath); rollbackErr != nil {
			return fmt.Errorf("install new executable: %v (rollback failed: %v, current binary preserved at %s)", err, rollbackErr, backupPath)
		}
		return fmt.Errorf("install new executable: %w", err)
	}
	if err := os.Remove(tmpPath); err != nil {
		logrus.Debugf("selfupdate: remove temp file: %v", err)
	}
	return nil
}

// installFile 把 src 复制为 dst（跨卷安全），写入失败时清理半成品
func installFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(dst)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(dst)
		return err
	}
	return nil
}

// CleanupBackup 清理上次升级留下的旧版备份文件（<exe>.old）。
// 探针与升级均不依赖清理结果，失败只记调试日志，下次启动重试。
func CleanupBackup() {
	exePath, err := exePathFn()
	if err != nil {
		logrus.Debugf("selfupdate: resolve current executable: %v", err)
		return
	}
	if err := os.Remove(exePath + ".old"); err != nil && !errors.Is(err, os.ErrNotExist) {
		logrus.Debugf("selfupdate: remove old binary backup: %v", err)
	}
}

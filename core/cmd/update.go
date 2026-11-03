package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xjtuer216/jdm/internal/selfupdate"
)

// updateCmd 实现 jdm update：主动检查并升级到最新版本，不受冷却期限制
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates and upgrade JDM to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		res, err := selfupdate.Check()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if !res.HasUpdate {
			fmt.Printf("JDM is up to date (v%s)\n", res.LatestVersion)
			return
		}
		fmt.Printf("Upgrading JDM to v%s...\n", res.LatestVersion)
		applied, err := selfupdate.Apply()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if applied.UpToDate {
			// Check 与 Apply 之间版本状态发生变化（如另一进程已完成升级）
			fmt.Printf("JDM is up to date (v%s)\n", applied.Version)
			return
		}
		fmt.Printf("Upgraded to v%s. Restart jdm for the new version to take effect.\n", applied.Version)
	},
}

// probeCooldown 是版本探针的冷却期，时间戳记录在 JDM 主目录下
const probeCooldown = 24 * time.Hour

// probeNoticeWait 是命令输出结束后等待后台探测结果的宽限期，
// 保证提示大概率可见，同时探测本身不拖慢命令
const probeNoticeWait = 2 * time.Second

// probeResultCh 承载后台探测结果；空字符串表示无更新或探测失败
var probeResultCh chan string

// beforeCommand 在每次命令执行前清理旧版备份并发起版本探针（update 命令除外）
func beforeCommand(cmd *cobra.Command, args []string) {
	selfupdate.CleanupBackup()
	if cmd.Name() == "update" {
		return
	}
	startUpdateProbe()
}

// afterCommand 在命令输出结束后打印升级提示
func afterCommand(cmd *cobra.Command, args []string) {
	if cmd.Name() == "update" || probeResultCh == nil {
		return
	}
	select {
	case version := <-probeResultCh:
		if version != "" {
			fmt.Fprintf(os.Stderr, "\nJDM v%s is available. Run `jdm update` to upgrade.\n", version)
		}
	case <-time.After(probeNoticeWait):
	}
}

// startUpdateProbe 检查冷却时间戳后异步发起版本探测。
// 无更新、无网络或探测失败时零输出，正常命令不受干扰。
func startUpdateProbe() {
	jdmHome := defaultJDMHome()
	if !cooldownExpired(jdmHome) {
		return
	}
	writeCooldown(jdmHome)

	probeResultCh = make(chan string, 1)
	go func() {
		res, err := selfupdate.Check()
		if err != nil || !res.HasUpdate {
			probeResultCh <- ""
			return
		}
		probeResultCh <- res.LatestVersion
	}()
}

// cooldownPath 返回 JDM 主目录下的冷却时间戳文件路径
func cooldownPath(jdmHome string) string {
	return filepath.Join(jdmHome, "update_probe")
}

func cooldownExpired(jdmHome string) bool {
	data, err := os.ReadFile(cooldownPath(jdmHome))
	if err != nil {
		return true
	}
	last, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return true
	}
	return time.Since(time.Unix(last, 0)) >= probeCooldown
}

// writeCooldown 在探测发起前写入时间戳，避免探测期间的每条命令重复联网。
// 时间戳是冷却机制的前提，目录缺失时创建；失败只记调试日志。
func writeCooldown(jdmHome string) {
	if err := os.MkdirAll(jdmHome, 0o755); err != nil {
		logrus.Debugf("selfupdate: create JDM home for probe cooldown: %v", err)
		return
	}
	data := strconv.FormatInt(time.Now().Unix(), 10)
	if err := os.WriteFile(cooldownPath(jdmHome), []byte(data), 0o644); err != nil {
		logrus.Debugf("selfupdate: write probe cooldown: %v", err)
	}
}

func init() {
	rootCmd.AddCommand(updateCmd)
	rootCmd.PersistentPreRun = beforeCommand
	rootCmd.PersistentPostRun = afterCommand
}

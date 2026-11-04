package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xjtuer216/jdm/internal/jdk"
)

// 回归测试（IKSYMI）：`jdm -v` 应输出 "jdm v<version>" 并以单个换行结尾，
// 不允许出现字面的反斜杠+n（raw string 模板导致的旧 bug）。
func TestVersionFlagOutput(t *testing.T) {
	var buf bytes.Buffer
	oldOut := rootCmd.OutOrStdout()
	rootCmd.SetOut(&buf)
	defer rootCmd.SetOut(oldOut)

	rootCmd.SetArgs([]string{"-v"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute jdm -v: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, `\n`) {
		t.Fatalf("version output contains literal backslash-n: %q", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("version output should end with a newline: %q", out)
	}
	if want := "jdm v" + jdk.GetVersion() + "\n"; out != want {
		t.Fatalf("version output = %q, want %q", out, want)
	}
}

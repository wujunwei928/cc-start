// cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wujunwei928/cc-start/internal/config"
	"github.com/wujunwei928/cc-start/internal/repl"
)

var (
	// 版本信息
	Version = "dev"
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "cc-start",
	Short: "AI 编程助手启动器 - 快速切换不同供应商",
	Long: `CC-Start 是一个 AI 编程助手启动器，帮助你管理多个 API 供应商配置。

使用方法:
  cc-start               进入交互式 REPL
  cc-start claude        启动 Claude Code CLI
  cc-start codex         启动 OpenAI Codex CLI
  cc-start opencode      启动 OpenCode AI 编程助手
  cc-start setup         配置向导
  cc-start list          列出所有配置`,
	Version: Version,
	RunE:    runRoot,
}

func init() {
	rootCmd.SetVersionTemplate("cc-start {{.Version}}\n")
}

func runRoot(cmd *cobra.Command, args []string) error {
	// 无参数时进入 REPL
	cfgPath := config.GetConfigPath()

	r, err := repl.New(cfgPath)
	if err != nil {
		return err
	}

	r.Run()
	return nil
}

// findDashSeparator 查找 -- 分隔符在 os.Args 中的位置
func findDashSeparator(args []string) int {
	for i, arg := range args {
		if arg == "--" {
			return i
		}
	}
	return -1
}

// isFlag 判断字符串是否是 flag（以 - 开头）
func isFlag(s string) bool {
	return len(s) > 0 && s[0] == '-'
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		return
	}
}

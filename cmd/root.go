// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wujunwei/cc-start/internal/config"
	"github.com/wujunwei/cc-start/internal/launcher"
)

var (
	// 版本信息
	Version = "dev"
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "cc-start [profile] [-- claude args...]",
	Short: "Claude Code 启动器 - 快速切换不同供应商",
	Long: `CC-Start 是一个 Claude Code 启动器，帮助你管理多个 API 供应商配置。

使用方法:
  cc-start                          使用默认配置启动
  cc-start moonshot                 使用 moonshot 配置启动
  cc-start minimax -- --help        查看 claude 帮助
  cc-start -- --dangerously-skip-permissions   传递参数给 claude

-- 之后的参数会原样传递给 claude 命令`,
	Version: Version,
	Args:    cobra.ArbitraryArgs,
	RunE:    runRoot,
}

func init() {
	rootCmd.SetVersionTemplate("cc-start {{.Version}}\n")
}

func runRoot(cmd *cobra.Command, args []string) error {
	// 解析参数：第一个是 profile 名（可选），-- 之后的是 claude 参数
	profileName := ""
	var claudeArgs []string

	// 查找 -- 分隔符的位置
	dashPos := findDashSeparator(os.Args)

	if dashPos != -1 {
		// 有 -- 分隔符
		// 解析 -- 之前的参数（cc-start 自己的参数）
		beforeDash := os.Args[1:dashPos]
		if len(beforeDash) > 0 && !isFlag(beforeDash[0]) {
			profileName = beforeDash[0]
		}
		// -- 之后的所有参数原样传递给 claude
		claudeArgs = os.Args[dashPos+1:]
	} else {
		// 没有 -- 分隔符，使用传统方式
		if len(args) > 0 {
			profileName = args[0]
		}
	}

	// 加载配置
	cfgPath := config.GetConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 获取配置
	profile, err := cfg.GetProfile(profileName)
	if err != nil {
		return fmt.Errorf("获取配置失败: %w\n\n运行 'cc-start setup' 创建配置", err)
	}

	// 启动
	return launcher.Launch(profile, claudeArgs)
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

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
  cc-start              使用默认配置启动
  cc-start moonshot     使用 moonshot 配置启动
  cc-start -- --help    传递参数给 claude`,
	Version: Version,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runRoot,
}

func init() {
	rootCmd.SetVersionTemplate("cc-start {{.Version}}\n")
}

func runRoot(cmd *cobra.Command, args []string) error {
	// 确定使用的配置名
	profileName := ""
	if len(args) > 0 {
		profileName = args[0]
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

	// 获取传递给 claude 的参数
	extraArgs := cmd.Flags().Args()
	if len(args) > 0 && len(extraArgs) > 0 && extraArgs[0] == args[0] {
		extraArgs = extraArgs[1:]
	}

	// 启动
	return launcher.Launch(profile, extraArgs)
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

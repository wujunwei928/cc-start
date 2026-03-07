// cmd/run.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wujunwei/cc-start/internal/config"
	"github.com/wujunwei/cc-start/internal/launcher"
)

// runCmd 启动 Claude Code 命令
var runCmd = &cobra.Command{
	Use:   "run [profile] [-- claude args...]",
	Short: "启动 Claude Code",
	Long: `使用指定配置启动 Claude Code。

使用方法:
  cc-start run                    使用默认配置启动
  cc-start run moonshot           使用 moonshot 配置启动
  cc-start run -- --help          查看 claude 帮助
  cc-start run moonshot -- --dangerously-skip-permissions

-- 之后的参数会原样传递给 claude 命令`,
	Args: cobra.ArbitraryArgs,
	RunE: runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	// 解析参数
	profileName := ""
	var claudeArgs []string

	dashPos := findDashSeparator(os.Args)

	if dashPos != -1 {
		beforeDash := os.Args[1:dashPos]
		// 跳过 "run" 命令本身
		for _, arg := range beforeDash {
			if arg != "run" && !isFlag(arg) {
				profileName = arg
				break
			}
		}
		claudeArgs = os.Args[dashPos+1:]
	} else {
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

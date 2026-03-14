// cmd/opencode.go
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wujunwei928/cc-start/internal/tools"
)

// opencodeCmd 启动 OpenCode AI 编程助手
var opencodeCmd = &cobra.Command{
	Use:   "opencode [profile] [flags] [-- tool-args]",
	Short: "启动 OpenCode AI 编程助手",
	Long: `使用 OpenCode AI 编程助手启动。

示例:
  cc-start opencode                      使用默认配置启动
  cc-start opencode deepseek             使用 deepseek 配置
  cc-start opencode -m deepseek-chat     指定模型
  cc-start opencode deepseek -- --help   传递参数给 opencode`,
	RunE: runOpencode,
}

func init() {
	rootCmd.AddCommand(opencodeCmd)

	opencodeCmd.Flags().StringVarP(&launchModel, "model", "m", "", "模型名称")
	opencodeCmd.Flags().StringVarP(&launchBaseURL, "base-url", "b", "", "API 基础地址")
	opencodeCmd.Flags().StringVarP(&launchToken, "token", "t", "", "认证令牌")
	opencodeCmd.Flags().StringArrayVarP(&launchEnv, "env", "e", nil, "环境变量 (格式: KEY=VALUE)")
}

func runOpencode(cmd *cobra.Command, args []string) error {
	toolName := "opencode"

	// 验证工具存在
	if _, err := tools.GetTool(toolName); err != nil {
		return err
	}

	return runLaunchWithTool(toolName, args, "opencode")
}

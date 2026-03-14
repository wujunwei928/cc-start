// cmd/codex.go
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wujunwei928/cc-start/internal/tools"
)

// codexCmd 启动 OpenAI Codex CLI
var codexCmd = &cobra.Command{
	Use:   "codex [profile] [flags] [-- tool-args]",
	Short: "启动 OpenAI Codex CLI",
	Long: `使用 OpenAI Codex CLI 启动编程助手。

示例:
  cc-start codex                      使用默认配置启动
  cc-start codex openai               使用 openai 配置
  cc-start codex -m gpt-4 -t sk-xxx   指定模型和令牌
  cc-start codex openai -- --help     传递参数给 codex`,
	RunE: runCodex,
}

func init() {
	rootCmd.AddCommand(codexCmd)

	codexCmd.Flags().StringVarP(&launchModel, "model", "m", "", "模型名称")
	codexCmd.Flags().StringVarP(&launchBaseURL, "base-url", "b", "", "API 基础地址")
	codexCmd.Flags().StringVarP(&launchToken, "token", "t", "", "认证令牌")
	codexCmd.Flags().StringArrayVarP(&launchEnv, "env", "e", nil, "环境变量 (格式: KEY=VALUE)")
}

func runCodex(cmd *cobra.Command, args []string) error {
	toolName := "codex"

	// 验证工具存在
	if _, err := tools.GetTool(toolName); err != nil {
		return err
	}

	return runLaunchWithTool(toolName, args, "codex")
}

// cmd/claude.go
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wujunwei928/cc-start/internal/tools"
)

// claudeCmd 启动 Claude Code CLI
var claudeCmd = &cobra.Command{
	Use:   "claude [profile] [flags] [-- tool-args]",
	Short: "启动 Claude Code CLI",
	Long: `使用 Claude Code CLI 启动编程助手。

示例:
  cc-start claude                      使用默认配置启动
  cc-start claude moonshot             使用 moonshot 配置
  cc-start claude -m claude-sonnet-4   指定模型
  cc-start claude moonshot -- --help   传递参数给 claude`,
	RunE: runClaude,
}

func init() {
	rootCmd.AddCommand(claudeCmd)

	claudeCmd.Flags().StringVarP(&launchModel, "model", "m", "", "模型名称")
	claudeCmd.Flags().StringVarP(&launchBaseURL, "base-url", "b", "", "API 基础地址")
	claudeCmd.Flags().StringVarP(&launchToken, "token", "t", "", "认证令牌")
	claudeCmd.Flags().StringArrayVarP(&launchEnv, "env", "e", nil, "环境变量 (格式: KEY=VALUE)")
}

func runClaude(cmd *cobra.Command, args []string) error {
	toolName := "claude"

	// 验证工具存在
	if _, err := tools.GetTool(toolName); err != nil {
		return err
	}

	return runLaunchWithTool(toolName, args, "claude")
}

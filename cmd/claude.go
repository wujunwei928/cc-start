// cmd/claude.go
package cmd

import (
	"github.com/spf13/cobra"
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
	claudeCmd.Flags().BoolVarP(&launchYolo, "yolo", "y", false, "自动接受所有操作（YOLO 模式）")
}

func runClaude(cmd *cobra.Command, args []string) error {
	return runLaunch(args, "claude")
}

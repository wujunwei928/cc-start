// cmd/list.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wujunwei928/cc-start/internal/config"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出所有配置",
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	cfgPath := config.GetConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println("暂无配置，运行 'cc-start setup' 创建配置")
		return nil
	}

	fmt.Println("已保存的配置:")
	fmt.Println()

	for _, p := range cfg.Profiles {
		marker := " "
		if p.Name == cfg.Default {
			marker = "*"
		}
		fmt.Printf("  %s %s\n", marker, p.Name)
		if p.AnthropicBaseURL != "" {
			fmt.Printf("      Anthropic URL: %s\n", p.AnthropicBaseURL)
		}
		if p.OpenAIBaseURL != "" {
			fmt.Printf("      OpenAI URL: %s\n", p.OpenAIBaseURL)
		}
		if p.Model != "" {
			fmt.Printf("      模型: %s\n", p.Model)
		}
		fmt.Printf("      Token: %s...\n\n", maskToken(p.Token))
	}

	return nil
}

// maskToken 隐藏 Token 大部分内容
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

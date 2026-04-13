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
			fmt.Printf("      Base URL: %s\n", p.AnthropicBaseURL)
		}
		if p.SonnetModel != "" {
			fmt.Printf("      主模型 (Sonnet): %s\n", p.SonnetModel)
		}
		if p.HaikuModel != "" {
			fmt.Printf("      快速模型 (Haiku): %s\n", p.HaikuModel)
		}
		if p.OpusModel != "" {
			fmt.Printf("      经济模型 (Opus): %s\n", p.OpusModel)
		}
		fmt.Printf("      Token: %s...\n\n", config.MaskToken(p.Token))
	}

	return nil
}

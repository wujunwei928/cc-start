// cmd/default.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wujunwei928/cc-start/internal/config"
)

var defaultCmd = &cobra.Command{
	Use:   "default <name>",
	Short: "设置默认配置",
	Args:  cobra.ExactArgs(1),
	RunE:  runDefault,
}

func init() {
	rootCmd.AddCommand(defaultCmd)
}

func runDefault(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfgPath := config.GetConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if err := cfg.SetDefault(name); err != nil {
		return err
	}

	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	fmt.Printf("✅ 已设置 '%s' 为默认配置\n", name)
	return nil
}

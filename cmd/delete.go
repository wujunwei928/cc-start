// cmd/delete.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wujunwei/cc-start/internal/config"
)

var (
	deleteForce bool
)

var deleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"rm"},
	Short:   "删除配置",
	Args:    cobra.ExactArgs(1),
	RunE:    runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "强制删除，不确认")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfgPath := config.GetConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 检查配置是否存在
	_, err = cfg.GetProfile(name)
	if err != nil {
		return err
	}

	// 确认删除
	if !deleteForce {
		fmt.Printf("确定要删除配置 '%s'? [y/N] ", name)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			fmt.Println("已取消")
			return nil
		}
	}

	if err := cfg.DeleteProfile(name); err != nil {
		return err
	}

	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	fmt.Printf("已删除配置 '%s'\n", name)
	return nil
}

// cmd/setup.go
package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/wujunwei928/cc-start/internal/tui/setup"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "交互式配置向导",
	RunE:  runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	m := setup.InitialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("启动 TUI 失败: %w", err)
	}

	if final, ok := result.(setup.Model); ok && final.Done() {
		fmt.Printf("\n✅ 配置 '%s' 已保存\n", final.GetName())
	}

	return nil
}

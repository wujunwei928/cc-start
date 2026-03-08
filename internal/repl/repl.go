// internal/repl/repl.go
package repl

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wujunwei/cc-start/internal/config"
	"github.com/wujunwei/cc-start/internal/launcher"
)

// REPL 交互式 REPL
type REPL struct {
	cfg         *config.Config
	cfgPath     string
	history     *History
	currentName string
}

// New 创建 REPL 实例
func New(cfgPath string) (*REPL, error) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return nil, err
	}

	repl := &REPL{
		cfg:     cfg,
		cfgPath: cfgPath,
		history: NewHistory(),
	}

	repl.currentName = cfg.Default

	return repl, nil
}

// Run 启动 REPL（使用 Bubble Tea）
func (r *REPL) Run() {
	r.printWelcome()

	model, err := NewModel(r.cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 同步状态
	model.currentProfile = r.currentName
	model.config = r.cfg
	model.history = r.history

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "启动失败: %v\n", err)
		os.Exit(1)
	}

	// 检查是否有待执行的启动命令
	if m, ok := finalModel.(Model); ok && m.PendingLaunch != nil {
		if err := launcher.Launch(&m.PendingLaunch.Profile, m.PendingLaunch.Args); err != nil {
			fmt.Fprintf(os.Stderr, "启动失败: %v\n", err)
			os.Exit(1)
		}
	}
}

func (r *REPL) printWelcome() {
	fmt.Println()
	fmt.Println("CC-Start REPL v2.0")
	fmt.Println("输入 '/help' 查看可用命令，'/exit' 退出。")
	fmt.Println("按 ctrl+p 打开命令面板。")
	fmt.Println()
}

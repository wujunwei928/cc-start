// internal/repl/repl.go
package repl

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wujunwei928/cc-start/internal/config"
	"github.com/wujunwei928/cc-start/internal/launcher"
	"github.com/wujunwei928/cc-start/internal/tui/setup"
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
	for {
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

		// 检查是否有待执行的 setup
		if m, ok := finalModel.(Model); ok && m.PendingSetup != nil {
			r.runSetup(m.PendingSetup)
			// setup 完成后重新加载配置并继续循环
			r.cfg, _ = config.LoadConfig(r.cfgPath)
			r.currentName = r.cfg.Default
			continue
		}

		// 检查是否有待执行的启动命令
		if m, ok := finalModel.(Model); ok && m.PendingLaunch != nil {
			if err := launcher.Launch(&m.PendingLaunch.Profile, m.PendingLaunch.Args); err != nil {
				fmt.Fprintf(os.Stderr, "启动失败: %v\n", err)
				os.Exit(1)
			}
		}
		break
	}
}

// runSetup 运行设置向导
func (r *REPL) runSetup(ps *PendingSetup) {
	var setupModel tea.Model
	if ps.IsEdit && ps.Profile != nil {
		setupModel = setup.InitialModelWithProfile(*ps.Profile)
	} else {
		setupModel = setup.InitialModel()
	}

	p := tea.NewProgram(setupModel, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "设置向导运行失败: %v\n", err)
		return
	}

	// 显示保存结果
	if m, ok := final.(setup.Model); ok && m.Done() {
		fmt.Printf("✓ 配置 '%s' 已保存\n", m.GetName())
	}
}

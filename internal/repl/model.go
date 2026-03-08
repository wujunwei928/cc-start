// internal/repl/model.go
package repl

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wujunwei/cc-start/internal/config"
)

// Focus 当前焦点状态
type Focus int

const (
	FocusInput Focus = iota
	FocusPalette
)

// keyMap 快捷键绑定
type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	CtrlP key.Binding
	CtrlC key.Binding
	CtrlL key.Binding
	Esc   key.Binding
}

// defaultKeyMap 默认快捷键
func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "上一条历史"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "下一条历史"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "执行命令"),
		),
		CtrlP: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "命令面板"),
		),
		CtrlC: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "退出"),
		),
		CtrlL: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "清屏"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "关闭面板"),
		),
	}
}

// Model REPL 主模型
type Model struct {
	// 配置
	config     *config.Config
	configPath string

	// 当前状态
	currentProfile string
	focus          Focus
	quitting       bool
	keys           keyMap

	// 组件
	input   textinput.Model
	output  *OutputBuffer
	palette *CommandPalette
	help    help.Model

	// 历史记录
	history *History
	histIdx int

	// 样式
	styles Styles

	// 窗口尺寸
	width  int
	height int
}

// NewModel 创建新的 REPL Model
func NewModel(cfgPath string) (Model, error) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return Model{}, err
	}

	ti := textinput.New()
	ti.Placeholder = "输入命令..."
	ti.Focus()

	h := help.New()
	hist := NewHistory()
	out := NewOutputBuffer(100)

	return Model{
		config:         cfg,
		configPath:     cfgPath,
		currentProfile: cfg.Default,
		focus:          FocusInput,
		keys:           defaultKeyMap(),
		input:          ti,
		output:         out,
		history:        hist,
		help:           h,
		styles:         DefaultStyles(),
	}, nil
}

// Init 初始化
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

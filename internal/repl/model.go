// internal/repl/model.go
package repl

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wujunwei928/cc-start/internal/config"
	"github.com/wujunwei928/cc-start/internal/i18n"
	"github.com/wujunwei928/cc-start/internal/theme"
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
	Space key.Binding
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
			key.WithHelp("ctrl+p", "系统设置"),
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
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "输入"),
		),
	}
}

// PendingLaunch 待执行的启动命令
type PendingLaunch struct {
	Profile config.Profile
	Args    []string
}

// Model REPL 主模型
type Model struct {
	config     *config.Config
	configPath string

	currentProfile string
	focus          Focus
	quitting       bool
	keys           keyMap

	input    textinput.Model
	output   *OutputBuffer
	palette  *CommandPalette
	settings *SettingsPanel
	help     help.Model

	history *History
	histIdx int

	Styles Styles

	width  int
	height int

	PendingLaunch *PendingLaunch

	I18n  *i18n.Manager
	Theme *theme.Theme
}

// NewModel 创建新的 REPL Model
func NewModel(cfgPath string) (Model, error) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return Model{}, err
	}

	i18nMgr := i18n.NewManager()
	if cfg.Settings.Language != "" {
		i18nMgr.SetLanguage(cfg.Settings.Language)
	}

	currentTheme, err := theme.GetTheme(cfg.Settings.Theme)
	if err != nil {
		currentTheme, _ = theme.GetTheme("default")
	}

	styles := NewStylesFromTheme(currentTheme)

	ti := textinput.New()
	ti.Placeholder = i18nMgr.T(i18n.MsgREPLInputPrompt)
	ti.Focus()
	ti.Prompt = ""

	h := help.New()
	hist := NewHistory()
	out := NewOutputBuffer(100)

	// 写入欢迎信息到输出缓冲区（在 TUI 内显示）
	out.WriteInfo("CC-Start REPL v2.0")
	out.Write("输入 '/' 打开命令面板，'/help' 查看帮助，'/exit' 退出。")
	out.Write("按 ctrl+p 打开系统设置。")

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
		Styles:         styles,
		I18n:           i18nMgr,
		Theme:          currentTheme,
	}, nil
}

// Init 初始化
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

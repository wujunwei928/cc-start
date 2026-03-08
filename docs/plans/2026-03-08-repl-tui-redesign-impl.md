# REPL TUI 重构实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 使用 Bubble Tea 框架重写 REPL，实现全屏交互式 UI

**Architecture:** 采用 MVU (Model-View-Update) 架构，主界面包含输入框、输出区和帮助栏，命令面板作为弹窗组件。样式系统使用 Lipgloss，模糊搜索使用 sahilm/fuzzy。

**Tech Stack:** Bubble Tea v0.27, Lipgloss v0.13, bubbles v0.20, sahilm/fuzzy

---

## Phase 1: 基础框架

### Task 1: 添加依赖

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

**Step 1: 添加新依赖**

```bash
cd /code/ai/cc-start/.worktrees/repl-tui-redesign
go get github.com/charmbracelet/bubbletea@v0.27.0
go get github.com/charmbracelet/lipgloss@v0.13.0
go get github.com/charmbracelet/bubbles@v0.20.0
go get github.com/sahilm/fuzzy@v0.1.0
```

**Step 2: 移除旧依赖**

```bash
go mod edit -droprequire=github.com/c-bata/go-prompt
go mod tidy
```

**Step 3: 验证依赖**

```bash
go mod verify
```

Expected: `all modules verified`

**Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: 切换依赖从 go-prompt 到 bubbletea"
```

---

### Task 2: 创建样式系统

**Files:**
- Create: `internal/repl/styles.go`

**Step 1: 创建样式文件**

```go
// internal/repl/styles.go
package repl

import "github.com/charmbracelet/lipgloss"

// 配色方案（参考 crush 深色主题）
var (
	// 背景色
	bgColor      = lipgloss.Color("#1a1a2e")
	surfaceColor = lipgloss.Color("#16213e")

	// 文字色
	textColor   = lipgloss.Color("#e0e0e0")
	mutedColor  = lipgloss.Color("#6c7086")
	accentColor = lipgloss.Color("#89b4fa")

	// 状态色
	successColor = lipgloss.Color("#a6e3a1")
	errorColor   = lipgloss.Color("#f38ba8")
	warningColor = lipgloss.Color("#fab387")
	infoColor    = lipgloss.Color("#89dceb")
)

// Styles 包含所有 UI 样式
type Styles struct {
	// 主界面
	App        lipgloss.Style
	Prefix     lipgloss.Style
	Input      lipgloss.Style
	Output     lipgloss.Style
	HelpBar    lipgloss.Style

	// 输出样式
	Success    lipgloss.Style
	Error      lipgloss.Style
	Warning    lipgloss.Style
	Info       lipgloss.Style

	// 命令面板
	Palette       lipgloss.Style
	PaletteTitle  lipgloss.Style
	PaletteInput  lipgloss.Style
	PaletteList   lipgloss.Style
	PaletteItem   lipgloss.Style
	PaletteActive lipgloss.Style
}

// DefaultStyles 返回默认样式
func DefaultStyles() Styles {
	return Styles{
		App: lipgloss.NewStyle().
			Padding(1, 2),

		Prefix: lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true),

		Input: lipgloss.NewStyle().
			Foreground(textColor),

		Output: lipgloss.NewStyle().
			Foreground(textColor).
			Padding(1, 0),

		HelpBar: lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(0, 1),

		Success: lipgloss.NewStyle().
			Foreground(successColor).
			SetString("✓"),

		Error: lipgloss.NewStyle().
			Foreground(errorColor).
			SetString("✗"),

		Warning: lipgloss.NewStyle().
			Foreground(warningColor).
			SetString("⚠"),

		Info: lipgloss.NewStyle().
			Foreground(infoColor).
			SetString("●"),

		Palette: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(0, 1).
			Width(50),

		PaletteTitle: lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Padding(0, 1),

		PaletteInput: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(mutedColor).
			Padding(0, 1).
			Margin(1, 0),

		PaletteItem: lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 2),

		PaletteActive: lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Padding(0, 2),
	}
}
```

**Step 2: 验证编译**

```bash
go build ./internal/repl/...
```

Expected: 无错误

**Step 3: Commit**

```bash
git add internal/repl/styles.go
git commit -m "feat(repl): 添加 Lipgloss 样式系统"
```

---

### Task 3: 创建消息类型

**Files:**
- Create: `internal/repl/messages.go`

**Step 1: 创建消息定义**

```go
// internal/repl/messages.go
package repl

// 消息类型定义
type (
	// CommandExecutedMsg 命令执行完成
	CommandExecutedMsg struct {
		Output string
		Err    error
	}

	// CommandSelectedMsg 从命令面板选择命令
	CommandSelectedMsg struct {
		Cmd  string
		Args []string
	}

	// ProfileChangedMsg 配置切换
	ProfileChangedMsg struct {
		Name string
	}

	// PaletteToggledMsg 切换命令面板显示
	PaletteToggledMsg struct{}

	// OutputClearedMsg 清空输出
	OutputClearedMsg struct{}
)
```

**Step 2: 验证编译**

```bash
go build ./internal/repl/...
```

Expected: 无错误

**Step 3: Commit**

```bash
git add internal/repl/messages.go
git commit -m "feat(repl): 添加 Bubble Tea 消息类型定义"
```

---

### Task 4: 创建 Model 定义

**Files:**
- Create: `internal/repl/model.go`

**Step 1: 创建 Model 结构**

```go
// internal/repl/model.go
package repl

import (
	"github.com/charmbracelet/bubbles/help"
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

// Model REPL 主模型
type Model struct {
	// 配置
	config     *config.Config
	configPath string

	// 当前状态
	currentProfile string
	focus          Focus
	quitting       bool

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

// keyMap 快捷键绑定
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	CtrlP    key.Binding
	CtrlC    key.Binding
	CtrlL    key.Binding
	Esc      key.Binding
	Tab      key.Binding
}

// NewModel 创建新的 REPL Model
func NewModel(cfgPath string) (Model, error) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return Model{}, err
	}

	// 初始化输入框
	ti := textinput.New()
	ti.Placeholder = "输入命令..."
	ti.Focus()

	// 初始化帮助
	h := help.New()

	// 初始化历史
	hist := NewHistory()

	// 初始化输出缓冲区
	out := NewOutputBuffer(100)

	return Model{
		config:         cfg,
		configPath:     cfgPath,
		currentProfile: cfg.Default,
		focus:          FocusInput,
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

// key 绑定需要引入
import "github.com/charmbracelet/bubbles/key"
```

**Step 2: 修复 import 并简化**

```go
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
```

**Step 3: 验证编译**

```bash
go build ./internal/repl/...
```

Expected: 编译错误（缺少 OutputBuffer 和 CommandPalette），这是预期的

**Step 4: Commit**

```bash
git add internal/repl/model.go
git commit -m "feat(repl): 添加 Model 结构定义"
```

---

### Task 5: 创建输出缓冲区

**Files:**
- Create: `internal/repl/output.go`

**Step 1: 创建输出缓冲区**

```go
// internal/repl/output.go
package repl

import (
	"strings"
)

// OutputLine 输出行
type OutputLine struct {
	Content string
	Type    OutputType
}

// OutputType 输出类型
type OutputType int

const (
	OutputNormal OutputType = iota
	OutputSuccess
	OutputError
	OutputWarning
	OutputInfo
)

// OutputBuffer 输出缓冲区
type OutputBuffer struct {
	lines   []OutputLine
	maxSize int
}

// NewOutputBuffer 创建输出缓冲区
func NewOutputBuffer(maxSize int) *OutputBuffer {
	return &OutputBuffer{
		lines:   make([]OutputLine, 0),
		maxSize: maxSize,
	}
}

// Write 写入普通输出
func (b *OutputBuffer) Write(content string) {
	b.writeLine(content, OutputNormal)
}

// WriteSuccess 写入成功输出
func (b *OutputBuffer) WriteSuccess(content string) {
	b.writeLine(content, OutputSuccess)
}

// WriteError 写入错误输出
func (b *OutputBuffer) WriteError(content string) {
	b.writeLine(content, OutputError)
}

// WriteWarning 写入警告输出
func (b *OutputBuffer) WriteWarning(content string) {
	b.writeLine(content, OutputWarning)
}

// WriteInfo 写入信息输出
func (b *OutputBuffer) WriteInfo(content string) {
	b.writeLine(content, OutputInfo)
}

func (b *OutputBuffer) writeLine(content string, t OutputType) {
	// 按换行分割
	for _, line := range strings.Split(content, "\n") {
		b.lines = append(b.lines, OutputLine{Content: line, Type: t})
	}

	// 超过最大行数时裁剪
	if len(b.lines) > b.maxSize {
		b.lines = b.lines[len(b.lines)-b.maxSize:]
	}
}

// Clear 清空缓冲区
func (b *OutputBuffer) Clear() {
	b.lines = make([]OutputLine, 0)
}

// Lines 获取所有行
func (b *OutputBuffer) Lines() []OutputLine {
	return b.lines
}

// Render 渲染输出
func (b *OutputBuffer) Render(styles Styles, width int) string {
	var sb strings.Builder
	for _, line := range b.lines {
		var styled string
		switch line.Type {
		case OutputSuccess:
			styled = styles.Success.Render(line.Content)
		case OutputError:
			styled = styles.Error.Render(line.Content)
		case OutputWarning:
			styled = styles.Warning.Render(line.Content)
		case OutputInfo:
			styled = styles.Info.Render(line.Content)
		default:
			styled = line.Content
		}
		sb.WriteString(styled + "\n")
	}
	return sb.String()
}
```

**Step 2: 验证编译**

```bash
go build ./internal/repl/...
```

Expected: 编译错误（缺少 CommandPalette）

**Step 3: Commit**

```bash
git add internal/repl/output.go
git commit -m "feat(repl): 添加输出缓冲区组件"
```

---

### Task 6: 创建命令面板（占位）

**Files:**
- Create: `internal/repl/palette.go`

**Step 1: 创建命令面板占位**

```go
// internal/repl/palette.go
package repl

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// CommandPalette 命令面板
type CommandPalette struct {
	visible bool
	query   string
	items   []PaletteItem
	selected int
	styles  Styles
	width   int
}

// PaletteItem 命令面板项
type PaletteItem struct {
	Cmd         string
	Description string
	Aliases     []string
	Group       string
}

// NewCommandPalette 创建命令面板
func NewCommandPalette(styles Styles) *CommandPalette {
	return &CommandPalette{
		styles:  styles,
		items:   getDefaultCommands(),
	}
}

func getDefaultCommands() []PaletteItem {
	return []PaletteItem{
		{Cmd: "/list", Description: "列出所有配置", Aliases: []string{"/ls"}, Group: "配置管理"},
		{Cmd: "/use", Description: "切换当前会话配置", Aliases: []string{"/switch"}, Group: "配置管理"},
		{Cmd: "/current", Description: "显示当前配置", Aliases: []string{"/status"}, Group: "配置管理"},
		{Cmd: "/default", Description: "设置默认配置", Group: "配置管理"},
		{Cmd: "/show", Description: "显示配置详情", Group: "配置管理"},
		{Cmd: "/edit", Description: "编辑配置", Group: "配置管理"},
		{Cmd: "/delete", Description: "删除配置", Aliases: []string{"/rm"}, Group: "配置管理"},
		{Cmd: "/copy", Description: "复制配置", Aliases: []string{"/cp"}, Group: "配置管理"},
		{Cmd: "/rename", Description: "重命名配置", Aliases: []string{"/mv"}, Group: "配置管理"},
		{Cmd: "/test", Description: "测试 API 连通性", Group: "测试"},
		{Cmd: "/export", Description: "导出配置", Group: "导入导出"},
		{Cmd: "/import", Description: "导入配置", Group: "导入导出"},
		{Cmd: "/history", Description: "显示命令历史", Group: "辅助"},
		{Cmd: "/help", Description: "显示帮助", Aliases: []string{"/?", "/h"}, Group: "辅助"},
		{Cmd: "/clear", Description: "清屏", Aliases: []string{"/cls"}, Group: "辅助"},
		{Cmd: "/run", Description: "启动 Claude Code", Group: "启动"},
		{Cmd: "/setup", Description: "运行配置向导", Group: "启动"},
		{Cmd: "/exit", Description: "退出 REPL", Aliases: []string{"/quit", "/q"}, Group: "启动"},
	}
}

// Toggle 切换显示状态
func (p *CommandPalette) Toggle() {
	p.visible = !p.visible
	if p.visible {
		p.query = ""
		p.selected = 0
	}
}

// IsVisible 返回是否可见
func (p *CommandPalette) IsVisible() bool {
	return p.visible
}

// SetWidth 设置宽度
func (p *CommandPalette) SetWidth(w int) {
	p.width = w
}

// Update 更新状态
func (p *CommandPalette) Update(msg tea.KeyMsg) tea.Cmd {
	if !p.visible {
		return nil
	}

	switch msg.String() {
	case "up":
		if p.selected > 0 {
			p.selected--
		}
	case "down":
		if p.selected < len(p.filteredItems())-1 {
			p.selected++
		}
	case "enter":
		// 返回选中的命令
		items := p.filteredItems()
		if p.selected < len(items) {
			// 由外部处理
		}
	case "esc":
		p.visible = false
	default:
		// 输入字符更新查询
		if len(msg.Runes) > 0 {
			p.query += string(msg.Runes)
			p.selected = 0
		} else if msg.String() == "backspace" && len(p.query) > 0 {
			p.query = p.query[:len(p.query)-1]
		}
	}
	return nil
}

// filteredItems 返回过滤后的项
func (p *CommandPalette) filteredItems() []PaletteItem {
	if p.query == "" {
		return p.items
	}
	// TODO: 使用 fuzzy 搜索
	var result []PaletteItem
	for _, item := range p.items {
		if contains(item.Cmd, p.query) {
			result = append(result, item)
		}
	}
	return result
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

// Render 渲染面板
func (p *CommandPalette) Render() string {
	if !p.visible {
		return ""
	}

	// 简单渲染
	var result string
	result += p.styles.PaletteTitle.Render("Commands") + "\n"
	result += p.styles.PaletteInput.Render("> " + p.query) + "\n"

	items := p.filteredItems()
	for i, item := range items {
		if i == p.selected {
			result += p.styles.PaletteActive.Render("● " + item.Cmd + "  " + item.Description) + "\n"
		} else {
			result += p.styles.PaletteItem.Render("  " + item.Cmd + "  " + item.Description) + "\n"
		}
	}

	return p.styles.Palette.Render(result)
}

// SelectedCommand 返回选中的命令
func (p *CommandPalette) SelectedCommand() string {
	items := p.filteredItems()
	if p.selected < len(items) {
		return items[p.selected].Cmd
	}
	return ""
}

// 键绑定（避免未使用警告）
var _ = key.Binding{}
```

**Step 2: 验证编译**

```bash
go build ./internal/repl/...
```

Expected: 编译成功

**Step 3: Commit**

```bash
git add internal/repl/palette.go
git commit -m "feat(repl): 添加命令面板组件（基础版）"
```

---

### Task 7: 创建 View 渲染

**Files:**
- Create: `internal/repl/view.go`

**Step 1: 创建 View 函数**

```go
// internal/repl/view.go
package repl

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View 渲染 UI
func (m Model) View() string {
	if m.quitting {
		return "再见!\n"
	}

	var sections []string

	// 输入区
	prefix := m.styles.Prefix.Render(m.getPromptPrefix())
	inputLine := lipgloss.JoinHorizontal(lipgloss.Left, prefix, " ", m.input.View())
	sections = append(sections, inputLine)

	// 输出区
	if len(m.output.Lines()) > 0 {
		outputContent := m.output.Render(m.styles, m.width)
		sections = append(sections, m.styles.Output.Render(outputContent))
	}

	// 命令面板（覆盖层）
	if m.palette != nil && m.palette.IsVisible() {
		paletteView := m.palette.Render()
		sections = append(sections, "\n"+paletteView)
	}

	// 帮助栏
	helpBar := m.renderHelpBar()
	sections = append(sections, helpBar)

	return lipgloss.JoinVertical(lipgloss.Left, sections...) + "\n"
}

func (m Model) getPromptPrefix() string {
	if m.currentProfile != "" {
		return fmt.Sprintf("cc-start [%s]>", m.currentProfile)
	}
	return "cc-start>"
}

func (m Model) renderHelpBar() string {
	var hints []string

	if m.palette != nil && m.palette.IsVisible() {
		hints = []string{"↑↓ 导航", "enter 确认", "esc 关闭"}
	} else {
		hints = []string{
			"ctrl+p 命令",
			"↑↓ 历史",
			"enter 执行",
			"ctrl+c 退出",
		}
	}

	return m.styles.HelpBar.Render(strings.Join(hints, "  "))
}
```

**Step 2: 验证编译**

```bash
go build ./internal/repl/...
```

Expected: 编译成功

**Step 3: Commit**

```bash
git add internal/repl/view.go
git commit -m "feat(repl): 添加 View 渲染函数"
```

---

### Task 8: 创建 Update 函数

**Files:**
- Create: `internal/repl/update.go`

**Step 1: 创建 Update 函数**

```go
// internal/repl/update.go
package repl

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Update 处理消息更新
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 命令面板激活时的处理
		if m.palette != nil && m.palette.IsVisible() {
			return m.updatePalette(msg)
		}

		// 主界面按键处理
		switch {
		case keyMatches(msg, m.keys.CtrlC):
			m.quitting = true
			return m, tea.Quit

		case keyMatches(msg, m.keys.CtrlP):
			if m.palette == nil {
				m.palette = NewCommandPalette(m.styles)
			}
			m.palette.Toggle()
			return m, nil

		case keyMatches(msg, m.keys.CtrlL):
			m.output.Clear()
			return m, nil

		case keyMatches(msg, m.keys.Enter):
			return m.executeInput()

		case keyMatches(msg, m.keys.Up):
			return m.navigateHistory(-1)

		case keyMatches(msg, m.keys.Down):
			return m.navigateHistory(1)

		default:
			// 更新输入框
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case CommandSelectedMsg:
		return m.executeCommand(msg.Cmd, msg.Args)

	case CommandExecutedMsg:
		if msg.Err != nil {
			m.output.WriteError(msg.Err.Error())
		} else {
			m.output.Write(msg.Output)
		}
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updatePalette(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		cmd := m.palette.SelectedCommand()
		m.palette.Toggle()
		return m.executeCommand(cmd, nil)
	case "esc":
		m.palette.Toggle()
		return m, nil
	default:
		m.palette.Update(msg)
		return m, nil
	}
}

func (m Model) executeInput() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.input.Value())
	if input == "" {
		return m, nil
	}

	m.history.Add(input)
	m.input.SetValue("")
	m.histIdx = 0

	// 解析命令
	parts := strings.Fields(input)
	cmd := parts[0]
	args := parts[1:]

	// 自动添加 / 前缀
	if !strings.HasPrefix(cmd, "/") {
		cmd = "/" + cmd
	}

	return m.executeCommand(cmd, args)
}

func (m Model) executeCommand(cmd string, args []string) (tea.Model, tea.Cmd) {
	// 调用现有的命令执行逻辑
	output := m.executeCommandInternal(cmd, args)
	m.output.Write(output)
	return m, nil
}

func (m Model) navigateHistory(dir int) (tea.Model, tea.Cmd) {
	cmds := m.history.GetCommands()
	if len(cmds) == 0 {
		return m, nil
	}

	newIdx := m.histIdx + dir
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx > len(cmds) {
		newIdx = len(cmds)
	}
	m.histIdx = newIdx

	if newIdx == 0 {
		m.input.SetValue("")
	} else {
		m.input.SetValue(cmds[newIdx-1])
	}
	m.input.CursorEnd()

	return m, nil
}

func keyMatches(msg tea.KeyMsg, binding interface{}) bool {
	// 简化版本，实际需要检查 binding
	return false // 临时返回，后续完善
}
```

**Step 2: 验证编译**

```bash
go build ./internal/repl/...
```

Expected: 编译错误（缺少 executeCommandInternal）

**Step 3: Commit**

```bash
git add internal/repl/update.go
git commit -m "feat(repl): 添加 Update 函数框架"
```

---

### Task 9: 集成命令执行逻辑

**Files:**
- Modify: `internal/repl/update.go`

**Step 1: 添加命令执行内部函数**

在 `update.go` 末尾添加：

```go
// executeCommandInternal 执行命令并返回输出
func (m *Model) executeCommandInternal(cmd string, args []string) string {
	var buf strings.Builder
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 调用现有命令逻辑
	m.ExecuteCommand(cmd, args)

	w.Close()
	os.Stdout = oldStdout

	io.Copy(&buf, r)
	return buf.String()
}
```

需要添加 import：
```go
import (
	"io"
	"os"
	// ...
)
```

**Step 2: 修复 keyMatches 函数**

```go
func keyMatches(msg tea.KeyMsg, binding key.Binding) bool {
	return binding.Enabled() && slices.Contains(binding.Keys(), msg.String())
}
```

需要添加 import：
```go
import "slices"
```

**Step 3: 验证编译**

```bash
go build ./internal/repl/...
```

Expected: 编译成功

**Step 4: Commit**

```bash
git add internal/repl/update.go
git commit -m "feat(repl): 集成命令执行逻辑"
```

---

### Task 10: 创建新的入口函数

**Files:**
- Modify: `internal/repl/repl.go`

**Step 1: 重写 REPL 入口**

```go
// internal/repl/repl.go
package repl

import (
	tea "github.com/charmbracelet/bubbletea"
)

// REPL 保留兼容性
type REPL struct {
	cfgPath string
}

// New 创建 REPL 实例（兼容旧接口）
func New(cfgPath string) (*REPL, error) {
	return &REPL{cfgPath: cfgPath}, nil
}

// Run 启动 REPL（使用 Bubble Tea）
func (r *REPL) Run() {
	model, err := NewModel(r.cfgPath)
	if err != nil {
		PrintError("加载配置失败: %v", err)
		return
	}

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		PrintError("启动失败: %v", err)
	}
}
```

**Step 2: 验证编译**

```bash
go build ./...
```

Expected: 编译成功

**Step 3: 运行测试**

```bash
go test ./internal/repl/... -v
```

Expected: 测试通过

**Step 4: Commit**

```bash
git add internal/repl/repl.go
git commit -m "feat(repl): 切换到 Bubble Tea 入口"
```

---

### Task 11: 清理旧代码

**Files:**
- Delete: `internal/repl/completer.go`
- Delete: `internal/repl/completer_test.go`
- Modify: `internal/repl/ui.go`

**Step 1: 删除不再需要的文件**

```bash
rm internal/repl/completer.go
rm internal/repl/completer_test.go
```

**Step 2: 简化 ui.go**

```go
// internal/repl/ui.go
package repl

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// 输出样式（供 commands.go 使用）
var (
	successStyle = lipgloss.NewStyle().Foreground(successColor)
	errorStyle   = lipgloss.NewStyle().Foreground(errorColor)
	warningStyle = lipgloss.NewStyle().Foreground(warningColor)
	infoStyle    = lipgloss.NewStyle().Foreground(infoColor)
)

// PrintSuccess 打印成功信息（兼容旧接口）
func PrintSuccess(format string, args ...interface{}) {
	printStyled(successStyle, "✓", format, args...)
}

// PrintError 打印错误信息
func PrintError(format string, args ...interface{}) {
	printStyled(errorStyle, "✗", format, args...)
}

// PrintWarning 打印警告信息
func PrintWarning(format string, args ...interface{}) {
	printStyled(warningStyle, "⚠", format, args...)
}

// PrintInfo 打印提示信息
func PrintInfo(format string, args ...interface{}) {
	printStyled(infoStyle, "●", format, args...)
}

// PrintCurrent 打印当前状态
func PrintCurrent(format string, args ...interface{}) {
	printStyled(infoStyle, "→", format, args...)
}

func printStyled(style lipgloss.Style, prefix, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(style.Render(prefix + " " + msg))
}

// NewTable 保留表格功能
// ... 保留现有的 NewTable 实现
```

**Step 3: 验证编译和测试**

```bash
go build ./...
go test ./internal/repl/... -v
```

Expected: 全部通过

**Step 4: Commit**

```bash
git add -A
git commit -m "refactor(repl): 移除 go-prompt 相关代码"
```

---

### Task 12: 手动测试

**Step 1: 编译并运行**

```bash
go build -o cc-start .
./cc-start
```

**Step 2: 验证功能**

- [ ] 程序启动显示输入框
- [ ] 输入 `/help` 显示帮助
- [ ] 输入 `/list` 显示配置列表
- [ ] `Ctrl+P` 打开命令面板
- [ ] 命令面板可以搜索和选择
- [ ] `↑↓` 可以导航历史
- [ ] `Ctrl+C` 可以退出

---

## 执行选项

**计划已保存到 `docs/plans/2026-03-08-repl-tui-redesign-impl.md`**

两种执行方式：

**1. Subagent-Driven (当前会话)** - 每个任务派发子代理，任务间审查，快速迭代

**2. Parallel Session (新会话)** - 在 worktree 中打开新会话，批量执行带检查点

**选择哪种方式？**

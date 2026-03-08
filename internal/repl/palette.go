// internal/repl/palette.go
package repl

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// CommandPalette 命令面板
type CommandPalette struct {
	visible  bool
	query    string
	items    []PaletteItem
	selected int
	styles   Styles
	width    int
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
		styles: styles,
		items:  getDefaultCommands(),
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

// HandleKey 处理按键
func (p *CommandPalette) HandleKey(s string) bool {
	switch s {
	case "up":
		if p.selected > 0 {
			p.selected--
		}
		return true
	case "down":
		if p.selected < len(p.filteredItems())-1 {
			p.selected++
		}
		return true
	case "backspace":
		if len(p.query) > 0 {
			p.query = p.query[:len(p.query)-1]
			p.selected = 0
		}
		return true
	default:
		// 单字符输入
		if len(s) == 1 {
			p.query += s
			p.selected = 0
			return true
		}
	}
	return false
}

// filteredItems 返回过滤后的项
func (p *CommandPalette) filteredItems() []PaletteItem {
	if p.query == "" {
		return p.items
	}

	// 使用 fuzzy 搜索
	var sources []string
	for _, item := range p.items {
		sources = append(sources, item.Cmd+" "+item.Description)
	}

	matches := fuzzy.Find(p.query, sources)

	var result []PaletteItem
	for _, match := range matches {
		result = append(result, p.items[match.Index])
	}
	return result
}

// Render 渲染面板
func (p *CommandPalette) Render() string {
	if !p.visible {
		return ""
	}

	var sections []string

	// 标题
	title := p.styles.PaletteTitle.Render("Commands")
	sections = append(sections, title)

	// 输入框
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(mutedColor).
		Padding(0, 1).
		Width(46)
	input := inputStyle.Render("> " + p.query)
	sections = append(sections, input)

	// 命令列表
	items := p.filteredItems()
	var listLines []string
	for i, item := range items {
		if i >= 10 {
			break // 最多显示 10 条
		}
		if i == p.selected {
			line := p.styles.PaletteActive.Render("● " + item.Cmd + "  " + item.Description)
			listLines = append(listLines, line)
		} else {
			line := p.styles.PaletteItem.Render("  " + item.Cmd + "  " + item.Description)
			listLines = append(listLines, line)
		}
	}
	if len(listLines) > 0 {
		sections = append(sections, strings.Join(listLines, "\n"))
	}

	// 提示
	hint := lipgloss.NewStyle().Foreground(mutedColor).Render("↑↓ 导航  enter 确认  esc 关闭")
	sections = append(sections, hint)

	return p.styles.Palette.Render(strings.Join(sections, "\n"))
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

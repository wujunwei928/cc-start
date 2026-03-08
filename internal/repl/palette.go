// internal/repl/palette.go
package repl

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"github.com/wujunwei/cc-start/internal/i18n"
)

// CommandPalette 命令面板
type CommandPalette struct {
	visible  bool
	query    string
	items    []PaletteItem
	selected int
	styles   Styles
	width    int
	i18n     *i18n.Manager
}

// PaletteItem 命令面板项
type PaletteItem struct {
	Cmd         string
	Description string
	Aliases     []string
	Group       string
}

// NewCommandPalette 创建命令面板
func NewCommandPalette(styles Styles, i18nMgr *i18n.Manager) *CommandPalette {
	return &CommandPalette{
		styles: styles,
		i18n:   i18nMgr,
		items:  getDefaultCommands(i18nMgr),
	}
}

func getDefaultCommands(i18nMgr *i18n.Manager) []PaletteItem {
	return []PaletteItem{
		{Cmd: "/list", Description: i18nMgr.T(i18n.MsgCmdList), Aliases: []string{"/ls"}, Group: "config"},
		{Cmd: "/use", Description: i18nMgr.T(i18n.MsgCmdUse), Aliases: []string{"/switch"}, Group: "config"},
		{Cmd: "/current", Description: i18nMgr.T(i18n.MsgCmdCurrent), Aliases: []string{"/status"}, Group: "config"},
		{Cmd: "/default", Description: i18nMgr.T(i18n.MsgCmdDefault), Group: "config"},
		{Cmd: "/show", Description: i18nMgr.T(i18n.MsgCmdShow), Group: "config"},
		{Cmd: "/edit", Description: i18nMgr.T(i18n.MsgCmdEdit), Group: "config"},
		{Cmd: "/delete", Description: i18nMgr.T(i18n.MsgCmdDelete), Aliases: []string{"/rm"}, Group: "config"},
		{Cmd: "/copy", Description: i18nMgr.T(i18n.MsgCmdCopy), Aliases: []string{"/cp"}, Group: "config"},
		{Cmd: "/rename", Description: i18nMgr.T(i18n.MsgCmdRename), Aliases: []string{"/mv"}, Group: "config"},
		{Cmd: "/test", Description: i18nMgr.T(i18n.MsgCmdTest), Group: "test"},
		{Cmd: "/export", Description: i18nMgr.T(i18n.MsgCmdExport), Group: "io"},
		{Cmd: "/import", Description: i18nMgr.T(i18n.MsgCmdImport), Group: "io"},
		{Cmd: "/history", Description: i18nMgr.T(i18n.MsgCmdHistory), Group: "util"},
		{Cmd: "/help", Description: i18nMgr.T(i18n.MsgCmdHelp), Aliases: []string{"/?", "/h"}, Group: "util"},
		{Cmd: "/clear", Description: i18nMgr.T(i18n.MsgCmdClear), Aliases: []string{"/cls"}, Group: "util"},
		{Cmd: "/run", Description: i18nMgr.T(i18n.MsgCmdRun), Group: "launch"},
		{Cmd: "/setup", Description: i18nMgr.T(i18n.MsgCmdSetup), Group: "launch"},
		{Cmd: "/exit", Description: i18nMgr.T(i18n.MsgCmdExit), Aliases: []string{"/quit", "/q"}, Group: "launch"},
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

	title := p.styles.PaletteTitle.Render(p.i18n.T(i18n.MsgPaletteTitle))
	sections = append(sections, title)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(mutedColor).
		Padding(0, 1).
		Width(46)
	input := inputStyle.Render("> " + p.query)
	sections = append(sections, input)

	items := p.filteredItems()
	var listLines []string
	for i, item := range items {
		if i >= 10 {
			break
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

	hint := lipgloss.NewStyle().Foreground(mutedColor).Render(p.i18n.T(i18n.MsgSettingsHint))
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

// SetI18n 设置 i18n 管理器
func (p *CommandPalette) SetI18n(i18nMgr *i18n.Manager) {
	p.i18n = i18nMgr
	p.items = getDefaultCommands(i18nMgr)
}

// SetStyles 设置样式
func (p *CommandPalette) SetStyles(styles Styles) {
	p.styles = styles
}

// 键绑定（避免未使用警告）
var _ = key.Binding{}

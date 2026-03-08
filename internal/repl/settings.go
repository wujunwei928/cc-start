// internal/repl/settings.go
package repl

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// SettingsPanel 系统设置面板
type SettingsPanel struct {
	visible  bool
	query    string
	items    []SettingsItem
	selected int
	styles   Styles
	width    int
}

// SettingsItem 设置项
type SettingsItem struct {
	Key         string
	Label       string
	Description string
	Value       string
	Action      string // 动作标识
}

// NewSettingsPanel 创建设置面板
func NewSettingsPanel(styles Styles) *SettingsPanel {
	return &SettingsPanel{
		styles: styles,
		items:  getDefaultSettings(),
	}
}

func getDefaultSettings() []SettingsItem {
	return []SettingsItem{
		{Key: "lang", Label: "语言 / Language", Description: "设置界面语言", Value: "中文", Action: "setting:lang"},
		{Key: "theme", Label: "主题 / Theme", Description: "设置显示主题", Value: "默认", Action: "setting:theme"},
		{Key: "editor", Label: "编辑器 / Editor", Description: "设置默认编辑器", Value: "系统默认", Action: "setting:editor"},
	}
}

// Toggle 切换显示状态
func (s *SettingsPanel) Toggle() {
	s.visible = !s.visible
	if s.visible {
		s.query = ""
		s.selected = 0
	}
}

// IsVisible 返回是否可见
func (s *SettingsPanel) IsVisible() bool {
	return s.visible
}

// SetWidth 设置宽度
func (s *SettingsPanel) SetWidth(w int) {
	s.width = w
}

// HandleKey 处理按键
func (s *SettingsPanel) HandleKey(key string) bool {
	switch key {
	case "up":
		if s.selected > 0 {
			s.selected--
		}
		return true
	case "down":
		if s.selected < len(s.filteredItems())-1 {
			s.selected++
		}
		return true
	case "backspace":
		if len(s.query) > 0 {
			s.query = s.query[:len(s.query)-1]
			s.selected = 0
		}
		return true
	default:
		// 单字符输入
		if len(key) == 1 {
			s.query += key
			s.selected = 0
			return true
		}
	}
	return false
}

// filteredItems 返回过滤后的项
func (s *SettingsPanel) filteredItems() []SettingsItem {
	if s.query == "" {
		return s.items
	}

	// 使用 fuzzy 搜索
	var sources []string
	for _, item := range s.items {
		sources = append(sources, item.Label+" "+item.Description)
	}

	matches := fuzzy.Find(s.query, sources)

	var result []SettingsItem
	for _, match := range matches {
		result = append(result, s.items[match.Index])
	}
	return result
}

// Render 渲染面板
func (s *SettingsPanel) Render() string {
	if !s.visible {
		return ""
	}

	var sections []string

	// 标题
	title := s.styles.PaletteTitle.Render("⚙ 系统设置 / Settings")
	sections = append(sections, title)

	// 输入框
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(mutedColor).
		Padding(0, 1).
		Width(46)
	input := inputStyle.Render("> " + s.query)
	sections = append(sections, input)

	// 设置列表
	items := s.filteredItems()
	var listLines []string
	for i, item := range items {
		if i >= 10 {
			break // 最多显示 10 条
		}
		valueStr := ""
		if item.Value != "" {
			valueStr = " [" + item.Value + "]"
		}
		if i == s.selected {
			line := s.styles.PaletteActive.Render("● " + item.Label + valueStr + "  " + item.Description)
			listLines = append(listLines, line)
		} else {
			line := s.styles.PaletteItem.Render("  " + item.Label + valueStr + "  " + item.Description)
			listLines = append(listLines, line)
		}
	}
	if len(listLines) > 0 {
		sections = append(sections, strings.Join(listLines, "\n"))
	}

	// 提示
	hint := lipgloss.NewStyle().Foreground(mutedColor).Render("↑↓ 导航  enter 确认  esc 关闭")
	sections = append(sections, hint)

	return s.styles.Palette.Render(strings.Join(sections, "\n"))
}

// SelectedAction 返回选中项的动作
func (s *SettingsPanel) SelectedAction() string {
	items := s.filteredItems()
	if s.selected < len(items) {
		return items[s.selected].Action
	}
	return ""
}

// SelectedItem 返回选中项
func (s *SettingsPanel) SelectedItem() *SettingsItem {
	items := s.filteredItems()
	if s.selected < len(items) {
		return &items[s.selected]
	}
	return nil
}

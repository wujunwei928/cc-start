// internal/repl/settings.go
package repl

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"github.com/wujunwei/cc-start/internal/i18n"
	"github.com/wujunwei/cc-start/internal/theme"
)

// SettingsMode 设置面板模式
type SettingsMode int

const (
	SettingsModeMain SettingsMode = iota
	SettingsModeLanguage
	SettingsModeTheme
)

// SettingsPanel 系统设置面板
type SettingsPanel struct {
	visible   bool
	query     string
	items     []SettingsItem
	selected  int
	styles    Styles
	width     int
	mode      SettingsMode
	subItems  []SettingsItem
	i18n      *i18n.Manager
	prevItems []SettingsItem
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
func NewSettingsPanel(styles Styles, i18nMgr *i18n.Manager) *SettingsPanel {
	return &SettingsPanel{
		styles: styles,
		i18n:   i18nMgr,
		items:  getMainSettings(i18nMgr),
		mode:   SettingsModeMain,
	}
}

func getMainSettings(i18nMgr *i18n.Manager) []SettingsItem {
	return []SettingsItem{
		{Key: "lang", Label: i18nMgr.T(i18n.MsgSettingsLanguage), Description: i18nMgr.T(i18n.MsgSettingsLanguage) + " Settings", Value: "", Action: "setting:lang"},
		{Key: "theme", Label: i18nMgr.T(i18n.MsgSettingsTheme), Description: i18nMgr.T(i18n.MsgSettingsTheme) + " Settings", Value: "", Action: "setting:theme"},
	}
}

func getLanguageOptions(i18nMgr *i18n.Manager) []SettingsItem {
	return []SettingsItem{
		{Key: "zh", Label: "中文", Description: "Chinese", Action: "lang:zh"},
		{Key: "en", Label: "English", Description: "English", Action: "lang:en"},
		{Key: "ja", Label: "日本語", Description: "Japanese", Action: "lang:ja"},
	}
}

func getThemeOptions(i18nMgr *i18n.Manager) []SettingsItem {
	themes := theme.GetAllThemes()
	items := make([]SettingsItem, len(themes))
	for i, t := range themes {
		items[i] = SettingsItem{
			Key:         t.Name,
			Label:       t.DisplayName,
			Description: "",
			Action:      "theme:" + t.Name,
		}
	}
	return items
}

// Toggle 切换显示状态
func (s *SettingsPanel) Toggle() {
	s.visible = !s.visible
	if s.visible {
		s.query = ""
		s.selected = 0
		s.mode = SettingsModeMain
		s.items = getMainSettings(s.i18n)
	}
}

// SetI18n 设置 i18n 管理器
func (s *SettingsPanel) SetI18n(i18nMgr *i18n.Manager) {
	s.i18n = i18nMgr
}

// SetStyles 设置样式
func (s *SettingsPanel) SetStyles(styles Styles) {
	s.styles = styles
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

	var title string
	switch s.mode {
	case SettingsModeLanguage:
		title = s.styles.PaletteTitle.Render("🌐 " + s.i18n.T(i18n.MsgSettingsLanguage))
	case SettingsModeTheme:
		title = s.styles.PaletteTitle.Render("🎨 " + s.i18n.T(i18n.MsgSettingsTheme))
	default:
		title = s.styles.PaletteTitle.Render(s.i18n.T(i18n.MsgSettingsTitle))
	}
	sections = append(sections, title)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(mutedColor).
		Padding(0, 1).
		Width(46)
	input := inputStyle.Render("> " + s.query)
	sections = append(sections, input)

	items := s.filteredItems()
	var listLines []string
	for i, item := range items {
		if i >= 15 {
			break
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

	// 添加空行间距
	sections = append(sections, "")

	hint := lipgloss.NewStyle().Foreground(mutedColor).Render(s.i18n.T(i18n.MsgSettingsHint))
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

// EnterSubMenu 进入子菜单
func (s *SettingsPanel) EnterSubMenu(mode SettingsMode) {
	s.prevItems = s.items
	s.mode = mode
	switch mode {
	case SettingsModeLanguage:
		s.items = getLanguageOptions(s.i18n)
	case SettingsModeTheme:
		s.items = getThemeOptions(s.i18n)
	}
	s.selected = 0
	s.query = ""
}

// BackToMain 返回主菜单
func (s *SettingsPanel) BackToMain() {
	s.mode = SettingsModeMain
	if len(s.prevItems) > 0 {
		s.items = s.prevItems
	} else {
		s.items = getMainSettings(s.i18n)
	}
	s.selected = 0
	s.query = ""
}

// GetMode 获取当前模式
func (s *SettingsPanel) GetMode() SettingsMode {
	return s.mode
}

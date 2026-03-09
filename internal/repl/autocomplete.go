// internal/repl/autocomplete.go
package repl

import (
	"strings"

	"github.com/wujunwei928/cc-start/internal/i18n"
)

// Autocomplete 内联命令自动补全组件
type Autocomplete struct {
	visible   bool          // 是否显示
	items     []PaletteItem // 所有命令（复用 PaletteItem）
	filtered  []PaletteItem // 过滤后的命令
	selected  int           // 当前选中索引（相对于 filtered）
	maxShow   int           // 最大显示数量
	scrollOff int           // 滚动偏移
	styles    Styles
	i18n      *i18n.Manager
}

// NewAutocomplete 创建自动补全组件
func NewAutocomplete(styles Styles, i18nMgr *i18n.Manager) *Autocomplete {
	return &Autocomplete{
		styles:  styles,
		i18n:    i18nMgr,
		items:   getDefaultCommands(i18nMgr),
		maxShow: 6,
	}
}

// Show 显示并按前缀过滤
func (a *Autocomplete) Show(prefix string) {
	a.visible = true
	a.selected = 0
	a.scrollOff = 0
	a.Filter(prefix)
}

// Hide 隐藏，重置状态
func (a *Autocomplete) Hide() {
	a.visible = false
	a.selected = 0
	a.scrollOff = 0
	a.filtered = nil
}

// IsVisible 返回可见状态
func (a *Autocomplete) IsVisible() bool {
	return a.visible
}

// Filter 根据前缀过滤命令
func (a *Autocomplete) Filter(prefix string) {
	if !a.visible {
		return
	}

	// 去掉前导 "/" 进行匹配，因为所有命令都以 "/" 开头
	searchPrefix := prefix
	if strings.HasPrefix(prefix, "/") {
		searchPrefix = prefix[1:] // 去掉 "/"
	}

	a.filtered = nil
	for _, item := range a.items {
		cmdWithoutSlash := strings.TrimPrefix(item.Cmd, "/")

		// 前缀匹配（不区分大小写）
		if strings.HasPrefix(strings.ToLower(cmdWithoutSlash), strings.ToLower(searchPrefix)) {
			a.filtered = append(a.filtered, item)
		}

		// 同时检查别名
		for _, alias := range item.Aliases {
			aliasWithoutSlash := strings.TrimPrefix(alias, "/")
			if strings.HasPrefix(strings.ToLower(aliasWithoutSlash), strings.ToLower(searchPrefix)) {
				// 避免重复添加
				found := false
				for _, f := range a.filtered {
					if f.Cmd == item.Cmd {
						found = true
						break
					}
				}
				if !found {
					a.filtered = append(a.filtered, item)
				}
				break
			}
		}
	}

	// 重置选中索引
	if a.selected >= len(a.filtered) {
		a.selected = 0
	}
}

// FilteredItems 返回过滤后的项（用于测试）
func (a *Autocomplete) FilteredItems() []PaletteItem {
	return a.filtered
}

// SelectedIndex 返回当前选中索引（用于测试）
func (a *Autocomplete) SelectedIndex() int {
	return a.selected
}

// SelectUp 向上选择
func (a *Autocomplete) SelectUp() {
	if a.selected > 0 {
		a.selected--
		// 处理滚动
		if a.selected < a.scrollOff {
			a.scrollOff = a.selected
		}
	}
}

// SelectDown 向下选择
func (a *Autocomplete) SelectDown() {
	if a.selected < len(a.filtered)-1 {
		a.selected++
		// 处理滚动
		if a.selected >= a.scrollOff+a.maxShow {
			a.scrollOff = a.selected - a.maxShow + 1
		}
	}
}

// SelectedCommand 返回选中的命令
func (a *Autocomplete) SelectedCommand() string {
	if !a.visible || len(a.filtered) == 0 || a.selected >= len(a.filtered) {
		return ""
	}
	return a.filtered[a.selected].Cmd
}

// Render 渲染下拉列表
func (a *Autocomplete) Render(width int) string {
	if !a.visible || len(a.filtered) == 0 {
		return ""
	}

	// 计算要显示的项
	start := a.scrollOff
	end := start + a.maxShow
	if end > len(a.filtered) {
		end = len(a.filtered)
	}

	var lines []string
	for i := start; i < end; i++ {
		item := a.filtered[i]
		text := item.Cmd + "  " + item.Description

		if i == a.selected {
			lines = append(lines, a.styles.AutocompleteActive.Render("● "+text))
		} else {
			lines = append(lines, a.styles.AutocompleteItem.Render("  "+text))
		}
	}

	content := strings.Join(lines, "\n")

	// 添加边框
	rendered := a.styles.AutocompleteBorder.
		Width(width - 2).
		Render(content)

	return rendered
}

// SetI18n 设置 i18n 管理器
func (a *Autocomplete) SetI18n(i18nMgr *i18n.Manager) {
	a.i18n = i18nMgr
	a.items = getDefaultCommands(i18nMgr)
}

// SetStyles 设置样式
func (a *Autocomplete) SetStyles(styles Styles) {
	a.styles = styles
}

// internal/repl/autocomplete.go
package repl

import (
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
	// TODO: 后续实现过滤逻辑
	a.filtered = a.items
}

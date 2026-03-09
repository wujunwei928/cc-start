# REPL 命令内联自动补全实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 REPL 的命令交互从模态弹窗改为内联下拉补全，提升交互流畅度。

**Architecture:** 创建独立的 Autocomplete 组件，在输入框下方渲染命令列表；用户输入 "/" 触发显示，Tab 补全选中命令，继续输入参数后回车执行。保持非模态设计，用户可以随时继续输入或按 Esc 关闭。

**Tech Stack:** Go 1.21+, Bubble Tea (tea), Lip Gloss (样式), 复用现有 PaletteItem 结构

---

## Task 1: 创建 Autocomplete 组件

**Files:**
- Create: `internal/repl/autocomplete.go`
- Test: `internal/repl/autocomplete_test.go`

### Step 1: 编写 Autocomplete 结构体和基础方法的测试

```go
// internal/repl/autocomplete_test.go
package repl

import (
	"testing"

	"github.com/wujunwei928/cc-start/internal/i18n"
)

func TestAutocompleteNew(t *testing.T) {
	i18nMgr := i18n.NewManager()
	styles := DefaultStyles()

	ac := NewAutocomplete(styles, i18nMgr)

	if ac == nil {
		t.Fatal("NewAutocomplete 返回 nil")
	}
	if ac.IsVisible() {
		t.Error("新建的 Autocomplete 不应该可见")
	}
	if ac.maxShow != 6 {
		t.Errorf("maxShow 应该是 6，实际是 %d", ac.maxShow)
	}
}

func TestAutocompleteShowHide(t *testing.T) {
	i18nMgr := i18n.NewManager()
	styles := DefaultStyles()

	ac := NewAutocomplete(styles, i18nMgr)

	// 测试 Show
	ac.Show("/")
	if !ac.IsVisible() {
		t.Error("Show 后应该可见")
	}

	// 测试 Hide
	ac.Hide()
	if ac.IsVisible() {
		t.Error("Hide 后不应该可见")
	}
}
```

### Step 2: 运行测试验证失败

```bash
cd /code/ai/cc-start && go test ./internal/repl/... -run TestAutocomplete -v
```

Expected: 编译失败，`NewAutocomplete` 未定义

### Step 3: 实现 Autocomplete 结构体和基础方法

```go
// internal/repl/autocomplete.go
package repl

import (
	"strings"

	"github.com/wujunwei928/cc-start/internal/i18n"
)

// Autocomplete 内联命令自动补全组件
type Autocomplete struct {
	visible  bool          // 是否显示
	items    []PaletteItem // 所有命令（复用 PaletteItem）
	filtered []PaletteItem // 过滤后的命令
	selected int           // 当前选中索引（相对于 filtered）
	maxShow  int           // 最大显示数量
	scrollOff int          // 滚动偏移
	styles   Styles
	i18n     *i18n.Manager
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
```

### Step 4: 运行测试验证通过

```bash
cd /code/ai/cc-start && go test ./internal/repl/... -run TestAutocompleteNew -v
```

Expected: PASS

### Step 5: 提交基础结构

```bash
git add internal/repl/autocomplete.go internal/repl/autocomplete_test.go
git commit -m "feat(autocomplete): 添加 Autocomplete 组件基础结构

- 定义 Autocomplete 结构体
- 实现 NewAutocomplete、Show、Hide、IsVisible 方法
- 添加基础单元测试"
```

---

## Task 2: 实现过滤和选择方法

**Files:**
- Modify: `internal/repl/autocomplete.go`
- Modify: `internal/repl/autocomplete_test.go`

### Step 1: 编写过滤和选择方法的测试

```go
// 添加到 internal/repl/autocomplete_test.go

func TestAutocompleteFilter(t *testing.T) {
	i18nMgr := i18n.NewManager()
	styles := DefaultStyles()

	ac := NewAutocomplete(styles, i18nMgr)

	// 测试 "/" 过滤 - 应该显示所有以 "/" 开头的命令
	ac.Show("/")
	filtered := ac.FilteredItems()
	if len(filtered) == 0 {
		t.Error("过滤 '/' 应该返回至少一个命令")
	}
	// 检查所有返回的命令都以 "/" 开头
	for _, item := range filtered {
		if !strings.HasPrefix(item.Cmd, "/") {
			t.Errorf("命令 %s 不以 '/' 开头", item.Cmd)
		}
	}

	// 测试 "/u" 过滤 - 应该只显示 /use 等
	ac.Filter("/u")
	filtered = ac.FilteredItems()
	if len(filtered) == 0 {
		t.Error("过滤 '/u' 应该返回至少一个命令")
	}
	for _, item := range filtered {
		if !strings.HasPrefix(item.Cmd, "/u") {
			t.Errorf("命令 %s 不以 '/u' 开头", item.Cmd)
		}
	}

	// 测试无匹配的情况
	ac.Filter("/xyz123notexist")
	filtered = ac.FilteredItems()
	if len(filtered) != 0 {
		t.Errorf("过滤不存在的命令应该返回空，实际返回 %d 条", len(filtered))
	}
}

func TestAutocompleteSelectUpAndDown(t *testing.T) {
	i18nMgr := i18n.NewManager()
	styles := DefaultStyles()

	ac := NewAutocomplete(styles, i18nMgr)
	ac.Show("/")

	// 初始选中应该是 0
	if ac.SelectedIndex() != 0 {
		t.Errorf("初始选中应该是 0，实际是 %d", ac.SelectedIndex())
	}

	// 在第一项时向上不应该越界
	ac.SelectUp()
	if ac.SelectedIndex() != 0 {
		t.Errorf("在第一项时向上应该保持在 0，实际是 %d", ac.SelectedIndex())
	}

	// 向下选择
	ac.SelectDown()
	if ac.SelectedIndex() != 1 {
		t.Errorf("向下选择后应该是 1，实际是 %d", ac.SelectedIndex())
	}

	// 再向上回到第一项
	ac.SelectUp()
	if ac.SelectedIndex() != 0 {
		t.Errorf("向上选择后应该是 0，实际是 %d", ac.SelectedIndex())
	}
}

func TestAutocompleteSelectedCommand(t *testing.T) {
	i18nMgr := i18n.NewManager()
	styles := DefaultStyles()

	ac := NewAutocomplete(styles, i18nMgr)
	ac.Show("/")

	// 选中第一个命令
	cmd := ac.SelectedCommand()
	if cmd == "" {
		t.Error("SelectedCommand 不应该返回空")
	}
	if !strings.HasPrefix(cmd, "/") {
		t.Errorf("命令应该以 '/' 开头，实际是 %s", cmd)
	}

	// 隐藏后应该返回空
	ac.Hide()
	cmd = ac.SelectedCommand()
	if cmd != "" {
		t.Errorf("隐藏后 SelectedCommand 应该返回空，实际返回 %s", cmd)
	}
}
```

### Step 2: 运行测试验证失败

```bash
cd /code/ai/cc-start && go test ./internal/repl/... -run "TestAutocompleteFilter|TestAutocompleteSelect" -v
```

Expected: 编译失败，方法未定义

### Step 3: 实现过滤和选择方法

```go
// 添加到 internal/repl/autocomplete.go

// Filter 前缀匹配过滤
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
```

### Step 4: 运行测试验证通过

```bash
cd /code/ai/cc-start && go test ./internal/repl/... -run "TestAutocompleteFilter|TestAutocompleteSelect" -v
```

Expected: PASS

### Step 5: 提交过滤和选择方法

```bash
git add internal/repl/autocomplete.go internal/repl/autocomplete_test.go
git commit -m "feat(autocomplete): 实现过滤和选择方法

- Filter: 前缀匹配过滤，支持命令和别名
- SelectUp/SelectDown: 上下选择，支持滚动
- SelectedCommand: 返回选中的命令"
```

---

## Task 3: 实现 Autocomplete 渲染方法

**Files:**
- Modify: `internal/repl/autocomplete.go`
- Modify: `internal/repl/autocomplete_test.go`
- Modify: `internal/repl/styles.go`

### Step 1: 编写渲染方法测试

```go
// 添加到 internal/repl/autocomplete_test.go

func TestAutocompleteRender(t *testing.T) {
	i18nMgr := i18n.NewManager()
	styles := DefaultStyles()

	ac := NewAutocomplete(styles, i18nMgr)

	// 隐藏时应该返回空
	rendered := ac.Render(80)
	if rendered != "" {
		t.Errorf("隐藏时 Render 应该返回空，实际返回: %s", rendered)
	}

	// 显示时应该返回内容
	ac.Show("/")
	rendered = ac.Render(80)
	if rendered == "" {
		t.Error("显示时 Render 应该返回内容")
	}

	// 应该包含命令
	if !strings.Contains(rendered, "/list") && !strings.Contains(rendered, "/use") {
		t.Error("渲染结果应该包含命令")
	}
}

func TestAutocompleteRenderWidth(t *testing.T) {
	i18nMgr := i18n.NewManager()
	styles := DefaultStyles()

	ac := NewAutocomplete(styles, i18nMgr)
	ac.Show("/")

	// 测试不同宽度
	rendered := ac.Render(40)
	if rendered == "" {
		t.Error("宽度 40 应该能渲染")
	}

	rendered = ac.Render(100)
	if rendered == "" {
		t.Error("宽度 100 应该能渲染")
	}
}
```

### Step 2: 添加 Autocomplete 样式到 Styles 结构体

```go
// 修改 internal/repl/styles.go

// 在 Styles 结构体中添加（PaletteActive 后面）:
	// 内联自动补全
	AutocompleteItem   lipgloss.Style
	AutocompleteActive lipgloss.Style
	AutocompleteBorder lipgloss.Style

// 在 DefaultStyles() 函数中添加（return 语句前）:
		AutocompleteItem: lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 1),

		AutocompleteActive: lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Padding(0, 1),

		AutocompleteBorder: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(mutedColor),

// 在 NewStylesFromTheme() 函数中添加（return 语句前）:
		AutocompleteItem: lipgloss.NewStyle().
			Foreground(fg).
			Padding(0, 1),

		AutocompleteActive: lipgloss.NewStyle().
			Foreground(highlight).
			Background(lipgloss.Color(t.Colors.PaletteActive)).
			Bold(true).
			Padding(0, 1),

		AutocompleteBorder: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(muted),
```

### Step 3: 运行测试验证失败

```bash
cd /code/ai/cc-start && go test ./internal/repl/... -run "TestAutocompleteRender" -v
```

Expected: 编译失败，`Render` 方法未定义

### Step 4: 实现 Render 方法

```go
// 添加到 internal/repl/autocomplete.go

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/wujunwei928/cc-start/internal/i18n"
)

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
```

### Step 5: 运行测试验证通过

```bash
cd /code/ai/cc-start && go test ./internal/repl/... -run "TestAutocompleteRender" -v
```

Expected: PASS

### Step 6: 提交渲染方法

```bash
git add internal/repl/autocomplete.go internal/repl/autocomplete_test.go internal/repl/styles.go
git commit -m "feat(autocomplete): 实现 Render 方法

- 添加 AutocompleteItem、AutocompleteActive、AutocompleteBorder 样式
- 支持滚动显示和选中高亮
- 限制显示数量为 6 条"
```

---

## Task 4: 添加 Tab 按键绑定到 Model

**Files:**
- Modify: `internal/repl/model.go`
- Modify: `internal/repl/autocomplete.go`

### Step 1: 修改 keyMap 结构体添加 Tab 绑定

```go
// 修改 internal/repl/model.go

// 在 keyMap 结构体中添加（Space 后面）:
	Tab key.Binding

// 在 defaultKeyMap() 函数中添加（Space 后面）:
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "补全命令"),
		),
```

### Step 2: 添加 autocomplete 字段到 Model 结构体

```go
// 修改 internal/repl/model.go

// 在 Model 结构体中添加（palette 后面）:
	autocomplete *Autocomplete
```

### Step 3: 在 NewModel 中初始化 autocomplete

```go
// 修改 internal/repl/model.go NewModel 函数

// 在 return 语句前添加:
	ac := NewAutocomplete(styles, i18nMgr)

// 修改 return 语句，添加:
		autocomplete: ac,
```

### Step 4: 添加 SetI18n 和 SetStyles 方法到 Autocomplete

```go
// 添加到 internal/repl/autocomplete.go

// SetI18n 设置 i18n 管理器
func (a *Autocomplete) SetI18n(i18nMgr *i18n.Manager) {
	a.i18n = i18nMgr
	a.items = getDefaultCommands(i18nMgr)
}

// SetStyles 设置样式
func (a *Autocomplete) SetStyles(styles Styles) {
	a.styles = styles
}
```

### Step 5: 运行测试验证编译通过

```bash
cd /code/ai/cc-start && go build ./...
```

Expected: 编译成功

### Step 6: 提交 Model 改动

```bash
git add internal/repl/model.go internal/repl/autocomplete.go
git commit -m "feat(autocomplete): 添加 autocomplete 字段和 Tab 按键绑定

- Model 添加 autocomplete *Autocomplete 字段
- keyMap 添加 Tab 绑定
- NewModel 初始化 autocomplete 组件"
```

---

## Task 5: 重构 Update 键盘事件处理

**Files:**
- Modify: `internal/repl/update.go`
- Modify: `internal/repl/autocomplete_test.go`

### Step 1: 编写集成测试验证键盘交互

```go
// 添加到 internal/repl/autocomplete_test.go

import (
	tea "github.com/charmbracelet/bubbletea"
)

func TestAutocompleteSlashTrigger(t *testing.T) {
	model, err := NewModel("")
	if err != nil {
		t.Fatalf("创建模型失败: %v", err)
	}

	// 模拟输入 "/"
	model.input.SetValue("")
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

	m := updatedModel.(Model)
	if m.autocomplete == nil || !m.autocomplete.IsVisible() {
		t.Error("输入 '/' 应该显示自动补全")
	}
}

func TestAutocompleteTabComplete(t *testing.T) {
	model, err := NewModel("")
	if err != nil {
		t.Fatalf("创建模型失败: %v", err)
	}

	// 设置初始状态：输入 "/" 并显示自动补全
	model.input.SetValue("/")
	if model.autocomplete == nil {
		model.autocomplete = NewAutocomplete(model.Styles, model.I18n)
	}
	model.autocomplete.Show("/")

	// 模拟 Tab 键
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})

	m := updatedModel.(Model)
	if m.autocomplete.IsVisible() {
		t.Error("Tab 补全后应该关闭自动补全")
	}

	inputValue := m.input.Value()
	if !strings.HasPrefix(inputValue, "/") {
		t.Errorf("Tab 补全后输入框应该包含命令，实际是: %s", inputValue)
	}
	// 应该以空格结尾，方便输入参数
	if !strings.HasSuffix(inputValue, " ") {
		t.Errorf("Tab 补全后应该以空格结尾，实际是: %s", inputValue)
	}
}

func TestAutocompleteEscapeClose(t *testing.T) {
	model, err := NewModel("")
	if err != nil {
		t.Fatalf("创建模型失败: %v", err)
	}

	// 设置初始状态
	model.input.SetValue("/")
	if model.autocomplete == nil {
		model.autocomplete = NewAutocomplete(model.Styles, model.I18n)
	}
	model.autocomplete.Show("/")

	// 模拟 Esc 键
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEscape})

	m := updatedModel.(Model)
	if m.autocomplete.IsVisible() {
		t.Error("Esc 应该关闭自动补全")
	}
	// 输入框内容应该保留
	if m.input.Value() != "/" {
		t.Errorf("Esc 后输入框内容应该保留，实际是: %s", m.input.Value())
	}
}
```

### Step 2: 运行测试验证失败

```bash
cd /code/ai/cc-start && go test ./internal/repl/... -run "TestAutocompleteSlash|TestAutocompleteTab|TestAutocompleteEscape" -v
```

Expected: 测试失败，键盘事件处理未实现

### Step 3: 重构 Update 函数处理自动补全

```go
// 修改 internal/repl/update.go Update 函数

// 在 tea.KeyMsg switch 的最开始添加（在设置面板检查之后）:
		// 自动补全激活时的特殊按键处理
		if m.autocomplete != nil && m.autocomplete.IsVisible() {
			switch {
			case keyMatches(msg, m.keys.Tab):
				// Tab 补全选中命令
				cmd := m.autocomplete.SelectedCommand()
				if cmd != "" {
					m.input.SetValue(cmd + " ")
					m.input.CursorEnd()
					m.autocomplete.Hide()
				}
				return m, nil

			case keyMatches(msg, m.keys.Up):
				m.autocomplete.SelectUp()
				return m, nil

			case keyMatches(msg, m.keys.Down):
				m.autocomplete.SelectDown()
				return m, nil

			case keyMatches(msg, m.keys.Esc):
				m.autocomplete.Hide()
				return m, nil

			case keyMatches(msg, m.keys.Enter):
				// 回车执行命令，先关闭自动补全
				m.autocomplete.Hide()
				return m.executeInput()
			}
		}
```

### Step 4: 修改默认字符输入处理

```go
// 修改 internal/repl/update.go Update 函数的 default 分支

// 替换现有的 default 分支:
		default:
			// 先更新输入框
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)

			// 检查是否需要触发/更新自动补全
			currentInput := m.input.Value()
			if strings.HasPrefix(currentInput, "/") {
				if m.autocomplete == nil {
					m.autocomplete = NewAutocomplete(m.Styles, m.I18n)
				}
				if !m.autocomplete.IsVisible() {
					m.autocomplete.Show(currentInput)
				} else {
					m.autocomplete.Filter(currentInput)
				}
			} else if m.autocomplete != nil && m.autocomplete.IsVisible() {
				m.autocomplete.Hide()
			}

			return m, cmd
```

### Step 5: 更新语言/主题切换时同步 autocomplete

```go
// 修改 internal/repl/update.go applyLanguageChange 函数

// 在现有的 if m.settings != nil 块后添加:
	if m.autocomplete != nil {
		m.autocomplete.SetI18n(m.I18n)
	}

// 修改 applyThemeChange 函数，在现有的 if m.palette != nil 块后添加:
	if m.autocomplete != nil {
		m.autocomplete.SetStyles(m.Styles)
	}
```

### Step 6: 运行测试验证通过

```bash
cd /code/ai/cc-start && go test ./internal/repl/... -run "TestAutocompleteSlash|TestAutocompleteTab|TestAutocompleteEscape" -v
```

Expected: PASS

### Step 7: 提交 Update 重构

```bash
git add internal/repl/update.go internal/repl/autocomplete_test.go
git commit -m "feat(autocomplete): 重构 Update 键盘事件处理

- Tab: 补全选中命令到输入框
- Up/Down: 在自动补全列表中导航
- Esc: 关闭自动补全，保留输入
- Enter: 执行命令前先关闭自动补全
- 字符输入: 自动触发/更新/隐藏自动补全"
```

---

## Task 6: 更新 View 渲染自动补全列表

**Files:**
- Modify: `internal/repl/view.go`
- Modify: `internal/repl/view_test.go`

### Step 1: 编写 View 渲染测试

```go
// 添加到 internal/repl/view_test.go

func TestAutocompleteInView(t *testing.T) {
	model, err := NewModel("")
	if err != nil {
		t.Fatalf("创建模型失败: %v", err)
	}

	model.currentProfile = "test"
	model.width = 80

	// 初始状态不应该有自动补全
	view := model.View()
	if strings.Contains(view, "● /") {
		t.Error("初始状态不应该显示自动补全")
	}

	// 触发自动补全
	model.input.SetValue("/")
	if model.autocomplete == nil {
		model.autocomplete = NewAutocomplete(model.Styles, model.I18n)
	}
	model.autocomplete.Show("/")

	view = model.View()
	if !strings.Contains(view, "/list") && !strings.Contains(view, "/use") {
		t.Error("显示自动补全时应该包含命令")
	}
}

func TestAutocompleteHelpBarDynamic(t *testing.T) {
	model, err := NewModel("")
	if err != nil {
		t.Fatalf("创建模型失败: %v", err)
	}

	model.currentProfile = "test"
	model.width = 80

	// 初始帮助栏
	helpBar := model.renderHelpBar()
	if !strings.Contains(helpBar, "/ commands") {
		t.Error("初始帮助栏应该包含 '/ commands'")
	}

	// 显示自动补全时的帮助栏
	if model.autocomplete == nil {
		model.autocomplete = NewAutocomplete(model.Styles, model.I18n)
	}
	model.autocomplete.Show("/")

	helpBar = model.renderHelpBar()
	if !strings.Contains(helpBar, "tab complete") {
		t.Errorf("自动补全显示时帮助栏应该包含 'tab complete'，实际是: %s", helpBar)
	}
}
```

### Step 2: 运行测试验证失败

```bash
cd /code/ai/cc-start && go test ./internal/repl/... -run "TestAutocompleteInView|TestAutocompleteHelpBarDynamic" -v
```

Expected: 测试失败

### Step 3: 修改 View 函数渲染自动补全

```go
// 修改 internal/repl/view.go View 函数

// 在输入区渲染后、帮助栏渲染前添加（在 sections = append(sections, inputLine) 后面）:
	// 自动补全列表（在输入框下方）
	if m.autocomplete != nil && m.autocomplete.IsVisible() {
		acView := m.autocomplete.Render(m.width)
		if acView != "" {
			sections = append(sections, acView)
		}
	}
```

### Step 4: 修改 renderHelpBar 函数支持动态提示

```go
// 修改 internal/repl/view.go renderHelpBar 函数

func (m Model) renderHelpBar() string {
	var hints []string

	if m.settings != nil && m.settings.IsVisible() {
		hints = []string{"up/down navigate", "enter confirm", "esc close"}
	} else if m.autocomplete != nil && m.autocomplete.IsVisible() {
		hints = []string{"↑↓ navigate", "tab complete", "esc close", "enter execute"}
	} else {
		hints = []string{
			"/ commands",
			"ctrl+p settings",
			"up/down history",
			"enter execute",
			"ctrl+c exit",
		}
	}

	return m.Styles.HelpBar.Render(strings.Join(hints, "  "))
}
```

### Step 5: 运行测试验证通过

```bash
cd /code/ai/cc-start && go test ./internal/repl/... -run "TestAutocompleteInView|TestAutocompleteHelpBarDynamic" -v
```

Expected: PASS

### Step 6: 提交 View 更新

```bash
git add internal/repl/view.go internal/repl/view_test.go
git commit -m "feat(autocomplete): 更新 View 渲染自动补全列表

- 在输入框下方渲染自动补全列表
- 帮助栏根据自动补全状态动态显示提示"
```

---

## Task 7: 运行完整测试套件

**Files:**
- 无文件修改，仅验证

### Step 1: 运行所有 REPL 测试

```bash
cd /code/ai/cc-start && go test ./internal/repl/... -v -timeout 60s
```

Expected: 所有测试 PASS

### Step 2: 运行完整项目测试

```bash
cd /code/ai/cc-start && go test ./... -v -timeout 60s
```

Expected: 所有测试 PASS

### Step 3: 构建验证

```bash
cd /code/ai/cc-start && go build ./...
```

Expected: 编译成功

### Step 4: 提交测试验证

```bash
git add -A
git commit -m "test(autocomplete): 验证所有测试通过

- 运行完整 REPL 测试套件
- 运行项目所有测试
- 构建验证通过"
```

---

## Task 8: 清理旧的模态面板逻辑（可选）

**Files:**
- Modify: `internal/repl/model.go`
- Modify: `internal/repl/update.go`
- Modify: `internal/repl/view.go`
- Keep: `internal/repl/palette.go` (保留供设置面板复用样式)

### Step 1: 移除 palette 字段和相关逻辑

**注意：** 由于 palette 样式被 settings 面板复用，只移除 palette 的使用逻辑，保留文件。

```go
// 修改 internal/repl/model.go
// 移除 Model 结构体中的 palette *CommandPalette 字段

// 修改 internal/repl/update.go
// 移除 updatePalette 函数及其调用
// 移除 palette 相关的检查逻辑
// 移除 applyLanguageChange 和 applyThemeChange 中的 palette.SetI18n/SetStyles 调用

// 修改 internal/repl/view.go
// 移除 View 函数中的 palette 渲染逻辑
```

### Step 2: 运行测试验证

```bash
cd /code/ai/cc-start && go test ./internal/repl/... -v -timeout 60s
```

Expected: 所有测试 PASS

### Step 3: 提交清理

```bash
git add internal/repl/model.go internal/repl/update.go internal/repl/view.go
git commit -m "refactor(autocomplete): 移除旧的模态面板逻辑

- 移除 palette 字段和相关调用
- 保留 palette.go 供设置面板复用样式"
```

---

## 实现后验证清单

- [ ] 输入 "/" 触发显示命令列表
- [ ] 输入 "/u" 过滤只显示 /use, /setup 等命令
- [ ] Tab 补全选中命令到输入框并添加空格
- [ ] ↑↓ 在列表中导航，不影响输入框光标
- [ ] Esc 关闭列表，保留输入框内容
- [ ] 回车执行命令
- [ ] 帮助栏根据状态动态显示提示
- [ ] 删除字符时自动更新过滤
- [ ] 删除 "/" 后自动关闭列表

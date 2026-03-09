// internal/repl/view.go
package repl

import (
	"fmt"
	"strings"
)

// View 渲染 UI
func (m Model) View() string {
	if m.quitting {
		return "再见!\n"
	}

	var sections []string

	// 输出区（限制高度，确保输入行始终可见）
	outputContent := m.renderOutput()
	if outputContent != "" {
		sections = append(sections, outputContent)
	}

	// 设置面板（覆盖层）
	if m.settings != nil && m.settings.IsVisible() {
		settingsView := m.settings.Render()
		sections = append(sections, "\n"+settingsView)
		return strings.Join(sections, "\n") + "\n"
	}

	// 命令面板（覆盖层）
	if m.palette != nil && m.palette.IsVisible() {
		paletteView := m.palette.Render()
		sections = append(sections, "\n"+paletteView)
		return strings.Join(sections, "\n") + "\n"
	}

	// 输入区 - 直接拼接，不使用 lipgloss 组合
	prefix := m.Styles.Prefix.Render(m.getPromptPrefix())
	inputLine := prefix + m.input.View()
	sections = append(sections, inputLine)

	// 自动补全列表（在输入框下方）
	if m.autocomplete != nil && m.autocomplete.IsVisible() {
		acView := m.autocomplete.Render(m.width)
		if acView != "" {
			sections = append(sections, acView)
		}
	}

	// 帮助栏
	helpBar := m.renderHelpBar()
	sections = append(sections, "", helpBar)

	return strings.Join(sections, "\n") + "\n"
}

// renderOutput 渲染输出区（限制高度）
func (m Model) renderOutput() string {
	outputContent := m.output.Render(m.Styles, m.width)
	if outputContent == "" {
		return ""
	}

	// 计算可用高度（保留空间给输入行和帮助栏）
	availableHeight := m.height - 3 // 输入行(1) + 空行(1) + 帮助栏(1)
	if availableHeight < 5 {
		availableHeight = 5 // 最小高度
	}

	// 按行分割
	lines := strings.Split(outputContent, "\n")

	// 移除末尾的空行（由strings.Split产生的）
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// 只显示最近的 N 行
	if len(lines) > availableHeight {
		lines = lines[len(lines)-availableHeight:]
	}

	// 重新组合
	if len(lines) == 0 {
		return ""
	}

	outputContent = strings.Join(lines, "\n") + "\n"
	return m.Styles.Output.Render(outputContent)
}

func (m Model) getPromptPrefix() string {
	if m.currentProfile != "" {
		return fmt.Sprintf("cc-start [%s]>", m.currentProfile)
	}
	return "cc-start>"
}

func (m Model) renderHelpBar() string {
	var hints []string

	if m.settings != nil && m.settings.IsVisible() {
		hints = []string{"up/down navigate", "enter confirm", "esc close"}
	} else if m.palette != nil && m.palette.IsVisible() {
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

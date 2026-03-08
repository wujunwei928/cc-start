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

	// 设置面板（覆盖层）
	if m.settings != nil && m.settings.IsVisible() {
		settingsView := m.settings.Render()
		sections = append(sections, "\n"+settingsView)
		return lipgloss.JoinVertical(lipgloss.Left, sections...) + "\n"
	}

	// 命令面板（覆盖层）
	if m.palette != nil && m.palette.IsVisible() {
		paletteView := m.palette.Render()
		sections = append(sections, "\n"+paletteView)
		return lipgloss.JoinVertical(lipgloss.Left, sections...) + "\n"
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

	if m.settings != nil && m.settings.IsVisible() {
		hints = []string{"up/down navigate", "enter confirm", "esc close"}
	} else if m.palette != nil && m.palette.IsVisible() {
		hints = []string{"up/down navigate", "enter confirm", "esc close"}
	} else {
		hints = []string{
			"/ commands",
			"ctrl+p settings",
			"up/down history",
			"enter execute",
			"ctrl+c exit",
		}
	}

	return m.styles.HelpBar.Render(strings.Join(hints, "  "))
}

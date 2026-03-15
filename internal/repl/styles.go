// internal/repl/styles.go
package repl

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/wujunwei928/cc-start/internal/theme"
)

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
	App     lipgloss.Style
	Prefix  lipgloss.Style
	Input   lipgloss.Style
	Output  lipgloss.Style
	HelpBar lipgloss.Style

	// 输出样式
	Success   lipgloss.Style
	Error     lipgloss.Style
	Warning   lipgloss.Style
	Info      lipgloss.Style
	Command   lipgloss.Style
	Separator lipgloss.Style

	// 命令面板
	Palette       lipgloss.Style
	PaletteTitle  lipgloss.Style
	PaletteInput  lipgloss.Style
	PaletteList   lipgloss.Style
	PaletteItem   lipgloss.Style
	PaletteActive lipgloss.Style

	// 内联自动补全
	AutocompleteItem   lipgloss.Style
	AutocompleteActive lipgloss.Style
	AutocompleteBorder lipgloss.Style

	// 基础颜色（用于动态样式）
	MutedColor lipgloss.TerminalColor
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
			Foreground(textColor),

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

		Command: lipgloss.NewStyle().
			Foreground(mutedColor).
			Bold(true),

		Separator: lipgloss.NewStyle().
			Foreground(mutedColor),

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

		MutedColor: mutedColor,
	}
}

// NewStylesFromTheme 从主题创建样式
func NewStylesFromTheme(t *theme.Theme) Styles {
	fg := lipgloss.Color(t.Colors.Foreground)
	muted := lipgloss.Color(t.Colors.Muted)
	accent := lipgloss.Color(t.Colors.Accent)
	success := lipgloss.Color(t.Colors.Success)
	errorCol := lipgloss.Color(t.Colors.Error)
	warning := lipgloss.Color(t.Colors.Warning)
	info := lipgloss.Color(t.Colors.Info)
	primary := lipgloss.Color(t.Colors.Primary)
	highlight := lipgloss.Color(t.Colors.Highlight)

	return Styles{
		App: lipgloss.NewStyle().
			Padding(1, 2),

		Prefix: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),

		Input: lipgloss.NewStyle().
			Foreground(fg),

		Output: lipgloss.NewStyle().
			Foreground(fg),

		HelpBar: lipgloss.NewStyle().
			Foreground(muted).
			Padding(0, 1),

		Success: lipgloss.NewStyle().
			Foreground(success).
			SetString("✓"),

		Error: lipgloss.NewStyle().
			Foreground(errorCol).
			SetString("✗"),

		Warning: lipgloss.NewStyle().
			Foreground(warning).
			SetString("⚠"),

		Info: lipgloss.NewStyle().
			Foreground(info).
			SetString("●"),

		Command: lipgloss.NewStyle().
			Foreground(muted).
			Bold(true),

		Separator: lipgloss.NewStyle().
			Foreground(muted),

		Palette: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(muted).
			Padding(0, 1).
			Width(50),

		PaletteTitle: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			Padding(0, 1),

		PaletteInput: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(muted).
			Padding(0, 1).
			Margin(1, 0),

		PaletteItem: lipgloss.NewStyle().
			Foreground(fg).
			Background(lipgloss.Color(t.Colors.PaletteInactive)).
			Padding(0, 2),

		PaletteActive: lipgloss.NewStyle().
			Foreground(highlight).
			Background(lipgloss.Color(t.Colors.PaletteActive)).
			Bold(true).
			Padding(0, 2),

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

		MutedColor: muted,
	}
}

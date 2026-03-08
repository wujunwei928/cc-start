// internal/theme/theme.go
package theme

import (
	"fmt"
)

// ColorScheme 颜色方案
type ColorScheme struct {
	Background string
	Foreground string
	Muted      string

	Primary  string
	Success  string
	Error    string
	Warning  string
	Info     string

	Border    string
	Accent    string
	Highlight string

	PaletteBg       string
	PaletteActive   string
	PaletteInactive string
}

// Theme 主题定义
type Theme struct {
	Name        string
	DisplayName string
	Colors      ColorScheme
}

// GetTheme 获取指定名称的主题
func GetTheme(name string) (*Theme, error) {
	for i := range presets {
		if presets[i].Name == name {
			return &presets[i], nil
		}
	}
	return nil, fmt.Errorf("theme '%s' not found", name)
}

// GetAllThemes 获取所有预设主题
func GetAllThemes() []Theme {
	return presets
}

// ApplyTheme 将主题颜色应用到样式映射
func ApplyTheme(t *Theme) map[string]string {
	return map[string]string{
		"background":     t.Colors.Background,
		"foreground":     t.Colors.Foreground,
		"primary":        t.Colors.Primary,
		"success":        t.Colors.Success,
		"error":          t.Colors.Error,
		"warning":        t.Colors.Warning,
		"info":           t.Colors.Info,
		"muted":          t.Colors.Muted,
		"border":         t.Colors.Border,
		"accent":         t.Colors.Accent,
		"highlight":      t.Colors.Highlight,
		"palette_bg":     t.Colors.PaletteBg,
		"palette_active": t.Colors.PaletteActive,
	}
}

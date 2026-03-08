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

// internal/theme/theme_test.go
package theme

import (
	"testing"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name    string
		theme   string
		wantErr bool
	}{
		{"default", "default", false},
		{"ocean", "ocean", false},
		{"forest", "forest", false},
		{"sunset", "sunset", false},
		{"light", "light", false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme, err := GetTheme(tt.theme)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTheme(%s) error = %v, wantErr %v", tt.theme, err, tt.wantErr)
				return
			}
			if !tt.wantErr && theme == nil {
				t.Errorf("GetTheme(%s) returned nil", tt.theme)
			}
		})
	}
}

func TestGetAllThemes(t *testing.T) {
	themes := GetAllThemes()

	if len(themes) != 5 {
		t.Errorf("GetAllThemes() returned %d themes, want 5", len(themes))
	}

	// 验证每个主题都有必要的字段
	for _, theme := range themes {
		if theme.Name == "" {
			t.Error("Theme has empty Name")
		}
		if theme.DisplayName == "" {
			t.Error("Theme has empty DisplayName")
		}
		if theme.Colors.Background == "" {
			t.Error("Theme has empty Background color")
		}
	}
}

func TestPresetThemesColors(t *testing.T) {
	themes := GetAllThemes()

	requiredColors := []string{
		"Background", "Foreground", "Primary", "Success", "Error",
		"Warning", "Info", "Border", "Accent", "Highlight",
	}

	for _, theme := range themes {
		t.Run(theme.Name, func(t *testing.T) {
			colors := map[string]string{
				"Background": theme.Colors.Background,
				"Foreground": theme.Colors.Foreground,
				"Primary":    theme.Colors.Primary,
				"Success":    theme.Colors.Success,
				"Error":      theme.Colors.Error,
				"Warning":    theme.Colors.Warning,
				"Info":       theme.Colors.Info,
				"Border":     theme.Colors.Border,
				"Accent":     theme.Colors.Accent,
				"Highlight":  theme.Colors.Highlight,
			}

			for _, colorName := range requiredColors {
				if colors[colorName] == "" {
					t.Errorf("Theme %s missing color: %s", theme.Name, colorName)
				}

				// 验证颜色格式（#RRGGBB 或 #RGB）
				color := colors[colorName]
				if len(color) != 7 && len(color) != 4 {
					t.Errorf("Theme %s color %s has invalid format: %s", theme.Name, colorName, color)
				}
				if color[0] != '#' {
					t.Errorf("Theme %s color %s must start with #", theme.Name, colorName)
				}
			}
		})
	}
}

// internal/config/presets_test.go
package config

import "testing"

func TestGetPresets(t *testing.T) {
	presets := GetPresets()

	expectedPresets := []string{"anthropic", "moonshot", "bigmodel", "deepseek", "minimax"}
	if len(presets) != len(expectedPresets) {
		t.Errorf("expected %d presets, got %d", len(expectedPresets), len(presets))
	}

	for _, name := range expectedPresets {
		found := false
		for _, p := range presets {
			if p.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("preset '%s' not found", name)
		}
	}
}

func TestGetPresetByName(t *testing.T) {
	tests := []struct {
		name     string
		expected *Profile
	}{
		{
			name: "anthropic",
			expected: &Profile{
				Name:             "anthropic",
				AnthropicBaseURL: "https://api.anthropic.com",
				Model:            "claude-sonnet-4-5-20250929",
			},
		},
		{
			name: "moonshot",
			expected: &Profile{
				Name:             "moonshot",
				AnthropicBaseURL: "https://api.kimi.com/coding/",
				Model:            "kimi-k2.5",
			},
		},
		{
			name: "bigmodel",
			expected: &Profile{
				Name:             "bigmodel",
				AnthropicBaseURL: "https://open.bigmodel.cn/api/anthropic",
				Model:            "glm-5",
			},
		},
		{
			name: "deepseek",
			expected: &Profile{
				Name:             "deepseek",
				AnthropicBaseURL: "https://api.deepseek.com/anthropic",
				Model:            "deepseek-chat",
			},
		},
		{
			name: "minimax",
			expected: &Profile{
				Name:             "minimax",
				AnthropicBaseURL: "https://api.minimaxi.com/anthropic",
				Model:            "MiniMax-M2.5",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetPresetByName(tt.name)
			if err != nil {
				t.Fatalf("GetPresetByName failed: %v", err)
			}
			if p.Name != tt.expected.Name {
				t.Errorf("expected name '%s', got '%s'", tt.expected.Name, p.Name)
			}
			if p.AnthropicBaseURL != tt.expected.AnthropicBaseURL {
				t.Errorf("expected baseURL '%s', got '%s'", tt.expected.AnthropicBaseURL, p.AnthropicBaseURL)
			}
			if p.Model != tt.expected.Model {
				t.Errorf("expected model '%s', got '%s'", tt.expected.Model, p.Model)
			}
		})
	}
}

func TestGetPresetByNameNotFound(t *testing.T) {
	_, err := GetPresetByName("notexist")
	if err == nil {
		t.Error("expected error for non-existent preset")
	}
}

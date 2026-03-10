// internal/config/presets.go
package config

import "fmt"

// presets 内置预设配置
var presets = []Profile{
	{
		Name:             "anthropic",
		AnthropicBaseURL: "https://api.anthropic.com",
		Model:            "claude-sonnet-4-5-20250929",
	},
	{
		Name:             "moonshot",
		AnthropicBaseURL: "https://api.kimi.com/coding/",
		Model:            "kimi-k2.5",
	},
	{
		Name:             "bigmodel",
		AnthropicBaseURL: "https://open.bigmodel.cn/api/anthropic",
		Model:            "glm-5",
	},
	{
		Name:             "deepseek",
		AnthropicBaseURL: "https://api.deepseek.com/anthropic",
		Model:            "deepseek-chat",
	},
	{
		Name:             "minimax",
		AnthropicBaseURL: "https://api.minimaxi.com/anthropic",
		Model:            "MiniMax-M2.5",
	},
}

// GetPresets 返回所有内置预设
func GetPresets() []Profile {
	return presets
}

// GetPresetByName 根据名称获取预设
func GetPresetByName(name string) (*Profile, error) {
	for i := range presets {
		if presets[i].Name == name {
			return &presets[i], nil
		}
	}
	return nil, fmt.Errorf("preset '%s' not found", name)
}

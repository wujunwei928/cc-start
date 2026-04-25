// internal/config/presets.go
package config

import "fmt"

// presets 内置预设配置
var presets = []Profile{
	{
		Name:             "anthropic",
		AnthropicBaseURL: "https://api.anthropic.com",
		HaikuModel:       "claude-haiku-4-5-20251001",
		SonnetModel:      "claude-sonnet-4-5-20250929",
		OpusModel:        "claude-opus-4-6",
	},
	{
		Name:             "moonshot",
		AnthropicBaseURL: "https://api.kimi.com/coding/",
		HaikuModel:       "kimi-k2.5",
		SonnetModel:      "kimi-k2.5",
		OpusModel:        "kimi-k2.5",
	},
	{
		Name:             "bigmodel",
		AnthropicBaseURL: "https://open.bigmodel.cn/api/anthropic",
		HaikuModel:       "glm-5-turbo",
		SonnetModel:      "glm-5-turbo",
		OpusModel:        "glm-5.1",
	},
	{
		Name:             "deepseek",
		AnthropicBaseURL: "https://api.deepseek.com/anthropic",
		HaikuModel:       "deepseek-v4-flash",
		SonnetModel:      "deepseek-v4-flash",
		OpusModel:        "deepseek-v4-pro",
	},
	{
		Name:             "minimax",
		AnthropicBaseURL: "https://api.minimaxi.com/anthropic",
		HaikuModel:       "MiniMax-M2.7",
		SonnetModel:      "MiniMax-M2.7",
		OpusModel:        "MiniMax-M2.7",
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

// internal/config/presets.go
package config

import "fmt"

// presets 内置预设配置
var presets = []Profile{
	{
		Name:    "anthropic",
		BaseURL: "https://api.anthropic.com",
		Model:   "claude-sonnet-4-5-20250929",
	},
	{
		Name:    "moonshot",
		BaseURL: "https://api.moonshot.cn/anthropic",
		Model:   "moonshot-v1-8k",
	},
	{
		Name:    "bigmodel",
		BaseURL: "https://open.bigmodel.cn/api/anthropic",
		Model:   "glm-4-plus",
	},
	{
		Name:    "deepseek",
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-chat",
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

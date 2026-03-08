// internal/config/config.go
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Profile 单个供应商配置
type Profile struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model,omitempty"`
	Token   string `json:"token"`
}

// Validate 验证配置项
func (p *Profile) Validate() error {
	if p.Name == "" {
		return errors.New("profile name is required")
	}
	if p.Token == "" {
		return errors.New("token is required")
	}
	return nil
}

// Config 完整配置
type Config struct {
	Profiles []Profile `json:"profiles"`
	Default  string    `json:"default,omitempty"`
	Settings Settings  `json:"settings,omitempty"`
}

// GetProfile 获取指定配置，name 为空时返回默认配置
func (c *Config) GetProfile(name string) (*Profile, error) {
	target := name
	if target == "" {
		target = c.Default
	}

	for i := range c.Profiles {
		if c.Profiles[i].Name == target {
			return &c.Profiles[i], nil
		}
	}

	if target == "" {
		return nil, errors.New("no profile specified and no default set")
	}
	return nil, fmt.Errorf("profile '%s' not found", target)
}

// AddProfile 添加配置
func (c *Config) AddProfile(p Profile) error {
	if err := p.Validate(); err != nil {
		return err
	}

	// 检查是否已存在
	for i, existing := range c.Profiles {
		if existing.Name == p.Name {
			c.Profiles[i] = p // 更新已存在的配置
			return nil
		}
	}

	c.Profiles = append(c.Profiles, p)
	return nil
}

// DeleteProfile 删除配置
func (c *Config) DeleteProfile(name string) error {
	for i, p := range c.Profiles {
		if p.Name == name {
			c.Profiles = append(c.Profiles[:i], c.Profiles[i+1:]...)
			if c.Default == name {
				c.Default = ""
			}
			return nil
		}
	}
	return fmt.Errorf("profile '%s' not found", name)
}

// SetDefault 设置默认配置
func (c *Config) SetDefault(name string) error {
	for _, p := range c.Profiles {
		if p.Name == name {
			c.Default = name
			return nil
		}
	}
	return fmt.Errorf("profile '%s' not found", name)
}

// UpdateSetting 更新设置项
func (c *Config) UpdateSetting(key, value string) {
	switch key {
	case "language":
		c.Settings.Language = value
	case "theme":
		c.Settings.Theme = value
	}
}

// Save 保存配置到文件
func (c *Config) Save(path string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Profiles: []Profile{}}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.Profiles == nil {
		cfg.Profiles = []Profile{}
	}

	// 迁移：如果 settings 为空，设置默认值
	if cfg.Settings.Language == "" {
		cfg.Settings.Language = "zh"
	}
	if cfg.Settings.Theme == "" {
		cfg.Settings.Theme = "default"
	}

	return &cfg, nil
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".cc-start", "profiles.json")
}

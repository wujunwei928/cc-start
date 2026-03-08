// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProfileValidation(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		wantErr bool
	}{
		{
			name: "valid profile",
			profile: Profile{
				Name:    "anthropic",
				BaseURL: "https://api.anthropic.com",
				Token:   "sk-ant-xxx",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			profile: Profile{
				BaseURL: "https://api.anthropic.com",
				Token:   "sk-ant-xxx",
			},
			wantErr: true,
		},
		{
			name: "missing token",
			profile: Profile{
				Name:    "anthropic",
				BaseURL: "https://api.anthropic.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Profile.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigLoadAndSave(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "cc-start-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "profiles.json")

	// 测试保存
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", BaseURL: "https://example.com", Token: "token123"},
		},
		Default: "test",
	}

	err = cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 测试加载
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(loaded.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(loaded.Profiles))
	}
	if loaded.Default != "test" {
		t.Errorf("expected default 'test', got '%s'", loaded.Default)
	}
}

func TestConfigGetProfile(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "anthropic", BaseURL: "https://api.anthropic.com", Token: "token1"},
			{Name: "moonshot", BaseURL: "https://api.kimi.com/coding/", Token: "token2"},
		},
		Default: "anthropic",
	}

	// 测试获取指定配置
	p, err := cfg.GetProfile("moonshot")
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}
	if p.Name != "moonshot" {
		t.Errorf("expected 'moonshot', got '%s'", p.Name)
	}

	// 测试获取默认配置
	p, err = cfg.GetProfile("")
	if err != nil {
		t.Fatalf("GetProfile(default) failed: %v", err)
	}
	if p.Name != "anthropic" {
		t.Errorf("expected default 'anthropic', got '%s'", p.Name)
	}

	// 测试获取不存在的配置
	_, err = cfg.GetProfile("notexist")
	if err == nil {
		t.Error("expected error for non-existent profile")
	}
}

func TestAddProfile(t *testing.T) {
	cfg := &Config{Profiles: []Profile{}}

	// 测试添加新配置
	p := Profile{Name: "test", BaseURL: "https://example.com", Token: "token123"}
	if err := cfg.AddProfile(p); err != nil {
		t.Errorf("AddProfile failed: %v", err)
	}

	if len(cfg.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(cfg.Profiles))
	}

	// 测试更新已存在的配置
	p.Token = "newtoken"
	if err := cfg.AddProfile(p); err != nil {
		t.Errorf("AddProfile update failed: %v", err)
	}

	// 验证更新成功
	profile, _ := cfg.GetProfile("test")
	if profile.Token != "newtoken" {
		t.Errorf("expected 'newtoken', got '%s'", profile.Token)
	}

	// 验证仍然是 1 个配置（更新而非添加）
	if len(cfg.Profiles) != 1 {
		t.Errorf("expected 1 profile after update, got %d", len(cfg.Profiles))
	}
}

func TestAddProfileValidation(t *testing.T) {
	cfg := &Config{Profiles: []Profile{}}

	// 测试无效配置（缺少 name）
	p := Profile{Token: "token"}
	if err := cfg.AddProfile(p); err == nil {
		t.Error("expected error for missing name")
	}

	// 测试无效配置（缺少 token）
	p = Profile{Name: "test"}
	if err := cfg.AddProfile(p); err == nil {
		t.Error("expected error for missing token")
	}
}

func TestDeleteProfile(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Token: "token"},
			{Name: "other", Token: "other-token"},
		},
		Default: "test",
	}

	// 测试删除配置
	if err := cfg.DeleteProfile("test"); err != nil {
		t.Errorf("DeleteProfile failed: %v", err)
	}

	// 验证配置已删除
	if len(cfg.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(cfg.Profiles))
	}

	// 验证默认值被清除
	if cfg.Default != "" {
		t.Errorf("default should be cleared after delete, got '%s'", cfg.Default)
	}

	// 测试删除不存在的配置
	if err := cfg.DeleteProfile("notexist"); err == nil {
		t.Error("expected error for non-existent profile")
	}
}

func TestSetDefault(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Token: "token"},
			{Name: "other", Token: "other-token"},
		},
	}

	// 测试设置默认
	if err := cfg.SetDefault("test"); err != nil {
		t.Errorf("SetDefault failed: %v", err)
	}

	if cfg.Default != "test" {
		t.Errorf("expected default 'test', got '%s'", cfg.Default)
	}

	// 测试设置不存在的配置为默认
	if err := cfg.SetDefault("notexist"); err == nil {
		t.Error("expected error for non-existent profile")
	}
}

func TestGetProfileNoDefault(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{{Name: "test", Token: "token"}},
		// Default 为空
	}

	// 测试无默认配置时的行为
	_, err := cfg.GetProfile("")
	if err == nil {
		t.Error("expected error when no default set")
	}
}

func TestConfigWithSettings(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Token: "xxx"},
		},
		Default: "test",
		Settings: Settings{
			Language: "en",
			Theme:    "ocean",
		},
	}

	if cfg.Settings.Language != "en" {
		t.Errorf("Settings.Language = %s, want en", cfg.Settings.Language)
	}

	if cfg.Settings.Theme != "ocean" {
		t.Errorf("Settings.Theme = %s, want ocean", cfg.Settings.Theme)
	}
}

func TestLoadConfigWithEmptySettings(t *testing.T) {
	// 创建临时配置文件（没有 settings 字段）
	tmpDir := t.TempDir()
	configPath := tmpDir + "/profiles.json"

	data := `{
		"profiles": [{"name": "test", "token": "xxx"}],
		"default": "test"
	}`
	if err := os.WriteFile(configPath, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	// 验证默认值
	if cfg.Settings.Language != "zh" {
		t.Errorf("Settings.Language = %s, want zh (default)", cfg.Settings.Language)
	}

	if cfg.Settings.Theme != "default" {
		t.Errorf("Settings.Theme = %s, want default", cfg.Settings.Theme)
	}
}

func TestUpdateSetting(t *testing.T) {
	cfg := &Config{}

	cfg.UpdateSetting("language", "en")
	if cfg.Settings.Language != "en" {
		t.Errorf("UpdateSetting(language, en) failed")
	}

	cfg.UpdateSetting("theme", "ocean")
	if cfg.Settings.Theme != "ocean" {
		t.Errorf("UpdateSetting(theme, ocean) failed")
	}
}

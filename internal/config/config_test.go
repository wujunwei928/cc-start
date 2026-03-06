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
			{Name: "moonshot", BaseURL: "https://api.moonshot.cn/anthropic", Token: "token2"},
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

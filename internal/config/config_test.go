// internal/config/config_test.go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
				Name:             "anthropic",
				AnthropicBaseURL: "https://api.anthropic.com",
				Token:            "sk-ant-xxx",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			profile: Profile{
				AnthropicBaseURL: "https://api.anthropic.com",
				Token:            "sk-ant-xxx",
			},
			wantErr: true,
		},
		{
			name: "missing token",
			profile: Profile{
				Name:             "anthropic",
				AnthropicBaseURL: "https://api.anthropic.com",
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

	configPath := filepath.Join(tmpDir, "settings.json")

	// 测试保存
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", AnthropicBaseURL: "https://example.com", Token: "token123"},
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
			{Name: "anthropic", AnthropicBaseURL: "https://api.anthropic.com", Token: "token1"},
			{Name: "moonshot", AnthropicBaseURL: "https://api.kimi.com/coding/", Token: "token2"},
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
	p := Profile{Name: "test", AnthropicBaseURL: "https://example.com", Token: "token123"}
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
	configPath := tmpDir + "/settings.json"

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

func TestProfileMigrate(t *testing.T) {
	// 旧格式 JSON 只有一个 model 字段
	oldJSON := `{"name":"test","model":"glm-5","token":"sk-xxx"}`
	var p Profile
	if err := json.Unmarshal([]byte(oldJSON), &p); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// 迁移前：LegacyModel 有值，SonnetModel 为空
	if p.LegacyModel != "glm-5" {
		t.Fatalf("LegacyModel should be 'glm-5', got '%s'", p.LegacyModel)
	}
	if p.SonnetModel != "" {
		t.Fatalf("SonnetModel should be empty before migrate, got '%s'", p.SonnetModel)
	}

	p.Migrate()

	// 迁移后：SonnetModel 有值，LegacyModel 清空
	if p.SonnetModel != "glm-5" {
		t.Errorf("SonnetModel should be 'glm-5' after migrate, got '%s'", p.SonnetModel)
	}
	if p.LegacyModel != "" {
		t.Errorf("LegacyModel should be empty after migrate, got '%s'", p.LegacyModel)
	}
}

func TestProfileMigrateNoOverwrite(t *testing.T) {
	// 新格式 JSON 已有 sonnet_model
	newJSON := `{"name":"test","sonnet_model":"new-model","model":"old-model","token":"sk-xxx"}`
	var p Profile
	if err := json.Unmarshal([]byte(newJSON), &p); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	p.Migrate()

	// 已有 sonnet_model 时，不应被 legacyModel 覆盖
	if p.SonnetModel != "new-model" {
		t.Errorf("SonnetModel should remain 'new-model', got '%s'", p.SonnetModel)
	}
}

func TestProfileMigrateEmpty(t *testing.T) {
	p := Profile{Name: "test", Token: "xxx"}
	p.Migrate() // 不应 panic
	if p.SonnetModel != "" {
		t.Errorf("SonnetModel should be empty, got '%s'", p.SonnetModel)
	}
}

func TestConfigMigrateAll(t *testing.T) {
	oldJSON := `{"profiles":[{"name":"a","model":"m1","token":"t1"},{"name":"b","model":"m2","token":"t2"}]}`
	var cfg Config
	if err := json.Unmarshal([]byte(oldJSON), &cfg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	cfg.MigrateAll()
	if cfg.Profiles[0].SonnetModel != "m1" {
		t.Errorf("profiles[0].SonnetModel should be 'm1', got '%s'", cfg.Profiles[0].SonnetModel)
	}
	if cfg.Profiles[1].SonnetModel != "m2" {
		t.Errorf("profiles[1].SonnetModel should be 'm2', got '%s'", cfg.Profiles[1].SonnetModel)
	}
}

func TestMigrateNotInOutput(t *testing.T) {
	oldJSON := `{"name":"test","model":"glm-5","token":"sk-xxx"}`
	var p Profile
	json.Unmarshal([]byte(oldJSON), &p)
	p.Migrate()

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if strings.Contains(string(data), `"model"`) {
		t.Errorf("migrated profile should not contain 'model' key in JSON output: %s", string(data))
	}
}

func TestNewProfileSerialization(t *testing.T) {
	p := Profile{
		Name:             "test",
		AnthropicBaseURL: "https://example.com",
		HaikuModel:       "haiku-model",
		SonnetModel:      "sonnet-model",
		OpusModel:        "opus-model",
		Token:            "sk-xxx",
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, `"haiku_model"`) || !strings.Contains(s, `"sonnet_model"`) || !strings.Contains(s, `"opus_model"`) {
		t.Errorf("new profile should contain haiku_model, sonnet_model, opus_model: %s", s)
	}
	if strings.Contains(s, `"model"`) {
		t.Errorf("new profile should not contain 'model' key: %s", s)
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

// internal/launcher/launcher_test.go
package launcher

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/wujunwei928/cc-start/internal/config"
)

func TestBuildSettings(t *testing.T) {
	tests := []struct {
		name     string
		profile  config.Profile
		wantKeys []string
	}{
		{
			name: "anthropic official",
			profile: config.Profile{
				Name:             "anthropic",
				AnthropicBaseURL: "https://api.anthropic.com",
				Token:            "sk-ant-xxx",
			},
			wantKeys: []string{"ANTHROPIC_AUTH_TOKEN"},
		},
		{
			name: "custom provider",
			profile: config.Profile{
				Name:             "moonshot",
				AnthropicBaseURL: "https://api.kimi.com/coding/",
				Token:            "sk-xxx",
			},
			wantKeys: []string{"ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_BASE_URL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := BuildSettings(&tt.profile)

			// 检查必需的键存在
			env, ok := settings["env"].(map[string]string)
			if !ok {
				t.Fatal("settings should have env map")
			}

			for _, key := range tt.wantKeys {
				if _, exists := env[key]; !exists {
					t.Errorf("missing key '%s' in settings", key)
				}
			}

			// 官方 API 不应该有 base_url
			if tt.profile.AnthropicBaseURL == "https://api.anthropic.com" {
				if _, exists := env["ANTHROPIC_BASE_URL"]; exists {
					t.Error("official API should not have ANTHROPIC_BASE_URL")
				}
			}
		})
	}
}

func TestBuildSettingsEmptyBaseURL(t *testing.T) {
	// 测试空 BaseURL 的情况
	profile := &config.Profile{
		Name:             "empty-url",
		AnthropicBaseURL: "",
		Token:            "token123",
	}

	settings := BuildSettings(profile)
	env, ok := settings["env"].(map[string]string)
	if !ok {
		t.Fatal("settings should have env map")
	}

	// 应该只有 token，没有 base URL
	if _, exists := env["ANTHROPIC_AUTH_TOKEN"]; !exists {
		t.Error("missing ANTHROPIC_AUTH_TOKEN")
	}
	if _, exists := env["ANTHROPIC_BASE_URL"]; exists {
		t.Error("should not have ANTHROPIC_BASE_URL when BaseURL is empty")
	}
}

func TestBuildSettingsJSON(t *testing.T) {
	profile := &config.Profile{
		Name:             "moonshot",
		AnthropicBaseURL: "https://api.kimi.com/coding/",
		Token:            "test-token",
		SonnetModel:      "kimi-k2.5",
	}

	settings := BuildSettings(profile)

	// 验证可以序列化为 JSON
	jsonData, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("failed to marshal settings: %v", err)
	}

	// 验证 JSON 格式正确
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("failed to unmarshal settings: %v", err)
	}

	env, ok := parsed["env"].(map[string]interface{})
	if !ok {
		t.Fatal("settings should have env map")
	}

	if env["ANTHROPIC_AUTH_TOKEN"] != "test-token" {
		t.Errorf("wrong token value")
	}
	if env["ANTHROPIC_BASE_URL"] != "https://api.kimi.com/coding/" {
		t.Errorf("wrong base URL value")
	}
}

func TestBuildSettingsTripleModel(t *testing.T) {
	profile := &config.Profile{
		Name:             "bigmodel",
		AnthropicBaseURL: "https://open.bigmodel.cn/api/anthropic",
		HaikuModel:       "glm-5-turbo",
		SonnetModel:      "glm-5-turbo",
		OpusModel:        "glm-5.1",
		Token:            "sk-xxx",
	}
	settings := BuildSettings(profile)
	env, ok := settings["env"].(map[string]string)
	if !ok {
		t.Fatal("settings should have env map")
	}
	if env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] != "glm-5-turbo" {
		t.Errorf("expected haiku model 'glm-5-turbo', got '%s'", env["ANTHROPIC_DEFAULT_HAIKU_MODEL"])
	}
	if env["ANTHROPIC_DEFAULT_SONNET_MODEL"] != "glm-5-turbo" {
		t.Errorf("expected sonnet model 'glm-5-turbo', got '%s'", env["ANTHROPIC_DEFAULT_SONNET_MODEL"])
	}
	if env["ANTHROPIC_DEFAULT_OPUS_MODEL"] != "glm-5.1" {
		t.Errorf("expected opus model 'glm-5.1', got '%s'", env["ANTHROPIC_DEFAULT_OPUS_MODEL"])
	}
}

func TestBuildSettingsPartialModel(t *testing.T) {
	profile := &config.Profile{
		Name:             "test",
		AnthropicBaseURL: "https://example.com",
		SonnetModel:      "only-sonnet",
		Token:            "sk-xxx",
	}
	settings := BuildSettings(profile)
	env, ok := settings["env"].(map[string]string)
	if !ok {
		t.Fatal("settings should have env map")
	}
	// 只设了 SonnetModel，其他两个不应出现
	if env["ANTHROPIC_DEFAULT_SONNET_MODEL"] != "only-sonnet" {
		t.Errorf("expected sonnet model 'only-sonnet', got '%s'", env["ANTHROPIC_DEFAULT_SONNET_MODEL"])
	}
	if _, exists := env["ANTHROPIC_DEFAULT_HAIKU_MODEL"]; exists {
		t.Error("haiku model should not be set when empty")
	}
	if _, exists := env["ANTHROPIC_DEFAULT_OPUS_MODEL"]; exists {
		t.Error("opus model should not be set when empty")
	}
}

func TestCheckClaudeInstalled(t *testing.T) {
	path, err := CheckClaudeInstalled()

	if err != nil {
		// 未安装时，应返回包含安装指引的错误消息
		if !strings.Contains(err.Error(), "npm install -g") {
			t.Errorf("error message should contain installation instructions, got: %v", err)
		}
		if path != "" {
			t.Errorf("path should be empty when claude is not found, got: %s", path)
		}
	} else {
		// 已安装时，应返回非空路径
		if path == "" {
			t.Error("path should not be empty when claude is found")
		}
	}
}

func TestMergeConfig(t *testing.T) {
	profile := &config.Profile{
		Name:             "test",
		AnthropicBaseURL: "https://anthropic.example.com",
		SonnetModel:      "model-v1",
		Token:            "profile-token",
	}

	tests := []struct {
		name      string
		cfg       *LaunchConfig
		wantModel string
		wantURL   string
		wantToken string
	}{
		{
			name: "profile provides defaults",
			cfg: &LaunchConfig{
				Profile: profile,
			},
			wantModel: "model-v1",
			wantURL:   "https://anthropic.example.com",
			wantToken: "profile-token",
		},
		{
			name: "command line overrides profile",
			cfg: &LaunchConfig{
				Profile: profile,
				Model:   "override-model",
				BaseURL: "https://override.com",
				Token:   "override-token",
			},
			wantModel: "override-model",
			wantURL:   "https://override.com",
			wantToken: "override-token",
		},
		{
			name: "partial override - model only",
			cfg: &LaunchConfig{
				Profile: profile,
				Model:   "new-model",
			},
			wantModel: "new-model",
			wantURL:   "https://anthropic.example.com",
			wantToken: "profile-token",
		},
		{
			name:      "no profile no override",
			cfg:       &LaunchConfig{},
			wantModel: "",
			wantURL:   "",
			wantToken: "",
		},
		{
			name: "no profile with command line values",
			cfg: &LaunchConfig{
				Model:   "cmd-model",
				BaseURL: "https://cmd.example.com",
				Token:   "cmd-token",
			},
			wantModel: "cmd-model",
			wantURL:   "https://cmd.example.com",
			wantToken: "cmd-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, baseURL, token := MergeConfig(tt.cfg)
			if model != tt.wantModel {
				t.Errorf("MergeConfig() model = %v, want %v", model, tt.wantModel)
			}
			if baseURL != tt.wantURL {
				t.Errorf("MergeConfig() baseURL = %v, want %v", baseURL, tt.wantURL)
			}
			if token != tt.wantToken {
				t.Errorf("MergeConfig() token = %v, want %v", token, tt.wantToken)
			}
		})
	}
}

// internal/launcher/launcher_test.go
package launcher

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/wujunwei928/cc-start/internal/config"
	"github.com/wujunwei928/cc-start/internal/tools"
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

func TestBuildCommand(t *testing.T) {
	profile := &config.Profile{
		Name:             "test",
		AnthropicBaseURL: "https://api.example.com",
		Token:            "token123",
		Model:            "test-model",
	}

	args := []string{"--dangerously-skip-permissions"}
	cmd := BuildCommand(profile, args)

	// 验证命令路径包含 claude
	if !strings.Contains(cmd.Path, "claude") {
		t.Errorf("expected path to contain 'claude', got '%s'", cmd.Path)
	}

	// 检查模型参数
	foundModel := false
	for _, arg := range cmd.Args {
		if arg == "--model" {
			foundModel = true
		}
	}
	if !foundModel {
		t.Error("command should include --model flag")
	}

	// 检查 --settings 参数存在
	foundSettings := false
	for _, arg := range cmd.Args {
		if arg == "--settings" {
			foundSettings = true
		}
	}
	if !foundSettings {
		t.Error("command should include --settings flag")
	}

	// 检查额外参数被包含
	foundDangerously := false
	for _, arg := range cmd.Args {
		if arg == "--dangerously-skip-permissions" {
			foundDangerously = true
		}
	}
	if !foundDangerously {
		t.Error("command should include extra args")
	}

	// 验证标准输入输出已设置
	if cmd.Stdin != os.Stdin {
		t.Error("command should have Stdin set to os.Stdin")
	}
	if cmd.Stdout != os.Stdout {
		t.Error("command should have Stdout set to os.Stdout")
	}
	if cmd.Stderr != os.Stderr {
		t.Error("command should have Stderr set to os.Stderr")
	}
}

func TestBuildCommandWithoutModel(t *testing.T) {
	// 测试没有指定模型的情况
	profile := &config.Profile{
		Name:             "no-model",
		AnthropicBaseURL: "https://api.anthropic.com",
		Token:            "token123",
		Model:            "", // 空模型
	}

	cmd := BuildCommand(profile, []string{})

	// 不应该有 --model 参数
	for i, arg := range cmd.Args {
		if arg == "--model" && i+1 < len(cmd.Args) && cmd.Args[i+1] != "" {
			t.Error("command should not include --model flag when model is empty")
		}
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
		Model:            "kimi-k2.5",
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

func TestMergeConfig(t *testing.T) {
	profile := &config.Profile{
		Name:             "test",
		AnthropicBaseURL: "https://anthropic.example.com",
		OpenAIBaseURL:    "https://openai.example.com",
		Model:            "model-v1",
		Token:            "profile-token",
	}

	tests := []struct {
		name       string
		cfg        *LaunchConfig
		toolFormat string
		wantModel  string
		wantURL    string
		wantToken  string
	}{
		{
			name: "anthropic format selects anthropic url",
			cfg: &LaunchConfig{
				Profile: profile,
			},
			toolFormat: tools.FormatAnthropic,
			wantModel:  "model-v1",
			wantURL:    "https://anthropic.example.com",
			wantToken:  "profile-token",
		},
		{
			name: "openai format selects openai url",
			cfg: &LaunchConfig{
				Profile: profile,
			},
			toolFormat: tools.FormatOpenAI,
			wantModel:  "model-v1",
			wantURL:    "https://openai.example.com",
			wantToken:  "profile-token",
		},
		{
			name: "command line overrides profile",
			cfg: &LaunchConfig{
				Profile: profile,
				Model:   "override-model",
				BaseURL: "https://override.com",
				Token:   "override-token",
			},
			toolFormat: tools.FormatAnthropic,
			wantModel:  "override-model",
			wantURL:    "https://override.com",
			wantToken:  "override-token",
		},
		{
			name: "partial override - model only",
			cfg: &LaunchConfig{
				Profile: profile,
				Model:   "new-model",
			},
			toolFormat: tools.FormatOpenAI,
			wantModel:  "new-model",
			wantURL:    "https://openai.example.com",
			wantToken:  "profile-token",
		},
		{
			name: "no profile no override",
			cfg: &LaunchConfig{
				Tool: "claude",
			},
			toolFormat: tools.FormatAnthropic,
			wantModel:  "",
			wantURL:    "",
			wantToken:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, baseURL, token := MergeConfig(tt.cfg, tt.toolFormat)
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

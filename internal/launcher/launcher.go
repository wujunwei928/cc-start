// internal/launcher/launcher.go
package launcher

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/wujunwei928/cc-start/internal/config"
)

// BuildSettings 构建 Claude Code 设置 JSON
func BuildSettings(profile *config.Profile) map[string]interface{} {
	env := map[string]string{
		"ANTHROPIC_AUTH_TOKEN": profile.Token,
	}

	// 非官方 API 需要设置 base URL
	if profile.AnthropicBaseURL != "" && profile.AnthropicBaseURL != "https://api.anthropic.com" {
		env["ANTHROPIC_BASE_URL"] = profile.AnthropicBaseURL
	}

	// 注入三个层级的模型映射
	if profile.HaikuModel != "" {
		env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = profile.HaikuModel
	}
	if profile.SonnetModel != "" {
		env["ANTHROPIC_DEFAULT_SONNET_MODEL"] = profile.SonnetModel
	}
	if profile.OpusModel != "" {
		env["ANTHROPIC_DEFAULT_OPUS_MODEL"] = profile.OpusModel
	}

	return map[string]interface{}{
		"env": env,
	}
}

// LaunchConfig 启动配置
type LaunchConfig struct {
	Profile  *config.Profile   // Profile 配置（可选）
	Model    string            // 命令行指定的模型
	BaseURL  string            // 命令行指定的 BaseURL
	Token    string            // 命令行指定的 Token
	Env      map[string]string // 额外环境变量
	ToolArgs []string          // 传递给 claude 的额外参数
	Yolo     bool              // 自动接受所有操作（YOLO 模式）
}

// MergeConfig 合并配置，返回最终参数
// 优先级: 命令行 > Profile > 默认值
func MergeConfig(cfg *LaunchConfig) (model, baseURL, token string) {
	// Profile 覆盖
	if cfg.Profile != nil {
		if cfg.Profile.SonnetModel != "" {
			model = cfg.Profile.SonnetModel
		}
		baseURL = cfg.Profile.AnthropicBaseURL
		if cfg.Profile.Token != "" {
			token = cfg.Profile.Token
		}
	}

	// 命令行覆盖
	if cfg.Model != "" {
		model = cfg.Model
	}
	if cfg.BaseURL != "" {
		baseURL = cfg.BaseURL
	}
	if cfg.Token != "" {
		token = cfg.Token
	}

	return
}

// Launch 使用合并后的配置启动 Claude Code
func Launch(cfg *LaunchConfig) error {
	model, baseURL, token := MergeConfig(cfg)

	// 校验 token 不为空
	if token == "" {
		return fmt.Errorf("未配置 token，请指定 profile 或通过 -t/--token 参数传入")
	}

	// 安全获取原始 profile（可能为 nil，如只传 -t 而不选 profile）
	var srcProfile config.Profile
	if cfg.Profile != nil {
		srcProfile = *cfg.Profile
	}

	effectiveProfile := &config.Profile{
		AnthropicBaseURL: baseURL,
		HaikuModel:       srcProfile.HaikuModel, // 继承原始值
		SonnetModel:      model,                 // CLI -m 覆盖 SonnetModel
		OpusModel:        srcProfile.OpusModel,  // 继承原始值
		Token:            token,
	}

	// 构建 --settings JSON
	settings := BuildSettings(effectiveProfile)
	settingsJSON, _ := json.Marshal(settings)

	args := []string{"--settings", string(settingsJSON)}

	// YOLO 模式：自动接受所有操作
	if cfg.Yolo {
		args = append(args, "--dangerously-skip-permissions")
	}

	// 添加工具原生参数
	args = append(args, cfg.ToolArgs...)

	// 创建命令
	cmd := exec.Command("claude", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 注入额外环境变量
	if len(cfg.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range cfg.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 打印启动信息（使用 effectiveProfile，反映 CLI 覆盖后的实际生效值）
	fmt.Printf("🚀 使用配置启动 Claude Code...\n")
	if effectiveProfile.HaikuModel != "" {
		fmt.Printf("   快速模型 (Haiku): %s\n", effectiveProfile.HaikuModel)
	}
	if effectiveProfile.SonnetModel != "" {
		fmt.Printf("   主模型 (Sonnet):  %s\n", effectiveProfile.SonnetModel)
	}
	if effectiveProfile.OpusModel != "" {
		fmt.Printf("   经济模型 (Opus):   %s\n", effectiveProfile.OpusModel)
	}
	if baseURL != "" {
		fmt.Printf("   Base URL:         %s\n", baseURL)
	}
	fmt.Println()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch claude: %w", err)
	}

	return nil
}

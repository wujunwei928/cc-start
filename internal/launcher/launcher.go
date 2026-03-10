// internal/launcher/launcher.go
package launcher

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/wujunwei928/cc-start/internal/config"
	"github.com/wujunwei928/cc-start/internal/tools"
)

// BuildSettings 构建 Claude Code 设置 JSON
func BuildSettings(profile *config.Profile) map[string]interface{} {
	env := map[string]string{
		"ANTHROPIC_AUTH_TOKEN": profile.Token,
	}

	// 非官方 API 需要设置 base URL
	if profile.BaseURL != "" && profile.BaseURL != "https://api.anthropic.com" {
		env["ANTHROPIC_BASE_URL"] = profile.BaseURL
	}

	return map[string]interface{}{
		"env": env,
	}
}

// BuildCommand 构建启动命令
func BuildCommand(profile *config.Profile, extraArgs []string) *exec.Cmd {
	settings := BuildSettings(profile)
	settingsJSON, _ := json.Marshal(settings)

	args := []string{"--settings", string(settingsJSON)}

	// 添加模型参数（如果指定）
	if profile.Model != "" {
		args = append(args, "--model", profile.Model)
	}

	// 添加额外参数
	args = append(args, extraArgs...)

	cmd := exec.Command("claude", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

// Launch 启动 Claude Code
func Launch(profile *config.Profile, extraArgs []string) error {
	cmd := BuildCommand(profile, extraArgs)

	fmt.Printf("🚀 使用配置 '%s' 启动 Claude Code...\n", profile.Name)
	if profile.Model != "" {
		fmt.Printf("   模型: %s\n", profile.Model)
	}
	fmt.Println()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch claude: %w", err)
	}

	return nil
}

// LaunchConfig 启动配置
type LaunchConfig struct {
	Tool     string            // 工具名称
	Profile  *config.Profile   // Profile 配置（可选）
	Model    string            // 命令行指定的模型
	BaseURL  string            // 命令行指定的 BaseURL
	Token    string            // 命令行指定的 Token
	Env      map[string]string // 额外环境变量
	ToolArgs []string          // 传递给工具的额外参数
}

// MergeConfig 合并配置，返回最终参数
// 优先级: 命令行 > Profile > 默认值
func MergeConfig(cfg *LaunchConfig) (model, baseURL, token string) {
	// 默认值（空）

	// Profile 覆盖
	if cfg.Profile != nil {
		if cfg.Profile.Model != "" {
			model = cfg.Profile.Model
		}
		if cfg.Profile.BaseURL != "" {
			baseURL = cfg.Profile.BaseURL
		}
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

// LaunchWithTool 使用指定工具启动
func LaunchWithTool(cfg *LaunchConfig) error {
	// 获取工具预设
	tool, err := tools.GetTool(cfg.Tool)
	if err != nil {
		return err
	}

	// 合并配置
	model, baseURL, token := MergeConfig(cfg)

	// 构建环境变量
	env := os.Environ()

	// 设置 Token 环境变量
	if token != "" {
		envName := tool.GetEnvName(tools.ParamToken)
		env = append(env, fmt.Sprintf("%s=%s", envName, token))
	}

	// 设置 BaseURL 环境变量
	if baseURL != "" {
		envName := tool.GetEnvName(tools.ParamBaseURL)
		env = append(env, fmt.Sprintf("%s=%s", envName, baseURL))
	}

	// 添加额外环境变量
	for k, v := range cfg.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// 构建命令参数
	args := []string{}

	// 对于 claude，使用 --settings 传递环境变量
	if cfg.Tool == "claude" {
		settingsEnv := make(map[string]string)
		if token != "" {
			settingsEnv["ANTHROPIC_AUTH_TOKEN"] = token
		}
		if baseURL != "" && baseURL != "https://api.anthropic.com" {
			settingsEnv["ANTHROPIC_BASE_URL"] = baseURL
		}
		if len(settingsEnv) > 0 {
			settings := map[string]interface{}{"env": settingsEnv}
			settingsJSON, _ := json.Marshal(settings)
			args = append(args, "--settings", string(settingsJSON))
		}
	}

	// 添加模型参数
	if model != "" {
		args = append(args, "--model", model)
	}

	// 添加工具原生参数
	args = append(args, cfg.ToolArgs...)

	// 创建命令
	cmd := exec.Command(tool.Executable, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	// 打印启动信息
	fmt.Printf("🚀 使用工具 '%s' 启动...\n", tool.Name)
	if model != "" {
		fmt.Printf("   模型: %s\n", model)
	}
	if baseURL != "" {
		fmt.Printf("   Base URL: %s\n", baseURL)
	}
	fmt.Println()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch %s: %w", tool.Name, err)
	}

	return nil
}

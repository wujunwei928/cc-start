# 移除 Codex/OpenCode 支持 实施计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除所有 codex/opencode 相关代码，简化为只支持 claude 的单一工具架构。

**Architecture:** 删除 `internal/tools` 包和 `cmd/codex.go`/`cmd/opencode.go` 死代码文件；从 `Profile` 结构体移除 `OpenAIBaseURL` 字段；将 `LaunchWithTool` + 旧 `Launch` 合并为统一的 `Launch(cfg *LaunchConfig)` 入口；清理所有 UI 层的 OpenAI URL 展示和 TUI setup 步骤。

**Tech Stack:** Go 1.24+, Cobra CLI, Bubble Tea TUI

---

### Task 1: 删除死代码文件和历史文档

**Files:**
- Delete: `cmd/codex.go`
- Delete: `cmd/opencode.go`
- Delete: `internal/tools/tools.go`
- Delete: `internal/tools/tools_test.go`
- Delete: `docs/plans/2026-03-10-launch-command-design.md`
- Delete: `docs/plans/2026-03-10-launch-command-impl.md`
- Delete: `docs/plans/2026-03-11-dual-baseurl-design.md`

- [ ] **Step 1: 删除文件**

```bash
rm cmd/codex.go cmd/opencode.go
rm -rf internal/tools/
rm docs/plans/2026-03-10-launch-command-design.md docs/plans/2026-03-10-launch-command-impl.md docs/plans/2026-03-11-dual-baseurl-design.md
```

- [ ] **Step 2: 验证编译失败（预期行为）**

Run: `go build ./...`
Expected: 编译失败，因为 `cmd/claude.go` 和 `internal/launcher/launcher.go` 仍引用 `internal/tools`

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "chore: 删除 codex/opencode 死代码文件、tools 包和历史设计文档"
```

---

### Task 2: 移除 Profile.OpenAIBaseURL 字段

**Files:**
- Modify: `internal/config/config.go:16`
- Modify: `internal/config/presets.go`

- [ ] **Step 1: 从 Profile 结构体移除 OpenAIBaseURL 字段**

在 `internal/config/config.go` 中，删除第 16 行：

```go
// 删除这一行：
OpenAIBaseURL    string `json:"openai_base_url,omitempty"`
```

- [ ] **Step 2: 从预设中移除 OpenAIBaseURL 赋值**

在 `internal/config/presets.go` 中，从 `bigmodel`、`deepseek`、`minimax` 预设中删除 `OpenAIBaseURL` 行：

```go
// 删除以下行：
// bigmodel 预设中：
OpenAIBaseURL:    "https://open.bigmodel.cn/api/coding/paas/v4",
// deepseek 预设中：
OpenAIBaseURL:    "https://api.deepseek.com/v1",
// minimax 预设中：
OpenAIBaseURL:    "https://api.minimaxi.com/v1",
```

- [ ] **Step 3: 验证编译**

Run: `go build ./...`
Expected: 编译失败（launcher、repl、tui 仍引用 OpenAIBaseURL，这是预期的，后续 Task 会修复）

- [ ] **Step 4: Commit**

```bash
git add internal/config/config.go internal/config/presets.go
git commit -m "refactor(config): 从 Profile 和预设中移除 OpenAIBaseURL 字段"
```

---

### Task 3: 简化 Launcher 层，移除 tools 依赖

**Files:**
- Modify: `internal/launcher/launcher.go`
- Modify: `internal/launcher/launcher_test.go`
- Modify: `cmd/claude.go`
- Modify: `cmd/launcher.go`

这是核心 Task。需要：
1. 重写 `internal/launcher/launcher.go` — 移除 `tools` 导入、`LaunchConfig.Tool`、`MergeConfig` 的 `toolFormat` 参数、整个 `LaunchWithTool` 函数，将旧 `Launch` 和 `LaunchWithTool` 合并为统一的 `Launch(cfg *LaunchConfig) error`
2. 重写 `internal/launcher/launcher_test.go` — 移除 `tools` 导入，重写 `TestMergeConfig`
3. 简化 `cmd/claude.go` — 移除 `tools` 导入和 `tools.GetTool` 调用
4. 简化 `cmd/launcher.go` — 移除 `toolName` 参数，调用 `launcher.Launch`

- [ ] **Step 1: 重写 `internal/launcher/launcher.go`**

完整替换文件内容：

```go
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

	return map[string]interface{}{
		"env": env,
	}
}

// BuildCommand 构建启动命令
func BuildCommand(profile *config.Profile, extraArgs []string) *exec.Cmd {
	settings := BuildSettings(profile)
	settingsJSON, _ := json.Marshal(settings)

	args := []string{"--settings", string(settingsJSON)}

	if profile.Model != "" {
		args = append(args, "--model", profile.Model)
	}

	args = append(args, extraArgs...)

	cmd := exec.Command("claude", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

// LaunchConfig 启动配置
type LaunchConfig struct {
	Profile  *config.Profile   // Profile 配置（可选）
	Model    string            // 命令行指定的模型
	BaseURL  string            // 命令行指定的 BaseURL
	Token    string            // 命令行指定的 Token
	Env      map[string]string // 额外环境变量
	ToolArgs []string          // 传递给工具的额外参数
	Yolo     bool              // 自动接受所有操作（YOLO 模式）
}

// MergeConfig 合并配置，返回最终参数
// 优先级: 命令行 > Profile > 默认值
func MergeConfig(cfg *LaunchConfig) (model, baseURL, token string) {
	if cfg.Profile != nil {
		if cfg.Profile.Model != "" {
			model = cfg.Profile.Model
		}
		baseURL = cfg.Profile.AnthropicBaseURL
		if cfg.Profile.Token != "" {
			token = cfg.Profile.Token
		}
	}

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

// Launch 使用 Claude Code 启动
func Launch(cfg *LaunchConfig) error {
	model, baseURL, token := MergeConfig(cfg)

	// 校验 Token
	if token == "" {
		return fmt.Errorf("未配置 API Token，无法启动")
	}

	// 构建合并后的 profile 用于 BuildSettings/BuildCommand
	effectiveProfile := &config.Profile{
		Name:             "default",
		AnthropicBaseURL: baseURL,
		Token:            token,
		Model:            model,
	}
	if cfg.Profile != nil {
		effectiveProfile.Name = cfg.Profile.Name
	}

	// 构建 --settings JSON（复用 BuildSettings）
	settings := BuildSettings(effectiveProfile)
	settingsJSON, _ := json.Marshal(settings)

	// 构建命令参数
	args := []string{"--settings", string(settingsJSON)}

	if model != "" {
		args = append(args, "--model", model)
	}

	if cfg.Yolo {
		args = append(args, "--dangerously-skip-permissions")
	}

	args = append(args, cfg.ToolArgs...)

	cmd := exec.Command("claude", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 打印启动信息
	profileName := effectiveProfile.Name
	fmt.Printf("🚀 使用配置 '%s' 启动 Claude Code...\n", profileName)
	if model != "" {
		fmt.Printf("   模型: %s\n", model)
	}
	if baseURL != "" {
		fmt.Printf("   Base URL: %s\n", baseURL)
	}
	fmt.Println()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch claude: %w", err)
	}

	return nil
}
```

- [ ] **Step 2: 重写 `internal/launcher/launcher_test.go`**

完整替换文件内容：

```go
package launcher

import (
	"encoding/json"
	"os"
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

			env, ok := settings["env"].(map[string]string)
			if !ok {
				t.Fatal("settings should have env map")
			}

			for _, key := range tt.wantKeys {
				if _, exists := env[key]; !exists {
					t.Errorf("missing key '%s' in settings", key)
				}
			}

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

	if !strings.Contains(cmd.Path, "claude") {
		t.Errorf("expected path to contain 'claude', got '%s'", cmd.Path)
	}

	foundModel := false
	for _, arg := range cmd.Args {
		if arg == "--model" {
			foundModel = true
		}
	}
	if !foundModel {
		t.Error("command should include --model flag")
	}

	foundSettings := false
	for _, arg := range cmd.Args {
		if arg == "--settings" {
			foundSettings = true
		}
	}
	if !foundSettings {
		t.Error("command should include --settings flag")
	}

	foundDangerously := false
	for _, arg := range cmd.Args {
		if arg == "--dangerously-skip-permissions" {
			foundDangerously = true
		}
	}
	if !foundDangerously {
		t.Error("command should include extra args")
	}

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
	profile := &config.Profile{
		Name:             "no-model",
		AnthropicBaseURL: "https://api.anthropic.com",
		Token:            "token123",
		Model:            "",
	}

	cmd := BuildCommand(profile, []string{})

	for i, arg := range cmd.Args {
		if arg == "--model" && i+1 < len(cmd.Args) && cmd.Args[i+1] != "" {
			t.Error("command should not include --model flag when model is empty")
		}
	}
}

func TestBuildSettingsEmptyBaseURL(t *testing.T) {
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

	jsonData, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("failed to marshal settings: %v", err)
	}

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
		Model:            "model-v1",
		Token:            "profile-token",
	}

	tests := []struct {
		name     string
		cfg      *LaunchConfig
		wantModel   string
		wantURL     string
		wantToken   string
	}{
		{
			name: "profile values used",
			cfg: &LaunchConfig{
				Profile: profile,
			},
			wantModel:   "model-v1",
			wantURL:     "https://anthropic.example.com",
			wantToken:   "profile-token",
		},
		{
			name: "command line overrides profile",
			cfg: &LaunchConfig{
				Profile: profile,
				Model:   "override-model",
				BaseURL: "https://override.com",
				Token:   "override-token",
			},
			wantModel:   "override-model",
			wantURL:     "https://override.com",
			wantToken:   "override-token",
		},
		{
			name: "partial override - model only",
			cfg: &LaunchConfig{
				Profile: profile,
				Model:   "new-model",
			},
			wantModel:   "new-model",
			wantURL:     "https://anthropic.example.com",
			wantToken:   "profile-token",
		},
		{
			name: "no profile no override",
			cfg: &LaunchConfig{},
			wantModel:   "",
			wantURL:     "",
			wantToken:   "",
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
```

- [ ] **Step 3: 简化 `cmd/claude.go`**

完整替换文件内容：

```go
// cmd/claude.go
package cmd

import (
	"github.com/spf13/cobra"
)

// claudeCmd 启动 Claude Code CLI
var claudeCmd = &cobra.Command{
	Use:   "claude [profile] [flags] [-- tool-args]",
	Short: "启动 Claude Code CLI",
	Long: `使用 Claude Code CLI 启动编程助手。

示例:
  cc-start claude                      使用默认配置启动
  cc-start claude moonshot             使用 moonshot 配置
  cc-start claude -m claude-sonnet-4   指定模型
  cc-start claude moonshot -- --help   传递参数给 claude`,
	RunE: runClaude,
}

func init() {
	rootCmd.AddCommand(claudeCmd)

	claudeCmd.Flags().StringVarP(&launchModel, "model", "m", "", "模型名称")
	claudeCmd.Flags().BoolVarP(&launchYolo, "yolo", "y", false, "自动接受所有操作（YOLO 模式）")
}

func runClaude(cmd *cobra.Command, args []string) error {
	return runLaunch(args, "claude")
}
```

- [ ] **Step 4: 简化 `cmd/launcher.go`**

完整替换文件内容：

```go
// cmd/launcher.go
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/wujunwei928/cc-start/internal/config"
	"github.com/wujunwei928/cc-start/internal/launcher"
)

var (
	launchModel   string
	launchBaseURL string
	launchToken   string
	launchEnv     []string
	launchYolo    bool
)

// runLaunch 执行启动逻辑
// cmdName 用于在 os.Args 中定位参数位置
func runLaunch(args []string, cmdName string) error {
	// 解析 profile 和工具参数
	var profileName string
	var toolArgs []string

	dashPos := findDashSeparator(os.Args)

	if dashPos != -1 {
		toolArgs = os.Args[dashPos+1:]
		for i := dashPos - 1; i >= 0; i-- {
			if os.Args[i] == cmdName {
				for j := i + 1; j < dashPos; j++ {
					arg := os.Args[j]
					if !isFlag(arg) && !isFlagValue(os.Args, j) {
						profileName = arg
						break
					}
				}
				break
			}
		}
	} else if len(args) > 0 {
		for _, arg := range args {
			if !isFlag(arg) && !isFlagValue(os.Args, findArgIndex(os.Args, arg)) {
				profileName = arg
				break
			}
		}
	}

	// 加载配置
	cfgPath := config.GetConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 获取 profile（未指定时使用默认配置）
	profile, err := cfg.GetProfile(profileName)
	if err != nil {
		return fmt.Errorf("获取配置失败: %w\n\n运行 'cc-start list' 查看可用配置", err)
	}

	// 解析环境变量
	envMap := make(map[string]string)
	for _, e := range launchEnv {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// 构建启动配置
	launchCfg := &launcher.LaunchConfig{
		Profile:  profile,
		Model:    launchModel,
		BaseURL:  launchBaseURL,
		Token:    launchToken,
		Env:      envMap,
		ToolArgs: toolArgs,
		Yolo:     launchYolo,
	}

	// 验证必要的配置
	if profile == nil && launchToken == "" {
		return fmt.Errorf("请指定 profile，运行 'cc-start list' 查看可用配置")
	}

	return launcher.Launch(launchCfg)
}

// findArgIndex 查找参数在数组中的索引
func findArgIndex(args []string, target string) int {
	for i, arg := range args {
		if arg == target {
			return i
		}
	}
	return -1
}

// isFlagValue 检查指定索引是否是某个 flag 的值
func isFlagValue(args []string, index int) bool {
	if index <= 0 || index >= len(args) {
		return false
	}
	return isFlag(args[index-1])
}
```

- [ ] **Step 5: 运行测试验证**

Run: `go test ./internal/launcher/... -v -timeout 30s`
Expected: 所有测试通过

- [ ] **Step 6: 验证编译**

Run: `go build ./...`
Expected: 编译通过（如果 repl 和 tui 还有 OpenAIBaseURL 引用会失败，这是预期的，后续 Task 修复）

- [ ] **Step 7: Commit**

```bash
git add internal/launcher/launcher.go internal/launcher/launcher_test.go cmd/claude.go cmd/launcher.go
git commit -m "refactor(launcher): 移除 tools 依赖，简化为统一的 Launch 入口"
```

---

### Task 4: 清理 CLI list 命令

**Files:**
- Modify: `cmd/list.go:46-48`

- [ ] **Step 1: 移除 OpenAI URL 展示并将标签改为 Base URL**

在 `cmd/list.go` 中删除第 46-48 行：

```go
// 删除以下代码：
if p.OpenAIBaseURL != "" {
    fmt.Printf("      OpenAI URL: %s\n", p.OpenAIBaseURL)
}
```

并将第 44 行的 `Anthropic URL` 改为 `Base URL`：
```go
fmt.Printf("      Base URL: %s\n", p.AnthropicBaseURL)
```

- [ ] **Step 2: 验证编译**

Run: `go build ./cmd/...`
Expected: 通过

- [ ] **Step 3: Commit**

```bash
git add cmd/list.go
git commit -m "refactor(cmd/list): 移除 OpenAI URL 展示"
```

---

### Task 5: 清理 REPL commands.go

**Files:**
- Modify: `internal/repl/commands.go:36,61,111,171,305,394-397`

- [ ] **Step 1: 修改 cmdList 表头和数据**

在 `cmdList` 函数中：

第 36 行，将表头从：
```go
table.Header([]string{"名称", "Anthropic URL", "OpenAI URL", "模型", "Token", "状态"})
```
改为：
```go
table.Header([]string{"名称", "Base URL", "模型", "Token", "状态"})
```

第 61 行，将 `p.OpenAIBaseURL` 改为删除该行，并修改 Anthropic URL 列名为 Base URL：
```go
table.Append([]string{
    name,
    p.AnthropicBaseURL,
    p.Model,
    maskAPIKey(p.Token),
    status,
})
```

- [ ] **Step 2: 修改 cmdCurrent**

第 110 行，将 `Anthropic URL` 改为 `Base URL`：
```go
fmt.Printf("  Base URL: %s\n", profile.AnthropicBaseURL)
```

第 111 行，删除：
```go
fmt.Printf("  OpenAI URL: %s\n", profile.OpenAIBaseURL)
```

- [ ] **Step 3: 修改 cmdShow**

第 170 行，将 `Anthropic URL` 改为 `Base URL`：
```go
fmt.Printf("Anthropic URL: %s\n", profile.AnthropicBaseURL)
```

第 171 行，删除：
```go
fmt.Printf("OpenAI URL: %s\n", profile.OpenAIBaseURL)
```

- [ ] **Step 4: 修改 cmdCopy**

第 305 行，从 Profile 创建中删除 `OpenAIBaseURL` 字段：
```go
newProfile := config.Profile{
    Name:             dstName,
    AnthropicBaseURL: src.AnthropicBaseURL,
    Model:            src.Model,
    Token:            src.Token,
}
```

- [ ] **Step 5: 修改 cmdTest**

第 394-397 行，删除 OpenAI URL 回退逻辑，只使用 AnthropicBaseURL：
```go
baseURL := profile.AnthropicBaseURL
if baseURL == "" {
    PrintWarning("未配置 Base URL")
    return
}
```

- [ ] **Step 6: 验证编译**

Run: `go build ./internal/repl/...`
Expected: 通过

- [ ] **Step 7: Commit**

```bash
git add internal/repl/commands.go
git commit -m "refactor(repl): 从 commands.go 移除 OpenAI URL 引用"
```

---

### Task 6: 清理 REPL update.go

**Files:**
- Modify: `internal/repl/update.go:443-444,515,598-599,880-881,906-907`

- [ ] **Step 1: 修改 cmdShow（update.go 版本）**

第 442 行，将 `Anthropic URL` 改为 `Base URL`：
```go
buf.WriteString(fmt.Sprintf("Base URL: %s\n", profile.AnthropicBaseURL))
```

第 443-444 行，删除：
```go
if profile.OpenAIBaseURL != "" {
    buf.WriteString(fmt.Sprintf("OpenAI URL: %s\n", profile.OpenAIBaseURL))
}
```

- [ ] **Step 2: 修改 cmdCopy（update.go 版本）**

第 515 行，从 Profile 创建中删除 `OpenAIBaseURL`：
```go
newProfile := config.Profile{
    Name:             dstName,
    AnthropicBaseURL: src.AnthropicBaseURL,
    Model:            src.Model,
    Token:            src.Token,
}
```

- [ ] **Step 3: 修改 cmdTest（update.go 版本）**

第 598-599 行，删除 OpenAI URL 回退：
```go
baseURL := profile.AnthropicBaseURL
if baseURL == "" {
    return result + "✗ 未配置 Base URL"
}
```

- [ ] **Step 4: 修改 formatProfileList**

第 880-881 行，删除：
```go
if p.OpenAIBaseURL != "" {
    buf.WriteString(fmt.Sprintf("    OpenAI URL: %s\n", p.OpenAIBaseURL))
}
```

并将第 877 行的 `Anthropic URL` 改为 `Base URL`：
```go
buf.WriteString(fmt.Sprintf("    Base URL: %s\n", p.AnthropicBaseURL))
```

- [ ] **Step 5: 修改 formatCurrentProfile**

第 906-907 行，删除：
```go
if profile.OpenAIBaseURL != "" {
    buf.WriteString(fmt.Sprintf("  OpenAI URL: %s\n", profile.OpenAIBaseURL))
}
```

并将第 903 行的 `Anthropic URL` 改为 `Base URL`：
```go
buf.WriteString(fmt.Sprintf("  Base URL: %s\n", profile.AnthropicBaseURL))
```

- [ ] **Step 6: 验证编译**

Run: `go build ./internal/repl/...`
Expected: 通过

- [ ] **Step 7: Commit**

```bash
git add internal/repl/update.go
git commit -m "refactor(repl): 从 update.go 移除 OpenAI URL 引用"
```

---

### Task 7: 清理 TUI Setup 向导

**Files:**
- Modify: `internal/tui/setup/model.go`

- [ ] **Step 1: 移除 stepInputOpenAIURL 步骤定义**

删除第 21 行：
```go
stepInputOpenAIURL
```

- [ ] **Step 2: 移除 openaiURLInput 字段和相关初始化**

从 `Model` 结构体中删除 `openaiURLInput textinput.Model`（第 56 行）。

从 `InitialModel()` 中删除 openaiURLInput 初始化（第 83-84 行和第 94 行）。

从 `InitialModelWithProfile()` 中删除 openaiURLInput 初始化（第 119-121 行和第 143 行）。

- [ ] **Step 3: 修改步骤流转**

在 `handleEnter()` 中：
- `stepInputAnthropicURL` case：改为直接跳到 `stepInputToken`
```go
case stepInputAnthropicURL:
    m.step = stepInputToken
    m.anthropicURLInput.Blur()
    m.tokenInput.Focus()
```
- 删除整个 `stepInputOpenAIURL` case

在 `handleGoBack()` 中：
- `stepInputToken` case：改为返回 `stepInputAnthropicURL`
```go
case stepInputToken:
    if m.isEdit || m.isCustom {
        m.step = stepInputAnthropicURL
        m.tokenInput.Blur()
        m.anthropicURLInput.Focus()
    } else {
        m.step = stepInputName
        m.tokenInput.Blur()
        m.nameInput.Focus()
    }
```
- 删除整个 `stepInputOpenAIURL` case

在 `Update()` 中：删除 `stepInputOpenAIURL` 的输入处理（第 195-198 行）。

在 `handleBackspace()` 中：删除 `stepInputOpenAIURL` 的处理（第 338-344 行）。

- [ ] **Step 4: 修改 saveProfile**

从 Profile 创建中删除 `OpenAIBaseURL`（第 367 行）：
```go
m.profile = &config.Profile{
    Name:             m.nameInput.Value(),
    AnthropicBaseURL: m.anthropicURLInput.Value(),
    Token:            m.tokenInput.Value(),
    Model:            m.modelInput.Value(),
}
```

- [ ] **Step 5: 修改 View 渲染**

删除整个 `stepInputOpenAIURL` case（第 438-445 行）。

在 `stepInputToken` case 中：删除 OpenAI URL 展示（第 452-454 行）。

在 `stepInputModel` case 中：删除 OpenAI URL 展示（第 464-466 行）。

- [ ] **Step 6: 验证编译**

Run: `go build ./internal/tui/setup/...`
Expected: 通过

- [ ] **Step 7: Commit**

```bash
git add internal/tui/setup/model.go
git commit -m "refactor(tui): 移除 OpenAI URL 输入步骤"
```

---

### Task 8: 更新 README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: 移除 codex/opencode 使用示例**

在 README.md 中删除第 54-61 行：
```markdown
# 启动 OpenAI Codex CLI（指定模型）
cc-start codex -m gpt-4

# 传递参数给工具
cc-start claude moonshot -- --dangerously-skip-permissions

# 启动 OpenCode
cc-start opencode deepseek
```

保留 `cc-start claude moonshot -- --dangerously-skip-permissions` 示例，但调整为：
```markdown
# 传递参数给 claude
cc-start claude moonshot -- --dangerously-skip-permissions
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs(readme): 移除 codex/opencode 使用示例"
```

---

### Task 9: 全量验证

**Files:** 无新变更，验证所有修改

- [ ] **Step 1: 运行完整测试**

Run: `go test ./... -timeout 30s`
Expected: 所有测试通过

- [ ] **Step 2: 编译验证**

Run: `go build ./...`
Expected: 编译通过，无错误

- [ ] **Step 3: 搜索残留引用**

Run: `grep -r "OpenAI" --include="*.go" .`
Expected: 无结果（排除 vendor）

Run: `grep -r "opencode\|codex" --include="*.go" .`
Expected: 无结果

Run: `grep -r "tools\." --include="*.go" ./cmd/ ./internal/`
Expected: 无对 `internal/tools` 包的引用

- [ ] **Step 4: 最终确认**

确认所有变更无误后，计划完成。

# Launch 命令实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 run 命令重构为 launch 命令，支持 claude、codex、opencode 三种 AI CLI 工具

**Architecture:** 新增 tools 模块定义工具预设，重构 launcher 模块支持多工具启动，新增 launch 命令，删除 run 命令

**Tech Stack:** Go, Cobra CLI, spf13/pflag

---

## Task 1: 创建 tools 模块 - 工具预设定义

**Files:**
- Create: `internal/tools/tools.go`

**Step 1: 编写工具预设模块**

```go
// internal/tools/tools.go
package tools

import (
	"fmt"
)

// 参数类型常量
const (
	ParamToken   = "token"
	ParamBaseURL = "base_url"
)

// Tool 工具预设
type Tool struct {
	Name       string            // 工具名
	Executable string            // 可执行文件名
	EnvMap     map[string]string // 参数到环境变量的映射
}

// GetEnvName 获取指定参数对应的环境变量名
func (t *Tool) GetEnvName(param string) string {
	return t.EnvMap[param]
}

// 内置工具预设
var builtInTools = map[string]Tool{
	"claude": {
		Name:       "claude",
		Executable: "claude",
		EnvMap: map[string]string{
			ParamToken:   "ANTHROPIC_AUTH_TOKEN",
			ParamBaseURL: "ANTHROPIC_BASE_URL",
		},
	},
	"codex": {
		Name:       "codex",
		Executable: "codex",
		EnvMap: map[string]string{
			ParamToken:   "OPENAI_API_KEY",
			ParamBaseURL: "OPENAI_BASE_URL",
		},
	},
	"opencode": {
		Name:       "opencode",
		Executable: "opencode",
		EnvMap: map[string]string{
			ParamToken:   "OPENAI_API_KEY",
			ParamBaseURL: "OPENAI_BASE_URL",
		},
	},
}

// GetTool 获取工具预设
func GetTool(name string) (*Tool, error) {
	tool, ok := builtInTools[name]
	if !ok {
		return nil, fmt.Errorf("未知工具: %s\n可用工具: claude, codex, opencode", name)
	}
	return &tool, nil
}

// ListTools 返回所有可用工具名
func ListTools() []string {
	names := make([]string, 0, len(builtInTools))
	for name := range builtInTools {
		names = append(names, name)
	}
	return names
}
```

**Step 2: 验证编译**

Run: `cd /code/ai/cc-start && go build ./internal/tools`
Expected: 无错误输出

**Step 3: Commit**

```bash
git add internal/tools/tools.go
git commit -m "feat(tools): 添加工具预设模块

支持 claude、codex、opencode 三种工具

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 2: 为 tools 模块编写单元测试

**Files:**
- Create: `internal/tools/tools_test.go`

**Step 1: 编写测试**

```go
// internal/tools/tools_test.go
package tools

import (
	"testing"
)

func TestGetTool(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		wantErr   bool
		wantExec  string
	}{
		{
			name:      "get claude tool",
			toolName:  "claude",
			wantErr:   false,
			wantExec:  "claude",
		},
		{
			name:      "get codex tool",
			toolName:  "codex",
			wantErr:   false,
			wantExec:  "codex",
		},
		{
			name:      "get opencode tool",
			toolName:  "opencode",
			wantErr:   false,
			wantExec:  "opencode",
		},
		{
			name:      "unknown tool",
			toolName:  "unknown",
			wantErr:   true,
			wantExec:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, err := GetTool(tt.toolName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tool.Executable != tt.wantExec {
				t.Errorf("GetTool().Executable = %v, want %v", tool.Executable, tt.wantExec)
			}
		})
	}
}

func TestToolGetEnvName(t *testing.T) {
	tool, _ := GetTool("claude")

	tests := []struct {
		param    string
		expected string
	}{
		{ParamToken, "ANTHROPIC_AUTH_TOKEN"},
		{ParamBaseURL, "ANTHROPIC_BASE_URL"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.param, func(t *testing.T) {
			got := tool.GetEnvName(tt.param)
			if got != tt.expected {
				t.Errorf("GetEnvName(%s) = %v, want %v", tt.param, got, tt.expected)
			}
		})
	}
}

func TestListTools(t *testing.T) {
	tools := ListTools()
	if len(tools) != 3 {
		t.Errorf("ListTools() returned %d tools, want 3", len(tools))
	}
}
```

**Step 2: 运行测试**

Run: `cd /code/ai/cc-start && go test ./internal/tools -v`
Expected: 所有测试通过

**Step 3: Commit**

```bash
git add internal/tools/tools_test.go
git commit -m "test(tools): 添加工具预设模块单元测试

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 3: 重构 launcher 模块 - 添加 LaunchConfig 结构

**Files:**
- Modify: `internal/launcher/launcher.go`

**Step 1: 添加 LaunchConfig 结构和合并逻辑**

在现有 `launcher.go` 文件中添加（在 `Launch` 函数之后）：

```go
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
```

**Step 2: 验证编译**

Run: `cd /code/ai/cc-start && go build ./internal/launcher`
Expected: 无错误输出

**Step 3: Commit**

```bash
git add internal/launcher/launcher.go
git commit -m "feat(launcher): 添加 LaunchConfig 和配置合并逻辑

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 4: 为 launcher 添加 LaunchWithTool 函数

**Files:**
- Modify: `internal/launcher/launcher.go`

**Step 1: 添加导入**

将 import 块修改为：

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/wujunwei928/cc-start/internal/config"
	"github.com/wujunwei928/cc-start/internal/tools"
)
```

**Step 2: 添加 LaunchWithTool 函数**

在文件末尾添加：

```go
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
```

**Step 2: 验证编译**

Run: `cd /code/ai/cc-start && go build ./internal/launcher`
Expected: 无错误输出

**Step 3: Commit**

```bash
git add internal/launcher/launcher.go
git commit -m "feat(launcher): 添加 LaunchWithTool 支持多工具启动

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 5: 为 launcher 添加测试

**Files:**
- Modify: `internal/launcher/launcher_test.go`

**Step 1: 添加测试**

在文件末尾添加：

```go
func TestMergeConfig(t *testing.T) {
	profile := &config.Profile{
		Name:    "test",
		BaseURL: "https://api.example.com",
		Model:   "model-v1",
		Token:   "profile-token",
	}

	tests := []struct {
		name      string
		cfg       *LaunchConfig
		wantModel string
		wantURL   string
		wantToken string
	}{
		{
			name: "only profile",
			cfg: &LaunchConfig{
				Profile: profile,
			},
			wantModel: "model-v1",
			wantURL:   "https://api.example.com",
			wantToken: "profile-token",
		},
		{
			name: "command line overrides profile",
			cfg: &LaunchConfig{
				Profile:  profile,
				Model:    "override-model",
				BaseURL:  "https://override.com",
				Token:    "override-token",
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
			wantURL:   "https://api.example.com",
			wantToken: "profile-token",
		},
		{
			name: "no profile no override",
			cfg: &LaunchConfig{
				Tool: "claude",
			},
			wantModel: "",
			wantURL:   "",
			wantToken: "",
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

**Step 2: 运行测试**

Run: `cd /code/ai/cc-start && go test ./internal/launcher -v -run TestMergeConfig`
Expected: 测试通过

**Step 3: Commit**

```bash
git add internal/launcher/launcher_test.go
git commit -m "test(launcher): 添加 MergeConfig 单元测试

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 6: 创建 launch 命令

**Files:**
- Create: `cmd/launch.go`

**Step 1: 编写 launch 命令**

```go
// cmd/launch.go
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wujunwei928/cc-start/internal/config"
	"github.com/wujunwei928/cc-start/internal/launcher"
	"github.com/wujunwei928/cc-start/internal/tools"
)

var (
	launchModel   string
	launchBaseURL string
	launchToken   string
	launchEnv     []string
)

// launchCmd 启动 AI 工具命令
var launchCmd = &cobra.Command{
	Use:   "launch <tool> [profile] [flags] [-- tool-args]",
	Short: "启动 AI 编程助手",
	Long: `使用指定的 AI 工具和配置启动编程助手。

工具:
  claude    Anthropic Claude Code CLI
  codex     OpenAI Codex CLI
  opencode  OpenCode AI 编程助手

示例:
  cc-start launch claude                      使用默认配置启动 claude
  cc-start launch claude moonshot             使用 moonshot 配置
  cc-start launch codex -m gpt-4 -t sk-xxx    指定模型和令牌
  cc-start launch claude moonshot -e DEBUG=true -- --help`,
	Args: cobra.MinimumNArgs(1),
	RunE: runLaunch,
}

func init() {
	rootCmd.AddCommand(launchCmd)

	launchCmd.Flags().StringVarP(&launchModel, "model", "m", "", "模型名称")
	launchCmd.Flags().StringVarP(&launchBaseURL, "base-url", "b", "", "API 基础地址")
	launchCmd.Flags().StringVarP(&launchToken, "token", "t", "", "认证令牌")
	launchCmd.Flags().StringArrayVarP(&launchEnv, "env", "e", nil, "环境变量 (格式: KEY=VALUE)")
}

func runLaunch(cmd *cobra.Command, args []string) error {
	// 第一个参数是工具名
	toolName := args[0]

	// 验证工具名
	if _, err := tools.GetTool(toolName); err != nil {
		return err
	}

	// 解析 profile 和工具参数
	var profileName string
	var toolArgs []string

	remainingArgs := args[1:]
	dashPos := findDashSeparator(os.Args)

	if dashPos != -1 {
		// 有 -- 分隔符
		toolArgs = os.Args[dashPos+1:]
		// 找 launch 之后、-- 之前的非 flag 参数作为 profile
		for i := dashPos - 1; i >= 0; i-- {
			if os.Args[i] == "launch" {
				for j := i + 1; j < dashPos; j++ {
					arg := os.Args[j]
					if !isFlag(arg) && arg != toolName && !isFlagValue(os.Args, j) {
						profileName = arg
						break
					}
				}
				break
			}
		}
	} else if len(remainingArgs) > 0 {
		// 无 -- 分隔符，第一个非 flag 参数是 profile
		for _, arg := range remainingArgs {
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

	// 获取 profile（如果指定）
	var profile *config.Profile
	if profileName != "" {
		profile, err = cfg.GetProfile(profileName)
		if err != nil {
			return fmt.Errorf("获取配置失败: %w\n\n运行 'cc-start list' 查看可用配置", err)
		}
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
		Tool:     toolName,
		Profile:  profile,
		Model:    launchModel,
		BaseURL:  launchBaseURL,
		Token:    launchToken,
		Env:      envMap,
		ToolArgs: toolArgs,
	}

	// 验证必要的配置
	if profile == nil && launchToken == "" {
		return fmt.Errorf("请通过 -t 指定令牌或指定 profile")
	}

	return launcher.LaunchWithTool(launchCfg)
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

**Step 2: 验证编译**

Run: `cd /code/ai/cc-start && go build ./cmd`
Expected: 无错误输出

**Step 3: Commit**

```bash
git add cmd/launch.go
git commit -m "feat(cmd): 添加 launch 命令

支持 claude、codex、opencode 三种工具
支持 -m/-b/-t/-e 短标记

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 7: 删除 run 命令

**Files:**
- Delete: `cmd/run.go`

**Step 1: 删除文件**

Run: `rm cmd/run.go`

**Step 2: 验证编译**

Run: `cd /code/ai/cc-start && go build ./...`
Expected: 无错误输出

**Step 3: Commit**

```bash
git add -A
git commit -m "refactor(cmd): 删除 run 命令

使用 launch 命令替代

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 8: 从 REPL 中删除 run 命令

**Files:**
- Modify: `internal/repl/commands.go`

**Step 1: 删除 cmdRun 函数**

删除 `cmdRun` 函数（约第 774-824 行）

**Step 2: 删除 ExecuteCommand 中的 run case**

在 `ExecuteCommand` 函数中删除：

```go
	// 启动
	case "/run":
		r.cmdRun(args)
```

**Step 3: 删除 cmdHelp 中的 run 帮助**

在 `cmdHelp` 函数中删除：

```go
	// 启动
	fmt.Println("启动 Claude Code:")
	fmt.Println("  /run [profile] [-- args...]  使用当前或指定配置启动")
	fmt.Println("  /setup              运行配置向导")
```

改为：

```go
	// 配置
	fmt.Println("配置:")
	fmt.Println("  /setup              运行配置向导")
```

**Step 4: 删除 showCommandHelp 中的 /run 帮助**

删除 `"/run"` 的帮助文本（约第 700-710 行）

**Step 5: 验证编译**

Run: `cd /code/ai/cc-start && go build ./...`
Expected: 无错误输出

**Step 6: Commit**

```bash
git add internal/repl/commands.go
git commit -m "refactor(repl): 删除 run 命令

启动功能移至命令行 launch 命令

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 9: 更新 cmd/root.go 帮助文本

**Files:**
- Modify: `cmd/root.go`

**Step 1: 更新 Long 描述**

将：

```go
	Long: `CC-Start 是一个 Claude Code 启动器，帮助你管理多个 API 供应商配置。

使用方法:
  cc-start           进入交互式 REPL
  cc-start run       启动 Claude Code
  cc-start setup     配置向导
  cc-start list      列出所有配置`,
```

改为：

```go
	Long: `CC-Start 是一个 AI 编程助手启动器，帮助你管理多个 API 供应商配置。

使用方法:
  cc-start              进入交互式 REPL
  cc-start launch       启动 AI 编程助手
  cc-start setup        配置向导
  cc-start list         列出所有配置`,
```

**Step 2: 更新 Short 描述**

将：

```go
	Short: "Claude Code 启动器 - 快速切换不同供应商",
```

改为：

```go
	Short: "AI 编程助手启动器 - 快速切换不同供应商",
```

**Step 3: 验证编译**

Run: `cd /code/ai/cc-start && go build ./...`
Expected: 无错误输出

**Step 4: Commit**

```bash
git add cmd/root.go
git commit -m "docs(cmd): 更新帮助文本

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 10: 更新 README.md

**Files:**
- Modify: `README.md`

**Step 1: 更新标题和描述**

将标题改为：

```markdown
# CC-Start

AI 编程助手启动器 - 快速切换不同供应商。
```

**Step 2: 更新直接启动部分**

将：

```markdown
### 直接启动 Claude Code

```bash
# 使用默认配置启动
cc-start run

# 使用指定配置启动
cc-start run moonshot

# 传递参数给 claude
cc-start run -- --dangerously-skip-permissions
```
```

改为：

```markdown
### 启动 AI 编程助手

```bash
# 使用默认配置启动 claude
cc-start launch claude

# 使用指定配置启动 claude
cc-start launch claude moonshot

# 指定模型和令牌启动 codex
cc-start launch codex -m gpt-4 -t sk-xxx

# 传递参数给工具
cc-start launch claude moonshot -- --dangerously-skip-permissions

# 添加环境变量
cc-start launch claude moonshot -e DEBUG=true
```
```

**Step 3: 更新 REPL 命令表**

删除 `run` 行：

```markdown
| `run [profile]` | 启动 Claude Code |
```

**Step 4: Commit**

```bash
git add README.md
git commit -m "docs: 更新 README 使用 launch 命令

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 11: 运行完整测试

**Step 1: 运行所有测试**

Run: `cd /code/ai/cc-start && go test ./... -v`
Expected: 所有测试通过

**Step 2: 运行静态检查**

Run: `cd /code/ai/cc-start && go vet ./...`
Expected: 无警告

**Step 3: 运行格式化**

Run: `cd /code/ai/cc-start && go fmt ./...`
Expected: 无输出（或格式已正确）

---

## Task 12: 最终验证

**Step 1: 检查变更**

Run: `git status`
Expected: 工作区干净

**Step 2: 检查提交历史**

Run: `git log --oneline -15`
Expected: 看到所有新提交

**Step 3: 构建并测试命令帮助**

Run: `cd /code/ai/cc-start && go build -o cc-start . && ./cc-start launch --help`
Expected: 显示 launch 命令帮助信息

**Step 4: 清理构建产物**

Run: `rm -f cc-start`

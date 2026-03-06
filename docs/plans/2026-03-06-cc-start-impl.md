# CC-Start 实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现一个 Go 语言的 Claude Code 启动器，支持多供应商配置管理和快速切换。

**Architecture:** 采用标准 Go 项目布局，CLI 使用 cobra 框架，TUI 使用 bubbletea。配置以 JSON 文件存储在 `~/.cc-start/profiles.json`。启动时构建 settings JSON 并传递给 claude 命令。

**Tech Stack:** Go 1.24+, cobra (CLI), bubbletea + bubbles (TUI), lipgloss (样式)

---

## Task 1: 配置数据模型

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: 编写配置结构测试**

```go
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
```

**Step 2: 运行测试验证失败**

Run: `cd /code/ai/cc-start && go test ./internal/config/... -v`
Expected: FAIL - 包不存在

**Step 3: 实现配置结构**

```go
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
```

**Step 4: 运行测试验证通过**

Run: `cd /code/ai/cc-start && go test ./internal/config/... -v`
Expected: PASS

**Step 5: 提交**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add config data model with load/save support"
```

---

## Task 2: 内置预设

**Files:**
- Create: `internal/config/presets.go`
- Create: `internal/config/presets_test.go`

**Step 1: 编写预设测试**

```go
// internal/config/presets_test.go
package config

import "testing"

func TestGetPresets(t *testing.T) {
	presets := GetPresets()

	expectedPresets := []string{"anthropic", "moonshot", "bigmodel", "deepseek"}
	if len(presets) != len(expectedPresets) {
		t.Errorf("expected %d presets, got %d", len(expectedPresets), len(presets))
	}

	for _, name := range expectedPresets {
		found := false
		for _, p := range presets {
			if p.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("preset '%s' not found", name)
		}
	}
}

func TestGetPresetByName(t *testing.T) {
	tests := []struct {
		name     string
		expected *Profile
	}{
		{
			name: "anthropic",
			expected: &Profile{
				Name:    "anthropic",
				BaseURL: "https://api.anthropic.com",
				Model:   "claude-sonnet-4-5-20250929",
			},
		},
		{
			name: "moonshot",
			expected: &Profile{
				Name:    "moonshot",
				BaseURL: "https://api.moonshot.cn/anthropic",
				Model:   "moonshot-v1-8k",
			},
		},
		{
			name: "bigmodel",
			expected: &Profile{
				Name:    "bigmodel",
				BaseURL: "https://open.bigmodel.cn/api/anthropic",
				Model:   "glm-4-plus",
			},
		},
		{
			name: "deepseek",
			expected: &Profile{
				Name:    "deepseek",
				BaseURL: "https://api.deepseek.com",
				Model:   "deepseek-chat",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetPresetByName(tt.name)
			if err != nil {
				t.Fatalf("GetPresetByName failed: %v", err)
			}
			if p.Name != tt.expected.Name {
				t.Errorf("expected name '%s', got '%s'", tt.expected.Name, p.Name)
			}
			if p.BaseURL != tt.expected.BaseURL {
				t.Errorf("expected baseURL '%s', got '%s'", tt.expected.BaseURL, p.BaseURL)
			}
			if p.Model != tt.expected.Model {
				t.Errorf("expected model '%s', got '%s'", tt.expected.Model, p.Model)
			}
		})
	}
}

func TestGetPresetByNameNotFound(t *testing.T) {
	_, err := GetPresetByName("notexist")
	if err == nil {
		t.Error("expected error for non-existent preset")
	}
}
```

**Step 2: 运行测试验证失败**

Run: `cd /code/ai/cc-start && go test ./internal/config/... -run TestGetPreset -v`
Expected: FAIL - 函数不存在

**Step 3: 实现预设**

```go
// internal/config/presets.go
package config

import "fmt"

// Preset 内置预设配置
var presets = []Profile{
	{
		Name:    "anthropic",
		BaseURL: "https://api.anthropic.com",
		Model:   "claude-sonnet-4-5-20250929",
	},
	{
		Name:    "moonshot",
		BaseURL: "https://api.moonshot.cn/anthropic",
		Model:   "moonshot-v1-8k",
	},
	{
		Name:    "bigmodel",
		BaseURL: "https://open.bigmodel.cn/api/anthropic",
		Model:   "glm-4-plus",
	},
	{
		Name:    "deepseek",
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-chat",
	},
}

// GetPresets 返回所有内置预设
func GetPresets() []Profile {
	return presets
}

// GetPresetByName 根据名称获取预设
func GetPresetByName(name string) (*Profile, error) {
	for i := range presets {
		if presets[i].Name == name {
			return &presets[i], nil
		}
	}
	return nil, fmt.Errorf("preset '%s' not found", name)
}
```

**Step 4: 运行测试验证通过**

Run: `cd /code/ai/cc-start && go test ./internal/config/... -v`
Expected: PASS

**Step 5: 提交**

```bash
git add internal/config/presets.go internal/config/presets_test.go
git commit -m "feat(config): add built-in presets for common providers"
```

---

## Task 3: 启动器核心逻辑

**Files:**
- Create: `internal/launcher/launcher.go`
- Create: `internal/launcher/launcher_test.go`

**Step 1: 编写启动器测试**

```go
// internal/launcher/launcher_test.go
package launcher

import (
	"testing"

	"github.com/wujunwei/cc-start/internal/config"
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
				Name:    "anthropic",
				BaseURL: "https://api.anthropic.com",
				Token:   "sk-ant-xxx",
			},
			wantKeys: []string{"ANTHROPIC_AUTH_TOKEN"},
		},
		{
			name: "custom provider",
			profile: config.Profile{
				Name:    "moonshot",
				BaseURL: "https://api.moonshot.cn/anthropic",
				Token:   "sk-xxx",
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
			if tt.profile.BaseURL == "https://api.anthropic.com" {
				if _, exists := env["ANTHROPIC_BASE_URL"]; exists {
					t.Error("official API should not have ANTHROPIC_BASE_URL")
				}
			}
		})
	}
}

func TestBuildCommand(t *testing.T) {
	profile := &config.Profile{
		Name:    "test",
		BaseURL: "https://api.example.com",
		Token:   "token123",
		Model:   "test-model",
	}

	args := []string{"--dangerously-skip-permissions"}
	cmd := BuildCommand(profile, args)

	// 验证命令包含必要的参数
	if cmd.Path != "claude" {
		t.Errorf("expected 'claude', got '%s'", cmd.Path)
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
}
```

**Step 2: 运行测试验证失败**

Run: `cd /code/ai/cc-start && go test ./internal/launcher/... -v`
Expected: FAIL - 包不存在

**Step 3: 实现启动器**

```go
// internal/launcher/launcher.go
package launcher

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/wujunwei/cc-start/internal/config"
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
```

**Step 4: 运行测试验证通过**

Run: `cd /code/ai/cc-start && go test ./internal/launcher/... -v`
Expected: PASS

**Step 5: 提交**

```bash
git add internal/launcher/launcher.go internal/launcher/launcher_test.go
git commit -m "feat(launcher): add claude code launcher with settings builder"
```

---

## Task 4: 根命令和入口

**Files:**
- Create: `cmd/root.go`
- Create: `main.go`

**Step 1: 创建根命令**

```go
// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wujunwei/cc-start/internal/config"
	"github.com/wujunwei/cc-start/internal/launcher"
)

var (
	// 版本信息
	Version = "dev"
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "cc-start [profile] [-- claude args...]",
	Short: "Claude Code 启动器 - 快速切换不同供应商",
	Long: `CC-Start 是一个 Claude Code 启动器，帮助你管理多个 API 供应商配置。

使用方法:
  cc-start              使用默认配置启动
  cc-start moonshot     使用 moonshot 配置启动
  cc-start -- --help    传递参数给 claude`,
	Version: Version,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runRoot,
}

func init() {
	rootCmd.SetVersionTemplate("cc-start {{.Version}}\n")
}

func runRoot(cmd *cobra.Command, args []string) error {
	// 确定使用的配置名
	profileName := ""
	if len(args) > 0 {
		profileName = args[0]
	}

	// 加载配置
	cfgPath := config.GetConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 获取配置
	profile, err := cfg.GetProfile(profileName)
	if err != nil {
		return fmt.Errorf("获取配置失败: %w\n\n运行 'cc-start setup' 创建配置", err)
	}

	// 获取传递给 claude 的参数
	extraArgs := cmd.Flags().Args()
	if len(args) > 0 && len(extraArgs) > 0 && extraArgs[0] == args[0] {
		extraArgs = extraArgs[1:]
	}

	// 启动
	return launcher.Launch(profile, extraArgs)
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

**Step 2: 创建入口文件**

```go
// main.go
package main

import "github.com/wujunwei/cc-start/cmd"

func main() {
	cmd.Execute()
}
```

**Step 3: 验证编译**

Run: `cd /code/ai/cc-start && go build -o cc-start .`
Expected: 编译成功，无错误

**Step 4: 验证命令帮助**

Run: `cd /code/ai/cc-start && ./cc-start --help`
Expected: 显示帮助信息

**Step 5: 提交**

```bash
git add cmd/root.go main.go
git commit -m "feat(cmd): add root command and main entry point"
```

---

## Task 5: list 命令

**Files:**
- Create: `cmd/list.go`

**Step 1: 实现 list 命令**

```go
// cmd/list.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wujunwei/cc-start/internal/config"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出所有配置",
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	cfgPath := config.GetConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println("暂无配置，运行 'cc-start setup' 创建配置")
		return nil
	}

	fmt.Println("已保存的配置:")
	fmt.Println()

	for _, p := range cfg.Profiles {
		marker := " "
		if p.Name == cfg.Default {
			marker = "*"
		}
		fmt.Printf("  %s %s\n", marker, p.Name)
		fmt.Printf("      URL: %s\n", p.BaseURL)
		if p.Model != "" {
			fmt.Printf("      模型: %s\n", p.Model)
		}
		fmt.Printf("      Token: %s...\n\n", maskToken(p.Token))
	}

	return nil
}

// maskToken 隐藏 Token 大部分内容
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
```

**Step 2: 验证编译和命令**

Run: `cd /code/ai/cc-start && go build -o cc-start . && ./cc-start list`
Expected: 显示"暂无配置"消息

**Step 3: 提交**

```bash
git add cmd/list.go
git commit -m "feat(cmd): add list command to show saved profiles"
```

---

## Task 6: default 命令

**Files:**
- Create: `cmd/default.go`

**Step 1: 实现 default 命令**

```go
// cmd/default.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wujunwei/cc-start/internal/config"
)

var defaultCmd = &cobra.Command{
	Use:   "default <name>",
	Short: "设置默认配置",
	Args:  cobra.ExactArgs(1),
	RunE:  runDefault,
}

func init() {
	rootCmd.AddCommand(defaultCmd)
}

func runDefault(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfgPath := config.GetConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if err := cfg.SetDefault(name); err != nil {
		return err
	}

	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	fmt.Printf("✅ 已设置 '%s' 为默认配置\n", name)
	return nil
}
```

**Step 2: 验证编译**

Run: `cd /code/ai/cc-start && go build -o cc-start .`
Expected: 编译成功

**Step 3: 提交**

```bash
git add cmd/default.go
git commit -m "feat(cmd): add default command to set default profile"
```

---

## Task 7: delete 命令

**Files:**
- Create: `cmd/delete.go`

**Step 1: 实现 delete 命令**

```go
// cmd/delete.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wujunwei/cc-start/internal/config"
)

var (
	deleteForce bool
)

var deleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"rm"},
	Short:   "删除配置",
	Args:    cobra.ExactArgs(1),
	RunE:    runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "强制删除，不确认")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfgPath := config.GetConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 检查配置是否存在
	_, err = cfg.GetProfile(name)
	if err != nil {
		return err
	}

	// 确认删除
	if !deleteForce {
		fmt.Printf("确定要删除配置 '%s'? [y/N] ", name)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			fmt.Println("已取消")
			return nil
		}
	}

	if err := cfg.DeleteProfile(name); err != nil {
		return err
	}

	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	fmt.Printf("✅ 已删除配置 '%s'\n", name)
	return nil
}
```

**Step 2: 验证编译**

Run: `cd /code/ai/cc-start && go build -o cc-start .`
Expected: 编译成功

**Step 3: 提交**

```bash
git add cmd/delete.go
git commit -m "feat(cmd): add delete command with confirmation"
```

---

## Task 8: setup 命令 TUI - 基础框架

**Files:**
- Create: `cmd/setup.go`
- Create: `internal/tui/setup/model.go`

**Step 1: 创建 TUI model 基础结构**

```go
// internal/tui/setup/model.go
package setup

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// 步骤状态
type step int

const (
	stepSelectPreset step = iota
	stepInputName
	stepInputToken
	stepInputModel
	stepConfirm
)

// Model setup TUI 模型
type Model struct {
	step        step
	presets     []string
	selected    int
	nameInput   textinput.Model
	tokenInput  textinput.Model
	modelInput  textinput.Model
	isCustom    bool
	err         error
}

// InitialModel 创建初始模型
func InitialModel() Model {
	nameInput := textinput.New()
	nameInput.Placeholder = "配置名称（如 my-api）"
	nameInput.Focus()

	tokenInput := textinput.New()
	tokenInput.Placeholder = "API Token"
	tokenInput.EchoMode = textinput.EchoPassword
	tokenInput.EchoCharacter = '•'

	modelInput := textinput.New()
	modelInput.Placeholder = "模型名称（可选，按回车跳过）"

	return Model{
		step:       stepSelectPreset,
		presets:    []string{"anthropic", "moonshot", "bigmodel", "deepseek", "自定义"},
		selected:   0,
		nameInput:  nameInput,
		tokenInput: tokenInput,
		modelInput: modelInput,
	}
}

// Init 初始化
func (m Model) Init() tea.Cmd {
	return nil
}
```

**Step 2: 创建 setup 命令框架**

```go
// cmd/setup.go
package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/wujunwei/cc-start/internal/tui/setup"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "交互式配置向导",
	RunE:  runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	m := setup.InitialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("启动 TUI 失败: %w", err)
	}

	if final, ok := result.(setup.Model); ok && final.Done() {
		fmt.Printf("\n✅ 配置 '%s' 已保存\n", final.GetName())
	}

	return nil
}
```

**Step 3: 验证编译**

Run: `cd /code/ai/cc-start && go build -o cc-start .`
Expected: 编译失败 - 缺少 Done() 和 GetName() 方法

**Step 4: 添加缺失的方法**

```go
// 在 internal/tui/setup/model.go 末尾添加

// Done 返回是否完成
func (m Model) Done() bool {
	return false // 后续实现
}

// GetName 返回配置名
func (m Model) GetName() string {
	return m.nameInput.Value()
}
```

**Step 5: 验证编译**

Run: `cd /code/ai/cc-start && go build -o cc-start .`
Expected: 编译成功

**Step 6: 提交**

```bash
git add cmd/setup.go internal/tui/setup/model.go
git commit -m "feat(cmd): add setup command with TUI skeleton"
```

---

## Task 9: setup 命令 TUI - 更新和视图

**Files:**
- Modify: `internal/tui/setup/model.go`

**Step 1: 实现 Update 和 View 方法**

```go
// internal/tui/setup/model.go (完整实现)
package setup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wujunwei/cc-start/internal/config"
)

// 步骤状态
type step int

const (
	stepSelectPreset step = iota
	stepInputName
	stepInputToken
	stepInputModel
	stepConfirm
	stepDone
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Padding(1, 0)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))
)

// Model setup TUI 模型
type Model struct {
	step        step
	presets     []string
	selected    int
	nameInput   textinput.Model
	tokenInput  textinput.Model
	modelInput  textinput.Model
	isCustom    bool
	presetName  string
	baseURL     string
	err         error
	profile     *config.Profile
}

// InitialModel 创建初始模型
func InitialModel() Model {
	nameInput := textinput.New()
	nameInput.Placeholder = "配置名称（如 my-api）"
	nameInput.Focus()

	tokenInput := textinput.New()
	tokenInput.Placeholder = "API Token"
	tokenInput.EchoMode = textinput.EchoPassword
	tokenInput.EchoCharacter = '•'

	modelInput := textinput.New()
	modelInput.Placeholder = "模型名称（可选，按回车跳过）"

	return Model{
		step:       stepSelectPreset,
		presets:    []string{"anthropic", "moonshot", "bigmodel", "deepseek", "自定义"},
		selected:   0,
		nameInput:  nameInput,
		tokenInput: tokenInput,
		modelInput: modelInput,
	}
}

// Init 初始化
func (m Model) Init() tea.Cmd {
	return nil
}

// Update 更新状态
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyUp:
			if m.step == stepSelectPreset && m.selected > 0 {
				m.selected--
			}
			return m, nil

		case tea.KeyDown:
			if m.step == stepSelectPreset && m.selected < len(m.presets)-1 {
				m.selected++
			}
			return m, nil

		case tea.KeyEnter:
			return m.handleEnter()

		case tea.KeyBackspace:
			return m.handleBackspace(msg)
		}
	}

	// 处理输入
	switch m.step {
	case stepInputName:
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	case stepInputToken:
		var cmd tea.Cmd
		m.tokenInput, cmd = m.tokenInput.Update(msg)
		return m, cmd
	case stepInputModel:
		var cmd tea.Cmd
		m.modelInput, cmd = m.modelInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepSelectPreset:
		m.presetName = m.presets[m.selected]
		if m.presetName == "自定义" {
			m.isCustom = true
			m.step = stepInputName
			m.nameInput.Focus()
		} else {
			// 使用预设
			preset, err := config.GetPresetByName(m.presetName)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.baseURL = preset.BaseURL
			m.modelInput.SetValue(preset.Model)
			m.nameInput.SetValue(preset.Name)
			m.step = stepInputToken
			m.tokenInput.Focus()
		}

	case stepInputName:
		if m.nameInput.Value() == "" {
			m.err = fmt.Errorf("配置名称不能为空")
			return m, nil
		}
		m.step = stepInputToken
		m.nameInput.Blur()
		m.tokenInput.Focus()

	case stepInputToken:
		if m.tokenInput.Value() == "" {
			m.err = fmt.Errorf("Token 不能为空")
			return m, nil
		}
		m.step = stepInputModel
		m.tokenInput.Blur()
		m.modelInput.Focus()

	case stepInputModel:
		m.saveProfile()
		return m, tea.Quit

	case stepConfirm:
		return m, tea.Quit
	}

	m.err = nil
	return m, nil
}

func (m *Model) handleBackspace(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 在输入步骤按 Backspace 可以返回上一步
	if m.step >= stepInputName && m.step < stepConfirm {
		// 检查输入框是否为空
		switch m.step {
		case stepInputName:
			if m.nameInput.Value() == "" {
				m.step = stepSelectPreset
				m.nameInput.Blur()
				return m, nil
			}
		case stepInputToken:
			if m.tokenInput.Value() == "" {
				if m.isCustom {
					m.step = stepInputName
					m.tokenInput.Blur()
					m.nameInput.Focus()
				} else {
					m.step = stepSelectPreset
					m.tokenInput.Blur()
				}
				return m, nil
			}
		case stepInputModel:
			if m.modelInput.Value() == "" {
				m.step = stepInputToken
				m.modelInput.Blur()
				m.tokenInput.Focus()
				return m, nil
			}
		}
	}
	return m, nil
}

func (m *Model) saveProfile() {
	m.profile = &config.Profile{
		Name:    m.nameInput.Value(),
		BaseURL: m.baseURL,
		Token:   m.tokenInput.Value(),
		Model:   m.modelInput.Value(),
	}

	// 保存到文件
	cfgPath := config.GetConfigPath()
	cfg, _ := config.LoadConfig(cfgPath)
	cfg.AddProfile(*m.profile)

	// 如果是第一个配置，设为默认
	if len(cfg.Profiles) == 1 {
		cfg.Default = m.profile.Name
	}

	cfg.Save(cfgPath)
	m.step = stepDone
}

// View 渲染视图
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🚀 CC-Start 配置向导"))
	b.WriteString("\n\n")

	switch m.step {
	case stepSelectPreset:
		b.WriteString("选择预设:\n\n")
		for i, preset := range m.presets {
			if i == m.selected {
				b.WriteString(selectedStyle.Render("  → " + preset))
			} else {
				b.WriteString(normalStyle.Render("    " + preset))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(normalStyle.Render("↑/↓ 选择，Enter 确认"))

	case stepInputName:
		b.WriteString("输入配置名称:\n\n")
		b.WriteString(fmt.Sprintf("  %s\n\n", m.nameInput.View()))
		b.WriteString(normalStyle.Render("Enter 确认"))

	case stepInputToken:
		b.WriteString(fmt.Sprintf("配置: %s\n", m.nameInput.Value()))
		b.WriteString(fmt.Sprintf("URL: %s\n\n", m.baseURL))
		b.WriteString("输入 API Token:\n\n")
		b.WriteString(fmt.Sprintf("  %s\n\n", m.tokenInput.View()))
		b.WriteString(normalStyle.Render("Enter 确认"))

	case stepInputModel:
		b.WriteString(fmt.Sprintf("配置: %s\n", m.nameInput.Value()))
		b.WriteString(fmt.Sprintf("URL: %s\n\n", m.baseURL))
		b.WriteString("输入模型名称（可选）:\n\n")
		b.WriteString(fmt.Sprintf("  %s\n\n", m.modelInput.View()))
		b.WriteString(normalStyle.Render("Enter 保存，留空使用默认"))
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("❌ %v", m.err)))
	}

	return b.String()
}

// Done 返回是否完成
func (m Model) Done() bool {
	return m.step == stepDone
}

// GetName 返回配置名
func (m Model) GetName() string {
	return m.nameInput.Value()
}
```

**Step 2: 验证编译**

Run: `cd /code/ai/cc-start && go build -o cc-start .`
Expected: 编译成功

**Step 3: 提交**

```bash
git add internal/tui/setup/model.go cmd/setup.go
git commit -m "feat(tui): implement setup wizard with preset selection"
```

---

## Task 10: 集成测试与最终验证

**Files:**
- Modify: `internal/launcher/launcher_test.go`

**Step 1: 添加集成测试**

```go
// 在 internal/launcher/launcher_test.go 添加

func TestBuildSettingsJSON(t *testing.T) {
	profile := &config.Profile{
		Name:    "moonshot",
		BaseURL: "https://api.moonshot.cn/anthropic",
		Token:   "test-token",
		Model:   "moonshot-v1-8k",
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
	if env["ANTHROPIC_BASE_URL"] != "https://api.moonshot.cn/anthropic" {
		t.Errorf("wrong base URL value")
	}
}
```

**Step 2: 运行所有测试**

Run: `cd /code/ai/cc-start && go test ./... -v`
Expected: PASS

**Step 3: 运行静态检查**

Run: `cd /code/ai/cc-start && go vet ./...`
Expected: 无错误

**Step 4: 运行格式化**

Run: `cd /code/ai/cc-start && go fmt ./...`
Expected: 格式化完成

**Step 5: 构建最终二进制**

Run: `cd /code/ai/cc-start && go build -ldflags "-s -w" -o cc-start .`
Expected: 编译成功

**Step 6: 提交**

```bash
git add -A
git commit -m "test: add integration tests and final verification"
```

---

## Task 11: 更新 .gitignore 和文档

**Files:**
- Modify: `.gitignore`
- Create: `README.md`

**Step 1: 更新 .gitignore**

```gitignore
# .gitignore
# 编译产物
cc-start
/cc-start
*.exe
*.exe~
*.dll
*.so
*.dylib

# 测试
*.test
*.out
coverage.txt

# IDE
.idea/
.vscode/
*.swp
*.swo

# 系统文件
.DS_Store
Thumbs.db
```

**Step 2: 创建 README**

```markdown
# CC-Start

Claude Code 启动器 - 快速切换不同 API 供应商。

## 安装

```bash
go install github.com/wujunwei/cc-start@latest
```

## 使用

### 首次配置

```bash
cc-start setup
```

### 启动 Claude Code

```bash
# 使用默认配置
cc-start

# 使用指定配置
cc-start moonshot

# 传递参数给 claude
cc-start -- --dangerously-skip-permissions
```

### 配置管理

```bash
# 列出所有配置
cc-start list

# 设置默认配置
cc-start default moonshot

# 删除配置
cc-start delete moonshot
```

## 支持的供应商

| 供应商 | Base URL | 默认模型 |
|--------|----------|----------|
| Anthropic | https://api.anthropic.com | claude-sonnet-4-5-20250929 |
| Moonshot | https://api.moonshot.cn/anthropic | moonshot-v1-8k |
| BigModel | https://open.bigmodel.cn/api/anthropic | glm-4-plus |
| DeepSeek | https://api.deepseek.com | deepseek-chat |

## 配置文件

配置存储在 `~/.cc-start/profiles.json`
```

**Step 3: 提交**

```bash
git add .gitignore README.md
git commit -m "docs: add README and update gitignore"
```

---

## 完成清单

- [ ] Task 1: 配置数据模型
- [ ] Task 2: 内置预设
- [ ] Task 3: 启动器核心逻辑
- [ ] Task 4: 根命令和入口
- [ ] Task 5: list 命令
- [ ] Task 6: default 命令
- [ ] Task 7: delete 命令
- [ ] Task 8: setup 命令 TUI - 基础框架
- [ ] Task 9: setup 命令 TUI - 更新和视图
- [ ] Task 10: 集成测试与最终验证
- [ ] Task 11: 更新 .gitignore 和文档

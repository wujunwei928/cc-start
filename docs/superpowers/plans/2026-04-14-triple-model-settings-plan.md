# 三模型设置 实施计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Profile 的单 Model 字段替换为 HaikuModel/SonnetModel/OpusModel 三字段，通过 `--settings` 注入 `ANTHROPIC_DEFAULT_HAIKU_MODEL`、`ANTHROPIC_DEFAULT_SONNET_MODEL`、`ANTHROPIC_DEFAULT_OPUS_MODEL` 三个环境变量。

**Architecture:** 自底向上修改——先改数据层（config），再改注入层（launcher），再改向导层（tui/setup），最后改显示层（repl/commands + cmd）。每层改完后跑测试确认不破坏已有功能。

**Tech Stack:** Go 1.24, Bubble Tea, encoding/json

---

## 文件结构

| 文件 | 职责 |
|------|------|
| `internal/config/config.go` | Profile 结构体、迁移逻辑、配置加载/保存 |
| `internal/config/config_test.go` | Profile 序列化、迁移测试 |
| `internal/config/presets.go` | 供应商预设定义 |
| `internal/config/presets_test.go` | 预设测试 |
| `internal/launcher/launcher.go` | BuildSettings 注入、Launch 启动 |
| `internal/launcher/launcher_test.go` | 注入测试 |
| `internal/tui/setup/model.go` | Setup 向导 TUI |
| `internal/tui/setup/model_test.go` | 向导测试 |
| `internal/repl/commands.go` | 旧 REPL 命令（/list、/copy、/import 等） |
| `internal/repl/update.go` | TUI REPL 命令（/use、/show、/current、/import 等） |
| `cmd/list.go` | CLI list 命令 |
| `cmd/claude.go` | CLI claude 命令（-m flag） |
| `cmd/launcher.go` | CLI 启动逻辑 |

---

### Task 1: Profile 结构体变更 + 迁移逻辑

**Files:**
- Modify: `internal/config/config.go:13-18` (Profile 结构体)
- Modify: `internal/config/config.go:130-158` (LoadConfig)
- Test: `internal/config/config_test.go`

- [ ] **Step 1: 写迁移失败测试**

在 `internal/config/config_test.go` 末尾添加：

```go
func TestProfileMigrate(t *testing.T) {
	// 旧格式 JSON 只有一个 model 字段
	oldJSON := `{"name":"test","model":"glm-5","token":"sk-xxx"}`
	var p config.Profile
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
	var p config.Profile
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
	p := config.Profile{Name: "test", Token: "xxx"}
	p.Migrate() // 不应 panic
	if p.SonnetModel != "" {
		t.Errorf("SonnetModel should be empty, got '%s'", p.SonnetModel)
	}
}

func TestConfigMigrateAll(t *testing.T) {
	oldJSON := `{"profiles":[{"name":"a","model":"m1","token":"t1"},{"name":"b","model":"m2","token":"t2"}]}`
	var cfg config.Config
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
	var p config.Profile
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
	p := config.Profile{
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
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd /code/ai/cc-start && go test ./internal/config/ -run TestProfile -v -count=1`
Expected: FAIL（`LegacyModel`、`Migrate`、`MigrateAll`、`HaikuModel`、`SonnetModel`、`OpusModel` 均不存在）

- [ ] **Step 3: 修改 Profile 结构体**

在 `internal/config/config.go` 中，将 Profile 结构体（第 13-18 行）替换为：

```go
// Profile 单个供应商配置
type Profile struct {
	Name             string `json:"name"`
	AnthropicBaseURL string `json:"anthropic_base_url,omitempty"`
	HaikuModel       string `json:"haiku_model,omitempty"`  // 快速模型
	SonnetModel      string `json:"sonnet_model,omitempty"` // 主模型
	OpusModel        string `json:"opus_model,omitempty"`   // 经济模型
	Token            string `json:"token"`
	LegacyModel      string `json:"model,omitempty"`        // 旧字段，仅用于迁移
}
```

- [ ] **Step 4: 添加 Migrate 和 MigrateAll 方法**

在 `Profile.Validate()` 方法后面添加：

```go
// Migrate 将旧 model 字段迁移到 SonnetModel
func (p *Profile) Migrate() {
	if p.LegacyModel != "" && p.SonnetModel == "" {
		p.SonnetModel = p.LegacyModel
	}
	p.LegacyModel = ""
}

// MigrateAll 对所有 Profile 执行迁移
func (c *Config) MigrateAll() {
	for i := range c.Profiles {
		c.Profiles[i].Migrate()
	}
}
```

- [ ] **Step 5: 在 LoadConfig 中调用 MigrateAll**

在 `internal/config/config.go` 的 `LoadConfig` 函数中，`return &cfg, nil` 之前添加：

```go
cfg.MigrateAll()
```

- [ ] **Step 6: 添加 strings import**

`config_test.go` 需要添加 `"strings"` 到 import。在文件顶部 import 块中添加。

- [ ] **Step 7: 运行测试确认通过**

Run: `cd /code/ai/cc-start && go test ./internal/config/ -v -count=1`
Expected: ALL PASS

- [ ] **Step 8: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "refactor(config): Profile 结构体替换为三模型字段，添加迁移逻辑"
```

---

### Task 2: 预设更新

**Files:**
- Modify: `internal/config/presets.go:7-33`
- Modify: `internal/config/presets_test.go:28-92`

- [ ] **Step 1: 更新预设定义**

在 `internal/config/presets.go` 中，将整个 `presets` 变量替换为：

```go
var presets = []Profile{
	{
		Name:             "anthropic",
		AnthropicBaseURL: "https://api.anthropic.com",
		HaikuModel:       "claude-haiku-4-5-20251001",
		SonnetModel:      "claude-sonnet-4-5-20250929",
		OpusModel:        "claude-opus-4-6",
	},
	{
		Name:             "moonshot",
		AnthropicBaseURL: "https://api.kimi.com/coding/",
		HaikuModel:       "kimi-k2.5",
		SonnetModel:      "kimi-k2.5",
		OpusModel:        "kimi-k2.5",
	},
	{
		Name:             "bigmodel",
		AnthropicBaseURL: "https://open.bigmodel.cn/api/anthropic",
		HaikuModel:       "glm-5-turbo",
		SonnetModel:      "glm-5-turbo",
		OpusModel:        "glm-5.1",
	},
	{
		Name:             "deepseek",
		AnthropicBaseURL: "https://api.deepseek.com/anthropic",
		HaikuModel:       "deepseek-chat",
		SonnetModel:      "deepseek-chat",
		OpusModel:        "deepseek-chat",
	},
	{
		Name:             "minimax",
		AnthropicBaseURL: "https://api.minimaxi.com/anthropic",
		HaikuModel:       "MiniMax-M2.7",
		SonnetModel:      "MiniMax-M2.7",
		OpusModel:        "MiniMax-M2.7",
	},
}
```

- [ ] **Step 2: 更新预设测试**

在 `internal/config/presets_test.go` 中，将 `TestGetPresetByName` 的 `tests` 表中每个 `expected` 的 `Model` 字段替换为三个模型字段。例如 anthropic 的 expected 改为：

```go
{
	name: "anthropic",
	expected: &Profile{
		Name:             "anthropic",
		AnthropicBaseURL: "https://api.anthropic.com",
		HaikuModel:       "claude-haiku-4-5-20251001",
		SonnetModel:      "claude-sonnet-4-5-20250929",
		OpusModel:        "claude-opus-4-6",
	},
},
```

其余四个预设类似修改。bigmodel 的值为 `HaikuModel: "glm-5-turbo"`, `SonnetModel: "glm-5-turbo"`, `OpusModel: "glm-5.1"`。deepseek 和 minimax 三个值相同。moonshot 三个值都是 `kimi-k2.5`。

同时将断言从 `p.Model != tt.expected.Model` 改为：

```go
if p.HaikuModel != tt.expected.HaikuModel {
    t.Errorf("expected haikuModel '%s', got '%s'", tt.expected.HaikuModel, p.HaikuModel)
}
if p.SonnetModel != tt.expected.SonnetModel {
    t.Errorf("expected sonnetModel '%s', got '%s'", tt.expected.SonnetModel, p.SonnetModel)
}
if p.OpusModel != tt.expected.OpusModel {
    t.Errorf("expected opusModel '%s', got '%s'", tt.expected.OpusModel, p.OpusModel)
}
```

- [ ] **Step 3: 运行测试**

Run: `cd /code/ai/cc-start && go test ./internal/config/ -v -count=1`
Expected: ALL PASS

- [ ] **Step 4: Commit**

```bash
git add internal/config/presets.go internal/config/presets_test.go
git commit -m "refactor(config): 预设更新为三模型字段 (HaikuModel/SonnetModel/OpusModel)"
```

---

### Task 3: BuildSettings 注入逻辑

**Files:**
- Modify: `internal/launcher/launcher.go:13-27` (BuildSettings)
- Modify: `internal/launcher/launcher.go:40-66` (MergeConfig)
- Modify: `internal/launcher/launcher.go:68-132` (Launch)
- Modify: `internal/launcher/launcher_test.go`

- [ ] **Step 1: 写 BuildSettings 三模型注入测试**

在 `internal/launcher/launcher_test.go` 中添加：

```go
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
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd /code/ai/cc-start && go test ./internal/launcher/ -run TestBuildSettingsTriple -v -count=1`
Expected: FAIL

- [ ] **Step 3: 修改 BuildSettings 函数**

在 `internal/launcher/launcher.go` 的 `BuildSettings` 函数中，在 `ANTHROPIC_BASE_URL` 设置块之后添加三个模型注入：

```go
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
```

- [ ] **Step 4: 修改 MergeConfig 函数**

在 `internal/launcher/launcher.go` 的 `MergeConfig` 函数中，将 `if cfg.Profile.Model != "" { model = cfg.Profile.Model }` 替换为：

```go
if cfg.Profile.SonnetModel != "" {
    model = cfg.Profile.SonnetModel
}
```

- [ ] **Step 5: 修改 Launch 函数**

在 `internal/launcher/launcher.go` 的 `Launch` 函数中：

1. 将 `effectiveProfile` 构建中的 `Model: model` 替换为 `SonnetModel: model`
2. 移除 `--model` 参数相关代码（第 90-93 行的 `if model != ""` 块）
3. 更新启动信息打印（第 118-125 行）替换为：

```go
// 打印启动信息
fmt.Printf("🚀 使用配置启动 Claude Code...\n")
if profile.HaikuModel != "" {
    fmt.Printf("   快速模型 (Haiku): %s\n", profile.HaikuModel)
}
if profile.SonnetModel != "" {
    fmt.Printf("   主模型 (Sonnet):  %s\n", profile.SonnetModel)
}
if profile.OpusModel != "" {
    fmt.Printf("   经济模型 (Opus):   %s\n", profile.OpusModel)
}
if baseURL != "" {
    fmt.Printf("   Base URL:         %s\n", baseURL)
}
fmt.Println()
```

注意：Launch 函数中仍然需要构建 merged `effectiveProfile`，因为 CLI `-t`、`--base-url`、`-m` 参数需要覆盖 Profile 中的值。不能直接传递原始 `profile`，否则 CLI 覆盖失效。将：

```go
effectiveProfile := &config.Profile{
    AnthropicBaseURL: baseURL,
    Model:            model,
    Token:            token,
}
```

改为：

```go
effectiveProfile := &config.Profile{
    AnthropicBaseURL: baseURL,
    HaikuModel:       profile.HaikuModel,   // 继承原始值
    SonnetModel:      model,                  // CLI -m 覆盖 SonnetModel
    OpusModel:        profile.OpusModel,     // 继承原始值
    Token:            token,
}
```

关键：`HaikuModel` 和 `OpusModel` 必须从原始 `profile` 继承，CLI 没有对应的覆盖参数。只有 `SonnetModel` 被 CLI `-m` 覆盖。

- [ ] **Step 6: 修复现有测试**

`TestMergeConfig` 中 `profile.Model` 改为 `profile.SonnetModel`。`TestBuildSettingsJSON` 中 `Model` 改为 `SonnetModel`。

- [ ] **Step 7: 运行所有测试**

Run: `cd /code/ai/cc-start && go test ./... -count=1`
Expected: ALL PASS

- [ ] **Step 8: Commit**

```bash
git add internal/launcher/launcher.go internal/launcher/launcher_test.go
git commit -m "feat(launcher): BuildSettings 注入三模型环境变量，移除 --model 参数"
```

---

### Task 4: Setup 向导三模型步骤

**Files:**
- Modify: `internal/tui/setup/model.go:15-25` (步骤常量)
- Modify: `internal/tui/setup/model.go:46-62` (Model 结构体)
- Modify: `internal/tui/setup/model.go:64-90` (InitialModel)
- Modify: `internal/tui/setup/model.go:92-136` (InitialModelWithProfile)
- Modify: `internal/tui/setup/model.go:197-257` (handleEnter)
- Modify: `internal/tui/setup/model.go:259-288` (handleGoBack)
- Modify: `internal/tui/setup/model.go:290-322` (handleBackspace)
- Modify: `internal/tui/setup/model.go:324-350` (saveProfile)
- Modify: `internal/tui/setup/model.go:352-419` (View)
- Modify: `internal/tui/setup/model_test.go`

- [ ] **Step 1: 添加步骤常量**

在 `internal/tui/setup/model.go` 的步骤常量块中，将 `stepInputModel` 替换为三个步骤：

```go
const (
	stepSelectPreset step = iota
	stepInputName
	stepInputAnthropicURL
	stepInputToken
	stepInputHaikuModel
	stepInputSonnetModel
	stepInputOpusModel
	stepConfirm
	stepDone
)
```

- [ ] **Step 2: 修改 Model 结构体**

将 `modelInput textinput.Model` 替换为三个输入框：

```go
type Model struct {
	step          step
	presets       []string
	selected      int
	nameInput     textinput.Model
	tokenInput    textinput.Model
	haikuInput    textinput.Model
	sonnetInput   textinput.Model
	opusInput     textinput.Model
	anthropicURLInput textinput.Model
	isCustom      bool
	presetName    string
	err           error
	profile       *config.Profile
	isEdit        bool
	originalName  string
}
```

- [ ] **Step 3: 修改 InitialModel**

在 `InitialModel` 函数中，将 `modelInput` 替换为三个输入框：

```go
haikuInput := textinput.New()
haikuInput.Placeholder = "快速模型 (Haiku)"

sonnetInput := textinput.New()
sonnetInput.Placeholder = "主模型 (Sonnet)"

opusInput := textinput.New()
opusInput.Placeholder = "经济模型 (Opus)"
```

返回值中用 `haikuInput`、`sonnetInput`、`opusInput` 替换 `modelInput`。

- [ ] **Step 4: 修改 InitialModelWithProfile**

将 `modelInput` 替换为三个输入框，预填充当前 Profile 的值：

```go
haikuInput := textinput.New()
haikuInput.Placeholder = "快速模型 (Haiku)"
haikuInput.SetValue(p.HaikuModel)

sonnetInput := textinput.New()
sonnetInput.Placeholder = "主模型 (Sonnet)"
sonnetInput.SetValue(p.SonnetModel)

opusInput := textinput.New()
opusInput.Placeholder = "经济模型 (Opus)"
opusInput.SetValue(p.OpusModel)
```

返回值中用 `haikuInput`、`sonnetInput`、`opusInput` 替换 `modelInput`。

- [ ] **Step 5: 修改 handleEnter — 选择预设时填充三个模型**

在 `stepSelectPreset` case 中，将 `m.modelInput.SetValue(preset.Model)` 替换为：

```go
m.haikuInput.SetValue(preset.HaikuModel)
m.sonnetInput.SetValue(preset.SonnetModel)
m.opusInput.SetValue(preset.OpusModel)
```

- [ ] **Step 6: 修改 handleEnter — Token 后进入 Haiku 步骤**

将 `stepInputModel` 的 case 替换为三个步骤：

```go
case stepInputToken:
    if m.tokenInput.Value() == "" {
        m.err = fmt.Errorf("Token 不能为空")
        return m, nil
    }
    m.step = stepInputHaikuModel
    m.tokenInput.Blur()
    m.haikuInput.Focus()

case stepInputHaikuModel:
    m.step = stepInputSonnetModel
    m.haikuInput.Blur()
    m.sonnetInput.Focus()

case stepInputSonnetModel:
    m.step = stepInputOpusModel
    m.sonnetInput.Blur()
    m.opusInput.Focus()

case stepInputOpusModel:
    m.saveProfile()
    return m, tea.Quit
```

- [ ] **Step 7: 修改 handleGoBack**

替换 `stepInputModel` case 为三个步骤：

```go
case stepInputOpusModel:
    m.step = stepInputSonnetModel
    m.opusInput.Blur()
    m.sonnetInput.Focus()
case stepInputSonnetModel:
    m.step = stepInputHaikuModel
    m.sonnetInput.Blur()
    m.haikuInput.Focus()
case stepInputHaikuModel:
    m.step = stepInputToken
    m.haikuInput.Blur()
    m.tokenInput.Focus()
```

同时更新 `stepInputToken` case 中的返回目标：原来是 `stepInputModel`，改为 `stepInputHaikuModel`。

- [ ] **Step 8: 修改 handleBackspace**

添加 `stepInputHaikuModel`、`stepInputSonnetModel`、`stepInputOpusModel` 的 Backspace 处理（与原 `stepInputModel` 逻辑相同，使用对应的 input 字段）。

- [ ] **Step 9: 修改 Update 中的输入处理**

将 `case stepInputModel` 替换为三个步骤：

```go
case stepInputHaikuModel:
    var cmd tea.Cmd
    m.haikuInput, cmd = m.haikuInput.Update(msg)
    return m, cmd
case stepInputSonnetModel:
    var cmd tea.Cmd
    m.sonnetInput, cmd = m.sonnetInput.Update(msg)
    return m, cmd
case stepInputOpusModel:
    var cmd tea.Cmd
    m.opusInput, cmd = m.opusInput.Update(msg)
    return m, cmd
```

- [ ] **Step 10: 修改 saveProfile**

替换 `saveProfile` 中的 Profile 构建：

```go
func (m *Model) saveProfile() {
    haiku := m.haikuInput.Value()
    sonnet := m.sonnetInput.Value()
    opus := m.opusInput.Value()

    // 创建模式：留空使用预设默认值
    if !m.isEdit {
        if haiku == "" && m.presetName != "" {
            if preset, err := config.GetPresetByName(m.presetName); err == nil {
                haiku = preset.HaikuModel
            }
        }
        if sonnet == "" && m.presetName != "" {
            if preset, err := config.GetPresetByName(m.presetName); err == nil {
                sonnet = preset.SonnetModel
            }
        }
        if opus == "" && m.presetName != "" {
            if preset, err := config.GetPresetByName(m.presetName); err == nil {
                opus = preset.OpusModel
            }
        }
    }
    // 编辑模式：直接使用输入值（包括空字符串，表示清空）

    m.profile = &config.Profile{
        Name:             m.nameInput.Value(),
        AnthropicBaseURL: m.anthropicURLInput.Value(),
        HaikuModel:       haiku,
        SonnetModel:      sonnet,
        OpusModel:        opus,
        Token:            m.tokenInput.Value(),
    }

    // ... 后续保存逻辑不变
}
```

- [ ] **Step 11: 修改 View**

将 `case stepInputModel` 替换为三个视图：

```go
case stepInputHaikuModel:
    fmt.Fprintf(&b, "配置: %s\n", m.nameInput.Value())
    if m.anthropicURLInput.Value() != "" {
        fmt.Fprintf(&b, "Base URL: %s\n", m.anthropicURLInput.Value())
    }
    b.WriteString("\n输入快速模型 (Haiku):\n\n")
    fmt.Fprintf(&b, "  %s\n\n", m.haikuInput.View())
    b.WriteString(normalStyle.Render("Enter 继续，ESC 返回"))

case stepInputSonnetModel:
    fmt.Fprintf(&b, "配置: %s\n", m.nameInput.Value())
    if m.anthropicURLInput.Value() != "" {
        fmt.Fprintf(&b, "Base URL: %s\n", m.anthropicURLInput.Value())
    }
    b.WriteString("\n输入主模型 (Sonnet):\n\n")
    fmt.Fprintf(&b, "  %s\n\n", m.sonnetInput.View())
    b.WriteString(normalStyle.Render("Enter 继续，ESC 返回"))

case stepInputOpusModel:
    fmt.Fprintf(&b, "配置: %s\n", m.nameInput.Value())
    if m.anthropicURLInput.Value() != "" {
        fmt.Fprintf(&b, "Base URL: %s\n", m.anthropicURLInput.Value())
    }
    b.WriteString("\n输入经济模型 (Opus):\n\n")
    fmt.Fprintf(&b, "  %s\n\n", m.opusInput.View())
    b.WriteString(normalStyle.Render("Enter 保存，ESC 返回"))
```

- [ ] **Step 12: 修复现有测试**

在 `internal/tui/setup/model_test.go` 中，所有引用 `m.modelInput` 的地方替换为 `m.sonnetInput`（因为 sonnet 是中间步骤，流程测试需要经过它）。具体需要修改的测试：

- `TestModelInputBackspace` → `TestSonnetInputBackspace`，将 `stepInputModel` 改为 `stepInputSonnetModel`，`m.modelInput` 改为 `m.sonnetInput`
- `TestModelInputCanType` → `TestSonnetInputCanType`，类似修改
- `TestEscGoBackFromModel` → `TestEscGoBackFromSonnet`，验证 ESC 从 Sonnet 返回到 Haiku
- `TestBackspaceNotGoBack` → 修改目标步骤为 `stepInputOpusModel`，使用 `m.opusInput`
- `TestBackspaceNotDeletePreviousStep` → 修改为从 Sonnet 步骤测试

- [ ] **Step 13: 运行测试**

Run: `cd /code/ai/cc-start && go test ./internal/tui/setup/ -v -count=1`
Expected: ALL PASS

- [ ] **Step 14: Commit**

```bash
git add internal/tui/setup/model.go internal/tui/setup/model_test.go
git commit -m "feat(setup): 向导支持三模型输入步骤 (Haiku/Sonnet/Opus)"
```

---

### Task 5: REPL 和 CLI 显示适配

**Files:**
- Modify: `internal/repl/commands.go` (cmdList、cmdUse、cmdCurrent、cmdShow、cmdCopy、cmdImport)
- Modify: `internal/repl/commands_test.go` (测试中的 `Model:` 字段)
- Modify: `internal/repl/update.go` (cmdUse、cmdShow、cmdCopy、cmdImport、formatProfileList、formatCurrentProfile)
- Modify: `cmd/list.go`
- Modify: `cmd/launcher.go`

- [ ] **Step 1: 修改旧 REPL commands_test.go**

在 `internal/repl/commands_test.go` 第 23-24 行，将 `Model: "model1"` 和 `Model: "model2"` 替换为 `SonnetModel: "model1"` 和 `SonnetModel: "model2"`。

- [ ] **Step 2: 修改旧 REPL commands.go**

在 `internal/repl/commands.go` 中：

**cmdList**（第 28 行）：将表格 header 从 `{"名称", "Base URL", "模型", "Token", "状态"}` 改为 `{"名称", "主模型", "快速模型", "经济模型", "Token", "状态"}`，将 `p.Model` 改为 `p.SonnetModel`，添加 `p.HaikuModel`、`p.OpusModel` 列。

**cmdUse**（第 80 行）：将 `profile.Model` 改为 `profile.SonnetModel`。

**cmdCurrent**（第 102 行）：将 `profile.Model` 替换为三个模型的显示。

**cmdShow**（第 161 行）：将 `profile.Model` 替换为三个模型的显示。

**cmdCopy**（第 294 行）：将 `Model: src.Model` 替换为 `HaikuModel: src.HaikuModel, SonnetModel: src.SonnetModel, OpusModel: src.OpusModel`。

**cmdImport**（第 462 行后）：在 `json.Unmarshal` 之后添加 `importCfg.MigrateAll()`。

- [ ] **Step 3: 修改 TUI REPL update.go**

在 `internal/repl/update.go` 中：

**cmdUse**（第 395 行）：将 `profile.Model` 改为 `profile.SonnetModel`。

**cmdShow**（第 443 行）：将 `profile.Model` 替换为三个模型的显示。

**cmdCopy**（第 512 行）：将 `Model: src.Model` 替换为三模型字段。

**cmdImport**（第 664 行后）：在 `json.Unmarshal` 之后添加 `importCfg.MigrateAll()`。

**formatProfileList**（第 873 行）：将 `p.Model` 替换为三个模型的显示。

**formatCurrentProfile**（第 896 行）：将 `profile.Model` 替换为三个模型的显示。

- [ ] **Step 4: 修改 cmd/list.go**

在 `cmd/list.go` 第 46 行，将 `p.Model` 替换为三模型显示。

- [ ] **Step 5: 运行所有测试**

Run: `cd /code/ai/cc-start && go test ./... -count=1`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add internal/repl/commands.go internal/repl/update.go cmd/list.go
git commit -m "feat(repl): 所有显示命令适配三模型，/import 添加迁移调用"
```

---

### Task 6: 修复旧 REPL 的 Setup/Edit 入口

**Files:**
- Modify: `internal/repl/commands.go:195` (cmdEdit — InitialModelWithProfile)
- Modify: `internal/repl/commands.go:291` (cmdCopy)

- [ ] **Step 1: 确认 cmdEdit 无需修改**

`cmdEdit` 调用 `setup.InitialModelWithProfile(*profile)`，传入的 profile 已经是新的三模型结构体，`InitialModelWithProfile` 在 Task 4 中已修改为读取 `HaikuModel`/`SonnetModel`/`OpusModel`。无需额外修改。

- [ ] **Step 2: 运行编译检查**

Run: `cd /code/ai/cc-start && go build ./...`
Expected: 编译成功，无错误

- [ ] **Step 3: Commit（如有变更）**

如果编译通过且无需额外修改，跳过此 commit。

---

### Task 7: 全量测试 + 编译验证

- [ ] **Step 1: 运行全量测试**

Run: `cd /code/ai/cc-start && go test ./... -count=1 -v`
Expected: ALL PASS

- [ ] **Step 2: 编译验证**

Run: `cd /code/ai/cc-start && go build -o cc-start .`
Expected: 编译成功

- [ ] **Step 3: 最终 commit（如有遗漏修复）**

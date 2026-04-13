# 三模型设置设计文档

**日期**: 2026-04-14
**状态**: 已确认

## 背景

cc-start 当前每个 Profile 只绑定一个 `Model` 字段，启动时通过 `--model` 参数传递给 Claude Code。但 Claude Code 内部有三个模型层级（Haiku/Sonnet/Opus），用户使用第三方供应商时需要将这三个层级分别映射到供应商的不同模型。

通过 `--settings` 的 `env` 字段注入 `ANTHROPIC_DEFAULT_HAIKU_MODEL`、`ANTHROPIC_DEFAULT_SONNET_MODEL`、`ANTHROPIC_DEFAULT_OPUS_MODEL` 三个环境变量，实现模型层级映射。

## 目标

1. 每个 Profile 支持配置三个层级的模型：Haiku（快速）、Sonnet（主）、Opus（经济）
2. 通过 Setup 向导编辑三个模型
3. 启动时通过 `--settings` 注入三个模型环境变量
4. 旧配置自动迁移（`model` → `sonnet_model`）

## 数据结构变更

### Profile 结构体

将原有的 `Model string` 字段替换为三个独立字段，同时保留一个导出的临时 `LegacyModel` 字段用于旧配置迁移：

```go
type Profile struct {
    Name             string `json:"name"`
    AnthropicBaseURL string `json:"anthropic_base_url,omitempty"`
    HaikuModel       string `json:"haiku_model,omitempty"`       // 快速模型
    SonnetModel      string `json:"sonnet_model,omitempty"`      // 主模型
    OpusModel        string `json:"opus_model,omitempty"`        // 经济模型
    Token            string `json:"token"`
    LegacyModel      string `json:"model,omitempty"`             // 旧字段，仅用于迁移
}
```

**注意**：`LegacyModel` 必须是导出字段（大写开头），因为 `encoding/json` 不会给未导出字段赋值。`LegacyModel` 不应在任何业务逻辑中读取或写入，仅在 `Migrate()` 中使用。迁移完成后该字段清空，序列化时因 `omitempty` 不会写入 JSON。

### 预设更新

每个供应商预设的三个模型默认值：

| 供应商 | HaikuModel | SonnetModel | OpusModel |
|--------|-----------|-------------|-----------|
| anthropic | claude-haiku-4-5-20251001 | claude-sonnet-4-5-20250929 | claude-opus-4-6 |
| moonshot | kimi-k2.5 | kimi-k2.5 | kimi-k2.5 |
| bigmodel | glm-5-turbo | glm-5-turbo | glm-5.1 |
| deepseek | deepseek-chat | deepseek-chat | deepseek-chat |
| minimax | MiniMax-M2.7 | MiniMax-M2.7 | MiniMax-M2.7 |

### 旧配置迁移

通过 `LegacyModel` 字段（`json:"model"` tag）接收旧 JSON 中的 `model` 值。`Migrate()` 方法将 `LegacyModel` 迁移到 `SonnetModel`：

```go
func (p *Profile) Migrate() {
    if p.LegacyModel != "" && p.SonnetModel == "" {
        p.SonnetModel = p.LegacyModel
    }
    p.LegacyModel = "" // 迁移后清空，下次序列化不再写入
}
```

**迁移触发点**（所有加载配置的路径都必须调用）：

1. `LoadConfig()` — 正常启动加载配置文件后，遍历 `cfg.Profiles` 调用 `Migrate()`
2. `cmdImport()` — `/import` 命令导入外部配置文件后，对导入的每个 Profile 调用 `Migrate()`（`internal/repl/commands.go:462`）
3. 或者：将迁移逻辑封装为 `Config.MigrateAll()` 方法，在 `LoadConfig` 和 `cmdImport` 两处统一调用

## BuildSettings 注入逻辑

在现有的 `env` map 中增加三个模型环境变量：

```go
func BuildSettings(profile *config.Profile) map[string]interface{} {
    env := map[string]string{
        "ANTHROPIC_AUTH_TOKEN": profile.Token,
    }

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

    return map[string]interface{}{"env": env}
}
```

### Launch 行为变更

- 移除 `--model` 参数（模型已通过 `--settings` 的 env 注入）
- CLI `-m` 参数仍可用于临时覆盖 `SonnetModel`（在 `BuildSettings` 中用 CLI 指定的模型覆盖 `SonnetModel`）
- 启动信息显示三个模型：

```
🚀 使用配置启动 Claude Code...
   主模型 (Sonnet):  glm-5-turbo
   快速模型 (Haiku):  glm-5-turbo
   经济模型 (Opus):   glm-5.1
   Base URL:         https://open.bigmodel.cn/api/anthropic
```

## Setup 向导变更

### 流程变更

```
选择预设 → 输入名称 → 输入 Base URL → 输入 Token → 输入快速模型 (Haiku) → 输入主模型 (Sonnet) → 输入经济模型 (Opus) → 完成
```

### 设计要点

- 预设自动填充：选择预设后，三个模型输入框自动填充预设默认值
- 可单独修改：用户可以只改其中一个或两个，留空的保持预设默认值
- 显示层级名称：每个输入框前标注用途（快速模型 Haiku / 主模型 Sonnet / 经济模型 Opus）
- 原来的 `stepInputModel` 拆为三个独立步骤：`stepInputHaikuModel`、`stepInputSonnetModel`、`stepInputOpusModel`

### 空输入合并规则

向导存在两种模式（通过 `PendingSetup.IsEdit` 区分），合并规则不同：

**创建模式**（`IsEdit == false`）— 从预设生成新 Profile：
- 留空 → 使用预设默认值
- 有输入 → 使用用户输入值

```go
// 创建模式：留空回退到预设默认值
HaikuModel  = mergeWithPreset(input, preset.HaikuModel)
SonnetModel = mergeWithPreset(input, preset.SonnetModel)
OpusModel   = mergeWithPreset(input, preset.OpusModel)
```

**编辑模式**（`IsEdit == true`）— 修改已有 Profile：
- 留空 → **保留当前已保存值**（不回退到预设）
- 有输入 → 使用用户输入值

```go
// 编辑模式：留空保留已有值
HaikuModel  = mergeWithExisting(input, existing.HaikuModel)
SonnetModel = mergeWithExisting(input, existing.SonnetModel)
OpusModel   = mergeWithExisting(input, existing.OpusModel)
```

```go
func mergeWithPreset(input, presetDefault string) string {
    if input == "" { return presetDefault }
    return input
}

func mergeWithExisting(input, currentValue string) string {
    if input == "" { return currentValue }
    return input
}
```

向导的模型输入框在编辑模式下应预填充当前 Profile 的已保存值，而非预设默认值。

## 显示适配

### /list 命令

表格增加快速模型和经济模型列：

```
  名称            主模型             快速模型            经济模型            Base URL
──────────────────────────────────────────────────────────────────────────────────────
  bigmodel        glm-5-turbo        glm-5-turbo        glm-5.1             open.bigmodel.cn
* anthropic       claude-sonnet...   claude-haiku...    claude-opus-4-6     api.anthropic.com
```

### /show 和 /current 命令

```
📦 当前配置: bigmodel
   主模型 (Sonnet):  glm-5-turbo
   快速模型 (Haiku):  glm-5-turbo
   经济模型 (Opus):   glm-5.1
   Base URL:         https://open.bigmodel.cn/api/anthropic
```

## 受影响文件

| 文件 | 变更类型 |
|------|----------|
| `internal/config/config.go` | Profile 结构体替换（含 legacyModel）、添加 Migrate()/MigrateAll() 方法 |
| `internal/config/presets.go` | 预设增加 HaikuModel/SonnetModel/OpusModel |
| `internal/launcher/launcher.go` | BuildSettings 注入三个模型、Launch 显示更新、移除 --model 参数 |
| `internal/tui/setup/model.go` | 向导增加三个模型输入步骤、空输入合并逻辑 |
| `internal/repl/update.go` | /use、/show、/current 显示三模型（:395、:443）、formatProfileList/formatCurrentProfile（:843、:881） |
| `internal/repl/commands.go` | /copy 复制三模型字段（:294、:512）、/import 调用迁移（:462） |
| `cmd/list.go` | 列表显示三个模型（:46） |
| `cmd/claude.go` | CLI -m 参数适配（:25，覆盖 SonnetModel） |
| `internal/config/config_test.go` | Profile 序列化、迁移测试 |
| `internal/config/presets_test.go` | 预设默认值测试 |
| `internal/launcher/launcher_test.go` | BuildSettings 注入测试 |
| `internal/tui/setup/model_test.go` | 向导流程测试 |

## 测试要点

1. Profile 序列化/反序列化：新格式正确读写
2. 旧配置迁移：`model` 字段通过 `legacyModel` 正确迁移到 `SonnetModel`
3. 迁移后序列化：迁移完成后 `model` 字段不再出现在 JSON 中
4. /import 迁移：导入旧格式配置文件后正确触发迁移
5. BuildSettings：三个模型环境变量正确注入到 env map
6. BuildSettings 空值：模型为空时不注入对应环境变量
7. 预设默认值：五个预设的三个模型字段正确填充
8. Setup 向导空输入（创建）：留空时使用预设默认值，不持久化空字符串
9. Setup 向导空输入（编辑）：留空时保留当前已保存值，不回退到预设
10. Setup 向导：三个模型步骤正确显示和编辑
11. 向后兼容：只有 `model` 字段的旧 JSON 文件能正确加载

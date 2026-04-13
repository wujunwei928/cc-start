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

将原有的 `Model string` 字段替换为三个独立字段：

```go
type Profile struct {
    Name             string `json:"name"`
    AnthropicBaseURL string `json:"anthropic_base_url,omitempty"`
    HaikuModel       string `json:"haiku_model,omitempty"`   // 快速模型
    SonnetModel      string `json:"sonnet_model,omitempty"`  // 主模型
    OpusModel        string `json:"opus_model,omitempty"`    // 经济模型
    Token            string `json:"token"`
}
```

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

加载配置时，如果 Profile 存在旧的 `model` 字段（JSON 反序列化时 `Model` 不为空而三个新字段为空）：
- `model` 值 → 迁移到 `SonnetModel`
- `HaikuModel` / `OpusModel` 留空
- 下次保存时自动写入新格式，旧 `model` 字段不再写入

实现方式：在 `Profile` 上添加一个 `Migrate()` 方法，在 `LoadConfig` 后调用。

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
| `internal/config/config.go` | Profile 结构体替换、添加 Migrate() 方法 |
| `internal/config/presets.go` | 预设增加 HaikuModel/SonnetModel/OpusModel |
| `internal/launcher/launcher.go` | BuildSettings 注入三个模型、Launch 显示更新、移除 --model 参数 |
| `internal/tui/setup/model.go` | 向导增加三个模型输入步骤 |
| `internal/repl/update.go` | /list、/show、/current 显示三个模型 |
| `internal/repl/view.go` | 显示渲染适配 |
| `cmd/launcher.go` | CLI -m 参数适配（覆盖 SonnetModel） |
| `internal/config/config_test.go` | Profile 序列化、迁移测试 |
| `internal/config/presets_test.go` | 预设默认值测试 |
| `internal/launcher/launcher_test.go` | BuildSettings 注入测试 |
| `internal/tui/setup/model_test.go` | 向导流程测试 |

## 测试要点

1. Profile 序列化/反序列化：新格式正确读写
2. 旧配置迁移：`model` 字段正确迁移到 `SonnetModel`
3. BuildSettings：三个模型环境变量正确注入到 env map
4. BuildSettings 空值：模型为空时不注入对应环境变量
5. 预设默认值：五个预设的三个模型字段正确填充
6. Setup 向导：三个模型步骤正确显示和编辑
7. 向后兼容：只有 `model` 字段的旧 JSON 文件能正确加载

# 移除 Codex 和 OpenCode 支持，只保留 Claude

## 背景

cc-start 最初设计为支持三种 AI CLI 工具（claude、codex、opencode），通过工具抽象层实现多工具切换。在 commit `f8adf23` 中已移除 codex/opencode 的子命令注册，但残留了大量死代码和引用。本项目决定彻底移除 codex/opencode 支持，只保留 claude。

## 目标

- 移除所有 codex/opencode 相关代码和配置
- 简化架构，移除不必要的抽象层
- 保持现有 claude 功能完全正常

## 设计决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| OpenAIBaseURL 字段 | 移除 | 只支持 claude，不需要 OpenAI 格式 URL |
| tools 抽象层 | 完全移除 | 单工具场景下抽象层是过度设计 |
| LaunchConfig.Tool 和 LaunchWithTool | 移除 | 工具轴不再有意义，简化为直接的 claude 启动 API |
| 历史设计文档 | 删除 | 保持文档与代码一致 |
| TUI Setup OpenAI URL 输入 | 移除 | 单格式不需要区分 URL 类型 |

## 变更清单

### 删除的文件

- `cmd/codex.go` — 死代码，codex 子命令已不注册
- `cmd/opencode.go` — 死代码，opencode 子命令已不注册
- `internal/tools/tools.go` — 工具抽象层不再需要
- `internal/tools/tools_test.go` — 配套测试
- `docs/plans/2026-03-10-launch-command-design.md` — 历史设计文档
- `docs/plans/2026-03-10-launch-command-impl.md` — 历史设计文档
- `docs/plans/2026-03-11-dual-baseurl-design.md` — 双 URL 设计文档

### 修改的文件

#### `internal/config/config.go`

移除 Profile 结构体中的 `OpenAIBaseURL` 字段。

```go
// Before
type Profile struct {
    Name             string `json:"name"`
    AnthropicBaseURL string `json:"anthropic_base_url,omitempty"`
    OpenAIBaseURL    string `json:"openai_base_url,omitempty"`    // 删除
    Model            string `json:"model,omitempty"`
    Token            string `json:"token"`
}

// After
type Profile struct {
    Name             string `json:"name"`
    AnthropicBaseURL string `json:"anthropic_base_url,omitempty"`
    Model            string `json:"model,omitempty"`
    Token            string `json:"token"`
}
```

旧配置兼容：Go JSON 反序列化自动忽略未知字段，已有 `settings.json` 中的 `openai_base_url` 字段会被静默忽略，无需迁移逻辑。

#### `internal/config/presets.go`

移除所有预设中的 `OpenAIBaseURL` 赋值。

#### `cmd/claude.go`

- 移除 `internal/tools` 包导入
- 移除 `tools.GetTool(toolName)` 调用和验证逻辑
- 直接调用 `runLaunch`，不再需要工具名称参数

#### `cmd/launcher.go`

- 移除 `internal/tools` 包导入
- 移除 `toolName` 参数和 `Tool` 字段设置
- 移除 `launcher.LaunchWithTool` 调用，改为调用简化后的 `launcher.Launch`
- `LaunchConfig` 中移除 `Tool` 字段

#### `internal/launcher/launcher.go`

- 移除 `tools` 包依赖
- 移除 `LaunchConfig.Tool` 字段（第 72 行）
- 移除 `MergeConfig` 函数中的 `toolFormat` 参数和 OpenAI 格式分支（第 85-98 行）
- 移除整个 `LaunchWithTool` 函数（第 118-208 行），该函数通过 `tools.GetTool` 做工具查找并设置通用环境变量，不再需要
- 简化为：`MergeConfig` 只接受 `*LaunchConfig`，直接使用 `profile.AnthropicBaseURL`
- **迁移现有 `Launch` 函数**：当前 `func Launch(profile *config.Profile, extraArgs []string) error`（第 54 行）是一个简单的 claude 启动函数。移除 `LaunchWithTool` 后，将其逻辑合并到现有 `Launch` 中，使 `Launch` 同时支持 `LaunchConfig` 参数（包含 profile、model override、yolo 等完整配置），并删除旧的 `(profile, extraArgs)` 签名版本

```go
// Before
func Launch(profile *config.Profile, extraArgs []string) error   // 第 54 行，现有简单版
func MergeConfig(cfg *LaunchConfig, toolFormat string) (model, baseURL, token string)
func LaunchWithTool(cfg *LaunchConfig) error                     // 第 118 行，通用工具版

// After
func MergeConfig(cfg *LaunchConfig) (model, baseURL, token string)
func Launch(cfg *LaunchConfig) error                             // 合并两者，统一入口
```

#### `internal/launcher/launcher_test.go`

- 移除 `internal/tools` 包导入（第 11 行）
- `TestMergeConfig`（第 204-279 行）：移除 `tools.FormatAnthropic`/`tools.FormatOpenAI` 引用，移除 `OpenAIBaseURL` 测试数据，移除 openai 格式测试用例，更新 `MergeConfig` 调用签名（不再传递 `toolFormat`）
- 移除 "openai format selects openai url" 和 "partial override - model only (openai)" 测试用例
- 更新 "no profile no override" 用例，移除 `Tool` 字段

#### `cmd/list.go`

- 移除 `p.OpenAIBaseURL` 的展示逻辑（第 46-48 行）

#### `internal/repl/commands.go`

- `cmdList`（第 36 行）：移除表格 "OpenAI URL" 列头和对应数据填充（第 61 行）
- `cmdCurrent`（第 111 行）：移除 OpenAI URL 展示
- `cmdShow`（第 171 行）：移除 OpenAI URL 展示

#### `internal/repl/update.go`

- `cmdShow`（第 443-444 行）：移除 OpenAI URL 展示
- `cmdCopy`（第 515 行）：移除 `OpenAIBaseURL` 字段的复制
- `cmdTest`（第 598-599 行）：移除回退到 OpenAI URL 的逻辑，只使用 `AnthropicBaseURL`
- `formatProfileList`（第 880-881 行）：移除 OpenAI URL 展示
- `formatCurrentProfile`（第 906-907 行）：移除 OpenAI URL 展示

#### `internal/tui/setup/model.go`

- 移除 `stepInputOpenAIURL` 步骤定义（第 21 行）
- 移除 OpenAI URL 输入相关的视图渲染和更新逻辑
- 修改提示文本，移除 "用于 Codex/OpenCode CLI" 引用（第 442 行）
- 调整步骤流程，从 `stepInputAnthropicURL` 直接跳到 `stepInputToken`

#### `README.md`

- 移除 `cc-start codex` 和 `cc-start opencode` 使用示例

## 边界处理

1. **旧配置兼容**：已有 `settings.json` 包含 `openai_base_url` → Go JSON 反序列化自动忽略
2. **JSON 序列化**：移除字段后输出不再包含 `openai_base_url`，符合预期
3. **测试**：删除 `tools_test.go`，重写 `launcher_test.go` 中的 `TestMergeConfig`（移除 tools 依赖和 OpenAI 格式用例），更新其他受影响的测试文件
4. **TUI 步骤编号**：移除 OpenAI URL 步骤后，后续步骤编号需要调整
5. **Launcher API 简化**：移除 `LaunchWithTool` → 统一使用 `Launch`，`cmd/launcher.go` 不再传递工具名称

## 风险评估

- **低风险**：所有变更都是确定性删除，不涉及新逻辑
- **测试验证**：变更后运行完整测试套件确保 claude 功能正常

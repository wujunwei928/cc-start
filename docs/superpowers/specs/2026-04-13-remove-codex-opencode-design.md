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
- 直接调用 `runLaunchWithTool`，不再需要工具名称参数

#### `cmd/launcher.go`

- 移除 `internal/tools` 包导入
- 移除 `tools.GetTool()` 调用
- 直接调用简化的 `launcher.MergeConfig(profile)`，不再传递 `tools.Tool`

#### `internal/launcher/launcher.go`

- 移除 `tools` 包依赖
- `MergeConfig` 不再接受 `tools.Tool` 参数，直接使用 `profile.AnthropicBaseURL`
- 移除 OpenAI 格式分支（`tool.URLFormat` 判断）
- Claude 的可执行文件名和环境变量映射直接硬编码

```go
// Before
func MergeConfig(profile config.Profile, tool tools.Tool) LaunchConfig

// After
func MergeConfig(profile config.Profile) LaunchConfig
```

#### `cmd/list.go`

- 移除 `p.OpenAIBaseURL` 的展示逻辑（第 46-48 行）

#### `internal/repl/commands.go`

- `cmdList`（第 36 行）：移除表格 "OpenAI URL" 列头和对应数据填充（第 61 行）
- `cmdCurrent`（第 111 行）：移除 OpenAI URL 展示
- `cmdShow`（第 171 行）：移除 OpenAI URL 展示
- `cmdCopy`（第 305 行）：移除 `OpenAIBaseURL` 字段的复制
- `cmdTest`（第 394-397 行）：移除回退到 OpenAI URL 的逻辑，只使用 `AnthropicBaseURL`

#### `internal/repl/update.go`

- TUI `/show` 命令（第 443-444 行）：移除 OpenAI URL 展示

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
3. **测试**：删除 `tools_test.go`，更新其他受影响的测试文件
4. **TUI 步骤编号**：移除 OpenAI URL 步骤后，后续步骤编号需要调整

## 风险评估

- **低风险**：所有变更都是确定性删除，不涉及新逻辑
- **测试验证**：变更后运行完整测试套件确保 claude 功能正常

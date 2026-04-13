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
    AnthropicBaseURL string `json:"anthropicBaseUrl"`
    OpenAIBaseURL    string `json:"openaiBaseUrl"`    // 删除
}

// After
type Profile struct {
    Name             string `json:"name"`
    AnthropicBaseURL string `json:"anthropicBaseUrl"`
}
```

旧配置兼容：Go JSON 反序列化自动忽略未知字段，无需迁移逻辑。

#### `internal/config/presets.go`

移除所有预设中的 `OpenAIBaseURL` 赋值。

#### `internal/launcher/launcher.go`

- 移除 `tools` 包依赖
- `MergeConfig` 不再接受 `tools.Tool` 参数，直接使用 `AnthropicBaseURL`
- 移除 OpenAI 格式分支
- Claude 的可执行文件名和环境变量映射直接硬编码

```go
// Before
func MergeConfig(profile config.Profile, tool tools.Tool) LaunchConfig

// After
func MergeConfig(profile config.Profile) LaunchConfig
```

#### `internal/tui/setup/model.go`

- 移除 OpenAI URL 输入字段
- 修改提示文本，移除 "用于 Codex/OpenCode CLI" 引用

#### `README.md`

- 移除 `cc-start codex` 和 `cc-start opencode` 使用示例

#### `cmd/launcher.go`

- 移除 `tools.GetTool()` 调用
- 直接调用简化的 `launcher.MergeConfig()`

## 边界处理

1. **旧配置兼容**：已有 `settings.json` 包含 `openaiBaseUrl` → 自动忽略
2. **JSON 序列化**：移除字段后输出不再包含 `openaiBaseUrl`，符合预期
3. **测试**：删除 `tools_test.go`，更新其他受影响的测试文件

## 风险评估

- **低风险**：所有变更都是确定性删除，不涉及新逻辑
- **测试验证**：变更后运行完整测试套件确保 claude 功能正常

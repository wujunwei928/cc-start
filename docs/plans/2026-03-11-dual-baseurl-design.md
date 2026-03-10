# 双 Base URL 配置设计

## 概述

支持配置两个独立的 base_url：
- **Anthropic 格式** → 用于 claude
- **OpenAI 格式** → 用于 codex 和 opencode

对应 base_url 为空时，不允许启动对应的 CLI 工具。

## 数据结构变更

### Profile 结构体

```go
type Profile struct {
    Name             string `json:"name"`
    AnthropicBaseURL string `json:"anthropic_base_url,omitempty"`
    OpenAIBaseURL    string `json:"openai_base_url,omitempty"`
    Model            string `json:"model,omitempty"`
    Token            string `json:"token"`
}
```

### Preset 结构体

```go
type Preset struct {
    Name             string
    AnthropicBaseURL string
    OpenAIBaseURL    string
    Model            string
}
```

### 向后兼容

加载旧配置时，如果 `base_url` 存在但新字段为空，迁移到 `anthropic_base_url`。

## 预设数据

| 预设名称 | Anthropic URL | OpenAI URL |
|---------|---------------|------------|
| anthropic | https://api.anthropic.com | （空） |
| moonshot | https://api.moonshot.cn/v1 | （空） |
| bigmodel | https://open.bigmodel.cn/api/paas/v4 | （空） |
| deepseek | https://api.deepseek.com | （空） |
| minimax | https://api.minimax.chat/v1 | （空） |

## Setup TUI 流程

### 预设模式
1. 选择预设 → 自动填充 URL
2. 输入 Token
3. 输入 Model（可选）
4. 保存

### 自定义模式
1. 输入名称
2. 同时显示两个 URL 输入框（均可留空）
3. 输入 Token
4. 输入 Model（可选）
5. 保存

## Launcher 校验逻辑

```go
// 根据工具类型选择 URL
if tool.URLFormat == tools.FormatAnthropic {
    baseURL = cfg.Profile.AnthropicBaseURL
} else {
    baseURL = cfg.Profile.OpenAIBaseURL
}

// 校验
if baseURL == "" {
    return fmt.Errorf("未配置 %s 格式的 base_url，无法启动 %s",
        tool.URLFormat, tool.Name)
}
```

## 需要修改的文件

| 文件 | 变更内容 |
|------|---------|
| `internal/config/config.go` | Profile 结构体变更 |
| `internal/config/presets.go` | Preset 结构体和预设数据更新 |
| `internal/tools/tools.go` | 新增 URLFormat 字段 |
| `internal/launcher/launcher.go` | URL 选择和校验逻辑 |
| `internal/tui/setup/model.go` | 自定义模式双 URL 输入 |

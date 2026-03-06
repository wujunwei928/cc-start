# CC-Start 设计文档

## 概述

CC-Start 是一个用 Golang 开发的 Claude Code 启动器，用于快速选择不同供应商启动 Claude Code。

## 技术选型

- **语言**: Go 1.21+
- **CLI 框架**: cobra
- **TUI 框架**: bubbletea + bubbles
- **配置存储**: JSON 文件 (`~/.cc-start/profiles.json`)

## 项目结构

```
cc-start/
├── cmd/                    # CLI 命令
│   ├── root.go            # 根命令 & 启动逻辑
│   ├── setup.go           # 创建/更新配置 (TUI)
│   ├── list.go            # 列出配置
│   ├── default.go         # 设置默认
│   └── delete.go          # 删除配置
├── internal/
│   ├── config/            # 配置管理
│   │   ├── config.go      # 配置结构 & 文件操作
│   │   └── presets.go     # 内置预设
│   └── launcher/          # 启动器
│       └── launcher.go    # Claude 启动逻辑
├── main.go                 # 入口
├── go.mod
└── go.sum
```

## 数据模型

### 配置文件结构 (`~/.cc-start/profiles.json`)

```json
{
  "profiles": [
    {
      "name": "anthropic",
      "base_url": "https://api.anthropic.com",
      "model": "claude-sonnet-4-5-20250929",
      "token": "sk-ant-xxx..."
    }
  ],
  "default": "anthropic"
}
```

### Profile 结构

```go
type Profile struct {
    Name    string `json:"name"`
    BaseURL  string `json:"base_url"`
    Model    string `json:"model,omitempty"`
    Token    string `json:"token"`
}

type Config struct {
    Profiles []Profile `json:"profiles"`
    Default  string    `json:"default,omitempty"`
}
```

## 内置预设

| 预设名 | Base URL | 默认模型 |
|--------|----------|----------|
| anthropic | `https://api.anthropic.com` | `claude-sonnet-4-5-20250929` |
| moonshot | `https://api.moonshot.cn/anthropic` | `moonshot-v1-8k` |
| bigmodel | `https://open.bigmodel.cn/api/anthropic` | `glm-4-plus` |
| deepseek | `https://api.deepseek.com` | `deepseek-chat` |

## CLI 命令设计

### 启动命令

```bash
# 使用默认配置启动
cc-start

# 使用指定配置启动
cc-start moonshot

# 传递参数给 Claude
cc-start moonshot -- --dangerously-skip-permissions
```

### 配置管理

```bash
# TUI 交互式配置
cc-start setup

# 列出所有配置
cc-start list

# 设置默认配置
cc-start default <name>

# 删除配置
cc-start delete <name>
```

## 启动逻辑

1. 解析命令行参数
2. 确定使用的配置（指定名 > 默认配置 > 报错）
3. 读取配置文件，获取 Token
4. 构建 settings JSON：
   ```json
   {
     "env": {
       "ANTHROPIC_AUTH_TOKEN": "<token>",
       "ANTHROPIC_BASE_URL": "<url>" // 仅非官方 API 时
     }
   }
   ```
5. 执行命令：`claude --settings '<json>' --model <model> [其他参数...]`

## 依赖

- github.com/spf13/cobra - CLI 框架
- github.com/charmbracelet/bubbletea - TUI 框架
- github.com/charmbracelet/bubbles - TUI 组件
- github.com/charmbracelet/lipgloss - 样式

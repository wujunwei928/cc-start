# Launch 命令设计文档

## 概述

将现有的 `run` 命令重构为 `launch` 命令，支持多种 AI CLI 工具（claude、codex、opencode），提供灵活的参数注入机制。

## 需求总结

| 项目 | 决定 |
|------|------|
| 核心功能 | 通用 AI CLI 启动器 |
| 支持工具 | claude、codex、opencode（内置预设） |
| 命令格式 | `cc-start launch <tool> [profile] [flags] [-- tool-args]` |
| 参数注入 | 混合模式：环境变量 + 结构化参数 + 配置覆盖 |
| 优先级 | profile 与命令行合并，命令行覆盖冲突项 |
| 旧命令处理 | 删除 `run`，无向后兼容 |
| REPL 处理 | 删除 `run`，不添加 `launch` |

## 命令结构

### 命令格式

```
cc-start launch <tool> [profile] [flags] [-- tool-args]
```

- `<tool>` — 必填，指定 AI 工具：`claude` | `codex` | `opencode`
- `[profile]` — 可选，使用预定义的配置名
- `[flags]` — 可选，注入参数
- `[-- tool-args]` — 可选，`--` 之后的参数原样传递给目标工具

### 参数列表

| 长标记 | 短标记 | 说明 |
|--------|--------|------|
| `--model` | `-m` | 指定模型 |
| `--base-url` | `-b` | API 基础地址 |
| `--token` | `-t` | 认证令牌 |
| `--env` | `-e` | 环境变量（格式: KEY=VALUE，可多次使用） |
| `--help` | `-h` | 显示帮助 |

### 使用示例

```bash
# 使用 profile 启动 claude
cc-start launch claude moonshot

# 使用 profile + 覆盖模型
cc-start launch claude moonshot -m claude-opus-4

# 无 profile，纯命令行参数
cc-start launch codex -b https://api.openai.com -t sk-xxx -m gpt-4

# 传递额外参数给工具
cc-start launch claude moonshot -- --dangerously-skip-permissions

# 添加环境变量
cc-start launch claude moonshot -e DEBUG=true -e LOG_LEVEL=info
```

## 工具配置预设

### 工具预设结构

```go
type Tool struct {
    Name       string            // 工具名
    Executable string            // 可执行文件名
    EnvMap     map[string]string // 参数到环境变量的映射
}
```

### 三工具预设

| 工具 | 可执行文件 | 环境变量映射 |
|------|-----------|-------------|
| **claude** | `claude` | `token` → `ANTHROPIC_AUTH_TOKEN`<br>`base-url` → `ANTHROPIC_BASE_URL` |
| **codex** | `codex` | `token` → `OPENAI_API_KEY`<br>`base-url` → `OPENAI_BASE_URL` |
| **opencode** | `opencode` | `token` → `OPENAI_API_KEY`<br>`base-url` → `OPENAI_BASE_URL` |

## 参数合并逻辑

### 合并规则

```
最终配置 = 工具默认值 + Profile 配置 + 命令行参数
```

- **基础字段**（`base_url`, `token`, `model`）：命令行覆盖 profile 覆盖 默认值
- **环境变量**（`-e`）：追加模式，不覆盖已有映射
- **工具原生平**（`--` 之后的参数）：直接追加到命令末尾

### 合并示例

```bash
cc-start launch claude moonshot -m claude-opus-4 -e DEBUG=true
```

```
工具默认值: { base_url: "", token: "", model: "" }
      ↓ 合并
Profile moonshot: { base_url: "https://api.kimi.com", token: "kimi-xxx", model: "moonshot-v1" }
      ↓ 合并
命令行参数: { model: "claude-opus-4", env: { DEBUG: "true" } }
      ↓ 最终配置
结果: { base_url: "https://api.kimi.com", token: "kimi-xxx", model: "claude-opus-4", env: { DEBUG: "true" } }
```

## 启动流程

```
1. 解析命令行参数
   ↓
2. 加载 Profile（如果指定）
   ↓
3. 合并配置（默认值 ← Profile ← 命令行）
   ↓
4. 构建环境变量（根据工具预设映射）
   ↓
5. 构建命令行参数
   ↓
6. 执行目标工具
```

## 文件结构变更

### 新增文件

```
internal/
├── tools/
│   ├── tools.go          # 工具注册和预设定义
│   └── tools_test.go     # 单元测试
```

### 修改文件

| 文件 | 变更 |
|------|------|
| `cmd/run.go` | **删除** |
| `cmd/launch.go` | **新增**，实现 launch 命令 |
| `internal/launcher/launcher.go` | 重构，支持多工具和参数合并 |
| `internal/repl/commands.go` | 删除 `run` 命令处理 |
| `internal/repl/messages.go` | 删除 `run` 相关消息 |
| `README.md` | 更新文档 |

## 核心接口设计

### internal/tools/tools.go

```go
type Tool struct {
    Name       string
    Executable string
    EnvMap     map[string]string // config key -> env name
}

var BuiltInTools = map[string]Tool{
    "claude":   { ... },
    "codex":    { ... },
    "opencode": { ... },
}

func GetTool(name string) (*Tool, error)
```

### internal/launcher/launcher.go

```go
type LaunchConfig struct {
    Tool     string
    Profile  string
    Model    string
    BaseURL  string
    Token    string
    Env      map[string]string
    ToolArgs []string
}

func Launch(cfg *LaunchConfig) error
```

## 帮助信息

```
使用指定的 AI 工具和配置启动编程助手。

用法:
  cc-start launch <tool> [profile] [flags] [-- tool-args]

工具:
  claude    Anthropic Claude Code CLI
  codex     OpenAI Codex CLI
  opencode  OpenCode AI 编程助手

参数:
  -m, --model string       模型名称
  -b, --base-url string    API 基础地址
  -t, --token string       认证令牌
  -e, --env stringArray    环境变量 (格式: KEY=VALUE)
  -h, --help               显示帮助

示例:
  cc-start launch claude                      # 使用默认配置启动 claude
  cc-start launch claude moonshot             # 使用 moonshot 配置
  cc-start launch codex -m gpt-4 -t sk-xxx    # 指定模型和令牌
  cc-start launch claude moonshot -e DEBUG=true -- --help
```

## 错误处理

| 场景 | 处理方式 |
|------|----------|
| 工具不存在 | 提示可用工具列表 |
| Profile 不存在 | 提示运行 `cc-start list` 查看 |
| 可执行文件未安装 | 提示安装方法 |
| Token 缺失 | 提示通过 `-t` 或 profile 提供 |

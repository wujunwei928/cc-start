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

### 交互式 REPL 模式

无参数运行进入 REPL：

```bash
cc-start
```

REPL 中可用命令：

| 命令 | 说明 |
|------|------|
| `list`, `ls` | 列出所有配置 |
| `use <name>` | 切换当前配置 |
| `current` | 显示当前配置 |
| `default <name>` | 设置默认配置 |
| `show <name>` | 显示配置详情 |
| `delete <name>` | 删除配置 |
| `copy <from> <to>` | 复制配置 |
| `rename <old> <new>` | 重命名配置 |
| `test [name]` | 测试 API 连通性 |
| `export [file]` | 导出配置 |
| `import <file>` | 导入配置 |
| `run [profile]` | 启动 Claude Code |
| `help` | 显示帮助 |
| `exit` | 退出 REPL |

### 直接启动 Claude Code

```bash
# 使用默认配置启动
cc-start run

# 使用指定配置启动
cc-start run moonshot

# 传递参数给 claude
cc-start run -- --dangerously-skip-permissions
```

### 命令行配置管理

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
| Moonshot | https://api.kimi.com/coding/ | moonshot-v1-8k |
| BigModel | https://open.bigmodel.cn/api/anthropic | glm-5 |
| DeepSeek | https://api.deepseek.com/anthropic | deepseek-chat |

## 配置文件

配置存储在 `~/.cc-start/profiles.json`

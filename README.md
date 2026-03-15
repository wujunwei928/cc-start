# CC-Start

AI 编程助手启动器 - 快速切换不同供应商。

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
| `help` | 显示帮助 |
| `exit` | 退出 REPL |

### 启动 AI 编程助手

```bash
# 启动 Claude Code CLI
cc-start claude

# 使用指定配置启动
cc-start claude moonshot

# 启动 OpenAI Codex CLI（指定模型）
cc-start codex -m gpt-4

# 传递参数给工具
cc-start claude moonshot -- --dangerously-skip-permissions

# 启动 OpenCode
cc-start opencode deepseek
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

## 系统设置

### 语言设置
支持中文、英文、日文三种语言，可通过 Ctrl+P 打开设置面板进行切换。

### 主题设置
提供 5 个预设主题：
- **默认 / Default** - 深色主题，高对比度，适合长时间使用
- **海洋 / Ocean** - 蓝绿色调，清爽明亮
- **森林 / Forest** - 绿色调，自然舒适
- **日落 / Sunset** - 暖色调，温馨浪漫
- **亮色 / Light** - 浅色主题，适合白天使用

### 使用方法
1. 在 REPL 中按 `Ctrl+P` 打开设置面板
2. 使用 ↑↓ 键选择设置项
3. 按 Enter 进入设置项
4. 选择新的值并按 Enter 确认
5. 按 Esc 关闭设置面板

设置会自动保存到 `~/.cc-start/profiles.json`，下次启动时自动应用。

### 示例

切换到英文界面：
```
1. 按 Ctrl+P
2. 选择 "语言 / Language"
3. 选择 "English"
4. 按 Enter 确认
```

切换到海洋主题：
```
1. 按 Ctrl+P
2. 选择 "主题 / Theme"
3. 选择 "海洋 / Ocean"
4. 按 Enter 确认
```

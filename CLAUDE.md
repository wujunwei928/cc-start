# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

CC-Start 是一个 Claude Code 启动器，用于管理多个 API 供应商配置并通过 `claude --settings` 注入环境变量来启动 Claude Code CLI。支持 REPL 交互模式、TUI 配置向导、主题系统和国际化（zh/en/ja）。

## 构建与测试

```bash
# 构建
go build -o cc-start .

# 运行所有测试
go test ./...

# 运行单个包的测试
go test ./internal/config/...
go test ./internal/repl/...

# 运行单个测试函数
go test ./internal/config -run TestProfileValidation

# 覆盖率 & 详细输出
go test -cover ./...
go test -v ./...

# 代码检查
go vet ./...
go fmt ./...
```

项目使用 pre-commit hooks（`.pre-commit-config.yaml`），包含 `go-unit-tests`、`go-fmt`、`go-vet`。

## 架构

### 启动流程

无参数运行 → `cmd/root.go` → `repl.New()` → `REPL.Run()` 循环 → 创建 Bubble Tea `Model` → `tea.Program.Run()`

`REPL.Run()` 维护一个 for 循环：TUI 退出后检查 `PendingSetup` 或 `PendingLaunch`，执行对应操作后决定继续循环还是退出。这使得 setup/edit 向导可以在独立 TUI 中运行后返回主 REPL。

### 双层 REPL

- **`REPL` 结构体**（`repl.go`）：外部循环管理器，持有配置和历史，协调 setup/launch 生命周期
- **`Model` 结构体**（`model.go`）：Bubble Tea Elm 架构模型（`Init/Update/View`），处理所有 TUI 交互
- 两个层级各自有命令实现：`REPL` 的命令在 `commands.go`（直接 fmt 输出），`Model` 的命令在 `update.go`（返回 string 写入 OutputBuffer）

### 三层模型映射

`Profile` 有三个模型字段：`HaikuModel`（快速）、`SonnetModel`（主模型）、`OpusModel`（经济模型）。旧 `model` 字段通过 `Migrate()` 自动迁移到 `SonnetModel`。

Launcher 中 `MergeConfig` 的优先级：命令行 `-m` > Profile > 默认值。CLI `-m` 只覆盖 `SonnetModel`，`HaikuModel` 和 `OpusModel` 始终继承原始 Profile 值。

### 配置系统

配置文件 `~/.cc-start/settings.json`，包含 `profiles[]`、`default`、`settings{language, theme}`。内置预设定义在 `internal/config/presets.go`。文件权限 `0600`，Token 显示必须遮蔽。

### 启动管道

`cmd/launcher.go` 解析参数 → `config.LoadConfig` → `cfg.GetProfile` → `launcher.LaunchConfig` → `launcher.MergeConfig` 合并覆盖 → `BuildSettings` 构建 `env` 映射 → `exec.Command("claude", "--settings", json)` 启动。环境变量注入 `ANTHROPIC_AUTH_TOKEN`、`ANTHROPIC_BASE_URL`、`ANTHROPIC_DEFAULT_HAIKU/SONNET/OPUS_MODEL`。

### Cobra CLI

`cmd/` 下每个文件对应一个子命令：`root`（REPL）、`claude`（启动）、`setup`（配置向导）、`list`、`default`、`delete`。通过 `init()` 注册到 `rootCmd`。

### 主题 & 国际化

- `internal/theme/presets.go`：5 个预设主题
- `internal/i18n/`：`messages.go` 定义常量 → `zh.go`/`en.go`/`ja.go` 各语言翻译
- 主题切换通过 `NewStylesFromTheme()` 重建样式，语言切换更新 `i18n.Manager` 并刷新所有子组件

## 代码风格

- 中文注释，导出类型/函数必须以名称开头注释
- 导入顺序：标准库 → 第三方库 → 本项目包，组间空行
- 方法接收者：单字母（`func (c *Config)`、`func (m Model)`）
- 错误处理：用 `fmt.Errorf` 包装，消息不首字母大写、不以标点结尾
- 测试：表驱动测试，`t.TempDir()` 处理临时目录
- 显式初始化切片：`cfg.Profiles = []Profile{}`

## 常见扩展点

- **添加 CLI 命令**：`cmd/` 创建文件 → `cobra.Command` → `init()` 注册到 `rootCmd`
- **添加 REPL 命令**：`commands.go` 添加 `cmdXxx` → `ExecuteCommand` switch 注册 → `cmdHelp` 更新
- **添加主题**：`internal/theme/presets.go` 的 `presets` 切片，确保对比度足够
- **添加 i18n 消息**：`messages.go` 常量 → `zh.go`/`en.go`/`ja.go` 翻译
- **添加供应商预设**：`internal/config/presets.go` 的 `presets` 切片

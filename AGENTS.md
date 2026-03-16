# AGENTS.md - CC-Start 项目指南

CC-Start 是一个 Claude Code 启动器，用于快速切换不同的 API 供应商。

## 构建和测试命令

```bash
# 构建
go build -o cc-start .
go install github.com/wujunwei/cc-start@latest

# 运行所有测试
go test ./...

# 运行单个包的测试
go test ./internal/config/...
go test ./internal/repl/...

# 运行单个测试函数
go test ./internal/config -run TestProfileValidation
go test ./internal/repl -run TestThemeSwitch

# 覆盖率和详细输出
go test -cover ./...
go test -v ./...

# 代码检查
go vet ./...
go fmt ./...
```

## 项目结构

```
cc-start/
├── main.go              # 程序入口
├── cmd/                 # CLI 命令 (Cobra)
└── internal/
    ├── config/          # 配置管理 (settings.json)
    ├── launcher/        # Claude Code 启动逻辑
    ├── repl/            # 交互式 REPL (Bubble Tea)
    ├── tui/setup/       # 配置向导 TUI
    ├── theme/           # 主题系统
    └── i18n/            # 国际化 (zh/en/ja)
```

## 代码风格规范

### 文件头部和导入

每个 Go 文件以包路径注释开头。导入顺序：标准库 → 第三方库 → 本项目内部包，组间空行分隔。

```go
// internal/config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/wujunwei/cc-start/internal/config"
)
```

### 命名约定

- **包名**: 小写单词 (`config`, `launcher`, `repl`, `i18n`)
- **导出类型/函数**: PascalCase，必须有注释
- **私有字段/函数**: camelCase
- **方法接收者**: 单字母 (`func (c *Config)`, `func (r *REPL)`)

### 注释规范

使用中文注释。导出的类型、函数必须有注释，注释以名称开头。

### 错误处理

永远不要忽略错误。使用 `fmt.Errorf` 包装错误添加上下文。错误消息不首字母大写，不以标点结尾。

```go
if err != nil {
	return fmt.Errorf("加载配置失败: %w", err)
}
```

### 零值处理

显式初始化切片：`cfg.Profiles = []Profile{}`

## 测试规范

使用表驱动测试。测试文件命名 `<filename>_test.go`，使用 `t.TempDir()` 处理临时目录。

```go
func TestProfileValidation(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		wantErr bool
	}{
		{name: "valid", profile: Profile{Name: "anthropic", Token: "sk-xxx"}, wantErr: false},
		{name: "missing name", profile: Profile{Token: "xxx"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

## 关键依赖

| 依赖 | 用途 |
|------|------|
| `github.com/spf13/cobra` | CLI 框架 |
| `github.com/charmbracelet/bubbletea` | TUI 框架 |
| `github.com/charmbracelet/lipgloss` | 终端样式 |
| `github.com/sahilm/fuzzy` | 模糊搜索 |

## 架构模式

### Bubble Tea TUI

REPL 和设置面板使用 Bubble Tea 架构（Elm 架构）：

```go
type Model struct { /* 状态 */ }
func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m Model) View() string
```

### 主题系统

主题定义在 `internal/theme/presets.go`，样式通过 `NewStylesFromTheme(t *theme.Theme) Styles` 创建。

### 国际化

支持 zh/en/ja 三种语言：

```go
i18nMgr := i18n.NewManager()
i18nMgr.SetLanguage("en")
msg := i18nMgr.T(i18n.MsgSettingsTheme)
```

## 配置文件

配置存储在 `~/.cc-start/settings.json`：

```json
{
  "profiles": [{"name": "anthropic", "base_url": "https://api.anthropic.com", "model": "claude-sonnet-4-5-20250929", "token": "sk-ant-xxx"}],
  "default": "anthropic",
  "settings": {"language": "zh", "theme": "default"}
}
```

## 常见任务

- **添加 CLI 命令**: 在 `cmd/` 创建文件 → 定义 `cobra.Command` → `init()` 注册到 `rootCmd`
- **添加 REPL 命令**: 在 `internal/repl/commands.go` 添加 `cmdXxx` → `ExecuteCommand` switch 注册 → 更新 `cmdHelp`
- **添加主题**: 在 `internal/theme/presets.go` 的 `presets` 切片添加，确保 `PaletteInactive` 背景色与前景色对比度足够
- **添加国际化消息**: 在 `internal/i18n/messages.go` 定义常量 → 在 `zh.go`/`en.go`/`ja.go` 添加翻译

## 安全注意事项

- Token 显示时必须遮蔽（使用 `maskAPIKey` 函数）
- 配置文件权限应为 `0600`
- 不要在日志中输出敏感信息

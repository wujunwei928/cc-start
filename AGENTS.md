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

# 运行单个测试函数
go test ./internal/config -run TestProfileValidation

# 运行带子测试
go test ./cmd -run TestFindDashSeparator

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
├── internal/
│   ├── config/          # 配置管理
│   ├── launcher/        # Claude Code 启动逻辑
│   ├── repl/            # 交互式 REPL (Bubble Tea)
│   └── tui/             # 终端 UI 组件
└── go.mod
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
	tea "github.com/charmbracelet/bubbletea"

	"github.com/wujunwei/cc-start/internal/config"
)
```

### 命名约定

- **包名**: 小写单词 (`config`, `launcher`, `repl`)
- **导出类型/函数**: PascalCase，必须有注释
- **私有字段/函数**: camelCase
- **方法接收者**: 单字母 (`func (c *Config)`, `func (r *REPL)`)

### 注释规范

使用中文注释。导出的类型、函数必须有注释，注释以名称开头。

```go
// Profile 单个供应商配置
type Profile struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model,omitempty"`
	Token   string `json:"token"`
}

// GetProfile 获取指定配置，name 为空时返回默认配置
func (c *Config) GetProfile(name string) (*Profile, error)
```

### 错误处理

永远不要忽略错误。使用 `fmt.Errorf` 包装错误添加上下文。错误消息不首字母大写，不以标点结尾。

```go
if err != nil {
	return fmt.Errorf("加载配置失败: %w", err)
}
return nil, fmt.Errorf("profile '%s' not found", name)
```

### 零值处理

显式初始化切片：`cfg.Profiles = []Profile{}`

## 测试规范

使用表驱动测试。测试文件命名 `<filename>_test.go`，测试函数 `Test<FunctionName>`。使用 `t.TempDir()` 或 `defer os.RemoveAll()` 处理临时目录。

```go
func TestProfileValidation(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		wantErr bool
	}{
		{name: "valid profile", profile: Profile{Name: "anthropic", Token: "sk-ant-xxx"}, wantErr: false},
		{name: "missing name", profile: Profile{Token: "xxx"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Profile.Validate() error = %v, wantErr %v", err, tt.wantErr)
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
| `github.com/olekukonko/tablewriter` | 表格输出 |

## 配置文件

配置存储在 `~/.cc-start/profiles.json`：

```json
{
  "profiles": [{"name": "anthropic", "base_url": "https://api.anthropic.com", "model": "claude-sonnet-4-5-20250929", "token": "sk-ant-xxx"}],
  "default": "anthropic"
}
```

## 常见任务

- **添加 CLI 命令**: 在 `cmd/` 创建文件 → 定义 `cobra.Command` → `init()` 注册到 `rootCmd` → 添加测试
- **添加 REPL 命令**: 在 `internal/repl/commands.go` 添加处理函数 → `ExecuteCommand` switch 注册 → 更新 `cmdHelp`
- **添加 API 预设**: 在 `internal/config/presets.go` 的 `presets` 切片添加

## 安全注意事项

- Token 显示时必须遮蔽（使用 `maskAPIKey`）
- 配置文件权限应为 `0600`
- 不要在日志中输出敏感信息

# REPL TUI 重构设计

> 基于 Bubble Tea 框架重写 REPL，实现全屏交互式 UI

## 背景

当前 REPL 使用 `go-prompt` 库，补全列表样式不够美观。参考 crush 项目的实现，使用 Bubble Tea 框架完全重写 REPL UI。

## 目标

- 全屏交互式 REPL UI
- 参考 crush 的视觉风格
- 命令面板支持模糊搜索
- 统一的样式系统

## 技术选型

| 组件 | 选择 | 说明 |
|------|------|------|
| 框架 | Bubble Tea v0.x | 公开稳定版 |
| 样式 | Lipgloss v0.x | 与 Bubble Tea 配套 |
| 组件 | bubbles v0.x | textinput、help 等现成组件 |
| 模糊搜索 | sahilm/fuzzy | 与 crush 一致 |

## 架构设计

### 整体架构

采用 Bubble Tea 的 **Model-View-Update (MVU)** 架构：

```
┌─────────────────────────────────────────────┐
│                  Main Model                  │
│  ┌───────────────────────────────────────┐  │
│  │            Input Component            │  │
│  │  (textinput + 前缀显示)                │  │
│  └───────────────────────────────────────┘  │
│  ┌───────────────────────────────────────┐  │
│  │          Output Component             │  │
│  │  (命令输出、表格、帮助信息)             │  │
│  └───────────────────────────────────────┘  │
│  ┌───────────────────────────────────────┐  │
│  │        Command Palette (弹窗)          │  │
│  │  (输入过滤 + 列表选择)                  │  │
│  └───────────────────────────────────────┘  │
│  ┌───────────────────────────────────────┐  │
│  │          Help Bar (底部)               │  │
│  │  (快捷键提示)                          │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

### 文件结构

```
internal/repl/
├── repl.go           # 主 Model 和入口
├── model.go          # Model 定义和状态管理
├── update.go         # Update 函数（消息处理）
├── view.go           # View 渲染
├── commands.go       # 命令定义和执行逻辑（保留现有）
├── history.go        # 历史记录（保留现有）
├── styles.go         # Lipgloss 样式定义
├── components/
│   ├── input.go      # 输入框组件
│   ├── output.go     # 输出区域组件（滚动缓冲区）
│   ├── palette.go    # 命令面板弹窗
│   └── helpbar.go    # 底部帮助栏
└── messages.go       # Bubble Tea 消息定义
```

### Model 状态定义

```go
type Model struct {
    // 配置
    config     *config.Config
    configPath string

    // 当前状态
    currentProfile string

    // 组件
    input    textinput.Model
    output   *OutputBuffer      // 输出缓冲区
    palette  *CommandPalette    // 命令面板（可为 nil）
    helpBar  help.Model

    // 历史记录
    history  *History
    histIdx  int               // 历史导航索引
}
```

### 消息流

```
用户输入 → KeyMsg → Update() → 返回 Cmd
                    ↓
              判断当前焦点
                    ↓
         ┌─────────┼─────────┐
         ↓         ↓         ↓
      主界面    命令面板    输入框
         │         │         │
         └─────────┴─────────┘
                   ↓
              更新 Model
                   ↓
              View() 渲染
```

### 关键消息类型

```go
type (
    // 命令执行结果
    CommandExecutedMsg struct { Output string; Err error }

    // 命令面板选择
    CommandSelectedMsg struct { Cmd string; Args []string }

    // 配置变更
    ProfileChangedMsg struct { Name string }
)
```

## 组件设计

### 命令面板

交互流程：

```
1. 按 Ctrl+P 或输入 / 触发面板
   ┌────────────────────────────────────────┐
   │ Commands                          [×]  │
   │ ┌──────────────────────────────────┐   │
   │ │ 🔍 Type to filter...             │   │
   │ └──────────────────────────────────┘   │
   │ ┌──────────────────────────────────┐   │
   │ │ ● /list      列出所有配置        │   │
   │ │ ○ /use       切换当前会话配置    │   │
   │ │ ○ /current   显示当前配置        │   │
   │ │ ○ /default   设置默认配置        │   │
   │ │ ...                              │   │
   │ └──────────────────────────────────┘   │
   │ tab 切换  ↑↓ 导航  enter 确认  esc 关闭│
   └────────────────────────────────────────┘

2. 输入过滤 + 上下选择 + 回车执行
```

### 命令分组

| 分组 | 命令 |
|------|------|
| **配置管理** | list, use, current, default, show, add, edit, delete, copy, rename |
| **测试/导入导出** | test, export, import |
| **辅助** | history, help, clear |
| **启动** | run, setup, exit |

## 样式设计

### 配色方案（参考 crush 深色主题）

```go
var (
    // 背景色
    bgColor      = lipgloss.Color("#1a1a2e")
    surfaceColor = lipgloss.Color("#16213e")

    // 文字色
    textColor   = lipgloss.Color("#e0e0e0")
    mutedColor  = lipgloss.Color("#6c7086")
    accentColor = lipgloss.Color("#89b4fa")  // 蓝色高亮

    // 状态色
    successColor = lipgloss.Color("#a6e3a1")  // 绿色
    errorColor   = lipgloss.Color("#f38ba8")  // 红色
    warningColor = lipgloss.Color("#fab387")  // 橙色
    infoColor    = lipgloss.Color("#89dceb")  // 青色
)
```

### 主界面布局

```
┌─────────────────────────────────────────────┐  ← 圆角边框
│ cc-start [moonshot]                         │  ← 前缀 + 输入区
├─────────────────────────────────────────────┤
│                                             │
│  ✓ 已切换到配置 'moonshot'                   │  ← 输出区
│  ● 模型: claude-3-sonnet                     │
│                                             │
├─────────────────────────────────────────────┤
│ ctrl+p 命令  ↑↓ 历史  enter 执行  ctrl+c 退出 │  ← 帮助栏
└─────────────────────────────────────────────┘
```

### 命令面板样式

```
┌─────────────────────────────────────────────┐
│ Commands                               [×]  │  ← 标题栏
│ ┌───────────────────────────────────────┐   │
│ │ > lis                                 │   │  ← 输入框
│ └───────────────────────────────────────┘   │
│ ┌───────────────────────────────────────┐   │
│ │ ● /list      列出所有配置              │   │  ← 选中项
│ │ ○ /ls        列出所有配置    →         │   │  ← 别名提示
│ └───────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

## 实施计划

### 阶段划分

```
Phase 1: 基础框架
├── 创建 Model 结构
├── 实现基本 View 渲染
├── 集成 textinput 组件
└── 命令执行逻辑迁移

Phase 2: 输出区域
├── 实现滚动缓冲区
├── 表格渲染美化
├── 彩色输出样式
└── 历史导航 (↑↓)

Phase 3: 命令面板
├── Palette 弹窗组件
├── 模糊搜索集成
├── 分组/别名支持
└── 快捷键绑定

Phase 4: 完善细节
├── Help 底部栏
├── 样式微调
├── 边界情况处理
└── 单元测试
```

### 依赖变更

新增依赖：

```go
require (
    github.com/charmbracelet/bubbletea v0.27.0
    github.com/charmbracelet/lipgloss v0.13.0
    github.com/charmbracelet/bubbles v0.20.0
    github.com/sahilm/fuzzy v0.1.0
)
```

移除依赖：

```go
github.com/c-bata/go-prompt
```

## 参考

- [crush 项目](/code/ai/crush) - 架构和风格参考
- [Bubble Tea 文档](https://github.com/charmbracelet/bubbletea)
- [Lipgloss 文档](https://github.com/charmbracelet/lipgloss)
- [bubbles 组件库](https://github.com/charmbracelet/bubbles)

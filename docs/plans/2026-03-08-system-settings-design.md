# 系统设置功能设计文档

**日期**: 2026-03-08
**作者**: AI Assistant
**状态**: 已批准

## 概述

本文档描述了 CC-Start 系统设置功能的完整设计方案，包括多语言支持和主题系统。移除了编辑器设置功能。

## 功能需求

### 核心功能
1. **多语言支持**
   - 支持中文、英文、日文等多种语言
   - 全界面翻译（标题、提示、命令说明等）
   - 内置翻译，无需外部文件
   - 运行时切换，立即生效

2. **预设主题**
   - 提供 5 个精心设计的颜色主题
   - 每个主题有独特的视觉风格
   - 实时预览和切换
   - 立即应用到整个界面

3. **配置持久化**
   - 合并到现有 `~/.cc-start/settings.json`
   - 添加 `settings` 字段存储用户偏好
   - 自动迁移旧配置

### 移除的功能
- ~~编辑器设置~~ - 根据用户需求移除

## 架构设计

### 整体架构

采用模块化设计，遵循单一职责原则：

```
internal/
├── i18n/              # 多语言支持
│   ├── i18n.go        # 核心接口和 Manager
│   ├── messages.go    # 所有翻译键定义
│   ├── zh.go          # 中文翻译
│   ├── en.go          # 英文翻译
│   ├── ja.go          # 日文翻译
│   └── i18n_test.go   # 单元测试
│
├── theme/             # 主题系统
│   ├── theme.go       # Theme 结构体和接口
│   ├── presets.go     # 5 个预设主题定义
│   └── theme_test.go  # 单元测试
│
└── config/            # 扩展现有配置
    ├── config.go      # 添加 Settings 字段
    └── settings.go    # Settings 结构体（新增）
```

### 依赖关系

```
repl 包 → 依赖 → i18n 包（翻译）
                → theme 包（主题样式）
                → config 包（配置管理）
```

### 配置文件结构

扩展后的 `~/.cc-start/settings.json`：

```json
{
  "profiles": [
    {
      "name": "anthropic",
      "base_url": "https://api.anthropic.com",
      "model": "claude-sonnet-4-5-20250929",
      "token": "sk-ant-xxx"
    }
  ],
  "default": "anthropic",
  "settings": {
    "language": "zh",
    "theme": "default"
  }
}
```

## 详细设计

### 1. 多语言系统（i18n）

#### 核心接口

```go
// internal/i18n/i18n.go

// Manager 多语言管理器
type Manager struct {
    currentLang string
    translations map[string]map[string]string // lang -> key -> text
}

// 支持的语言列表
const (
    LangZH = "zh"  // 中文
    LangEN = "en"  // 英文
    LangJA = "ja"  // 日文
)

// 核心方法
func NewManager() *Manager
func (m *Manager) SetLanguage(lang string) error
func (m *Manager) T(key string) string              // 翻译文本
func (m *Manager) TWithData(key string, data map[string]interface{}) string // 带变量的翻译
func (m *Manager) GetSupportedLanguages() []string
```

#### 翻译键命名规范

使用点分命名，按功能模块组织：

```go
// internal/i18n/messages.go
const (
    // 通用
    MsgCommonSuccess = "common.success"
    MsgCommonError   = "common.error"

    // 设置面板
    MsgSettingsTitle     = "settings.title"
    MsgSettingsLanguage  = "settings.language"
    MsgSettingsTheme     = "settings.theme"
    MsgSettingsHint      = "settings.hint"

    // 命令面板
    MsgPaletteTitle      = "palette.title"
    MsgPaletteSearchHint = "palette.search_hint"

    // REPL 界面
    MsgREPLInputPrompt = "repl.input_prompt"
    MsgREPLWelcome     = "repl.welcome"

    // 命令描述
    MsgCmdList   = "cmd.list"
    MsgCmdUse    = "cmd.use"
    MsgCmdSetup  = "cmd.setup"
    // ... 更多命令
)
```

#### 翻译数据结构

每个语言一个文件，使用 map 存储翻译：

```go
// internal/i18n/zh.go
var zhTranslations = map[string]string{
    "common.success":    "成功",
    "common.error":      "错误",
    "settings.title":    "⚙ 系统设置",
    "settings.language": "语言 / Language",
    "settings.theme":    "主题 / Theme",
    "palette.title":     "命令面板",
    "repl.welcome":      "欢迎使用 CC-Start",
    "cmd.list":          "列出所有配置",
    // ... 完整翻译
}

// internal/i18n/en.go
var enTranslations = map[string]string{
    "common.success":    "Success",
    "common.error":      "Error",
    "settings.title":    "⚙ Settings",
    "settings.language": "Language",
    "settings.theme":    "Theme",
    // ... 完整翻译
}
```

#### 使用方式

```go
// 在 Model 中添加 i18n manager
type Model struct {
    // ...
    i18n *i18n.Manager
    // ...
}

// 使用翻译
output := m.i18n.T(i18n.MsgSettingsTitle)
```

### 2. 主题系统（theme）

#### 核心结构

```go
// internal/theme/theme.go

// Theme 主题定义
type Theme struct {
    Name        string
    DisplayName string
    Colors      ColorScheme
}

// ColorScheme 颜色方案
type ColorScheme struct {
    // 基础颜色
    Background string
    Foreground string
    Muted      string

    // 状态颜色
    Primary    string
    Success    string
    Error      string
    Warning    string
    Info       string

    // 组件颜色
    Border     string
    Accent     string
    Highlight  string

    // 面板专用
    PaletteBg      string
    PaletteActive  string
    PaletteInactive string
}

// 核心方法
func GetTheme(name string) (*Theme, error)
func GetAllThemes() []Theme
func ApplyTheme(theme *Theme, styles *Styles) *Styles  // 将主题应用到样式
```

#### 5 个预设主题

```go
// internal/theme/presets.go

var presets = []Theme{
    {
        Name:        "default",
        DisplayName: "默认 / Default",
        Colors: ColorScheme{
            Background: "#1a1a1a",
            Foreground: "#ffffff",
            Muted:      "#626262",
            Primary:    "#00d7ff",
            Success:    "#00ff00",
            Error:      "#ff0000",
            Warning:    "#ffff00",
            Info:       "#00d7ff",
            Border:     "#4a4a4a",
            Accent:     "#ff6b35",
            Highlight:  "#ffff00",
        },
    },
    {
        Name:        "ocean",
        DisplayName: "海洋 / Ocean",
        Colors: ColorScheme{
            Background: "#0c2340",
            Foreground: "#e0f0ff",
            Primary:    "#00bfff",
            Success:    "#00ff88",
            // ... 海洋蓝绿色调
        },
    },
    {
        Name:        "forest",
        DisplayName: "森林 / Forest",
        Colors: ColorScheme{
            Background: "#1a2819",
            Foreground: "#e8f5e9",
            Primary:    "#4caf50",
            Success:    "#76ff03",
            // ... 森林绿色调
        },
    },
    {
        Name:        "sunset",
        DisplayName: "日落 / Sunset",
        Colors: ColorScheme{
            Background: "#2d1b2d",
            Foreground: "#fff8e1",
            Primary:    "#ff6b35",
            Success:    "#ffeb3b",
            // ... 暖色橙红色调
        },
    },
    {
        Name:        "light",
        DisplayName: "亮色 / Light",
        Colors: ColorScheme{
            Background: "#ffffff",
            Foreground: "#1a1a1a",
            Primary:    "#0066cc",
            Success:    "#008800",
            // ... 亮色主题
        },
    },
}
```

#### 主题应用机制

```go
// 将主题颜色应用到 Bubble Tea 样式
func ApplyTheme(theme *Theme, styles *Styles) *Styles {
    styles.PaletteTitle = lipgloss.NewStyle().
        Foreground(lipgloss.Color(theme.Colors.Primary)).
        Bold(true).
        Padding(0, 1)

    styles.PaletteActive = lipgloss.NewStyle().
        Foreground(lipgloss.Color(theme.Colors.Highlight)).
        Background(lipgloss.Color(theme.Colors.Background))

    // ... 应用所有样式

    return styles
}
```

### 3. 设置面板交互

#### 状态机设计

```go
// internal/repl/settings.go 扩展

type SettingsPanel struct {
    visible      bool
    mode         SettingsMode  // 新增：当前模式
    query        string
    items        []SettingsItem
    selected     int
    subItems     []SettingsItem  // 新增：子选项（如语言列表）
    styles       Styles
    width        int
    i18n         *i18n.Manager   // 新增
}

type SettingsMode int

const (
    SettingsModeMain    SettingsMode = iota  // 主设置列表
    SettingsModeLanguage                      // 语言选择
    SettingsModeTheme                         // 主题选择
)
```

#### 交互流程

```
1. 用户按 Ctrl+P
   ↓
2. 显示主设置面板
   - 语言设置 [中文]
   - 主题设置 [默认]
   ↓
3. 用户选择"语言设置"按 Enter
   ↓
4. 进入语言选择模式
   - ● 中文
   - English
   - 日本語
   ↓
5. 用户选择语言并按 Enter
   ↓
6. 立即应用语言更改
   - 切换 i18n.Manager 的语言
   - 更新所有 UI 文本
   - 保存到配置文件
   - 返回主设置面板（已更新语言）
   ↓
7. 用户按 Esc 关闭设置面板
```

#### 主题预览

```go
// 选择主题时提供实时预览
func (s *SettingsPanel) renderThemeList() string {
    themes := theme.GetAllThemes()
    var lines []string

    for i, t := range themes {
        // 使用主题颜色渲染预览文本
        preview := fmt.Sprintf("● %s  %s",
            t.DisplayName,
            renderColorSample(t.Colors))

        if i == s.selected {
            lines = append(lines,
                lipgloss.NewStyle().
                    Foreground(lipgloss.Color(t.Colors.Highlight)).
                    Render(preview))
        } else {
            lines = append(lines,
                lipgloss.NewStyle().
                    Foreground(lipgloss.Color(t.Colors.Foreground)).
                    Render(preview))
        }
    }

    return strings.Join(lines, "\n")
}
```

### 4. 配置管理

#### Settings 结构体

```go
// internal/config/settings.go

type Settings struct {
    Language string `json:"language"`
    Theme    string `json:"theme"`
}

// 在 Config 结构中添加
type Config struct {
    Profiles []Profile `json:"profiles"`
    Default  string    `json:"default,omitempty"`
    Settings Settings  `json:"settings,omitempty"`  // 新增
}

// 保存设置
func (c *Config) SaveSettings(path string) error {
    return c.Save(path)
}

// 更新单个设置项
func (c *Config) UpdateSetting(key, value string) {
    switch key {
    case "language":
        c.Settings.Language = value
    case "theme":
        c.Settings.Theme = value
    }
}
```

#### 配置迁移

```go
// internal/config/config.go

func LoadConfig(path string) (*Config, error) {
    // 读取文件
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    // 迁移：如果 settings 为空，设置默认值
    if cfg.Settings.Language == "" {
        cfg.Settings.Language = "zh"
    }
    if cfg.Settings.Theme == "" {
        cfg.Settings.Theme = "default"
    }

    return &cfg, nil
}
```

### 5. 数据流和状态管理

#### 初始化流程

```
程序启动
    ↓
1. 加载配置文件
   - 读取 ~/.cc-start/settings.json
   - 如果没有 settings，使用默认值
    ↓
2. 初始化 i18n Manager
   - 根据配置中的 language 设置
   - 默认为 "zh"
    ↓
3. 初始化 Theme
   - 根据配置中的 theme 设置
   - 默认为 "default"
    ↓
4. 应用主题到 Styles
   - theme.ApplyTheme(theme, &styles)
    ↓
5. 创建 REPL Model
   - 注入 i18n manager
   - 注入 theme
   - 注入 styles
    ↓
6. 启动 Bubble Tea 程序
```

#### 状态变更流程

```
用户选择新语言（如 "en"）
    ↓
1. UI 层：settings panel 发出 action "lang:en"
    ↓
2. Controller 层：handleSettingAction 处理
    ↓
3. 业务逻辑层：
   a. i18n.SetLanguage("en")  // 更新内存状态
   b. config.UpdateSetting("language", "en")
   c. config.Save(configPath)  // 持久化
    ↓
4. UI 更新：
   a. 所有 i18n.T() 调用返回英文
   b. 设置面板标题变为 "⚙ Settings"
   c. 命令描述变为英文
    ↓
5. 下次启动时：
   - 从配置文件读取 language="en"
   - 自动应用英文界面
```

### 6. 错误处理和边界情况

#### 错误处理策略

```go
// 1. 无效的语言代码
func (m *Manager) SetLanguage(lang string) error {
    if !isValidLanguage(lang) {
        return fmt.Errorf("unsupported language: %s", lang)
    }
    m.currentLang = lang
    return nil
}

// 2. 缺失的翻译键 - 多级回退
func (m *Manager) T(key string) string {
    // 1. 当前语言
    if text, ok := m.translations[m.currentLang][key]; ok {
        return text
    }
    // 2. 英文
    if text, ok := m.translations[LangEN][key]; ok {
        return text
    }
    // 3. 键名本身
    return fmt.Sprintf("[%s]", key)
}

// 3. 无效的主题名称
func GetTheme(name string) (*Theme, error) {
    for _, t := range presets {
        if t.Name == name {
            return &t, nil
        }
    }
    return nil, fmt.Errorf("theme '%s' not found", name)
}

// 4. 配置文件损坏
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            // 文件不存在，返回默认配置
            return &Config{
                Settings: Settings{
                    Language: "zh",
                    Theme:    "default",
                },
            }, nil
        }
        return nil, fmt.Errorf("读取配置失败: %w", err)
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        // 配置文件损坏，备份并创建新的
        backupPath := path + ".backup"
        os.Rename(path, backupPath)
        return &Config{
            Settings: Settings{
                Language: "zh",
                Theme:    "default",
            },
        }, fmt.Errorf("配置文件已损坏，已备份到 %s", backupPath)
    }

    return &cfg, nil
}
```

#### 边界情况处理

```go
// 1. 用户快速连续切换设置 - 使用防抖
func (m *Model) saveSettingsDebounced() {
    time.AfterFunc(100*time.Millisecond, func() {
        m.config.Save(m.configPath)
    })
}

// 2. 并发访问配置文件 - 原子写入
func (c Config) Save(path string) error {
    tmpPath := path + ".tmp"
    data, _ := json.MarshalIndent(c, "", "  ")
    if err := os.WriteFile(tmpPath, data, 0600); err != nil {
        return err
    }
    return os.Rename(tmpPath, path)
}

// 3. 终端不支持颜色
func ApplyTheme(theme *Theme, styles *Styles) *Styles {
    if !term.HasColor() {
        return applyMonochromeStyles(styles)
    }
    return applyColoredStyles(theme, styles)
}
```

#### 用户友好的错误提示

所有错误信息都通过 i18n 翻译：

```go
const (
    MsgErrConfigLoad      = "error.config_load"
    MsgErrConfigSave      = "error.config_save"
    MsgErrInvalidLanguage = "error.invalid_language"
    MsgErrInvalidTheme    = "error.invalid_theme"
)

// 使用示例
if err := m.config.Save(m.configPath); err != nil {
    m.output.WriteError(
        fmt.Sprintf(m.i18n.T(i18n.MsgErrConfigSave), err.Error()))
}
```

## 测试策略

### 单元测试

1. **i18n 包测试**
   - 测试语言切换
   - 测试翻译键回退机制
   - 测试无效语言代码处理

2. **theme 包测试**
   - 测试主题获取
   - 测试主题应用
   - 测试无效主题名称处理

3. **config 包测试**
   - 测试配置加载和保存
   - 测试配置迁移
   - 测试并发写入

### 集成测试

1. **设置面板交互**
   - 测试完整的语言切换流程
   - 测试完整的主题切换流程
   - 测试设置持久化

2. **配置文件兼容性**
   - 测试旧配置文件迁移
   - 测试配置文件损坏恢复

## 实现计划

详见实现计划文档（待生成）。

## 风险和缓解措施

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 翻译不完整 | 部分文本未翻译 | 实现多级回退机制，至少显示英文或键名 |
| 主题颜色在部分终端显示异常 | 视觉体验差 | 检测终端能力，不支持颜色时使用单色模式 |
| 配置文件并发写入冲突 | 数据丢失 | 使用原子写入（临时文件+重命名） |
| 用户快速切换设置导致频繁 IO | 性能问题 | 实现防抖机制，延迟保存 |

## 未来扩展

1. **更多语言**：添加更多语言支持（韩语、法语等）
2. **自定义主题**：允许用户创建和保存自定义主题
3. **主题导入导出**：支持主题配置的导入和导出
4. **更多设置项**：添加字体大小、快捷键绑定等设置

## 参考资料

- [Bubble Tea 文档](https://github.com/charmbracelet/bubbletea)
- [Lipgloss 文档](https://github.com/charmbracelet/lipgloss)
- [Go i18n 最佳实践](https://go.dev/blog/i18n)

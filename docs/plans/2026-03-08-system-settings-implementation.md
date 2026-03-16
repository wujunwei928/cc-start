# System Settings Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现系统设置功能，包括多语言支持（中文、英文、日文）和 5 个预设主题

**Architecture:** 采用模块化设计，创建独立的 i18n 和 theme 包，扩展 config 包添加 Settings 字段，更新 REPL 集成 i18n 和 theme

**Tech Stack:** Go, Bubble Tea, Lipgloss

---

## Task 1: 创建 i18n 包基础结构和接口

**Files:**
- Create: `internal/i18n/i18n.go`
- Create: `internal/i18n/messages.go`
- Create: `internal/i18n/i18n_test.go`

**Step 1: 写失败的测试**

```go
// internal/i18n/i18n_test.go
package i18n

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}

	if m.currentLang != "zh" {
		t.Errorf("default language = %s, want zh", m.currentLang)
	}
}

func TestSetLanguage(t *testing.T) {
	m := NewManager()

	tests := []struct {
		name    string
		lang    string
		wantErr bool
	}{
		{"valid zh", "zh", false},
		{"valid en", "en", false},
		{"valid ja", "ja", false},
		{"invalid", "fr", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.SetLanguage(tt.lang)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetLanguage(%s) error = %v, wantErr %v", tt.lang, err, tt.wantErr)
			}
		})
	}
}

func TestT(t *testing.T) {
	m := NewManager()

	// 测试翻译键
	text := m.T("test.key")
	if text == "" {
		t.Error("T() returned empty string")
	}

	// 测试不存在的键 - 应该返回键名
	missingKey := m.T("nonexistent.key")
	if missingKey != "nonexistent.key" {
		t.Errorf("T(nonexistent) = %s, want nonexistent.key", missingKey)
	}
}

func TestGetSupportedLanguages(t *testing.T) {
	m := NewManager()
	langs := m.GetSupportedLanguages()

	if len(langs) == 0 {
		t.Fatal("GetSupportedLanguages() returned empty list")
	}

	expected := []string{"zh", "en", "ja"}
	for _, exp := range expected {
		found := false
		for _, lang := range langs {
			if lang == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetSupportedLanguages() missing %s", exp)
		}
	}
}
```

**Step 2: 运行测试验证失败**

```bash
go test ./internal/i18n -v
```

Expected: FAIL - 包不存在

**Step 3: 实现最小代码使测试通过**

```go
// internal/i18n/i18n.go
package i18n

import (
	"fmt"
)

const (
	LangZH = "zh"
	LangEN = "en"
	LangJA = "ja"
)

var supportedLanguages = []string{LangZH, LangEN, LangJA}

type Manager struct {
	currentLang   string
	translations  map[string]map[string]string
}

func NewManager() *Manager {
	m := &Manager{
		currentLang:  LangZH,
		translations: make(map[string]map[string]string),
	}

	// 加载翻译
	m.loadTranslations()

	return m
}

func (m *Manager) SetLanguage(lang string) error {
	for _, supported := range supportedLanguages {
		if lang == supported {
			m.currentLang = lang
			return nil
		}
	}
	return fmt.Errorf("unsupported language: %s", lang)
}

func (m *Manager) T(key string) string {
	// 1. 尝试当前语言
	if trans, ok := m.translations[m.currentLang]; ok {
		if text, ok := trans[key]; ok {
			return text
		}
	}

	// 2. 回退到英文
	if trans, ok := m.translations[LangEN]; ok {
		if text, ok := trans[key]; ok {
			return text
		}
	}

	// 3. 返回键名
	return key
}

func (m *Manager) TWithData(key string, data map[string]interface{}) string {
	text := m.T(key)
	// TODO: 实现变量替换
	return text
}

func (m *Manager) GetSupportedLanguages() []string {
	return supportedLanguages
}

func (m *Manager) loadTranslations() {
	// 将在后续任务中实现
	m.translations[LangZH] = getZhTranslations()
	m.translations[LangEN] = getEnTranslations()
	m.translations[LangJA] = getJaTranslations()
}
```

```go
// internal/i18n/messages.go
package i18n

const (
	// 通用
	MsgCommonSuccess = "common.success"
	MsgCommonError   = "common.error"
	MsgCommonInfo    = "common.info"
	MsgCommonWarning = "common.warning"

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
	MsgREPLHint        = "repl.hint"

	// 命令描述
	MsgCmdList   = "cmd.list"
	MsgCmdUse    = "cmd.use"
	MsgCmdSetup  = "cmd.setup"
	MsgCmdEdit   = "cmd.edit"
	MsgCmdDelete = "cmd.delete"
	MsgCmdCopy   = "cmd.copy"
	MsgCmdRename = "cmd.rename"
	MsgCmdTest   = "cmd.test"
	MsgCmdExport = "cmd.export"
	MsgCmdImport = "cmd.import"
	MsgCmdRun    = "cmd.run"
	MsgCmdHelp   = "cmd.help"
	MsgCmdExit   = "cmd.exit"
	MsgCmdClear  = "cmd.clear"
	MsgCmdHistory = "cmd.history"
	MsgCmdDefault = "cmd.default"
	MsgCmdShow   = "cmd.show"
	MsgCmdCurrent = "cmd.current"

	// 错误消息
	MsgErrConfigLoad      = "error.config_load"
	MsgErrConfigSave      = "error.config_save"
	MsgErrInvalidLanguage = "error.invalid_language"
	MsgErrInvalidTheme    = "error.invalid_theme"
	MsgErrProfileNotFound = "error.profile_not_found"

	// 测试键（仅用于测试）
	"test.key": "test.key",
)
```

**Step 4: 运行测试验证通过**

```bash
go test ./internal/i18n -v
```

Expected: PASS

**Step 5: 提交**

```bash
git add internal/i18n/i18n.go internal/i18n/messages.go internal/i18n/i18n_test.go
git commit -m "feat(i18n): 添加 i18n 基础结构和接口"
```

---

## Task 2: 实现中文翻译

**Files:**
- Create: `internal/i18n/zh.go`
- Modify: `internal/i18n/i18n.go:loadTranslations()`

**Step 1: 写失败的测试**

```go
// 在 internal/i18n/i18n_test.go 添加
func TestZhTranslations(t *testing.T) {
	m := NewManager()
	m.SetLanguage(LangZH)

	tests := []struct {
		key  string
		want string
	}{
		{MsgCommonSuccess, "成功"},
		{MsgCommonError, "错误"},
		{MsgSettingsTitle, "⚙ 系统设置"},
		{MsgSettingsLanguage, "语言 / Language"},
		{MsgSettingsTheme, "主题 / Theme"},
		{MsgCmdList, "列出所有配置"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := m.T(tt.key)
			if got != tt.want {
				t.Errorf("T(%s) = %s, want %s", tt.key, got, tt.want)
			}
		})
	}
}
```

**Step 2: 运行测试验证失败**

```bash
go test ./internal/i18n -run TestZhTranslations -v
```

Expected: FAIL - getZhTranslations 未定义

**Step 3: 实现中文翻译**

```go
// internal/i18n/zh.go
package i18n

func getZhTranslations() map[string]string {
	return map[string]string{
		// 通用
		MsgCommonSuccess: "成功",
		MsgCommonError:   "错误",
		MsgCommonInfo:    "信息",
		MsgCommonWarning: "警告",

		// 设置面板
		MsgSettingsTitle:     "⚙ 系统设置",
		MsgSettingsLanguage:  "语言 / Language",
		MsgSettingsTheme:     "主题 / Theme",
		MsgSettingsHint:      "↑↓ 导航  enter 确认  esc 关闭",

		// 命令面板
		MsgPaletteTitle:      "命令面板",
		MsgPaletteSearchHint: "输入搜索命令...",

		// REPL 界面
		MsgREPLInputPrompt: "输入命令...",
		MsgREPLWelcome:     "欢迎使用 CC-Start",
		MsgREPLHint:        "输入 /help 查看帮助",

		// 命令描述
		MsgCmdList:    "列出所有配置",
		MsgCmdUse:     "切换当前会话配置",
		MsgCmdSetup:   "运行配置向导",
		MsgCmdEdit:    "编辑配置",
		MsgCmdDelete:  "删除配置",
		MsgCmdCopy:    "复制配置",
		MsgCmdRename:  "重命名配置",
		MsgCmdTest:    "测试 API 连通性",
		MsgCmdExport:  "导出配置到 stdout 或文件",
		MsgCmdImport:  "从文件导入配置",
		MsgCmdRun:     "使用当前或指定配置启动",
		MsgCmdHelp:    "显示帮助",
		MsgCmdExit:    "退出",
		MsgCmdClear:   "清屏",
		MsgCmdHistory: "显示命令历史",
		MsgCmdDefault: "设置默认配置",
		MsgCmdShow:    "显示配置详情",
		MsgCmdCurrent: "显示当前配置",

		// 错误消息
		MsgErrConfigLoad:      "加载配置失败: %s",
		MsgErrConfigSave:      "保存配置失败: %s",
		MsgErrInvalidLanguage: "不支持的语言: %s",
		MsgErrInvalidTheme:    "不支持的主题: %s",
		MsgErrProfileNotFound: "配置 '%s' 不存在",
	}
}
```

**Step 4: 运行测试验证通过**

```bash
go test ./internal/i18n -run TestZhTranslations -v
```

Expected: PASS

**Step 5: 提交**

```bash
git add internal/i18n/zh.go internal/i18n/i18n_test.go
git commit -m "feat(i18n): 添加中文翻译"
```

---

## Task 3: 实现英文翻译

**Files:**
- Create: `internal/i18n/en.go`
- Modify: `internal/i18n/i18n_test.go`

**Step 1: 写失败的测试**

```go
// 在 internal/i18n/i18n_test.go 添加
func TestEnTranslations(t *testing.T) {
	m := NewManager()
	m.SetLanguage(LangEN)

	tests := []struct {
		key  string
		want string
	}{
		{MsgCommonSuccess, "Success"},
		{MsgCommonError, "Error"},
		{MsgSettingsTitle, "⚙ Settings"},
		{MsgSettingsLanguage, "Language"},
		{MsgSettingsTheme, "Theme"},
		{MsgCmdList, "List all profiles"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := m.T(tt.key)
			if got != tt.want {
				t.Errorf("T(%s) = %s, want %s", tt.key, got, tt.want)
			}
		})
	}
}
```

**Step 2: 运行测试验证失败**

```bash
go test ./internal/i18n -run TestEnTranslations -v
```

Expected: FAIL

**Step 3: 实现英文翻译**

```go
// internal/i18n/en.go
package i18n

func getEnTranslations() map[string]string {
	return map[string]string{
		// Common
		MsgCommonSuccess: "Success",
		MsgCommonError:   "Error",
		MsgCommonInfo:    "Info",
		MsgCommonWarning: "Warning",

		// Settings Panel
		MsgSettingsTitle:     "⚙ Settings",
		MsgSettingsLanguage:  "Language",
		MsgSettingsTheme:     "Theme",
		MsgSettingsHint:      "↑↓ Navigate  enter Confirm  esc Close",

		// Command Palette
		MsgPaletteTitle:      "Command Palette",
		MsgPaletteSearchHint: "Type to search commands...",

		// REPL Interface
		MsgREPLInputPrompt: "Enter command...",
		MsgREPLWelcome:     "Welcome to CC-Start",
		MsgREPLHint:        "Type /help for available commands",

		// Command Descriptions
		MsgCmdList:    "List all profiles",
		MsgCmdUse:     "Switch current profile",
		MsgCmdSetup:   "Run setup wizard",
		MsgCmdEdit:    "Edit profile",
		MsgCmdDelete:  "Delete profile",
		MsgCmdCopy:    "Copy profile",
		MsgCmdRename:  "Rename profile",
		MsgCmdTest:    "Test API connectivity",
		MsgCmdExport:  "Export config to stdout or file",
		MsgCmdImport:  "Import config from file",
		MsgCmdRun:     "Launch with current or specified profile",
		MsgCmdHelp:    "Show help",
		MsgCmdExit:    "Exit",
		MsgCmdClear:   "Clear screen",
		MsgCmdHistory: "Show command history",
		MsgCmdDefault: "Set default profile",
		MsgCmdShow:    "Show profile details",
		MsgCmdCurrent: "Show current profile",

		// Error Messages
		MsgErrConfigLoad:      "Failed to load config: %s",
		MsgErrConfigSave:      "Failed to save config: %s",
		MsgErrInvalidLanguage: "Unsupported language: %s",
		MsgErrInvalidTheme:    "Unsupported theme: %s",
		MsgErrProfileNotFound: "Profile '%s' not found",
	}
}
```

**Step 4: 运行测试验证通过**

```bash
go test ./internal/i18n -run TestEnTranslations -v
```

Expected: PASS

**Step 5: 提交**

```bash
git add internal/i18n/en.go internal/i18n/i18n_test.go
git commit -m "feat(i18n): 添加英文翻译"
```

---

## Task 4: 实现日文翻译

**Files:**
- Create: `internal/i18n/ja.go`
- Modify: `internal/i18n/i18n_test.go`

**Step 1: 写失败的测试**

```go
// 在 internal/i18n/i18n_test.go 添加
func TestJaTranslations(t *testing.T) {
	m := NewManager()
	m.SetLanguage(LangJA)

	tests := []struct {
		key  string
		want string
	}{
		{MsgCommonSuccess, "成功"},
		{MsgCommonError, "エラー"},
		{MsgSettingsTitle, "⚙ 設定"},
		{MsgSettingsLanguage, "言語"},
		{MsgSettingsTheme, "テーマ"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := m.T(tt.key)
			if got != tt.want {
				t.Errorf("T(%s) = %s, want %s", tt.key, got, tt.want)
			}
		})
	}
}
```

**Step 2: 运行测试验证失败**

```bash
go test ./internal/i18n -run TestJaTranslations -v
```

Expected: FAIL

**Step 3: 实现日文翻译**

```go
// internal/i18n/ja.go
package i18n

func getJaTranslations() map[string]string {
	return map[string]string{
		// 共通
		MsgCommonSuccess: "成功",
		MsgCommonError:   "エラー",
		MsgCommonInfo:    "情報",
		MsgCommonWarning: "警告",

		// 設定パネル
		MsgSettingsTitle:     "⚙ 設定",
		MsgSettingsLanguage:  "言語",
		MsgSettingsTheme:     "テーマ",
		MsgSettingsHint:      "↑↓ 移動  enter 確定  esc 閉じる",

		// コマンドパレット
		MsgPaletteTitle:      "コマンドパレット",
		MsgPaletteSearchHint: "コマンドを検索...",

		// REPL インターフェース
		MsgREPLInputPrompt: "コマンドを入力...",
		MsgREPLWelcome:     "CC-Startへようこそ",
		MsgREPLHint:        "/help でヘルプを表示",

		// コマンド説明
		MsgCmdList:    "すべてのプロファイルを一覧表示",
		MsgCmdUse:     "現在のプロファイルを切り替え",
		MsgCmdSetup:   "セットアップウィザードを実行",
		MsgCmdEdit:    "プロファイルを編集",
		MsgCmdDelete:  "プロファイルを削除",
		MsgCmdCopy:    "プロファイルをコピー",
		MsgCmdRename:  "プロファイル名を変更",
		MsgCmdTest:    "API接続をテスト",
		MsgCmdExport:  "設定をstdoutまたはファイルにエクスポート",
		MsgCmdImport:  "ファイルから設定をインポート",
		MsgCmdRun:     "現在または指定されたプロファイルで起動",
		MsgCmdHelp:    "ヘルプを表示",
		MsgCmdExit:    "終了",
		MsgCmdClear:   "画面をクリア",
		MsgCmdHistory: "コマンド履歴を表示",
		MsgCmdDefault: "デフォルトプロファイルを設定",
		MsgCmdShow:    "プロファイル詳細を表示",
		MsgCmdCurrent: "現在のプロファイルを表示",

		// エラーメッセージ
		MsgErrConfigLoad:      "設定の読み込みに失敗: %s",
		MsgErrConfigSave:      "設定の保存に失敗: %s",
		MsgErrInvalidLanguage: "サポートされていない言語: %s",
		MsgErrInvalidTheme:    "サポートされていないテーマ: %s",
		MsgErrProfileNotFound: "プロファイル '%s' が見つかりません",
	}
}
```

**Step 4: 运行测试验证通过**

```bash
go test ./internal/i18n -run TestJaTranslations -v
```

Expected: PASS

**Step 5: 提交**

```bash
git add internal/i18n/ja.go internal/i18n/i18n_test.go
git commit -m "feat(i18n): 添加日文翻译"
```

---

## Task 5: 创建 theme 包基础结构和接口

**Files:**
- Create: `internal/theme/theme.go`
- Create: `internal/theme/theme_test.go`

**Step 1: 写失败的测试**

```go
// internal/theme/theme_test.go
package theme

import (
	"testing"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name    string
		theme   string
		wantErr bool
	}{
		{"default", "default", false},
		{"ocean", "ocean", false},
		{"forest", "forest", false},
		{"sunset", "sunset", false},
		{"light", "light", false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme, err := GetTheme(tt.theme)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTheme(%s) error = %v, wantErr %v", tt.theme, err, tt.wantErr)
				return
			}
			if !tt.wantErr && theme == nil {
				t.Errorf("GetTheme(%s) returned nil", tt.theme)
			}
		})
	}
}

func TestGetAllThemes(t *testing.T) {
	themes := GetAllThemes()

	if len(themes) != 5 {
		t.Errorf("GetAllThemes() returned %d themes, want 5", len(themes))
	}

	// 验证每个主题都有必要的字段
	for _, theme := range themes {
		if theme.Name == "" {
			t.Error("Theme has empty Name")
		}
		if theme.DisplayName == "" {
			t.Error("Theme has empty DisplayName")
		}
		if theme.Colors.Background == "" {
			t.Error("Theme has empty Background color")
		}
	}
}

func TestThemeStructure(t *testing.T) {
	theme, err := GetTheme("default")
	if err != nil {
		t.Fatalf("GetTheme(default) error: %v", err)
	}

	// 验证默认主题的颜色
	if theme.Colors.Primary == "" {
		t.Error("Default theme missing Primary color")
	}
	if theme.Colors.Success == "" {
		t.Error("Default theme missing Success color")
	}
}
```

**Step 2: 运行测试验证失败**

```bash
go test ./internal/theme -v
```

Expected: FAIL - 包不存在

**Step 3: 实现最小代码使测试通过**

```go
// internal/theme/theme.go
package theme

import (
	"fmt"
)

type ColorScheme struct {
	Background string
	Foreground string
	Muted      string

	Primary   string
	Success   string
	Error     string
	Warning   string
	Info      string

	Border     string
	Accent     string
	Highlight  string

	PaletteBg       string
	PaletteActive   string
	PaletteInactive string
}

type Theme struct {
	Name        string
	DisplayName string
	Colors      ColorScheme
}

func GetTheme(name string) (*Theme, error) {
	for i := range presets {
		if presets[i].Name == name {
			return &presets[i], nil
		}
	}
	return nil, fmt.Errorf("theme '%s' not found", name)
}

func GetAllThemes() []Theme {
	return presets
}
```

**Step 4: 运行测试验证失败（presets 未定义）**

```bash
go test ./internal/theme -v
```

Expected: FAIL - presets 未定义

**Step 5: 提交**

```bash
git add internal/theme/theme.go internal/theme/theme_test.go
git commit -m "feat(theme): 添加 theme 基础结构"
```

---

## Task 6: 实现 5 个预设主题

**Files:**
- Create: `internal/theme/presets.go`
- Modify: `internal/theme/theme_test.go` (添加颜色验证)

**Step 1: 写失败的测试**

```go
// 在 internal/theme/theme_test.go 添加
func TestPresetThemesColors(t *testing.T) {
	themes := GetAllThemes()

	requiredColors := []string{
		"Background", "Foreground", "Primary", "Success", "Error",
		"Warning", "Info", "Border", "Accent", "Highlight",
	}

	for _, theme := range themes {
		t.Run(theme.Name, func(t *testing.T) {
			colors := map[string]string{
				"Background": theme.Colors.Background,
				"Foreground": theme.Colors.Foreground,
				"Primary":    theme.Colors.Primary,
				"Success":    theme.Colors.Success,
				"Error":      theme.Colors.Error,
				"Warning":    theme.Colors.Warning,
				"Info":       theme.Colors.Info,
				"Border":     theme.Colors.Border,
				"Accent":     theme.Colors.Accent,
				"Highlight":  theme.Colors.Highlight,
			}

			for _, colorName := range requiredColors {
				if colors[colorName] == "" {
					t.Errorf("Theme %s missing color: %s", theme.Name, colorName)
				}

				// 验证颜色格式（#RRGGBB 或 #RGB）
				color := colors[colorName]
				if len(color) != 7 && len(color) != 4 {
					t.Errorf("Theme %s color %s has invalid format: %s", theme.Name, colorName, color)
				}
				if color[0] != '#' {
					t.Errorf("Theme %s color %s must start with #", theme.Name, colorName)
				}
			}
		})
	}
}
```

**Step 2: 运行测试验证失败**

```bash
go test ./internal/theme -v
```

Expected: FAIL - presets 未定义

**Step 3: 实现 5 个预设主题**

```go
// internal/theme/presets.go
package theme

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
			PaletteBg:       "#2a2a2a",
			PaletteActive:   "#3a3a3a",
			PaletteInactive: "#1a1a1a",
		},
	},
	{
		Name:        "ocean",
		DisplayName: "海洋 / Ocean",
		Colors: ColorScheme{
			Background: "#0c2340",
			Foreground: "#e0f0ff",
			Muted:      "#5a8ab0",
			Primary:    "#00bfff",
			Success:    "#00ff88",
			Error:      "#ff4d4d",
			Warning:    "#ffd700",
			Info:       "#00bfff",
			Border:     "#1e4d7b",
			Accent:     "#00ced1",
			Highlight:  "#00ffff",
			PaletteBg:       "#1a3a5c",
			PaletteActive:   "#2a4a6c",
			PaletteInactive: "#0c2340",
		},
	},
	{
		Name:        "forest",
		DisplayName: "森林 / Forest",
		Colors: ColorScheme{
			Background: "#1a2819",
			Foreground: "#e8f5e9",
			Muted:      "#6b8e6b",
			Primary:    "#4caf50",
			Success:    "#76ff03",
			Error:      "#ff5252",
			Warning:    "#ffeb3b",
			Info:       "#4caf50",
			Border:     "#2e4d2e",
			Accent:     "#8bc34a",
			Highlight:  "#c6ff00",
			PaletteBg:       "#2a3a29",
			PaletteActive:   "#3a4a39",
			PaletteInactive: "#1a2819",
		},
	},
	{
		Name:        "sunset",
		DisplayName: "日落 / Sunset",
		Colors: ColorScheme{
			Background: "#2d1b2d",
			Foreground: "#fff8e1",
			Muted:      "#a08080",
			Primary:    "#ff6b35",
			Success:    "#ffeb3b",
			Error:      "#ff1744",
			Warning:    "#ff9800",
			Info:       "#ff6b35",
			Border:     "#4a2a4a",
			Accent:     "#ff4081",
			Highlight:  "#ffd740",
			PaletteBg:       "#3d2b3d",
			PaletteActive:   "#4d3b4d",
			PaletteInactive: "#2d1b2d",
		},
	},
	{
		Name:        "light",
		DisplayName: "亮色 / Light",
		Colors: ColorScheme{
			Background: "#ffffff",
			Foreground: "#1a1a1a",
			Muted:      "#757575",
			Primary:    "#0066cc",
			Success:    "#008800",
			Error:      "#cc0000",
			Warning:    "#cc8800",
			Info:       "#0066cc",
			Border:     "#e0e0e0",
			Accent:     "#ff6b35",
			Highlight:  "#0066cc",
			PaletteBg:       "#f5f5f5",
			PaletteActive:   "#e8e8e8",
			PaletteInactive: "#ffffff",
		},
	},
}
```

**Step 4: 运行测试验证通过**

```bash
go test ./internal/theme -v
```

Expected: PASS

**Step 5: 提交**

```bash
git add internal/theme/presets.go internal/theme/theme_test.go
git commit -m "feat(theme): 实现 5 个预设主题"
```

---

## Task 7: 扩展 config 包添加 Settings

**Files:**
- Create: `internal/config/settings.go`
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

**Step 1: 写失败的测试**

```go
// 在 internal/config/config_test.go 添加
func TestConfigWithSettings(t *testing.T) {
	cfg := &Config{
		Profiles: []Profile{
			{Name: "test", Token: "xxx"},
		},
		Default: "test",
		Settings: Settings{
			Language: "en",
			Theme:    "ocean",
		},
	}

	if cfg.Settings.Language != "en" {
		t.Errorf("Settings.Language = %s, want en", cfg.Settings.Language)
	}

	if cfg.Settings.Theme != "ocean" {
		t.Errorf("Settings.Theme = %s, want ocean", cfg.Settings.Theme)
	}
}

func TestLoadConfigWithEmptySettings(t *testing.T) {
	// 创建临时配置文件（没有 settings 字段）
	tmpDir := t.TempDir()
	configPath := tmpDir + "/settings.json"

	data := `{
		"profiles": [{"name": "test", "token": "xxx"}],
		"default": "test"
	}`
	if err := os.WriteFile(configPath, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	// 验证默认值
	if cfg.Settings.Language != "zh" {
		t.Errorf("Settings.Language = %s, want zh (default)", cfg.Settings.Language)
	}

	if cfg.Settings.Theme != "default" {
		t.Errorf("Settings.Theme = %s, want default", cfg.Settings.Theme)
	}
}

func TestUpdateSetting(t *testing.T) {
	cfg := &Config{}

	cfg.UpdateSetting("language", "en")
	if cfg.Settings.Language != "en" {
		t.Errorf("UpdateSetting(language, en) failed")
	}

	cfg.UpdateSetting("theme", "ocean")
	if cfg.Settings.Theme != "ocean" {
		t.Errorf("UpdateSetting(theme, ocean) failed")
	}
}
```

**Step 2: 运行测试验证失败**

```bash
go test ./internal/config -v -run "TestConfigWith|TestLoadConfigWith|TestUpdateSetting"
```

Expected: FAIL - Settings 未定义

**Step 3: 实现 Settings 结构体和扩展 Config**

```go
// internal/config/settings.go
package config

type Settings struct {
	Language string `json:"language"`
	Theme    string `json:"theme"`
}
```

```go
// 在 internal/config/config.go 修改
type Config struct {
	Profiles []Profile `json:"profiles"`
	Default  string    `json:"default,omitempty"`
	Settings Settings  `json:"settings,omitempty"`  // 新增
}

// 在 LoadConfig 函数中添加迁移逻辑
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
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

// 添加 UpdateSetting 方法
func (c *Config) UpdateSetting(key, value string) {
	switch key {
	case "language":
		c.Settings.Language = value
	case "theme":
		c.Settings.Theme = value
	}
}
```

**Step 4: 运行测试验证通过**

```bash
go test ./internal/config -v -run "TestConfigWith|TestLoadConfigWith|TestUpdateSetting"
```

Expected: PASS

**Step 5: 提交**

```bash
git add internal/config/settings.go internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): 添加 Settings 支持和迁移逻辑"
```

---

## Task 8: 在 REPL Model 中集成 i18n 和 theme

**Files:**
- Modify: `internal/repl/model.go`
- Modify: `internal/repl/repl.go`

**Step 1: 写失败的测试（手动测试）**

由于这是集成任务，我们将通过运行程序来测试。

**Step 2: 修改 Model 结构体**

```go
// 在 internal/repl/model.go 修改
import (
	// ... 现有导入
	"github.com/wujunwei/cc-start/internal/i18n"
	"github.com/wujunwei/cc-start/internal/theme"
)

type Model struct {
	// 配置
	config     *config.Config
	configPath string

	// 当前状态
	currentProfile string
	focus          Focus
	quitting       bool
	keys           keyMap

	// 组件
	input    textinput.Model
	output   *OutputBuffer
	palette  *CommandPalette
	settings *SettingsPanel
	help     help.Model

	// 历史记录
	history *History
	histIdx int

	// 样式和国际化（新增）
	styles Styles
	i18n   *i18n.Manager  // 新增
	theme  *theme.Theme   // 新增

	// 窗口尺寸
	width  int
	height int

	// 待执行的启动命令
	PendingLaunch *PendingLaunch
}
```

**Step 3: 修改 NewModel 初始化逻辑**

```go
// 在 internal/repl/model.go 修改 NewModel
func NewModel(cfgPath string) (Model, error) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return Model{}, err
	}

	// 初始化 i18n
	i18nMgr := i18n.NewManager()
	if cfg.Settings.Language != "" {
		i18nMgr.SetLanguage(cfg.Settings.Language)
	}

	// 初始化主题
	currentTheme, err := theme.GetTheme(cfg.Settings.Theme)
	if err != nil {
		currentTheme, _ = theme.GetTheme("default")
	}

	ti := textinput.New()
	ti.Placeholder = i18nMgr.T(i18n.MsgREPLInputPrompt)
	ti.Focus()
	ti.Prompt = ""

	h := help.New()
	hist := NewHistory()
	out := NewOutputBuffer(100)

	// 应用主题到样式
	styles := DefaultStyles()
	styles = *theme.ApplyTheme(currentTheme, &styles)

	return Model{
		config:         cfg,
		configPath:     cfgPath,
		currentProfile: cfg.Default,
		focus:          FocusInput,
		keys:           defaultKeyMap(),
		input:          ti,
		output:         out,
		history:        hist,
		help:           h,
		styles:         styles,
		i18n:           i18nMgr,
		theme:          currentTheme,
	}, nil
}
```

**Step 4: 在 theme 包中添加 ApplyTheme 函数**

```go
// 在 internal/theme/theme.go 添加
import (
	"github.com/charmbracelet/lipgloss"
)

// ApplyTheme 将主题应用到样式（这里返回简化的样式，实际使用时需要适配 repl.Styles）
func ApplyTheme(t *Theme, baseColors map[string]string) map[string]string {
	// 返回主题颜色映射，供 repl 包使用
	return map[string]string{
		"background": t.Colors.Background,
		"foreground": t.Colors.Foreground,
		"primary":    t.Colors.Primary,
		"success":    t.Colors.Success,
		"error":      t.Colors.Error,
		"warning":    t.Colors.Warning,
		"info":       t.Colors.Info,
		"muted":      t.Colors.Muted,
		"border":     t.Colors.Border,
		"accent":     t.Colors.Accent,
		"highlight":  t.Colors.Highlight,
	}
}
```

**Step 5: 手动测试**

```bash
go build -o cc-start .
./cc-start
```

Expected: 程序能正常启动，无编译错误

**Step 6: 提交**

```bash
git add internal/repl/model.go internal/theme/theme.go
git commit -m "feat(repl): 集成 i18n 和 theme 到 Model"
```

---

## Task 9: 更新设置面板支持二级选择

**Files:**
- Modify: `internal/repl/settings.go`
- Modify: `internal/repl/update.go`

**Step 1: 修改 SettingsPanel 结构体**

```go
// 在 internal/repl/settings.go 修改
type SettingsMode int

const (
	SettingsModeMain     SettingsMode = iota
	SettingsModeLanguage
	SettingsModeTheme
)

type SettingsPanel struct {
	visible      bool
	mode         SettingsMode  // 新增
	query        string
	items        []SettingsItem
	selected     int
	subItems     []SettingsItem  // 新增：子选项
	styles       Styles
	width        int
	i18n         *i18n.Manager   // 新增
}

// 修改 NewSettingsPanel
func NewSettingsPanel(styles Styles, i18nMgr *i18n.Manager) *SettingsPanel {
	return &SettingsPanel{
		styles: styles,
		i18n:   i18nMgr,
		mode:   SettingsModeMain,
		items:  getDefaultSettings(i18nMgr),
	}
}

func getDefaultSettings(i18nMgr *i18n.Manager) []SettingsItem {
	return []SettingsItem{
		{
			Key:         "lang",
			Label:       i18nMgr.T(i18n.MsgSettingsLanguage),
			Description: "设置界面语言",
			Value:       i18nMgr.T(i18n.MsgCommonInfo), // TODO: 显示当前语言
			Action:      "setting:lang",
		},
		{
			Key:         "theme",
			Label:       i18nMgr.T(i18n.MsgSettingsTheme),
			Description: "设置显示主题",
			Value:       "默认", // TODO: 显示当前主题
			Action:      "setting:theme",
		},
	}
}

// 添加获取语言选项的方法
func (s *SettingsPanel) getLanguageOptions() []SettingsItem {
	return []SettingsItem{
		{Key: "zh", Label: "中文", Action: "lang:zh"},
		{Key: "en", Label: "English", Action: "lang:en"},
		{Key: "ja", Label: "日本語", Action: "lang:ja"},
	}
}

// 添加获取主题选项的方法
func (s *SettingsPanel) getThemeOptions() []SettingsItem {
	themes := theme.GetAllThemes()
	items := make([]SettingsItem, len(themes))
	for i, t := range themes {
		items[i] = SettingsItem{
			Key:    t.Name,
			Label:  t.DisplayName,
			Action: "theme:" + t.Name,
		}
	}
	return items
}

// 修改 Render 方法支持不同模式
func (s *SettingsPanel) Render() string {
	if !s.visible {
		return ""
	}

	var sections []string

	// 根据模式选择标题
	var title string
	switch s.mode {
	case SettingsModeLanguage:
		title = s.styles.PaletteTitle.Render("语言设置 / Language")
	case SettingsModeTheme:
		title = s.styles.PaletteTitle.Render("主题设置 / Theme")
	default:
		title = s.styles.PaletteTitle.Render(s.i18n.T(i18n.MsgSettingsTitle))
	}
	sections = append(sections, title)

	// 输入框
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#626262")).
		Padding(0, 1).
		Width(46)
	input := inputStyle.Render("> " + s.query)
	sections = append(sections, input)

	// 根据模式选择显示的项
	var displayItems []SettingsItem
	switch s.mode {
	case SettingsModeLanguage:
		displayItems = s.getLanguageOptions()
	case SettingsModeTheme:
		displayItems = s.getThemeOptions()
	default:
		displayItems = s.filteredItems()
	}

	// 渲染列表
	var listLines []string
	for i, item := range displayItems {
		if i >= 10 {
			break
		}

		var line string
		if i == s.selected {
			line = s.styles.PaletteActive.Render("● " + item.Label)
		} else {
			line = s.styles.PaletteItem.Render("  " + item.Label)
		}
		listLines = append(listLines, line)
	}

	if len(listLines) > 0 {
		sections = append(sections, strings.Join(listLines, "\n"))
	}

	// 提示
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(
		s.i18n.T(i18n.MsgSettingsHint))
	sections = append(sections, hint)

	return s.styles.Palette.Render(strings.Join(sections, "\n"))
}
```

**Step 2: 修改 update.go 中的设置处理**

```go
// 在 internal/repl/update.go 修改 handleSettingAction
func (m Model) handleSettingAction(action string) (tea.Model, tea.Cmd) {
	switch action {
	case "setting:lang":
		// 进入语言选择模式
		m.settings.mode = SettingsModeLanguage
		m.settings.selected = 0
		m.settings.query = ""
		return m, nil

	case "setting:theme":
		// 进入主题选择模式
		m.settings.mode = SettingsModeTheme
		m.settings.selected = 0
		m.settings.query = ""
		return m, nil

	case "lang:zh", "lang:en", "lang:ja":
		// 应用语言更改
		lang := strings.TrimPrefix(action, "lang:")
		if err := m.i18n.SetLanguage(lang); err != nil {
			m.output.WriteError(err.Error())
			return m, nil
		}
		m.config.UpdateSetting("language", lang)
		if err := m.config.Save(m.configPath); err != nil {
			m.output.WriteError("保存配置失败: " + err.Error())
			return m, nil
		}
		m.output.WriteSuccess(fmt.Sprintf("语言已切换为: %s", lang))
		// 返回主设置面板
		m.settings.mode = SettingsModeMain
		m.settings.selected = 0
		m.settings.query = ""
		return m, nil

	case "theme:default", "theme:ocean", "theme:forest", "theme:sunset", "theme:light":
		// 应用主题更改
		themeName := strings.TrimPrefix(action, "theme:")
		newTheme, err := theme.GetTheme(themeName)
		if err != nil {
			m.output.WriteError(err.Error())
			return m, nil
		}
		m.theme = newTheme
		m.styles = *applyThemeToStyles(newTheme, &m.styles)
		m.config.UpdateSetting("theme", themeName)
		if err := m.config.Save(m.configPath); err != nil {
			m.output.WriteError("保存配置失败: " + err.Error())
			return m, nil
		}
		m.output.WriteSuccess(fmt.Sprintf("主题已切换为: %s", newTheme.DisplayName))
		// 返回主设置面板
		m.settings.mode = SettingsModeMain
		m.settings.selected = 0
		m.settings.query = ""
		return m, nil

	default:
		m.output.Write("● 未知设置项: " + action)
	}
	return m, nil
}

// 添加辅助函数
func applyThemeToStyles(t *theme.Theme, styles *Styles) *Styles {
	// 应用主题颜色到样式
	// 这里需要根据实际的 Styles 结构进行调整
	return styles
}
```

**Step 3: 手动测试**

```bash
go build -o cc-start .
./cc-start
# 按 Ctrl+P 打开设置
# 选择语言设置，验证语言列表显示
# 选择主题设置，验证主题列表显示
```

Expected: 设置面板能显示二级选项

**Step 4: 提交**

```bash
git add internal/repl/settings.go internal/repl/update.go
git commit -m "feat(repl): 设置面板支持二级选择"
```

---

## Task 10: 实现主题应用到样式

**Files:**
- Modify: `internal/repl/styles.go`
- Modify: `internal/repl/update.go` (完善 applyThemeToStyles)

**Step 1: 修改 Styles 结构体使用主题颜色**

```go
// 在 internal/repl/styles.go 添加
import "github.com/charmbracelet/lipgloss"

// DefaultStyles 创建默认样式（使用默认主题）
func DefaultStyles() Styles {
	return Styles{
		// ... 现有样式定义
	}
}

// NewStylesFromTheme 从主题创建样式
func NewStylesFromTheme(t *theme.Theme) Styles {
	return Styles{
		PaletteTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Colors.Primary)).
			Bold(true).
			Padding(0, 1),

		PaletteItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Colors.Foreground)),

		PaletteActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Colors.Highlight)).
			Background(lipgloss.Color(t.Colors.PaletteActive)).
			Bold(true),

		// ... 应用其他样式
	}
}
```

**Step 2: 完善主题应用逻辑**

```go
// 在 internal/repl/update.go 修改
func applyThemeToStyles(t *theme.Theme, styles *Styles) *Styles {
	*styles = NewStylesFromTheme(t)
	return styles
}
```

**Step 3: 手动测试**

```bash
go build -o cc-start .
./cc-start
# 按 Ctrl+P 打开设置
# 选择主题设置
# 切换不同主题，验证颜色变化
```

Expected: 切换主题后，UI 颜色立即改变

**Step 4: 提交**

```bash
git add internal/repl/styles.go internal/repl/update.go
git commit -m "feat(repl): 实现主题应用到样式"
```

---

## Task 11: 应用 i18n 到所有 UI 文本

**Files:**
- Modify: `internal/repl/view.go`
- Modify: `internal/repl/commands.go`
- Modify: `internal/repl/palette.go`

**Step 1: 修改命令面板使用 i18n**

```go
// 在 internal/repl/palette.go 修改
func NewCommandPalette(styles Styles, i18nMgr *i18n.Manager) *CommandPalette {
	return &CommandPalette{
		styles: styles,
		i18n:   i18nMgr,
		items:  getDefaultCommands(i18nMgr),
	}
}

func getDefaultCommands(i18nMgr *i18n.Manager) []PaletteItem {
	return []PaletteItem{
		{Cmd: "/list", Label: i18nMgr.T(i18n.MsgCmdList)},
		{Cmd: "/use", Label: i18nMgr.T(i18n.MsgCmdUse)},
		{Cmd: "/setup", Label: i18nMgr.T(i18n.MsgCmdSetup)},
		// ... 其他命令
	}
}
```

**Step 2: 修改命令描述使用 i18n**

```go
// 在 internal/repl/commands.go 修改
func (m *Model) formatHelp() string {
	return fmt.Sprintf(`
%s:

%s:
  /list, /ls          %s
  /use, /switch       %s
  /current, /status   %s
  // ... 其他命令
`,
		m.i18n.T(i18n.MsgCmdHelp),
		m.i18n.T("category.config_management"),
		m.i18n.T(i18n.MsgCmdList),
		m.i18n.T(i18n.MsgCmdUse),
		m.i18n.T(i18n.MsgCmdCurrent),
		// ...
	)
}
```

**Step 3: 手动测试**

```bash
go build -o cc-start .
./cc-start
# 按 Ctrl+P 打开设置
# 切换语言为英文
# 验证所有文本变为英文
```

Expected: 所有 UI 文本根据语言设置显示

**Step 4: 提交**

```bash
git add internal/repl/view.go internal/repl/commands.go internal/repl/palette.go
git commit -m "feat(repl): 应用 i18n 到所有 UI 文本"
```

---

## Task 12: 集成测试和文档更新

**Files:**
- Create: `internal/repl/integration_test.go`
- Modify: `README.md`

**Step 1: 写集成测试**

```go
// internal/repl/integration_test.go
package repl

import (
	"os"
	"testing"

	"github.com/wujunwei/cc-start/internal/config"
	"github.com/wujunwei/cc-start/internal/i18n"
	"github.com/wujunwei/cc-start/internal/theme"
)

func TestLanguageSwitch(t *testing.T) {
	// 创建临时配置
	tmpDir := t.TempDir()
	cfgPath := tmpDir + "/settings.json"

	cfg := &config.Config{
		Profiles: []config.Profile{
			{Name: "test", Token: "xxx"},
		},
		Settings: config.Settings{
			Language: "zh",
			Theme:    "default",
		},
	}
	cfg.Save(cfgPath)

	// 创建 Model
	model, err := NewModel(cfgPath)
	if err != nil {
		t.Fatalf("NewModel() error: %v", err)
	}

	// 验证初始语言
	if model.i18n.T(i18n.MsgCommonSuccess) != "成功" {
		t.Errorf("Initial language not zh")
	}

	// 切换语言
	model.i18n.SetLanguage("en")
	if model.i18n.T(i18n.MsgCommonSuccess) != "Success" {
		t.Errorf("Language switch to en failed")
	}
}

func TestThemeSwitch(t *testing.T) {
	// 测试主题切换
	model, _ := NewModel(t.TempDir() + "/settings.json")

	// 切换主题
	oceanTheme, err := theme.GetTheme("ocean")
	if err != nil {
		t.Fatalf("GetTheme(ocean) error: %v", err)
	}

	model.theme = oceanTheme
	model.styles = *applyThemeToStyles(oceanTheme, &model.styles)

	// 验证主题已应用
	if model.theme.Name != "ocean" {
		t.Errorf("Theme not switched to ocean")
	}
}
```

**Step 2: 运行集成测试**

```bash
go test ./internal/repl -v -run Integration
```

Expected: PASS

**Step 3: 更新 README.md**

```markdown
# 在 README.md 添加
## 系统设置

### 语言设置
支持中文、英文、日文三种语言，可通过 Ctrl+P 打开设置面板进行切换。

### 主题设置
提供 5 个预设主题：
- 默认 / Default - 深色主题，高对比度
- 海洋 / Ocean - 蓝绿色调，清爽
- 森林 / Forest - 绿色调，自然
- 日落 / Sunset - 暖色调，温馨
- 亮色 / Light - 浅色主题，适合白天使用

### 使用方法
1. 在 REPL 中按 `Ctrl+P` 打开设置面板
2. 使用 ↑↓ 键选择设置项
3. 按 Enter 进入设置项
4. 选择新的值并按 Enter 确认
5. 按 Esc 关闭设置面板

设置会自动保存到 `~/.cc-start/settings.json`，下次启动时自动应用。
```

**Step 4: 提交**

```bash
git add internal/repl/integration_test.go README.md
git commit -m "test: 添加集成测试和更新文档"
```

---

## Task 13: 最终验证和清理

**Step 1: 运行所有测试**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 2: 运行代码检查**

```bash
go vet ./...
go fmt ./...
```

Expected: 无错误

**Step 3: 构建并手动测试**

```bash
go build -o cc-start .
./cc-start
```

手动测试清单：
- [ ] 按 Ctrl+P 打开设置面板
- [ ] 切换语言为英文，验证所有文本变为英文
- [ ] 切换语言为日文，验证所有文本变为日文
- [ ] 切换主题为 ocean，验证颜色变化
- [ ] 切换主题为 forest，验证颜色变化
- [ ] 退出并重新启动，验证设置已保存
- [ ] 测试命令面板（/）的文本显示

**Step 4: 最终提交**

```bash
git add .
git commit -m "feat: 完成系统设置功能实现

- 实现多语言支持（中文、英文、日文）
- 实现 5 个预设主题
- 配置持久化到 settings.json
- 设置面板支持二级选择
- 完整的测试覆盖"
```

---

## 总结

本实现计划遵循 TDD 原则，分 13 个任务完成系统设置功能：

1. ✅ i18n 基础结构
2. ✅ 中文翻译
3. ✅ 英文翻译
4. ✅ 日文翻译
5. ✅ theme 基础结构
6. ✅ 5 个预设主题
7. ✅ config 扩展 Settings
8. ✅ REPL 集成 i18n 和 theme
9. ✅ 设置面板二级选择
10. ✅ 主题应用
11. ✅ i18n 应用到 UI
12. ✅ 集成测试
13. ✅ 最终验证

每个任务都包含：
- 失败的测试
- 最小实现
- 通过测试
- 提交代码

遵循 DRY、YAGNI、TDD 原则，频繁提交。
